package workflows

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"

	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_workflows "gitlab.com/fynbos/backend/providers/machnet/workflows"
	"go.temporal.io/api/enums/v1"
	temporal_client "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

func StartOutgoingPayment(ctx context.Context, b Backends, args openpayments.CreateOutgoingPaymentArgs) (*openpayments.OutgoingPayment, error) {
	// Validate the incoming, outgoing provider accounts exist
	q, err := ops.GetQuote(ctx, b, args.QuoteID)
	if err != nil {
		return nil, err
	}

	recvPPURL, _, err := ops.ExtractPaymentPointer(q.IncomingPayment)
	if err != nil {
		return nil, err
	}

	// Check that the recv payment pointer can receive
	_, err = getProviderLinkedAccount(ctx, b, recvPPURL, machnet.ProviderName, machnet.TypeWallet)
	if err != nil {
		return nil, err
	}

	// Check that the sending payment pointer has the provider types
	_, err = getProviderLinkedAccount(ctx, b, q.PaymentPointer, machnet.ProviderName, machnet.TypeSendCard)
	if err != nil {
		return nil, err
	}

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
		if isNonRetryableError(err) {
			innerErr := workflow.ExecuteActivity(ctx, a.FailOutgoingPayment, outgoingID).Get(ctx, nil)
			if innerErr != nil {
				return "", err
			}
		}
		logger.Error("GetProviderArgs Activity failed.", "Error", err)
		return "", err
	}

	childWorkflowOptions := workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_REQUEST_CANCEL,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 4,
		},
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	var extID string
	err = workflow.ExecuteChildWorkflow(ctx, machnet_workflows.CreateTransactionWorkflow, tArgs).Get(ctx, &extID)
	if err != nil {
		logger.Error("CreateTransactionWorkflow child workflow failed.", "Error", err)
		if isNonRetryableError(err) {
			innerErr := workflow.ExecuteActivity(ctx, a.FailOutgoingPayment, outgoingID).Get(ctx, nil)
			if innerErr != nil {
				return "", err
			}
		}
		return "", err
	}

	// Update Outgoing payment
	err = workflow.ExecuteActivity(ctx, a.CompleteOutgoingPayment, outgoingID, extID).Get(ctx, nil)
	if err != nil {
		logger.Error("GetProviderArgs Activity failed.", "Error", err)
		return "", err
	}

	return extID, nil
}
