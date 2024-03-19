package ops

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

func CreateUser(ctx context.Context, b Backends, walletID string) (gatehub.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                    "gatehub_create_user_" + walletID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateGatehubUserWorkflow, walletID)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return await.Get, nil
}

func GetOnboardingWidget(ctx context.Context, b Backends, ec external.Client, walletID string) (string, error) {
	externalUserID, err := getExternalUserID(ctx, b, walletID)
	if errors.Is(err, gatehub.ErrNotFound) {
		await, innerErr := CreateUser(ctx, b, walletID)
		if innerErr != nil {
			return "", innerErr
		}

		innerErr = await(ctx, &externalUserID)
		if innerErr != nil {
			return "", fmt.Errorf("%w %s", gatehub.ErrInternal, innerErr)
		}
	} else if err != nil {
		return "", err
	}

	widget, err := ec.GetOnboardingWidget(ctx, externalUserID)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return widget, nil
}
