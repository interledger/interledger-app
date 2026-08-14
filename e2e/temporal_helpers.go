package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

const (
	temporalHostPort  = "localhost:7233"
	temporalNamespace = "default"
	backendTaskQueue  = "backend"
)

// waitForBackendWorker polls Temporal until at least one workflow worker is
// registered on the backend task queue, or the context expires. The
// prerequisite() phase in godog_test.go only does this on GateHub runs (see
// needsGateHubPrerequisite), so @xago-tagged scenarios that drive a workflow
// directly must guard themselves.
func waitForBackendWorker(ctx context.Context, c client.Client) error {
	for {
		resp, err := c.DescribeTaskQueue(ctx, backendTaskQueue, enums.TASK_QUEUE_TYPE_WORKFLOW)
		if err == nil && len(resp.Pollers) > 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waitForBackendWorker: no worker registered on %q before deadline: %w", backendTaskQueue, ctx.Err())
		case <-time.After(2 * time.Second):
		}
	}
}

// runAgreementNotifyWorkflow executes NotifyAgreementChangedWorkflow on the
// backend task queue with the same argument shape used by main.go's startup
// trigger, then blocks until the workflow completes. The workflow itself is
// referenced by name so the e2e module does not need to import the workflow
// type. Argument order must mirror jobs.NotifyAgreementChangedWorkflow:
//
//	(agreementIDs []string, deadlineDate string, startOffset int, cachedChanges, cachedMetadata)
//
// The two cache args are nil on the initial call — the workflow loads them.
func (sc *E2EContext) runAgreementNotifyWorkflow(ctx context.Context, agreementIDs []string, deadlineDate string) error {
	c, err := client.Dial(client.Options{Namespace: temporalNamespace, HostPort: temporalHostPort})
	if err != nil {
		return fmt.Errorf("runAgreementNotifyWorkflow: dial temporal: %w", err)
	}
	defer c.Close()

	// Bound the worker wait so a missing backend surfaces as a clear
	// 'no worker' error rather than a 10-minute WorkflowExecutionTimeout below.
	workerCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if err := waitForBackendWorker(workerCtx, c); err != nil {
		return fmt.Errorf("runAgreementNotifyWorkflow: %w", err)
	}

	wo := client.StartWorkflowOptions{
		ID:        "e2e_agreement_notify_" + uuid.NewString(),
		TaskQueue: backendTaskQueue,
		// Strictly larger than the workflow's per-activity StartToCloseTimeout
		// (5min) so an activity at its limit surfaces as the real activity
		// error, not a 'workflow execution timeout' from this ceiling.
		// ContinueAsNew inherits this — bump if the affected-user set grows
		// past one 500-user page in test environments.
		WorkflowExecutionTimeout: 10 * time.Minute,
	}

	run, err := c.ExecuteWorkflow(ctx, wo, "NotifyAgreementChangedWorkflow",
		agreementIDs, deadlineDate, 0, nil, nil,
	)
	if err != nil {
		return fmt.Errorf("runAgreementNotifyWorkflow: start workflow: %w", err)
	}
	if err := run.Get(ctx, nil); err != nil {
		return fmt.Errorf("runAgreementNotifyWorkflow: workflow failed: %w", err)
	}
	return nil
}

// runMigrationEmailJob executes SendMigrationEmailJob on the backend task queue
// and returns the addresses it failed to send to. Params are a plain map keyed
// by the workflow's JSON field names — the same input an operator types into the
// Temporal portal, and no import of the jobs package.
//
// An error means the job itself failed, e.g. an address matched no user.
func (sc *E2EContext) runMigrationEmailJob(ctx context.Context, params map[string]any) ([]string, error) {
	c, err := client.Dial(client.Options{Namespace: temporalNamespace, HostPort: temporalHostPort})
	if err != nil {
		return nil, fmt.Errorf("runMigrationEmailJob: dial temporal: %w", err)
	}
	defer c.Close()

	// Same guard as runAgreementNotifyWorkflow: @xago scenarios skip the
	// GateHub prerequisite, so nothing else has waited for the worker.
	workerCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if err := waitForBackendWorker(workerCtx, c); err != nil {
		return nil, fmt.Errorf("runMigrationEmailJob: %w", err)
	}

	wo := client.StartWorkflowOptions{
		ID:        "e2e_migration_email_" + uuid.NewString(),
		TaskQueue: backendTaskQueue,
		// One address means one Kratos lookup and one send, so a hang must fail
		// here well inside the suite's own `go test` timeout.
		WorkflowExecutionTimeout: 5 * time.Minute,
	}

	run, err := c.ExecuteWorkflow(ctx, wo, "SendMigrationEmailJob", params)
	if err != nil {
		return nil, fmt.Errorf("runMigrationEmailJob: start workflow: %w", err)
	}

	var failed []string
	if err := run.Get(ctx, &failed); err != nil {
		return nil, fmt.Errorf("runMigrationEmailJob: workflow failed: %w", err)
	}
	return failed, nil
}
