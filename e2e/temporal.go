package main

import (
	"context"
	"fmt"
	"time"

	temporalclient "go.temporal.io/sdk/client"
)

// runTemporalWorkflowByName dials Temporal, executes the named workflow on the backend
// task queue, and blocks until it completes. Used by card transaction polling steps.
func (sc *E2EContext) runTemporalWorkflowByName(workflowName string, args ...interface{}) error {
	c, err := temporalclient.Dial(temporalclient.Options{
		Namespace: "default",
	})
	if err != nil {
		return fmt.Errorf("runTemporalWorkflowByName: connect: %w", err)
	}
	defer c.Close()

	wo := temporalclient.StartWorkflowOptions{
		ID:                       fmt.Sprintf("e2e-%s-%d", workflowName, time.Now().UnixNano()),
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: 2 * time.Minute,
	}
	run, err := c.ExecuteWorkflow(context.Background(), wo, workflowName, args...)
	if err != nil {
		return fmt.Errorf("runTemporalWorkflowByName: execute %s: %w", workflowName, err)
	}
	return run.Get(context.Background(), nil)
}
