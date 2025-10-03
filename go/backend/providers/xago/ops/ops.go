package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

func LookupSubAccount(ctx context.Context, b Backends, walletID string) (*xago.SubAccount, error) {
	var entry xago.SubAccount
	err := b.DB().GetContext(ctx, &entry, "SELECT id, wallet_id, account_id, deposit_address, deposit_tag, deposit_reference FROM xago_sub_accounts WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, xago.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}
	entry.Details, err = getDepositDetails(ctx, b, entry.ID)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func LookupByAccountID(ctx context.Context, b Backends, accountID string) (*xago.SubAccount, error) {
	var entry xago.SubAccount
	err := b.DB().GetContext(ctx, &entry, "SELECT id, wallet_id, account_id, deposit_address, deposit_tag, deposit_reference FROM xago_sub_accounts WHERE account_id=$1", accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, xago.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}
	entry.Details, err = getDepositDetails(ctx, b, accountID)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func getDepositDetails(ctx context.Context, b Backends, accountID string) ([]xago.DepositDetails, error) {
	var entries []xago.DepositDetails
	err := b.DB().SelectContext(ctx, &entries, "SELECT id, wallet_id, sub_account_id, currency, bank_name, account_name, account_number, branch_code FROM xago_deposit_details WHERE sub_account_id=$1", accountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	return entries, nil
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

	txID, err := b.External().CreateTransaction(ctx, args.Amount, args.TransactionID, la.ProviderID, args.Reference)
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

func CreateBalanceAccount(ctx context.Context, b Backends, args xago.CreateBalanceAccArgs) (xago.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                       "xago_create_balance_acc_" + args.WalletID + "_" + args.Currency.String(),
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, CreateBalanceAccountWorkflow, args)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	return await.Get, nil
}

func LookupWithdrawal(ctx context.Context, b Backends, id string) (*xago.Withdrawal, error) {
	w, err := b.External().GetWithdrawal(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	return &xago.Withdrawal{
		ID:     w.ID,
		Amount: currency.FromFloat64(w.Amount, currency.ParseCurrency(w.Currency)),
		Status: w.Status,
	}, nil
}

func UpdateInquiryLink(ctx context.Context, b Backends, args xago.InquiryLinkUpdate) error {
	wo := client.StartWorkflowOptions{
		ID:                       "xago_update_inquiry_link_" + args.WalletID + "_" + args.AccountID,
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
		return fmt.Errorf("%w %s", xago.ErrInternal, err)
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, UpdateInquiryLinkWorkflow, args)
	}
	if executeErr != nil {
		return fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	return await.Get(ctx, nil)

}

func GetBalance(ctx context.Context, b Backends, linkedAccountID string) (*xago.Balance, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	if la.Provider != xago.ProviderName || la.Type != xago.AccTypeBalance {
		return nil, fmt.Errorf("%w linked account not correct type", xago.ErrNotFound)
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", xago.ErrNotFound)
	}

	return &xago.Balance{
		Total:     currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted, la.SendCurrency),
		Available: currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted-accs[0].DebitsPending, la.SendCurrency),
	}, nil
}

func ReserveBalance(ctx context.Context, b Backends, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*xago.Balance, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	if la.Provider != xago.ProviderName || la.Type != xago.AccTypeBalance {
		return nil, fmt.Errorf("%w linked account not correct type", xago.ErrNotFound)
	}

	ledger := xago.LedgerIDUSD
	opsAcc := xago.USDOpsAccount
	if la.SendCurrency == currency.ZAR {
		ledger = xago.LedgerIDZAR
		opsAcc = xago.ZAROpsAccount
	}

	tx, err := b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              txID,
			Amount:          amt.Value,
			DebitAccountID:  la.ID,
			CreditAccountID: opsAcc,
			Pending:         true,
			Code:            1,
			Timeout:         uint64(timeout),
			Ledger:          ledger,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}
	if len(tx) > 0 {
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return nil, fmt.Errorf("%w insufficient balance code (%s)", xago.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return nil, fmt.Errorf("%w non success code (%s)", xago.ErrInternal, tx[0].Code.String())
		}
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", xago.ErrNotFound)
	}

	return &xago.Balance{
		Total:     currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted, la.SendCurrency),
		Available: currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted-accs[0].DebitsPending, la.SendCurrency),
	}, nil
}

func FinaliseReserve(ctx context.Context, b Backends, txID string) error {
	tx, err := b.Pacioli().PostTransfers(ctx, []string{txID})
	if err != nil {
		return fmt.Errorf("%w %s", xago.ErrInternal, err)
	}
	if len(tx) == 0 {
		return nil
	}
	if tx[0].Code == pacioli.TransferPendingTransferAlreadyPosted {
		return nil
	}
	if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
		return fmt.Errorf("%w insufficient balance code (%s)", xago.ErrInsufficientBalance, tx[0].Code.String())
	}
	if tx[0].Code == pacioli.TransferPendingTransferExpired {
		return fmt.Errorf("%w transfer timed out code (%s)", xago.ErrTimedOut, tx[0].Code.String())
	}
	if tx[0].Code == pacioli.TransferPendingTransferNotPending {
		return fmt.Errorf("%w pending transfer not found code (%s)", xago.ErrNotFound, tx[0].Code.String())
	}
	if tx[0].Code != 0 {
		return fmt.Errorf("%w non success code (%s)", xago.ErrInternal, tx[0].Code.String())
	}

	return nil
}

func RollbackReserve(ctx context.Context, b Backends, txID string) error {
	tx, err := b.Pacioli().VoidTransfers(ctx, []string{txID})
	if err != nil {
		slack.SendToChannel(ctx, slack.ChannelNotifyErrors, "wallet-info-bot", fmt.Sprintf("*:::[XAGO ERROR]:::* \n *RollbackReserve txID:* %s,\n *error:* %s", txID, err))
		return fmt.Errorf("%w %s", xago.ErrInternal, err)
	}
	if len(tx) == 0 {
		return nil
	}
	if tx[0].Code == pacioli.TransferPendingTransferNotFound ||
		tx[0].Code == pacioli.TransferPendingTransferAlreadyVoided ||
		tx[0].Code == pacioli.TransferPendingTransferExpired {
		return nil
	}
	if tx[0].Code != 0 {
		return fmt.Errorf("%w non success code (%s)", xago.ErrInternal, tx[0].Code.String())
	}

	return nil
}

func AssignBalance(ctx context.Context, b Backends, linkedAccountID, txID string, amt currency.Amount) (*xago.Balance, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	if la.Provider != xago.ProviderName || la.Type != xago.AccTypeBalance {
		return nil, fmt.Errorf("%w linked account not correct type", xago.ErrNotFound)
	}

	ledger := xago.LedgerIDUSD
	opsAcc := xago.USDOpsAccount
	if la.SendCurrency == currency.ZAR {
		ledger = xago.LedgerIDZAR
		opsAcc = xago.ZAROpsAccount
	}

	tx, err := b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              txID,
			Amount:          amt.Value,
			CreditAccountID: la.ID,
			DebitAccountID:  opsAcc,
			Pending:         false,
			Code:            1,
			Ledger:          ledger,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}
	if len(tx) != 0 {
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return nil, fmt.Errorf("%w insufficiens balance cod (%s)", xago.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return nil, fmt.Errorf("%w non success code (%s)", xago.ErrInternal, tx[0].Code.String())
		}
	}

	accs, err := b.Pacioli().GetAccounts(ctx, []string{la.ID})
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	if len(accs) != 1 {
		return nil, fmt.Errorf("%w account not found", xago.ErrNotFound)
	}

	return &xago.Balance{
		Total:     currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted, la.SendCurrency),
		Available: currency.FromUInt64(accs[0].CreditsPosted-accs[0].DebitsPosted-accs[0].DebitsPending, la.SendCurrency),
	}, nil
}

func GetBankAccount(ctx context.Context, b Backends) (*xago.DepositDetails, error) {
	accounts, err := b.External().BankAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w %s", xago.ErrInternal, err)
	}
	var found *external.BankProvider
	for _, a := range *accounts {
		if a.CurrencyCode == currency.ZAR.String() && a.DepositEnabled {
			found = &a.BankingProviders[0]
			break
		}
	}
	if found == nil {
		return nil, fmt.Errorf("%w no bank account found", xago.ErrNotFound)
	}
	return &xago.DepositDetails{
		BankName:      found.DepositFields.BankName,
		AccountName:   found.DepositFields.AccountName,
		AccountNumber: found.DepositFields.AccountNumber,
		BranchCode:    found.DepositFields.BranchCode,
		CurrencyCode:  currency.ZAR,
	}, nil
}

// TestDeposit is only going to make the POST request. The deposit is going to
// be processed by the cronjob that is polling Xago deposits or by using the webhook listerner.
func TestDeposit(ctx context.Context, b Backends, sa xago.SubAccount) error {
	reqStruct := external.TestDepositReq{
		RunTestDeposit:    true,
		Amount:            200.00,
		DepositReference:  sa.DepositReference,
		BankTransactionID: uuid.New().String(),
		CurrencyCode:      string(currency.ZAR),
	}

	err := b.External().TestDeposit(ctx, reqStruct)

	if err != nil {
		return fmt.Errorf("%w %s", xago.ErrInternal, err)
	}

	return err
}
