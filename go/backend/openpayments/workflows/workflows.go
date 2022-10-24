package workflows

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	temporal_client "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_workflows "gitlab.com/fynbos/backend/providers/machnet/workflows"
	"go.temporal.io/api/enums/v1"

	"go.temporal.io/sdk/workflow"

	"gitlab.com/fynbos/backend/openpayments/ops"

	"gitlab.com/fynbos/backend/openpayments"
)

func StartOutgoingPayment(ctx context.Context, b Backends, args openpayments.CreateOutgoingPaymentArgs) (*openpayments.OutgoingPayment, error) {
	id, err := ops.CreateOutgoingPayment(ctx, b, args)
	if err != nil {
		return nil, err
	}

	workflowOptions := temporal_client.StartWorkflowOptions{
		ID:        "openpayments_execute_outgoing_payment_" + id,
		TaskQueue: "backend",
	}

	_, err = b.Temporal().ExecuteWorkflow(ctx, workflowOptions, OutgoingTransactionWorkflow, id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", openpayments.ErrInternal, err)
	}

	return ops.GetOutgoingPayment(ctx, b, id)
}

type Activity struct {
	b Backends
}

func (a *Activity) GetProviderArgs(ctx context.Context, outgoingID string) (*machnet.CreateTransactionArgs, error) {
	op, err := ops.GetOutgoingPayment(ctx, a.b, outgoingID)
	if err != nil {
		return nil, err
	}

	recvPPURL, _, err := ops.ExtractPaymentPointer(op.Receiver)

	recvPP, err := ops.GetPaymentPointer(ctx, a.b, recvPPURL)
	if err != nil {
		return nil, err
	}

	recvAccs, err := a.b.LinkedAccounts().ListByWalletId(ctx, recvPP.WalletID)
	if err != nil {
		return nil, err
	}

	var found bool
	var recvAcc linkedaccounts.LinkedAccount
	for _, ra := range recvAccs {
		if ra.Provider != machnet.ProviderName ||
			ra.Type != machnet.TypeReceiveBankAccount {
			continue
		}
		found = true
		recvAcc = ra
		break
	}

	if !found {
		return nil, openpayments.ErrPaymentPointerNotFound // TODO: Non retry
	}

	sendPP, err := ops.GetPaymentPointer(ctx, a.b, op.PaymentPointer) // TODO
	if err != nil {
		return nil, err
	}

	sendAccs, err := a.b.LinkedAccounts().ListByWalletId(ctx, sendPP.WalletID)
	if err != nil {
		return nil, err
	}

	found = false
	var sendAcc linkedaccounts.LinkedAccount
	for _, sa := range sendAccs {
		if sa.Provider != machnet.ProviderName ||
			sa.Type != machnet.TypeSendCard {
			continue
		}
		found = true
		sendAcc = sa
		break
	}

	if !found {
		return nil, openpayments.ErrPaymentPointerNotFound // TODO: Non retry
	}

	// TODO: Check this in unit tests
	amnt := float64(op.SendAmount.Value)
	if op.SendAmount.AssetScale > 0 {
		amnt /= math.Pow(10, float64(op.SendAmount.AssetScale))
	}

	return &machnet.CreateTransactionArgs{
		FromLinkedAccountID: sendAcc.ID,
		ToLinkedAccountID:   recvAcc.ID,
		Amount:              amnt,
		Currency:            op.SendAmount.Asset,
	}, nil
}

func OutgoingTransactionWorkflow(ctx workflow.Context, outgoingID string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("OutgoingTransactionWorkflow workflow started", "outgoingID", outgoingID)

	var tArgs machnet.CreateTransactionArgs
	err := workflow.ExecuteActivity(ctx, a.GetProviderArgs, outgoingID).Get(ctx, &tArgs)
	if err != nil {
		logger.Error("GetProviderArgs Activity failed.", "Error", err)
		return "", err
	}

	childWorkflowOptions := workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_REQUEST_CANCEL,
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	var extID string
	err = workflow.ExecuteChildWorkflow(ctx, machnet_workflows.CreateTransactionWorkflow, tArgs).Get(ctx, &extID)
	if err != nil {
		if isNonRetryableError(err) {
			// Update outgoing payment failed
		}
		return "", err
	}

	// Update Outgoing payment
	return "resp", nil
}

func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if !temporal.IsApplicationError(err) {
		return false
	}

	var applicationError *temporal.ApplicationError
	errors.As(err, &applicationError)

	return applicationError.NonRetryable()
}
