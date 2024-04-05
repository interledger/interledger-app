package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/currency"
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

func getExternalUserID(ctx context.Context, b Backends, walletID string) (string, error) {
	var externalID string
	err := b.DB().GetContext(ctx, &externalID, "SELECT external_id FROM gatehub_users WHERE wallet_id=$1;", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	} else if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return externalID, err
}

func GetBalance(ctx context.Context, b Backends, linkedAccountID string) (*gatehub.Balance, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if la.Provider != gatehub.ProviderName || la.Type != gatehub.AccTypeBalance {
		return nil, fmt.Errorf("%w linked account not correct type", gatehub.ErrNotFound)
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", gatehub.ErrNotFound)
	}

	return &gatehub.Balance{
		Total:     currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted, la.SendCurrency),
		Available: currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted-accs[0].DebitsPending, la.SendCurrency),
	}, nil
}
