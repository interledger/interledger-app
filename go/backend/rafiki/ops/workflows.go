package ops

import (
	"context"
	"time"

	"gitlab.com/fynbos/log"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	RafikiGatehubSignalChannel = "rafiki_gatehub_signal"
	errMsgStoreTransferMapping = "failed to store gatehub transfer mapping"
	errMsgCancelOutgoingFailed = "failed to cancel outgoing payment"
)

type RafikiIncomingPaymentFinalizedArgs struct {
	IncomingPayment incomingPaymentData
	WebhookType     string
}

func StartRafikiIncomingPaymentsPolling(b ActivityBackends) {
	// This workflow ID can be user business logic identifier as well.
	workflowID := "cron_rafiki_web_monetization_payouts_pti"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "backend",
		CronSchedule:          "0 0 * * *",                                         // Every day at midnight
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING, // There can be only one
	}

	we, err := b.Temporal().ExecuteWorkflow(context.Background(), workflowOptions, WebMonetizationPaymentsWorkflow)
	if err != nil {
		log.Fatal("Unable to execute workflow", zap.Error(err))
	}
	log.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}

func WebMonetizationPaymentsWorkflow(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	var payouts []RafikiPayment
	err := workflow.ExecuteActivity(ctx, a.ListPaymentsToMake).Get(ctx, &payouts)
	if err != nil {
		logger.Error("failed to list outgoing payments set for payout", "err", err)
		return err
	}

	for _, p := range payouts {
		var paymentID string
		err = workflow.ExecuteActivity(ctx, a.CreateWebMonetizationPayment, p).Get(ctx, &paymentID)
		if err != nil {
			logger.Error("failed to create payment for web monetization", "err", err)
			// Don't return try the next one, we'll come back later and retry
			continue
		}

		err = workflow.ExecuteActivity(ctx, a.ConfirmPayment, paymentID).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to confirm payment for web monetization", "err", err)
			// Don't return try the next one, we'll come back later and retry
			continue
		}

		err = workflow.ExecuteActivity(ctx, a.AddWebMonetizationPayment, p, paymentID).Get(ctx, nil)
		if err != nil {
			logger.Error("failed to add payment ref to incoming payment payout", "err", err)
			return err
		}
	}

	return nil
}

// Handles `incoming_payment.completed` and `incoming_payment.expired`.
// It transfers funds from the intermediary GateHub account to the receiver's account,
// waits for the GateHub webhook signal, creates a transaction record and ledger transfer,
// and finally withdraws liquidity from Rafiki.
func RafikiIncomingPaymentFinalizedWorkflow(ctx workflow.Context, args RafikiIncomingPaymentFinalizedArgs) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	ip := args.IncomingPayment

	receivedAmt, err := parseAmountValue(ip.ReceivedAmount.Value)
	if err != nil {
		return err
	}

	if receivedAmt == 0 {
		logger.Info("incoming payment has no received amount, nothing to do", "paymentId", ip.ID)
		return nil
	}

	var accountInfo GatehubLinkedAccountInfo
	err = workflow.ExecuteActivity(ctx, a.GetGatehubLinkedAccountInfo, ip.WalletAddressID).Get(ctx, &accountInfo)
	if err != nil {
		logger.Error("failed to get gatehub linked account info", "paymentId", ip.ID, "err", err)
		return err
	}

	var gatehubTxID string
	err = workflow.ExecuteActivity(ctx, a.TransferFromIntermediaryToUser, accountInfo, ip.ReceivedAmount).Get(ctx, &gatehubTxID)
	if err != nil {
		logger.Error("failed to transfer from intermediary to user", "paymentId", ip.ID, "err", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.StoreGatehubTransferMapping, gatehubTxID, workflow.GetInfo(ctx).WorkflowExecution.ID).Get(ctx, nil)
	if err != nil {
		logger.Error(errMsgStoreTransferMapping, "paymentId", ip.ID, "err", err)
		return err
	}

	signalCh := workflow.GetSignalChannel(ctx, RafikiGatehubSignalChannel)
	signalCh.Receive(ctx, nil)

	err = workflow.ExecuteActivity(ctx, a.CreateIncomingPaymentTransaction, ip).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to create incoming payment transaction", "paymentId", ip.ID, "err", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateAndPostLedgerTransferForIncoming, ip).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to create and post ledger transfer", "paymentId", ip.ID, "err", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.WithdrawIncomingPaymentLiquidity, ip.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to withdraw incoming payment liquidity", "paymentId", ip.ID, "err", err)
		return err
	}

	return nil
}

// Handles `outgoing_payment.created`.
// It validates the user (KYC, receiver locality, currency match), transfers funds from
// the user's GateHub account to the intermediary, waits for the GateHub webhook signal,
// creates a transaction, reserves balance in the ledger, and deposits liquidity into Rafiki.
func RafikiOutgoingPaymentCreatedWorkflow(ctx workflow.Context, op outgoingPaymentData) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	var cancelReason string
	defer func() {
		if cancelReason == "" {
			return
		}
		err := workflow.ExecuteActivity(ctx, a.CancelOutgoingPayment, op.ID, cancelReason).Get(ctx, nil)
		if err != nil {
			logger.Error(errMsgCancelOutgoingFailed, "paymentId", op.ID, "err", err)
		}
	}()

	var validationResult ValidationResult
	err := workflow.ExecuteActivity(ctx, a.ValidateOutgoingPayment, op).Get(ctx, &validationResult)
	if err != nil {
		logger.Error("failed to validate outgoing payment", "paymentId", op.ID, "err", err)
		cancelReason = "validation error"
		return err
	}

	if !validationResult.Valid {
		logger.Warn("outgoing payment invalid, cancelling",
			"paymentId", op.ID,
			"reason", validationResult.Reason)
		cancelReason = validationResult.Reason
		// Don't treat this as a workflow error, just cancel the payment
		return nil
	}

	var gatehubTxID string
	err = workflow.ExecuteActivity(ctx, a.TransferFromUserToIntermediary, op.WalletAddressID, op.DebitAmount).Get(ctx, &gatehubTxID)
	if err != nil {
		logger.Error("failed to transfer from user to intermediary", "paymentId", op.ID, "err", err)
		cancelReason = "failed to transfer from user to intermediary"
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.StoreGatehubTransferMapping, gatehubTxID, workflow.GetInfo(ctx).WorkflowExecution.ID).Get(ctx, nil)
	if err != nil {
		logger.Error(errMsgStoreTransferMapping, "paymentId", op.ID, "err", err)
		cancelReason = errMsgStoreTransferMapping
		return err
	}

	signalCh := workflow.GetSignalChannel(ctx, RafikiGatehubSignalChannel)
	signalCh.Receive(ctx, nil)

	err = workflow.ExecuteActivity(ctx, a.CreateOutgoingPaymentTransaction, op).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to create outgoing payment transaction", "paymentId", op.ID, "err", err)
		cancelReason = "failed to create outgoing payment transaction"
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.ReserveBalanceForOutgoing, op).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to reserve balance for outgoing payment", "paymentId", op.ID, "err", err)
		cancelReason = "failed to reserve balance for outgoing payment"
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.DepositOutgoingPaymentLiquidity, op.ID).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to deposit outgoing payment liquidity", "paymentId", op.ID, "err", err)
		cancelReason = "failed to deposit outgoing payment liquidity"
		return err
	}

	return nil
}

// Handles `outgoing_payment.completed`.
// It marks the transaction as completed, posts the pending ledger transfer (finalizing it),
// and withdraws the outgoing payment liquidity from Rafiki.
func RafikiOutgoingPaymentCompletedWorkflow(ctx workflow.Context, op outgoingPaymentData) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	err := workflow.ExecuteActivity(ctx, a.UpdateOutgoingPaymentTransactionState, op, "Completed").Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update transaction state", "paymentId", op.ID, "err", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.PostLedgerTransferForOutgoing, op).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to post ledger transfer", "paymentId", op.ID, "err", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.WithdrawOutgoingPaymentLiquidity, op).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to withdraw outgoing payment liquidity", "paymentId", op.ID, "err", err)
		return err
	}

	return nil
}

// Handles `outgoing_payment.failed`.
// It reverses the user's funds (intermediary -> user), waits for the GateHub webhook signal,
// marks the transaction as failed, voids the ledger transfer, and withdraws liquidity from Rafiki.
func RafikiOutgoingPaymentFailedWorkflow(ctx workflow.Context, op outgoingPaymentData) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	var accountInfo GatehubLinkedAccountInfo
	err := workflow.ExecuteActivity(ctx, a.GetGatehubLinkedAccountInfo, op.WalletAddressID).Get(ctx, &accountInfo)
	if err != nil {
		logger.Error("failed to get gatehub linked account info", "paymentId", op.ID, "err", err)
		return err
	}

	var gatehubTxID string
	err = workflow.ExecuteActivity(ctx, a.TransferFromIntermediaryToUser, accountInfo, op.DebitAmount).Get(ctx, &gatehubTxID)
	if err != nil {
		logger.Error("failed to transfer from intermediary to user (refund)", "paymentId", op.ID, "err", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.StoreGatehubTransferMapping, gatehubTxID, workflow.GetInfo(ctx).WorkflowExecution.ID).Get(ctx, nil)
	if err != nil {
		logger.Error(errMsgStoreTransferMapping, "paymentId", op.ID, "err", err)
		return err
	}

	signalCh := workflow.GetSignalChannel(ctx, RafikiGatehubSignalChannel)
	signalCh.Receive(ctx, nil)

	err = workflow.ExecuteActivity(ctx, a.UpdateOutgoingPaymentTransactionState, op, "Failed").Get(ctx, nil)
	if err != nil {
		logger.Error("failed to update transaction state to failed", "paymentId", op.ID, "err", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.VoidLedgerTransferForOutgoing, op).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to void ledger transfer", "paymentId", op.ID, "err", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.WithdrawOutgoingPaymentLiquidity, op).Get(ctx, nil)
	if err != nil {
		logger.Error("failed to withdraw outgoing payment liquidity", "paymentId", op.ID, "err", err)
		return err
	}

	return nil
}
