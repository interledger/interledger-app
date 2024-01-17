package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
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
		ID:                    "pti_create_wallet_" + args.WalletID + "_" + args.Currency.String(),
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateWalletWorkflow, args)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return await.Get, nil
}

func GetUser(ctx context.Context, b Backends, walletID string) (*pti.User, error) {
	var user pti.User
	err := b.DB().GetContext(ctx, &user, fmt.Sprintf("SELECT %s from pti_users where wallet_id=$1;", userFields), walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pti.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetWallet(ctx context.Context, b Backends, external external.Client, linkedAccountID string) (*pti.Wallet, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	if la.Provider != pti.ProviderName {
		return nil, pti.ErrNotFound
	}

	externalUser, err := GetUser(ctx, b, la.WalletID)
	if err != nil {
		return nil, err
	}

	w, err := external.GetWallet(ctx, externalUser.ExternalID, la.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return &pti.Wallet{
		ID:        w.WalletID,
		UserID:    externalUser.ExternalID,
		Reference: w.Reference,
		Balance:   currency.FromFloat64(w.Balance, currency.ParseCurrency(w.Currency)),
	}, nil
}

func DepositToWallet(ctx context.Context, b Backends, ec external.Client, args pti.TransactionArgs) (string, error) {
	externalUser, err := GetUser(ctx, b, args.WalletID)
	if err != nil {
		return "", err
	}

	la, err := b.LinkedAccounts().Get(ctx, args.LinkedAccountID)
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	if la.Provider != pti.ProviderName || la.WalletID != args.WalletID {
		return "", pti.ErrNotFound
	}

	txID, err := ec.WalletDeposit(ctx, external.DepositArgs{
		RequestID:        args.PaymentID,
		ScenarioID:       pti.ScenarioDeposit,
		UserID:           externalUser.ExternalID,
		ExternalWalletID: la.ProviderID,
		Amount:           args.Amount,
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return txID, nil
}

func WithdrawFromWallet(ctx context.Context, b Backends, ec external.Client, args pti.TransactionArgs) (string, error) {
	externalUser, err := GetUser(ctx, b, args.WalletID)
	if err != nil {
		return "", err
	}

	la, err := b.LinkedAccounts().Get(ctx, args.LinkedAccountID)
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	if la.Provider != pti.ProviderName || la.WalletID != args.WalletID {
		return "", pti.ErrNotFound
	}

	txID, err := ec.WalletWithdrawal(ctx, external.WithdrawalArgs{
		RequestID:        args.PaymentID,
		ScenarioID:       pti.ScenarioWithdrawal,
		UserID:           externalUser.ExternalID,
		ExternalWalletID: la.ProviderID,
		Amount:           args.Amount,
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return txID, nil
}

func UpdateTransactionStatus(ctx context.Context, b Backends, ex external.Client, args pti.TransactionStatusArgs) error {
	_, err := ex.UpdateTransactionStatus(ctx, external.UpdateTxStatusArgs{
		RequestID:     args.PaymentID,
		TransactionID: args.TransactionID,
		Feedback:      args.Status,
		Date:          time.Now(),
	})
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return nil
}
