package workflows

import (
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"

	"go.temporal.io/sdk/workflow"
)

const (
	TransactionEventsChannel         = "machnet_transaction_events"
	TransactionDeliveryEventsChannel = "machnet_delivery_events"
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
	err := workflow.ExecuteActivity(ctx, a.CreateExternalSendUser, walletID).Get(ctx, &externalUserID)
	if err != nil {
		logger.Error("CreateExternalSendUser Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.CreateUser, walletID, externalUserID).Get(ctx, nil)
	if err != nil {
		logger.Error("CreateExternalSendUser Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.StartExternalKYC, externalUserID).Get(ctx, nil)
	if err != nil {
		logger.Error("StartExternalKYC Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("CreateSendUserWorkflow completed.", "external_user_id", externalUserID)

	return externalUserID, nil
}

func CreateTransactionWorkflow(ctx workflow.Context, args machnet.CreateTransactionArgs) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("CreateTransactionWorkflow workflow started", "From", args.FromLinkedAccountID, "To", args.ToLinkedAccountID, "Amount", args.Amount)

	var to TransactionTo
	err := workflow.ExecuteActivity(ctx, a.GetOrCreateReceiveUser, args).Get(ctx, &to)
	if err != nil {
		logger.Error("GetOrCreateReceiveUser Activity failed.", "Error", err)
		return "", err
	}

	var trxID string
	err = workflow.ExecuteActivity(ctx, a.CreateExternalTransaction, args, to).Get(ctx, &trxID)
	if err != nil {
		logger.Error("CreateTransaction Activity failed.", "Error", err)
		return "", err
	}

	trxChan := workflow.GetSignalChannel(ctx, TransactionEventsChannel)
	var transactionCreatedSuccessfully bool
	for {
		var transactionEvent external.Event
		trxChan.Receive(ctx, &transactionEvent)
		logger.Info("status event: transactionID=", transactionEvent.ResourceID, "status=", transactionEvent.EventName)
		if transactionEvent.ResourceID != trxID { // not for this transaction
			logger.Error("Received notification for different transaction.")
			continue
		}

		if external.TransactionProcessedEvent == transactionEvent.EventName {
			transactionCreatedSuccessfully = true
			break
		}
	}
	if !transactionCreatedSuccessfully {
		return "", fmt.Errorf("%w Transaction failed.", machnet.ErrInternal)
	}

	err = workflow.ExecuteActivity(ctx, a.DeliverTransaction, args.FromLinkedAccountID, trxID).Get(ctx, nil)
	if err != nil {
		logger.Error("DeliverTransaction Activity failed.", "Error", err)
		return "", err
	}

	deliveryChan := workflow.GetSignalChannel(ctx, TransactionDeliveryEventsChannel)
	var deliverySuccessful bool
	for {
		var deliveryEvent external.Event
		deliveryChan.Receive(ctx, &deliveryEvent)
		logger.Info("delivery status event: transactionID=", deliveryEvent.ResourceID, "status=", deliveryEvent.EventName)
		if deliveryEvent.ResourceID != trxID { // not for this transaction
			logger.Error("Received delivery notification for different transaction.")
			continue
		}

		if external.TransactionDeliveredEvent == deliveryEvent.EventName {
			deliverySuccessful = true
			break
		}

		if external.TransactionDeliveryFailed == deliveryEvent.EventName {
			deliverySuccessful = false
			break
		}
	}
	if !deliverySuccessful {
		return "", fmt.Errorf("%w Transaction delivery failed.", machnet.ErrInternal)
	}

	logger.Info("CreateTransactionWorkflow completed.", "external_transaction_id", trxID)

	return trxID, nil
}
