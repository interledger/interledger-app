package client

import (
	"context"

	"gitlab.com/fynbos/backend/providers/gmt"
	"gitlab.com/fynbos/backend/providers/gmt/ops"
	"go.temporal.io/api/enums/v1"
	temporal "go.temporal.io/sdk/client"
)

var _ gmt.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) gmt.Client {
	return &client{b: b}
}

func (c client) StartUserOnboarding(ctx context.Context, walletID string) (gmt.Await, error) {
	workflowOptions := temporal.StartWorkflowOptions{
		ID:                    "gmt_onboard_user_" + walletID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	wf, err := c.b.Temporal().ExecuteWorkflow(ctx, workflowOptions, ops.OnboardUserWorkflow, walletID)
	if err != nil {
		return nil, err
	}

	return wf.Get, nil
}
