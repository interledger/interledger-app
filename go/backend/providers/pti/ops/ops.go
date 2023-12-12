package ops

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/pti"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

var (
	userFields       = "id, external_id, wallet_id, status, assessment_status, created_at, updated_at"
	userInsertFields = "external_id, wallet_id"
)

func CreateWallet(ctx context.Context, b Backends, args pti.CreateWalletArgs) (pti.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                       "pti_create_wallet_" + args.WalletID + "_" + args.Currency.String(),
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: 5 * time.Minute,
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// return workflow if it's running
	var await client.WorkflowRun
	var executeErr error
	if workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		await = b.Temporal().GetWorkflow(ctx, wo.ID, "")
	} else {
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateWalletWorkflow, pti.CreateWalletArgs{})
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return await.Get, nil
}
