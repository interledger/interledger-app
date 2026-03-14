package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/env"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	agreementChangePageSize     = 500
	agreementChangeConcurrency  = 50
	agreementChangeActivityTimeout = 2 * time.Minute
)

// AgreementMetadata is the display name and terms URL for an agreement (keyed by agreement name in metadata map).
type AgreementMetadata struct {
	DisplayName string
	TermsURL    string
}

// AgreementChangeMetadataResult is the result type of LoadAgreementChangeMetadata.
type AgreementChangeMetadataResult struct {
	Changes  []agreements.AgreementChange
	Metadata map[string]AgreementMetadata
}

// NotifyAgreementChangedWorkflow notifies users who signed previous versions of any of the changed agreements.
// agreementIDs are the newly inserted agreement IDs (e.g. privacy_policy-2.0.0). deadlineDate is the last date
// to close account at no charge. startOffset is the pagination cursor; pass 0 on the initial call.
// cachedChanges and cachedMetadata are pre-loaded on the first run and carried through ContinueAsNew so that
// LoadAgreementChangeMetadata is only called once, not once per page.
func NotifyAgreementChangedWorkflow(ctx workflow.Context, agreementIDs []string, deadlineDate string, startOffset int, cachedChanges []agreements.AgreementChange, cachedMetadata map[string]AgreementMetadata) error {
	if len(agreementIDs) == 0 {
		return nil
	}
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var a *Activity
	changes := cachedChanges
	metadata := cachedMetadata
	if len(changes) == 0 {
		var metaResult AgreementChangeMetadataResult
		if err := workflow.ExecuteActivity(ctx, a.LoadAgreementChangeMetadata, agreementIDs).Get(ctx, &metaResult); err != nil {
			return err
		}
		changes = metaResult.Changes
		metadata = metaResult.Metadata
		if len(changes) == 0 {
			return nil
		}
	}
	var userIDs []string
	if err := workflow.ExecuteActivity(ctx, a.GetNextPageOfAffectedUserIDs, changes, agreementChangePageSize, startOffset).Get(ctx, &userIDs); err != nil {
		return err
	}
	if len(userIDs) == 0 {
		return nil
	}

	var namesByUser map[string][]string
	if err := workflow.ExecuteActivity(ctx, a.GetAgreementNamesForUserBatch, userIDs, changes).Get(ctx, &namesByUser); err != nil {
		return err
	}

	sendOpts := workflow.ActivityOptions{
		StartToCloseTimeout: agreementChangeActivityTimeout,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 3},
	}
	type pendingEmail struct {
		userID string
		future workflow.Future
	}
	var pending []pendingEmail
	drainPending := func() {
		for _, p := range pending {
			if err := p.future.Get(ctx, nil); err != nil {
				workflow.GetLogger(ctx).Warn("failed to send agreement changed email after retries", zap.String("userID", p.userID), zap.Error(err))
			}
		}
		pending = nil
	}
	for _, userID := range userIDs {
		names := namesByUser[userID]
		if len(names) == 0 {
			continue
		}
		agreementLinks := make([]email.AgreementLink, 0, len(names))
		for _, name := range names {
			if m, ok := metadata[name]; ok {
				agreementLinks = append(agreementLinks, email.AgreementLink{DisplayName: m.DisplayName, TermsURL: m.TermsURL})
			}
		}
		if len(agreementLinks) == 0 {
			continue
		}
		pending = append(pending, pendingEmail{
			userID: userID,
			future: workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, sendOpts), a.SendAgreementChangedEmailActivity, userID, agreementLinks, deadlineDate),
		})
		if len(pending) >= agreementChangeConcurrency {
			drainPending()
		}
	}
	drainPending()

	// If we got a full page there may be more — checkpoint via ContinueAsNew.
	// A partial page means we've reached the end; return nil to complete the workflow.
	if len(userIDs) < agreementChangePageSize {
		return nil
	}
	return workflow.NewContinueAsNewError(ctx, NotifyAgreementChangedWorkflow, agreementIDs, deadlineDate, startOffset+agreementChangePageSize, changes, metadata)
}

func (a *Activity) LoadAgreementChangeMetadata(ctx context.Context, agreementIDs []string) (AgreementChangeMetadataResult, error) {
	baseURL := strings.TrimSuffix(env.GetUrl(), "/")
	changes := make([]agreements.AgreementChange, 0, len(agreementIDs))
	metadata := make(map[string]AgreementMetadata)
	for _, id := range agreementIDs {
		ag, err := a.b.Agreements().Get(ctx, id)
		if err != nil {
			return AgreementChangeMetadataResult{}, err
		}
		changes = append(changes, agreements.AgreementChange{Name: ag.Name, ExceptID: id})
		slug := strings.ReplaceAll(ag.Name, "_", "-")
		metadata[ag.Name] = AgreementMetadata{
			DisplayName: agreementDisplayName(ag.Name),
			TermsURL:    fmt.Sprintf("%s/legal/%s", baseURL, slug),
		}
	}
	return AgreementChangeMetadataResult{Changes: changes, Metadata: metadata}, nil
}

func agreementDisplayName(name string) string {
	switch name {
	case "privacy_policy":
		return "Privacy Policy"
	case "terms_of_service":
		return "Terms of Service"
	case "user_policy":
		return "User Policy"
	default:
		parts := strings.Split(name, "_")
		for i, p := range parts {
			if len(p) > 0 {
				parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
			}
		}
		return strings.Join(parts, " ")
	}
}

func (a *Activity) GetNextPageOfAffectedUserIDs(ctx context.Context, changes []agreements.AgreementChange, limit, offset int) ([]string, error) {
	return a.b.Agreements().ListAffectedUserIDsPaginated(ctx, changes, limit, offset)
}

func (a *Activity) GetAgreementNamesForUserBatch(ctx context.Context, userIDs []string, changes []agreements.AgreementChange) (map[string][]string, error) {
	return a.b.Agreements().GetAgreementNamesSignedByUsersFromSet(ctx, userIDs, changes)
}

func (a *Activity) SendAgreementChangedEmailActivity(ctx context.Context, userID string, agreements []email.AgreementLink, deadlineDate string) error {
	a.b.Email().SendAgreementChangedEmail(ctx, userID, agreements, deadlineDate)
	return nil
}
