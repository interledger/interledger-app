package main

import (
	"context"
	"fmt"
	"time"
)

// agreementWaitTimeout is the longest we wait for grpc.CompleteSignup to
// populate signups.user_id and write agreement_signatures. CompleteSignup runs
// server-side after Kratos identity creation; theSignupShouldBeSubmitted only
// waits for the Kratos row, so this poll closes the window documented in
// e2e/AGENTS.md (signups.user_id NULL race).
const agreementWaitTimeout = 15 * time.Second

// anAgreementSignatureShouldExistForMyselfFor polls until the current user has
// a signature recorded for the given agreement_id, or until the timeout. The
// poll covers two consecutive windows: signups.user_id getting populated, and
// the agreement_signatures INSERT in the same gRPC call.
//
// Usage: Then an agreement signature should exist for myself for "privacy_policy-0.0.0"
func (sc *E2EContext) anAgreementSignatureShouldExistForMyselfFor(agreementID string) error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("anAgreementSignatureShouldExistForMyselfFor: %w", err)
	}
	return sc.waitForAgreementSignature(email, agreementID, agreementWaitTimeout)
}

// aNewAgreementVersionIsPublished inserts a fresh agreement row sharing the name
// of an existing agreement, with notified=false (i.e. the state the change-notify
// trigger expects). The agreement ID is stashed on the context so the workflow
// step and the assertions can reach it.
//
// Tests parameterise the version with the test identifier so concurrent runs
// don't collide on a unique-by-id agreement row. Pass the bare version (e.g.
// "9.9.9") — the step rewrites it as "<name>-<version>-<testIdentifier>".
//
// Rejects a second publish in the same scenario: the After hook only cleans
// the latest pendingAgreementID, so a silent overwrite would leak the first
// row and let main.go's startup trigger re-fire the workflow against it.
//
// Usage: Given a new "privacy_policy" agreement version "9.9.9" is published
func (sc *E2EContext) aNewAgreementVersionIsPublished(name, version string) error {
	if sc.testIdentifier == "" {
		return fmt.Errorf("aNewAgreementVersionIsPublished: no testIdentifier set — call 'a random test identifier is generated' in the Background")
	}
	if sc.pendingAgreementID != "" {
		return fmt.Errorf("aNewAgreementVersionIsPublished: an agreement (%s) is already pending in this scenario — publishing a second one would leak the first because cleanup only tracks one ID", sc.pendingAgreementID)
	}
	uniqueVersion := fmt.Sprintf("%s-%s", version, sc.testIdentifier)
	agreementID := fmt.Sprintf("%s-%s", name, uniqueVersion)
	content := fmt.Sprintf("e2e test content for %s", agreementID)
	if err := sc.insertAgreement(agreementID, name, uniqueVersion, content); err != nil {
		return err
	}
	sc.pendingAgreementID = agreementID
	debugPrintf("   📜 Published test agreement %s\n", agreementID)
	return nil
}

// theAgreementChangeNotificationWorkflowRuns triggers NotifyAgreementChangedWorkflow
// against the most recently published agreement and waits for it to complete.
// Uses a 30-day deadline string to mirror the production trigger in main.go.
//
// Blast radius: the workflow is NOT scoped to the impersonated user — it marks
// (and noop-emails) every user in the DB who signed an older version of the
// published agreement's name (e.g. all "privacy_policy" signers), paginating
// 500/page via ContinueAsNew. cleanupTestAgreement resets all of those markers
// (they all point at pendingAgreementID) on teardown, so state self-heals; but
// under CI concurrency, users from other in-flight scenarios can be swept in.
// Watch this as the shared test DB grows.
//
// Usage: When the agreement change notification workflow runs
func (sc *E2EContext) theAgreementChangeNotificationWorkflowRuns() error {
	if sc.pendingAgreementID == "" {
		return fmt.Errorf("theAgreementChangeNotificationWorkflowRuns: no published agreement on the context — call 'a new ... agreement version ... is published' first")
	}
	deadlineDate := time.Now().UTC().AddDate(0, 0, 30).Format("January 2, 2006")
	// Slightly above runAgreementNotifyWorkflow's 10min WorkflowExecutionTimeout
	// so the ctx doesn't fire first and mask the workflow-side error.
	ctx, cancel := context.WithTimeout(context.Background(), 11*time.Minute)
	defer cancel()
	return sc.runAgreementNotifyWorkflow(ctx, []string{sc.pendingAgreementID}, deadlineDate)
}

// iShouldBeMarkedNotifiedForTheNewAgreement asserts the current user has at least
// one agreement_signatures row whose last_notified_agreement_id equals the most
// recently published agreement ID. This is the production marker proving the
// workflow ran end to end and reached MarkUsersNotifiedActivity, which only runs
// for users whose SendAgreementChangedEmailActivity returned success.
//
// Note: in the default e2e env EMAIL_ENABLED=false, so the send activity is the
// noop email client (returns nil without dispatching). This proves the workflow
// + DB-marker path, NOT that real SMTP delivery works.
//
// Still polls signups.user_id (via pollKratosUserID) — although the workflow
// has committed by the time this step runs, the step can be authored without a
// prior 'an agreement signature should exist' step, in which case the
// CompleteSignup user_id race (see pollKratosUserID) would otherwise crash on
// a NULL scan. The signature read itself is single-shot.
//
// Usage: Then I should be marked notified for the new agreement
func (sc *E2EContext) iShouldBeMarkedNotifiedForTheNewAgreement() error {
	if sc.pendingAgreementID == "" {
		return fmt.Errorf("iShouldBeMarkedNotifiedForTheNewAgreement: no published agreement on the context")
	}
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return err
	}
	userID, err := sc.pollKratosUserID(email, agreementWaitTimeout)
	if err != nil {
		return err
	}
	sigs, err := sc.getAgreementSignaturesForUser(userID)
	if err != nil {
		return err
	}
	notifiedOn := make([]string, 0, len(sigs))
	for _, s := range sigs {
		if s.LastNotifiedAgreementID.Valid {
			notifiedOn = append(notifiedOn, s.LastNotifiedAgreementID.String)
			if s.LastNotifiedAgreementID.String == sc.pendingAgreementID {
				return nil
			}
		}
	}
	return fmt.Errorf("expected user %s to be marked notified for agreement %s, but no signature has that last_notified_agreement_id (have %d signatures, notified markers: %v)",
		userID, sc.pendingAgreementID, len(sigs), notifiedOn)
}
