package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/xago"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

func LookupSubAccount(ctx context.Context, b Backends, walletID string) (*xago.SubAccount, error) {
	var entry xago.SubAccount
	err := b.DB().GetContext(ctx, &entry, "SELECT id, wallet_id, account_id, deposit_address, deposit_tag FROM xago_sub_accounts WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, xago.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	return &entry, nil
}

func LookupByAccountID(ctx context.Context, b Backends, accountID string) (*xago.SubAccount, error) {
	var entry xago.SubAccount
	err := b.DB().GetContext(ctx, &entry, "SELECT id, wallet_id, account_id, deposit_address, deposit_tag FROM xago_sub_accounts WHERE account_id=$1", accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, xago.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	return &entry, nil
}

func CreateBeneficiary(ctx context.Context, b Backends, bankAcc xago.CreateBankAccountArgs) (xago.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                       "xago_create_beneficiary_" + bankAcc.WalletID + "_" + bankAcc.AccountNumber,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: 2 * time.Minute,
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateBeneficiaryWorkflow, bankAcc)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	return await.Get, nil
}

func CreateTransaction(ctx context.Context, b Backends, args xago.CreateTransactionArgs) (*xago.Transaction, error) {
	la, err := b.LinkedAccounts().Get(ctx, args.LinkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	if la.WalletID != args.WalletID {
		return nil, fmt.Errorf("%w linked account not found", xago.ErrNotFound)
	}

	txID, err := b.External().CreateTransaction(ctx, args.Amount, args.TransactionID, la.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO xago_transactions (id, wallet_id, linked_account_id, transaction_id, amount, currency) VALUES ($1, $2, $3, $4, $5, $6)",
		txID, args.WalletID, args.LinkedAccountID, args.TransactionID, args.Amount.Value, args.Amount.Currency)
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	return &xago.Transaction{
		ID:              txID,
		WalletID:        args.WalletID,
		LinkedAccountID: la.ID,
		TransactionID:   args.TransactionID,
		Amount:          args.Amount,
	}, nil
}
