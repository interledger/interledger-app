package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/pacioli"
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

func GetUser(ctx context.Context, b Backends, ec external.Client, walletID string) (*gatehub.User, error) {
	externalUserID, err := getExternalUserID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	usr, err := ec.GetUser(ctx, externalUserID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return usr, nil
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

func GetOnOffRampWidget(ctx context.Context, b Backends, ec external.Client, walletID string, isDeposit bool) (string, error) {
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

	widget, err := ec.GetOnOffRampWidget(ctx, externalUserID, isDeposit)
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

func getWalletID(ctx context.Context, b Backends, externalUserID string) (string, error) {
	var walletID string
	err := b.DB().GetContext(ctx, &walletID, "SELECT wallet_id FROM gatehub_users WHERE external_id=$1;", externalUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	} else if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return walletID, nil
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

func CreateWithdrawal(ctx context.Context, b Backends, ec external.Client, walletID, externalTransactionID string) (string, error) {
	existingWithdrawal, err := b.Transactions().GetTransactionByForeignID(ctx, walletID, externalTransactionID)
	if err != nil && !errors.Is(err, transactions.ErrNotFound) {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if existingWithdrawal != nil {
		_, err = processWithdrawal(ctx, b, walletID, existingWithdrawal.ID)
		if err != nil {
			return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
		}
		return existingWithdrawal.ID, nil
	}

	amount, balanceAccount, fee, err := validateWithdrawal(ctx, b, ec, walletID, externalTransactionID)
	if err != nil {
		return "", err
	}

	wallet, err := b.Wallets().Get(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	trxID, err := b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID:                walletID,
		ForeignID:               externalTransactionID,
		Provider:                gatehub.ProviderName,
		State:                   transactions.StatePending,
		ForeignType:             transactions.TransactionTypeWithdrawal,
		Source:                  wallet.AddressString(),
		Destination:             wallet.AddressString(),
		Title:                   "Withdrawal",
		DestinationIdentity:     walletID,
		ProviderFee:             &fee,
		DestinationIdentityType: payments.IdentityTypeWalletID.String(),
		Amount:                  amount,
		LinkedAccountTitle:      "EUR Balance",
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: balanceAccount.ID,
				ForeignID:       externalTransactionID,
				Amount:          amount,
				Type:            transactions.TransferTypeDebitBalance,
				State:           transactions.StatePending,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	_, err = processWithdrawal(ctx, b, walletID, trxID)
	if err != nil {
		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return trxID, err
}

func validateWithdrawal(ctx context.Context, b Backends, ec external.Client, walletID, externalTransactionID string) (currency.Amount, *linkedaccounts.LinkedAccount, currency.Amount, error) {
	externalUserID, err := getExternalUserID(ctx, b, walletID)
	if err != nil {
		return currency.Amount{}, nil, currency.Amount{}, err
	}

	trx, err := ec.GetTransaction(ctx, externalUserID, externalTransactionID)
	if err != nil {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	fee, err := StringToScaledUInt(trx.Fee)
	if err != nil {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w invalid fee", gatehub.ErrInternal)
	}

	if err != nil {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if trx.Type != external.TransactionTypeWithdrawal {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w Transaction is not a withdrawal", gatehub.ErrInternal)
	}
	cc := currency.ParseCurrency(trx.Vault.AssetCode)
	if cc != currency.EUR {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w Invalid currency", gatehub.ErrInternal)
	}

	balances, err := b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	var balance *linkedaccounts.LinkedAccount
	for _, bal := range balances {
		if bal.Provider == gatehub.ProviderName && bal.Type == gatehub.AccTypeBalance {
			balance = &bal
			break
		}
	}
	if balance == nil {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w Gatehub balance linked account not found", gatehub.ErrNotFound)
	}
	if balance.ProviderID != trx.SendingWallet.Address {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w Gatehub withdrawal is not for this Interledger wallet", gatehub.ErrInternal)
	}

	parts := strings.Split(trx.Amount, ".")
	if len(parts) < 1 {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w invalid amount", gatehub.ErrInternal)
	}

	value, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return currency.Amount{}, nil, currency.Amount{}, fmt.Errorf("%w invalid amount", gatehub.ErrInternal)
	}

	return currency.Amount{
			Value:    value * 100, // EUR scale = 2
			Currency: cc,
		}, balance, currency.Amount{
			Value:    fee,
			Currency: cc,
			Scale:    2,
		}, nil
}

func processWithdrawal(ctx context.Context, b Backends, walletID, transactionID string) (gatehub.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("gatehub_create_withdrawal_%s", transactionID),
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
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, ProcessGatehubWithdrawal, walletID, transactionID)
	}
	if executeErr != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return await.Get, nil
}

func ReserveBalance(ctx context.Context, b Backends, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*gatehub.Balance, error) {
	if amt.Currency != currency.EUR {
		return nil, fmt.Errorf("%w %s not supported.", gatehub.ErrInternal, amt.Currency)
	}

	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, err
	}
	if la.Provider != gatehub.ProviderName || la.Type != gatehub.AccTypeBalance {
		return nil, fmt.Errorf("%w Not a Gatehub linked account", gatehub.ErrInternal)
	}

	opsAcc := gatehub.EUROpsAccount
	ledger := gatehub.LedgerIDEUR
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
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(tx) > 0 {
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return nil, fmt.Errorf("%w insufficiens balance cod (%s)", gatehub.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return nil, fmt.Errorf("%w non success code (%s)", gatehub.ErrInternal, tx[0].Code.String())
		}
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

func FinaliseReserve(ctx context.Context, b Backends, trxID string) error {
	tx, err := b.Pacioli().PostTransfers(ctx, []string{trxID})
	if err != nil {
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(tx) == 0 {
		return nil
	}
	if tx[0].Code == pacioli.TransferPendingTransferAlreadyPosted {
		return nil
	}
	if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
		return fmt.Errorf("%w insufficient balance code (%s)", gatehub.ErrInsufficientBalance, tx[0].Code.String())
	}
	if tx[0].Code == pacioli.TransferPendingTransferExpired {
		return fmt.Errorf("%w transfer timed out code (%s)", gatehub.ErrTimedOut, tx[0].Code.String())
	}
	if tx[0].Code == pacioli.TransferPendingTransferNotPending {
		return fmt.Errorf("%w pending transfer not found code (%s)", gatehub.ErrNotFound, tx[0].Code.String())
	}
	if tx[0].Code != 0 {
		return fmt.Errorf("%w non success code (%s)", gatehub.ErrInternal, tx[0].Code.String())
	}

	return nil
}

func RollbackReserve(ctx context.Context, b Backends, txID string) error {
	tx, err := b.Pacioli().VoidTransfers(ctx, []string{txID})
	if err != nil {
		slack.SendToChannel(ctx, slack.ChannelNotifyErrors, "wallet-info-bot", fmt.Sprintf("*:::[GateHub ERROR]:::* \n *RollbackReserve  txID:* %s,\n *error:* %s", txID, err))
		return fmt.Errorf("%w %s", gatehub.ErrInternal, err)
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
		return fmt.Errorf("%w non success code (%s)", gatehub.ErrInternal, tx[0].Code.String())
	}

	return nil
}

func AssignBalance(ctx context.Context, b Backends, linkedAccountID, txID string, amt currency.Amount) (*gatehub.Balance, error) {
	la, err := b.LinkedAccounts().Get(ctx, linkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if la.Provider != gatehub.ProviderName || la.Type != gatehub.AccTypeBalance {
		return nil, fmt.Errorf("%w linked account not correct type", gatehub.ErrNotFound)
	}

	opsAcc := gatehub.EUROpsAccount
	ledger := gatehub.LedgerIDEUR
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
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	if len(tx) != 0 {
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return nil, fmt.Errorf("%w insufficiens balance cod (%s)", gatehub.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return nil, fmt.Errorf("%w non success code (%s)", gatehub.ErrInternal, tx[0].Code.String())
		}
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

func CreateTransfer(ctx context.Context, b Backends, ec external.Client, args gatehub.CreateTransferArgs) (*external.Transaction, error) {
	sendLA, err := b.LinkedAccounts().Get(ctx, args.SendingLinkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	recvLA, err := b.LinkedAccounts().Get(ctx, args.ReceivingLinkedAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	recvWallet, err := b.Wallets().Get(ctx, recvLA.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	sendingUser, err := getExternalUserID(ctx, b, sendLA.WalletID)
	if err != nil {
		return nil, err
	}

	externalTx, err := ec.CreateTransaction(ctx, external.CreateTransactionRequest{
		SendingUserID:    sendingUser,
		SendingAddress:   sendLA.ProviderID,
		ReceivingAddress: recvLA.ProviderID,
		Amount:           args.Amount.Float64(),
		Message:          fmt.Sprintf("Payment to %s", recvWallet.Name),
		Type:             external.TransactionTypeHosted,
		VaultID:          ec.GetVaultID(),
	})
	if errors.Is(err, external.ErrNotFound) {
		return nil, fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return externalTx, nil
}

func GetTransaction(ctx context.Context, b Backends, ec external.Client, walletID, id string) (*external.Transaction, error) {
	externalUser, err := getExternalUserID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	return ec.GetTransaction(ctx, externalUser, id)
}
