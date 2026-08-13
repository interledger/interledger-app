package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// migrationEmailJobTimeout covers the worker wait plus runMigrationEmailJob's
// own 5min WorkflowExecutionTimeout, so the workflow-side error surfaces first.
const migrationEmailJobTimeout = 6 * time.Minute

// migrationEmailParams is the operator input for SendMigrationEmailJob, minus
// the targeting fields each step fills in.
func migrationEmailParams() map[string]any {
	return map[string]any{
		"subject": "We are moving your account",
		"paragraphs": []map[string]any{
			{"heading": "What is changing"},
			{"paragraph": "Your wallet is moving to a new payment provider."},
		},
	}
}

// theMigrationEmailJobRunsForMyEmail targets the signed-up user by address, the
// way an operator does a test send. The job fails when an address matches no
// user, so a clean run proves it found this user's Kratos identity.
//
// Usage: When the migration email job runs for my email
func (sc *E2EContext) theMigrationEmailJobRunsForMyEmail() error {
	email, err := sc.getCurrentUserEmail()
	if err != nil {
		return fmt.Errorf("theMigrationEmailJobRunsForMyEmail: %w", err)
	}

	params := migrationEmailParams()
	params["email"] = email
	ctx, cancel := context.WithTimeout(context.Background(), migrationEmailJobTimeout)
	defer cancel()
	sc.migrationEmailFailures, sc.migrationEmailErr = sc.runMigrationEmailJob(ctx, params)
	debugPrintf("📧 migration email job for %s: failures=%v err=%v\n", email, sc.migrationEmailFailures, sc.migrationEmailErr)
	return nil
}

// theMigrationEmailJobRunsForAnUnknownAddress targets an address no user has,
// so the job should refuse the whole run rather than silently send nothing.
//
// Usage: When the migration email job runs for an unknown address
func (sc *E2EContext) theMigrationEmailJobRunsForAnUnknownAddress() error {
	params := migrationEmailParams()
	params["email"] = "missing-" + uuid.NewString() + "@example.com"
	ctx, cancel := context.WithTimeout(context.Background(), migrationEmailJobTimeout)
	defer cancel()
	sc.migrationEmailFailures, sc.migrationEmailErr = sc.runMigrationEmailJob(ctx, params)
	debugPrintf("📧 migration email job for an unknown address: err=%v\n", sc.migrationEmailErr)
	return nil
}

// theMigrationEmailJobShouldReportNoFailures asserts the last run completed and
// sent to every recipient.
//
// With email.enabled=false in the e2e env the send activity is the noop client,
// so this proves recipient listing and the workflow path, not SMTP delivery.
//
// Usage: Then the migration email job should report no failures
func (sc *E2EContext) theMigrationEmailJobShouldReportNoFailures() error {
	if sc.migrationEmailErr != nil {
		return fmt.Errorf("migration email job failed: %w", sc.migrationEmailErr)
	}
	if len(sc.migrationEmailFailures) > 0 {
		return fmt.Errorf("migration email job reported failures: %v", sc.migrationEmailFailures)
	}
	return nil
}

// theMigrationEmailJobShouldFailWith asserts the last run failed, naming the
// address it could not resolve.
//
// Usage: Then the migration email job should fail with "no user found for:"
func (sc *E2EContext) theMigrationEmailJobShouldFailWith(want string) error {
	if sc.migrationEmailErr == nil {
		return fmt.Errorf("expected the migration email job to fail with %q, but it succeeded (failures: %v)", want, sc.migrationEmailFailures)
	}
	if !strings.Contains(sc.migrationEmailErr.Error(), want) {
		return fmt.Errorf("expected the migration email job to fail with %q, got: %v", want, sc.migrationEmailErr)
	}
	return nil
}
