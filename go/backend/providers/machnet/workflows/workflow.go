package workflows

import (
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/providers/machnet/ops"

	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
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

func CreateTransactionWorkflow(ctx workflow.Context, args machnet.CreateTransactionArgs) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("CreateTransactionWorkflow workflow started", "From", args.FromLinkedAccountID, "To", args.ToLinkedAccountID, "Amount", args.Amount)

	var fundWallet bool
	err := workflow.ExecuteActivity(ctx, a.ShouldFundWallet, args).Get(ctx, &fundWallet)
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

	transactionAmount := transactions.Amount{
		Value:      uint64(args.Amount * 100), // TODO: scale
		Asset:      args.Currency,
		AssetScale: 2, // TODO: Asset scale
	}

	// The fund wallet has 7 days to complete
	timeoutFuture := workflow.NewTimer(ctx, time.Hour*24*8)

	trxChan := workflow.GetSignalChannel(ctx, ops.TransactionEventsChannel)

	var doBreak bool
	var errToReturn error

	var fundWalletTX FundWalletResponse
	if fundWallet {
		err = workflow.ExecuteActivity(ctx, a.FundUserWalletFromCard, FundWalletArgs{
			CreateTransactionArgs: args,
			WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
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

		err = workflow.ExecuteActivity(ctx, a.AddTransactionTransfer, []transactions.TransferArgs{
			{
				WalletID:             transactionWallets.FromWalletID,
				TransactionForeignID: args.FromForeignID,
				LinkedAccountID:      args.FromLinkedAccountID,
				ForeignID:            fundWalletTX.FundTX,
				Type:                 transactions.TransferTypeDebitCard,
				Amount:               transactionAmount,
				State:                transactions.StatePending,
			},
			{
				WalletID:             transactionWallets.FromWalletID,
				TransactionForeignID: args.FromForeignID,
				LinkedAccountID:      fundWalletTX.FromWalletLinkedAcc,
				ForeignID:            fundWalletTX.FundTX,
				Type:                 transactions.TransferTypeCreditMachnetWallet,
				Amount:               transactionAmount,
				State:                transactions.StatePending,
			},
		}).Get(ctx, nil)
		if err != nil {
			logger.Error("AddTransaction Activity failed.", "Error", err)
			return "", err
		}

		for {
			if doBreak {
				break
			}
			selector := workflow.NewSelector(ctx)
			selector.AddFuture(timeoutFuture, func(f workflow.Future) {
				doBreak = true
				errToReturn = temporal.NewNonRetryableApplicationError("fund user wallet transaction has timed out", "ErrTimeout", machnet.ErrInternal)
			})

			selector.AddReceive(trxChan, func(c workflow.ReceiveChannel, _ bool) {
				var transaction external.Transaction
				trxChan.Receive(ctx, &transaction)
				logger.Info("status event: transactionID=", transaction.ID, "status=", transaction.Status)
				if transaction.ID != fundWalletTX.FundTX { // not for this transaction
					logger.Error("Received notification for different transaction.")
					return
				}

				if external.TransactionProcessed == transaction.Status {
					doBreak = true
					return
				}

				if transaction.Status == external.TransactionFailed ||
					transaction.Status == external.TransactionCancelled {
					doBreak = true
					errToReturn = temporal.NewNonRetryableApplicationError(fmt.Sprintf("fund user wallet transaction failed event(%s)", transaction.Status), "ErrInternal", external.ErrInternal)
				}
			})

			// Wait the timer or the transaction to complete on machnet side.
			selector.Select(ctx)
		}

		transactionState := transactions.StateCompleted
		if errToReturn != nil {
			transactionState = transactions.StateFailed
		}

		err = workflow.ExecuteActivity(ctx, a.UpdateTransactionTransfer, []transactions.TransferArgs{{
			WalletID:             transactionWallets.FromWalletID,
			TransactionForeignID: args.FromForeignID,
			ForeignID:            fundWalletTX.FundTX,
			Type:                 transactions.TransferTypeDebitCard,
			Amount:               transactionAmount,
			State:                transactionState,
		}, {
			WalletID:             transactionWallets.FromWalletID,
			TransactionForeignID: args.FromForeignID,
			ForeignID:            fundWalletTX.FundTX,
			Type:                 transactions.TransferTypeCreditMachnetWallet,
			Amount:               transactionAmount,
			State:                transactionState,
		},
		}).Get(ctx, nil)
		if err != nil {
			logger.Error("UpdateTransactionTransfer Activity failed.", "Error", err)
			return "", err
		}

		if errToReturn != nil {
			return "", errToReturn
		}
	}

	var transferID string
	err = workflow.ExecuteActivity(
		ctx,
		a.StartWalletTransfer,
		StartWalletTransferArgs{
			CreateTransactionArgs: args,
			WorkflowID:            workflow.GetInfo(ctx).WorkflowExecution.ID,
			FundingTx:             fundWalletTX,
		},
	).Get(ctx, &transferID)
	if err != nil {
		logger.Error("StartWalletTransfer Activity failed.", "Error", err)
		return "", err
	}

	var trxRefFromId string
	if fundWallet {
		trxRefFromId = fundWalletTX.FromWalletLinkedAcc
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

	err = workflow.ExecuteActivity(ctx, a.AddTransactionTransfer, []transactions.TransferArgs{
		{
			WalletID:             transactionWallets.FromWalletID,
			TransactionForeignID: args.FromForeignID,
			ForeignID:            transferID,
			LinkedAccountID:      fundWalletTX.FromWalletLinkedAcc,
			Type:                 transactions.TransferTypeDebitMachnetWallet,
			Amount:               transactionAmount,
			State:                transactions.StatePending,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("AddTransaction Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.AddTransaction, transactions.CreateTransactionArgs{
		WalletID:    transactionWallets.ToWalletID,
		ForeignID:   args.ToForeignID,
		ForeignType: transactions.TransactionTypeOpenPaymentIncoming,
		Provider:    transactions.ProviderOpenPayments,
		State:       transactions.StatePending,
		Source:      args.FromPaymentPointer,
		Destination: args.ToPaymentPointer,
		Amount:      transactionAmount,
		Transfers: []transactions.TransferArgs{
			{
				WalletID:             transactionWallets.ToWalletID,
				TransactionForeignID: args.ToForeignID,
				ForeignID:            transferID,
				LinkedAccountID:      args.ToLinkedAccountID,
				Type:                 transactions.TransferTypeCreditMachnetWallet,
				Amount:               transactionAmount,
				State:                transactions.StatePending,
			},
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("AddTransaction Activity failed.", "Error", err)
		return "", err
	}

	// Wait for webhook to say if transfer is successful
	doBreak = false
	for {
		if doBreak {
			break
		}
		selector := workflow.NewSelector(ctx)
		selector.AddFuture(timeoutFuture, func(f workflow.Future) {
			doBreak = true
			errToReturn = temporal.NewNonRetryableApplicationError("wallet to wallet transaction has timed out", "ErrTimeout", machnet.ErrInternal)
		})

		selector.AddReceive(trxChan, func(c workflow.ReceiveChannel, _ bool) {
			var transaction external.Transaction
			trxChan.Receive(ctx, &transaction)
			logger.Info("status event: transactionID=", transaction.ID, "status=", transaction.Status)
			if transaction.ID != transferID { // not for this transaction
				logger.Error("Received notification for different transaction.")
				return
			}

			if external.TransactionProcessed == transaction.Status {
				doBreak = true
				return
			}

			if transaction.Status == external.TransactionFailed ||
				transaction.Status == external.TransactionCancelled {
				doBreak = true
				errToReturn = temporal.NewNonRetryableApplicationError(fmt.Sprintf("wallet transfer failed failed event(%s)", transaction.Status), "ErrInternal", external.ErrInternal)
			}
		})

		// Wait the timer or the transaction to complete on machnet side.
		selector.Select(ctx)
	}
	transactionState := transactions.StateCompleted
	if errToReturn != nil {
		transactionState = transactions.StateFailed
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateTransactionTransfer, []transactions.TransferArgs{
		{
			WalletID:             transactionWallets.FromWalletID,
			TransactionForeignID: args.FromForeignID,
			ForeignID:            transferID,
			Type:                 transactions.TransferTypeDebitMachnetWallet,
			Amount:               transactionAmount,
			State:                transactionState,
		},
		{
			WalletID:             transactionWallets.ToWalletID,
			TransactionForeignID: args.ToForeignID,
			ForeignID:            transferID,
			Type:                 transactions.TransferTypeCreditMachnetWallet,
			Amount:               transactionAmount,
			State:                transactionState,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("UpdateTransactionTransfer Activity failed.", "Error", err)
		return "", err
	}

	if errToReturn != nil {
		return "", errToReturn
	}

	logger.Info("CreateTransactionWorkflow completed.", "fund_transfer_id", fundWalletTX.FundTX, "external_transfer_id", transferID)

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
	logger.Info("CreateWalletTopupWorkflow workflow started", "From", args.FromLinkedAccountID, "To Wallet", args.WalletLinkedAccountID, "Amount", args.Amount)

	floatAmount := float64(args.Amount / 100) // TODO: asset scale

	var fundWallet bool
	err := workflow.ExecuteActivity(ctx, a.ShouldFundWallet, machnet.CreateTransactionArgs{
		FromLinkedAccountID: args.FromLinkedAccountID,
		ToLinkedAccountID:   args.WalletLinkedAccountID,
		Amount:              floatAmount,
		Currency:            args.Currency,
		IPAddress:           args.IpAddress,
	}).Get(ctx, &fundWallet)
	if err != nil {
		logger.Error("ShouldFundWallet Activity failed.", "Error", err)
		return "", err
	}
	if !fundWallet {
		err = fmt.Errorf("Cannot fund wallet from linked account. (id=%s)", args.FromLinkedAccountID)
		logger.Error(err.Error(), "Error", err)
		return "", err
	}

	// The fund wallet has 7 days to complete
	timeoutFuture := workflow.NewTimer(ctx, time.Hour*24*7)

	trxChan := workflow.GetSignalChannel(ctx, ops.TransactionEventsChannel)

	var doBreak bool
	var errToReturn error

	var fundWalletTX FundWalletResponse
	err = workflow.ExecuteActivity(ctx, a.FundUserWalletFromCard, FundWalletArgs{
		CreateTransactionArgs: machnet.CreateTransactionArgs{
			FromLinkedAccountID: args.FromLinkedAccountID,
			ToLinkedAccountID:   args.WalletLinkedAccountID,
			Amount:              floatAmount,
			Currency:            args.Currency,
		},
		WorkflowID: workflow.GetInfo(ctx).WorkflowExecution.ID,
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

	transactionAmount := transactions.Amount{
		Value:      args.Amount,
		Asset:      args.Currency,
		AssetScale: 2, // TODO: Asset scale
	}

	err = workflow.ExecuteActivity(ctx, a.AddTransaction, transactions.CreateTransactionArgs{
		WalletID:    args.WalletID,
		ForeignID:   fundWalletTX.FundTX,
		ForeignType: transactions.TransactionTypeMachnetWalletTopUp,
		Provider:    transactions.ProviderMachnet,
		State:       transactions.StatePending,
		Amount:      transactionAmount,
		Transfers: []transactions.TransferArgs{
			{
				WalletID:             args.WalletID,
				TransactionForeignID: fundWalletTX.FundTX,
				ForeignID:            fundWalletTX.FundTX,
				LinkedAccountID:      args.FromLinkedAccountID,
				Type:                 transactions.TransferTypeDebitCard,
				Amount:               transactionAmount,
				State:                transactions.StatePending,
			},
			{
				WalletID:             args.WalletID,
				TransactionForeignID: fundWalletTX.FundTX,
				ForeignID:            fundWalletTX.FundTX,
				LinkedAccountID:      args.WalletLinkedAccountID,
				Type:                 transactions.TransferTypeCreditMachnetWallet,
				Amount:               transactionAmount,
				State:                transactions.StatePending,
			},
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("AddTransaction Activity failed.", "Error", err)
		return "", err
	}

	for {
		if doBreak {
			break
		}
		selector := workflow.NewSelector(ctx)
		selector.AddFuture(timeoutFuture, func(f workflow.Future) {
			doBreak = true
			errToReturn = temporal.NewNonRetryableApplicationError("fund user wallet transaction has timed out", "ErrTimeout", machnet.ErrInternal)
		})

		selector.AddReceive(trxChan, func(c workflow.ReceiveChannel, _ bool) {
			var transaction external.Transaction
			trxChan.Receive(ctx, &transaction)
			logger.Info("status event: transactionID=", transaction.ID, "status=", transaction.Status)
			if transaction.ID != fundWalletTX.FundTX { // not for this transaction
				logger.Error("Received notification for different transaction.")
				return
			}

			if external.TransactionProcessed == transaction.Status {
				doBreak = true
				return
			}

			if transaction.Status == external.TransactionFailed ||
				transaction.Status == external.TransactionCancelled {
				doBreak = true
				errToReturn = temporal.NewNonRetryableApplicationError(fmt.Sprintf("fund user wallet transaction failed event(%s)", transaction.Status), "ErrInternal", external.ErrInternal)
			}
		})

		// Wait the timer or the transaction to complete on machnet side.
		selector.Select(ctx)
	}

	transactionState := transactions.StateCompleted
	if errToReturn != nil {
		transactionState = transactions.StateFailed
	}

	transactionArgs := transactions.UpdateTransactionArgs{
		WalletID:  args.WalletID,
		ForeignID: fundWalletTX.FundTX,
		State:     transactionState,
		Amount:    transactionAmount,
		UpdateTransfers: []transactions.TransferArgs{{
			TransactionForeignID: fundWalletTX.FundTX,
			ForeignID:            fundWalletTX.FundTX,
			WalletID:             args.WalletID,
			Type:                 transactions.TransferTypeDebitCard,
			State:                transactionState,
			Amount:               transactionAmount,
		}, {
			TransactionForeignID: fundWalletTX.FundTX,
			ForeignID:            fundWalletTX.FundTX,
			WalletID:             args.WalletID,
			Type:                 transactions.TransferTypeCreditMachnetWallet,
			State:                transactionState,
			Amount:               transactionAmount,
		}},
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, transactionArgs).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update transaction state", "error", err, "state", transactionState)
		return "", err
	}

	if errToReturn != nil {
		return "", errToReturn
	}

	logger.Info("CreateWalletTopupWorkflow completed.", "fund_transfer_id", fundWalletTX.FundTX)

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

	trxChan := workflow.GetSignalChannel(ctx, ops.TransactionEventsChannel)

	var doBreak bool
	var errToReturn error

	var withdrawal machnet.WalletWithdrawal
	err := workflow.ExecuteActivity(ctx, a.WithdrawFromWallet, args).Get(ctx, &withdrawal)
	if err != nil {
		logger.Error("WithdrawFromWallet Activity failed.", "Error", err)
		return "", err
	}

	// The fund wallet has 4 days to complete
	timeoutFuture := workflow.NewTimer(ctx, time.Hour*24*4)

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

	transactionAmount := transactions.Amount{
		Value:      args.Amount,
		Asset:      "USD",
		AssetScale: 2, // TODO: Asset scale
	}

	err = workflow.ExecuteActivity(ctx, a.AddTransaction, transactions.CreateTransactionArgs{
		WalletID:    args.WalletID,
		ForeignID:   withdrawal.ID,
		ForeignType: transactions.TransactionTypeMachnetWalletWithdrawal,
		Provider:    transactions.ProviderMachnet,
		State:       transactions.StatePending,
		Amount:      transactionAmount,
		Transfers: []transactions.TransferArgs{
			{
				WalletID:             args.WalletID,
				TransactionForeignID: withdrawal.ID,
				ForeignID:            withdrawal.ID,
				LinkedAccountID:      args.WalletLinkedAccountID,
				Type:                 transactions.TransferTypeDebitMachnetWallet,
				Amount:               transactionAmount,
				State:                transactions.StatePending,
			},
			{
				WalletID:             args.WalletID,
				TransactionForeignID: withdrawal.ID,
				ForeignID:            withdrawal.ID,
				LinkedAccountID:      args.ToLinkedAccountID,
				Type:                 transactions.TransferTypeCreditBankAccount,
				Amount:               transactionAmount,
				State:                transactions.StatePending,
			},
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("AddTransaction Activity failed.", "Error", err)
		return "", err
	}

	for {
		if doBreak {
			break
		}
		selector := workflow.NewSelector(ctx)
		selector.AddFuture(timeoutFuture, func(f workflow.Future) {
			doBreak = true
			errToReturn = temporal.NewNonRetryableApplicationError("withdraw from wallet to bank account has timed out", "ErrTimeout", machnet.ErrInternal)
		})

		selector.AddReceive(trxChan, func(c workflow.ReceiveChannel, _ bool) {
			var transactionEvent external.Event
			trxChan.Receive(ctx, &transactionEvent)
			logger.Info("status event: transactionID=", transactionEvent.ResourceID, "status=", transactionEvent.EventName)
			if transactionEvent.ResourceID != withdrawal.ID { // not for this transaction
				logger.Error("Received notification for different transaction.")
				return
			}

			if external.TransactionProcessedEvent == transactionEvent.EventName {
				doBreak = true
				return
			}

			if transactionEvent.EventName == external.TransactionFailedEvent ||
				transactionEvent.EventName == external.TransactionCancelledEvent {
				doBreak = true
				errToReturn = temporal.NewNonRetryableApplicationError(fmt.Sprintf("withdraw from wallet to bank account failed event(%s)", transactionEvent.EventName), "ErrInternal", external.ErrInternal)
			}
		})

		// Wait the timer or the transaction to complete on machnet side.
		selector.Select(ctx)
	}

	transactionState := transactions.StateCompleted
	if errToReturn != nil {
		transactionState = transactions.StateFailed
	}

	transactionArgs := transactions.UpdateTransactionArgs{
		WalletID:  args.WalletID,
		ForeignID: withdrawal.ID,
		State:     transactionState,
		Amount:    transactionAmount,
		UpdateTransfers: []transactions.TransferArgs{{
			TransactionForeignID: withdrawal.ID,
			ForeignID:            withdrawal.ID,
			WalletID:             args.WalletID,
			Type:                 transactions.TransferTypeDebitMachnetWallet,
			State:                transactionState,
			Amount:               transactionAmount,
		}, {
			TransactionForeignID: withdrawal.ID,
			ForeignID:            withdrawal.ID,
			WalletID:             args.WalletID,
			Type:                 transactions.TransferTypeCreditBankAccount,
			State:                transactionState,
			Amount:               transactionAmount,
		}},
	}

	err = workflow.ExecuteActivity(ctx, a.UpdateTransactionState, transactionArgs).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update transaction state", "error", err, "state", transactionState)
		return "", err
	}

	if errToReturn != nil {
		return "", errToReturn
	}

	logger.Info("CreateWalletWithdrawalWorkflow completed.", "withdrawal_id", withdrawal.ID)

	return withdrawal.ID, nil
}
