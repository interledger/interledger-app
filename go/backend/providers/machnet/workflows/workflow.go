package workflows

import (
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	"gitlab.com/fynbos/backend/transactions"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func CreateSendUserWorkflow(ctx workflow.Context, walletID string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("CreateSendUserWorkflow workflow started", "walletID", walletID)

	var externalUserID string
	err := workflow.ExecuteActivity(ctx, a.UpsertExternalSendUser, walletID).Get(ctx, &externalUserID)
	if err != nil {
		logger.Error("UpsertExternalSendUser Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateUser, walletID, externalUserID).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateUser Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.StartExternalKYC, externalUserID).Get(ctx, nil)
	if err != nil {
		logger.Error("StartExternalKYC Activity failed.", "Error", err, "externalID", externalUserID)
		return "", err
	}

	// Wait for KYC passing
	workflowArgs := CreateUserWorkflowRefArgs{
		ExternalUserID: externalUserID,
		WorkflowID:     workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID:  workflow.GetInfo(ctx).WorkflowExecution.RunID,
		ActivityName:   "StartExternalKYC",
	}
	err = workflow.ExecuteActivity(ctx, a.CreateUserWorkflowRef, workflowArgs).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateUserWorkflowRef Activity failed.", "Error", err)
		return "", err
	}

	userChan := workflow.GetSignalChannel(ctx, ops.UserEventsChannel)
	for {
		var user external.User
		userChan.Receive(ctx, &user)
		logger.Info("status event: external user ID=", user.ID, "status=", user.Status)
		if user.ID != externalUserID { // not for this user
			logger.Error("Received notification for different user.")
			continue
		}

		if external.UserKYCVerified == user.Status {
			break
		}

		if user.Status == external.UserKYCRetry ||
			user.Status == external.UserKYCSuspended {
			err = workflow.ExecuteActivity(ctx, a.CompleteUserWorkflowRef, workflowArgs).Get(ctx, nil)
			if err != nil {
				logger.Error("CompleteUserWorkflowRef Activity failed for failed KYC event.", "Error", err)
			}
			return "", temporal.NewNonRetryableApplicationError(fmt.Sprintf("user (%s) KYC failed (%s)", externalUserID, user.Status), "ErrInternal", external.ErrInternal)
		}
	}

	err = workflow.ExecuteActivity(ctx, a.CompleteUserWorkflowRef, workflowArgs).Get(ctx, nil)
	if err != nil {
		logger.Error("CompleteUserWorkflowRef Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateWallet, externalUserID).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateWallet Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("CreateSendUserWorkflow completed.", "external_user_id", externalUserID)

	return externalUserID, nil
}

func CreateTransactionWorkflow(ctx workflow.Context, args machnet.CreateTransactionArgs, trxID string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("CreateTransactionWorkflow workflow started", "From", args.FromLinkedAccountID, "To", args.ToLinkedAccountID, "Amount", args.Amount)

	var fundWallet bool
	err := workflow.ExecuteActivity(ctx, a.ShouldFundWallet, args.FromLinkedAccountID).Get(ctx, &fundWallet)
	if err != nil {
		logger.Error("FundUserWalletFromCard Activity failed.", "Error", err)
		return "", err
	}

	var transactionWallets TransactionWalletIDs
	err = workflow.ExecuteActivity(ctx, a.GetTransactionsWallets, args).Get(ctx, &transactionWallets)
	if err != nil {
		logger.Error("GetTransactionsWallets Activity failed.", "Error", err)
		return "", err
	}

	var topupTX string
	if fundWallet {
		childWorkflowOptions := workflow.ChildWorkflowOptions{
			WorkflowID:            "transaction_wallet_top_up_" + trxID,
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
			ParentClosePolicy:     enums.PARENT_CLOSE_POLICY_ABANDON,
		}
		childCtx := workflow.WithChildOptions(ctx, childWorkflowOptions)

		err = workflow.ExecuteChildWorkflow(childCtx, ExecuteWalletTopupWorkflow, ExecuteTopupArgs{
			Amount:              args.Amount,
			TransactionID:       trxID,
			UpdateTransaction:   false,
			FromLinkedAccountID: args.FromLinkedAccountID,
			WalletID:            transactionWallets.FromWalletID,
			IPAddress:           args.IPAddress,
		}).Get(childCtx, &topupTX)
		if err != nil {
			logger.Error("ExecuteChildWorkflow failed to start for wallet top up", "Error", err)
			return "", err
		}
	}

	var transferID string
	err = workflow.ExecuteActivity(
		ctx,
		a.StartWalletTransfer,
		StartWalletTransferArgs{
			CreateTransactionArgs: args,
			WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
		},
	).Get(ctx, &transferID)
	if err != nil {
		logger.Error("StartWalletTransfer Activity failed.", "Error", err)
		return "", err
	}

	var trxRefFromId string
	if fundWallet {
		trxRefFromId = transactionWallets.FromWalletLinkedAcc
	} else {
		trxRefFromId = args.FromLinkedAccountID
	}

	err = workflow.ExecuteActivity(ctx, a.CreateTransactionWorkflowRef, CreateTransactionWorkflowRefArgs{
		FromLinkedAccountID:   trxRefFromId,
		ExternalTransactionID: transferID,
		WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID:         workflow.GetInfo(ctx).WorkflowExecution.RunID,
		AcitivityName:         "StartWalletTransfer",
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateTransactionWorkflowRef Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.AddTransactionTransfer, trxID, []transactions.TransferArgs{
		{
			ForeignID:       transferID,
			LinkedAccountID: trxRefFromId,
			Type:            transactions.TransferTypeDebitMachnetWallet,
			Amount:          args.Amount,
			State:           transactions.StatePending,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("AddTransactionTransfer Activity failed.", "Error", err)
		return "", err
	}

	var recvTrxID string
	err = workflow.ExecuteActivity(ctx, a.AddTransaction, transactions.CreateTransactionArgs{
		WalletID:    transactionWallets.ToWalletID,
		ForeignID:   args.ToForeignID,
		ForeignType: transactions.TransactionTypeOpenPaymentIncoming,
		Provider:    transactions.ProviderOpenPayments,
		State:       transactions.StatePending,
		Source:      args.FromPaymentPointer,
		Destination: args.ToPaymentPointer,
		Amount:      args.Amount,
		Transfers: []transactions.TransferArgs{
			{
				ForeignID:       transferID,
				LinkedAccountID: args.ToLinkedAccountID,
				Type:            transactions.TransferTypeCreditMachnetWallet,
				Amount:          args.Amount,
				State:           transactions.StatePending,
			},
		},
	}).Get(ctx, &recvTrxID)
	if err != nil {
		logger.Error("AddTransaction Activity failed.", "Error", err)
		return "", err
	}

	errToReturn := awaitTransactionState(ctx, time.Hour*24*8, transferID)

	transactionState := transactions.StateCompleted
	if errToReturn != nil {
		transactionState = transactions.StateFailed
	}

	// update send and receive transaction state.
	err = workflow.ExecuteActivity(ctx, a.UpdateTransferStateByType, trxID, transactionWallets.FromWalletID, transactions.TransferTypeDebitMachnetWallet, transactionState).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update transaction state", "error", err, "state", transactionState)
		return "", err
	}
	err = workflow.ExecuteActivity(ctx, a.UpdateTransferStateByType, recvTrxID, transactionWallets.ToWalletID, transactions.TransferTypeCreditMachnetWallet, transactionState).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update transaction state", "error", err, "state", transactionState)
		return "", err
	}

	if errToReturn != nil {
		return "", errToReturn
	}

	logger.Info("CreateTransactionWorkflow completed.", "fund_transfer_id", topupTX, "external_transfer_id", transferID)

	return transferID, nil
}

func DeleteAccountWorkflow(ctx workflow.Context, linkedAccountID string) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("DeleteAccountWorkflow workflow started", "linkedAccountID", linkedAccountID)

	err := workflow.ExecuteActivity(ctx, a.DeleteUserFundSource, linkedAccountID).Get(ctx, nil)
	if err != nil {
		logger.Error("DeleteUserFundSource Activity failed.", "Error", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.DeleteLinkedAccount, linkedAccountID).Get(ctx, nil)
	if err != nil {
		logger.Error("DeleteLinkedAccount Activity failed.", "Error", err)
		return err
	}

	return nil
}

func CreateWalletTopupWorkflow(ctx workflow.Context, args machnet.StartWalletTopupArgs) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("CreateWalletTopupWorkflow workflow started", "From", args.FromLinkedAccountID, "Amount", args.Amount)

	var trxID string
	err := workflow.ExecuteActivity(ctx, a.AddTransaction, transactions.CreateTransactionArgs{
		WalletID:    args.WalletID,
		ForeignType: transactions.TransactionTypeMachnetWalletTopUp,
		Provider:    transactions.ProviderMachnet,
		State:       transactions.StatePending,
		Amount:      args.Amount,
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: args.FromLinkedAccountID,
				Type:            transactions.TransferTypeDebitCard,
				Amount:          args.Amount,
				State:           transactions.StatePending,
			},
			{
				LinkedAccountID: args.WalletLinkedAccountID,
				Type:            transactions.TransferTypeCreditMachnetWallet,
				Amount:          args.Amount,
				State:           transactions.StatePending,
			},
		},
	}).Get(ctx, &trxID)
	if err != nil {
		logger.Error("AddTransaction Activity failed.", "Error", err)
		return "", err
	}

	childWorkflowOptions := workflow.ChildWorkflowOptions{
		WorkflowID:            "execute_wallet_top_up" + trxID,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		ParentClosePolicy:     enums.PARENT_CLOSE_POLICY_ABANDON,
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	var we workflow.Execution
	err = workflow.ExecuteChildWorkflow(ctx, ExecuteWalletTopupWorkflow, ExecuteTopupArgs{
		TransactionID:       trxID,
		UpdateTransaction:   true,
		WalletID:            args.WalletID,
		Amount:              args.Amount,
		FromLinkedAccountID: args.FromLinkedAccountID,
		IPAddress:           args.IpAddress,
	}).GetChildWorkflowExecution().Get(ctx, &we)
	if err != nil {
		logger.Error("ExecuteChildWorkflow failed to start", "Error", err)
		return "", err
	}
	// Child workflow execution has started. We can return

	return trxID, nil
}

type ExecuteTopupArgs struct {
	TransactionID       string
	UpdateTransaction   bool // Set to `true` for top-ups, `false` indicates it is part of a send transaction and the transaction is not complete when the top up is completed.
	WalletID            string
	Amount              currency.Amount
	FromLinkedAccountID string
	IPAddress           string
}

func ExecuteWalletTopupWorkflow(ctx workflow.Context, args ExecuteTopupArgs) (string, error) {
	var a *Activity

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("ExecuteWalletTopupWorkflow workflow started", "From", args.FromLinkedAccountID, "Amount", args.Amount)

	var fundWallet bool
	err := workflow.ExecuteActivity(ctx, a.ShouldFundWallet, args.FromLinkedAccountID).Get(ctx, &fundWallet)
	if err != nil {
		logger.Error("ShouldFundWallet Activity failed.", "Error", err)
		return "", err
	}
	if !fundWallet {
		// TODO: This should be a permanent error...
		err = fmt.Errorf("cannot fund wallet from linked account. (id=%s)", args.FromLinkedAccountID)
		logger.Error(err.Error(), "Error", err)
		return "", err
	}

	var trx transactions.Transaction
	err = workflow.ExecuteActivity(ctx, a.GetTransaction, args.WalletID, args.TransactionID).Get(ctx, &trx)
	if err != nil {
		logger.Error("error getting transaction.", "Error", err)
		return "", err
	}

	var fundWalletTX FundWalletResponse
	err = workflow.ExecuteActivity(ctx, a.FundUserWalletFromCard, FundWalletArgs{
		ExecuteTopupArgs: args,
		WorkflowID:       workflow.GetInfo(ctx).WorkflowExecution.ID,
		Transaction:      trx,
	}).Get(ctx, &fundWalletTX)
	if err != nil {
		logger.Error("FundUserWalletFromCard Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateTransactionWorkflowRef, CreateTransactionWorkflowRefArgs{
		FromLinkedAccountID:   args.FromLinkedAccountID,
		ExternalTransactionID: fundWalletTX.FundTX,
		WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID:         workflow.GetInfo(ctx).WorkflowExecution.RunID,
		AcitivityName:         "FundUserWalletFromCard",
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateTransactionWorkflowRef Activity failed.", "Error", err)
		return "", err
	}

	errToReturn := awaitTransactionState(ctx, time.Hour*24*7, fundWalletTX.FundTX)

	transactionState := transactions.StateCompleted
	if errToReturn != nil {
		transactionState = transactions.StateFailed
	}

	if args.UpdateTransaction {
		err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, trx.ID, transactionState).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to update transaction state", "error", err, "state", transactionState)
			return "", err
		}
	}
	for _, tfr := range trx.Transfers {
		// Ignore transfers that aren't related to the top-up.
		if tfr.Type != transactions.TransferTypeDebitCard &&
			tfr.Type != transactions.TransferTypeCreditMachnetWallet {
			continue
		}
		err = workflow.ExecuteActivity(ctx, a.UpdateTransferState, tfr.ID, transactionState).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to update transaction state", "error", err, "state", transactionState)
			return "", err
		}
	}

	// TODO: Don't fail
	if errToReturn != nil {
		return "", errToReturn
	}

	logger.Info("ExecuteWalletTopupWorkflow completed.", "fund_transfer_id", fundWalletTX.FundTX)

	return fundWalletTX.FundTX, nil
}

func CreateWalletWithdrawalWorkflow(ctx workflow.Context, args machnet.WithdrawFromWalletArgs) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("CreateWalletWithdrawalWorkflow workflow started", "From Wallet", args.WalletLinkedAccountID, "To ", args.ToLinkedAccountID, "Amount", args.Amount)

	transactionAmount := currency.Amount{
		Value:    args.Amount,
		Currency: currency.USD,
		Scale:    currency.USD.Scale(),
	}

	var trxID string
	err := workflow.ExecuteActivity(ctx, a.AddTransaction, transactions.CreateTransactionArgs{
		WalletID:    args.WalletID,
		ForeignType: transactions.TransactionTypeMachnetWalletWithdrawal,
		Provider:    transactions.ProviderMachnet,
		State:       transactions.StatePending,
		Amount:      transactionAmount,
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: args.WalletLinkedAccountID,
				Type:            transactions.TransferTypeDebitMachnetWallet,
				Amount:          transactionAmount,
				State:           transactions.StatePending,
			},
			{
				LinkedAccountID: args.ToLinkedAccountID,
				Type:            transactions.TransferTypeCreditBankAccount,
				Amount:          transactionAmount,
				State:           transactions.StatePending,
			},
		},
	}).Get(ctx, &trxID)
	if err != nil {
		logger.Error("AddTransaction Activity failed.", "Error", err)
		return "", err
	}

	childWorkflowOptions := workflow.ChildWorkflowOptions{
		WorkflowID:            "execute_wallet_withdrawal_" + trxID,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		ParentClosePolicy:     enums.PARENT_CLOSE_POLICY_ABANDON,
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	var we workflow.Execution
	err = workflow.ExecuteChildWorkflow(ctx, ExecuteWalletWithdrawalWorkflow, trxID, args).GetChildWorkflowExecution().Get(ctx, &we)
	if err != nil {
		logger.Error("ExecuteChildWorkflow failed to start", "Error", err)
		return "", err
	}
	// Child workflow execution has started. We can return

	return trxID, nil
}

func ExecuteWalletWithdrawalWorkflow(ctx workflow.Context, trxID string, args machnet.WithdrawFromWalletArgs) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("ExecuteWalletWithdrawalWorkflow workflow started", "From Wallet", args.WalletLinkedAccountID, "To ", args.ToLinkedAccountID, "Amount", args.Amount)

	var trx transactions.Transaction
	err := workflow.ExecuteActivity(ctx, a.GetTransaction, args.WalletID, trxID).Get(ctx, &trx)
	if err != nil {
		logger.Error("error getting transaction.", "Error", err)
		return "", err
	}

	var withdrawal machnet.WalletWithdrawal
	err = workflow.ExecuteActivity(ctx, a.WithdrawFromWallet, trx, args).Get(ctx, &withdrawal)
	if err != nil {
		logger.Error("WithdrawFromWallet Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateTransactionWorkflowRef, CreateTransactionWorkflowRefArgs{
		FromLinkedAccountID:   args.WalletLinkedAccountID,
		ExternalTransactionID: withdrawal.ID,
		WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
		WorkflowRunID:         workflow.GetInfo(ctx).WorkflowExecution.RunID,
		AcitivityName:         "WithdrawFromWallet",
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateTransactionWorkflowRef Activity failed.", "Error", err)
		return "", err
	}

	errToReturn := awaitTransactionState(ctx, time.Hour*24*4, withdrawal.ID)

	transactionState := transactions.StateCompleted
	if errToReturn != nil {
		transactionState = transactions.StateFailed
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, trx.ID, transactionState).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update transaction state", "error", err, "state", transactionState)
		return "", err
	}
	for _, tfr := range trx.Transfers {
		err = workflow.ExecuteActivity(ctx, a.UpdateTransferState, tfr.ID, transactionState).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to update transfer state", "error", err, "state", transactionState)
			return "", err
		}
	}

	if errToReturn != nil {
		return "", errToReturn
	}

	logger.Info("CreateWalletWithdrawalWorkflow completed.", "withdrawal_id", withdrawal.ID)

	return withdrawal.ID, nil
}

func awaitTransactionState(ctx workflow.Context, timeout time.Duration, transactionID string) error {
	logger := workflow.GetLogger(ctx)

	trxChan := workflow.GetSignalChannel(ctx, ops.TransactionEventsChannel)
	var doBreak bool
	var errToReturn error

	// The fund wallet has 4 days to complete
	timeoutFuture := workflow.NewTimer(ctx, timeout)

	for {
		if doBreak {
			break
		}
		selector := workflow.NewSelector(ctx)
		selector.AddFuture(timeoutFuture, func(f workflow.Future) {
			doBreak = true
			errToReturn = temporal.NewNonRetryableApplicationError("transaction has timed out", "ErrTimeout", machnet.ErrInternal)
		})

		selector.AddReceive(trxChan, func(c workflow.ReceiveChannel, _ bool) {
			var transaction external.Transaction
			trxChan.Receive(ctx, &transaction)
			logger.Info("status event: transactionID=", transaction.ID, "status=", transaction.Status)
			if transaction.ID != transactionID { // not for this transaction
				logger.Error("Received notification for different transaction.")
				return
			}

			if external.TransactionProcessed == transaction.Status {
				doBreak = true
				return
			}

			if transaction.Status == external.TransactionFailed ||
				transaction.Status == external.TransactionCancelled ||
				transaction.Status == external.TransactionReturned {
				doBreak = true
				errToReturn = temporal.NewNonRetryableApplicationError(fmt.Sprintf("transaction (%s) has failed with status(%s)", transactionID, transaction.Status), "ErrInternal", external.ErrInternal)
			}
		})

		// Wait the timer or the transaction to complete on machnet side.
		selector.Select(ctx)
	}

	return errToReturn
}
