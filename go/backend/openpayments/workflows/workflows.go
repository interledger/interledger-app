package workflows

import (
	"context"
	"errors"
	"time"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_workflows "gitlab.com/fynbos/backend/providers/machnet/workflows"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func StartOutgoingTransaction(ctx context.Context, b Backends, quoteID string) {

}

type Activity struct {
	b Backends
}

func (a *Activity) GetProviderArgs(ctx context.Context, qID string) (*machnet.CreateTransactionArgs, error) {
	q, err := ops.GetQuote(ctx, a.b, qID)
	if err != nil {
		return nil, err
	}

	ops.ExtractPaymentPointer(q.IncomingPayment)

	recvPP, err := ops.GetPaymentPointer(ctx, a.b, q.PaymentPointer) // TODO
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
		return nil, openpayments.ErrPaymentPointerNotFound // TODO
	}

	sendPP, err := ops.GetPaymentPointer(ctx, a.b, q.PaymentPointer) // TODO
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
		return nil, openpayments.ErrPaymentPointerNotFound // TODO
	}

	return &machnet.CreateTransactionArgs{
		FromLinkedAccountID: sendAcc.ID,
		ToLinkedAccountID:   recvAcc.ID,
		Amount:              200,   // TODO
		Currency:            "USD", // TODO
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
