package workflows

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"

	"gitlab.com/fynbos/backend/providers"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	gmt_workflows "gitlab.com/fynbos/backend/providers/gmt/ops"
	temporal_utils "gitlab.com/fynbos/backend/temporal/utils"
	"gitlab.com/fynbos/backend/transactions"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/api/enums/v1"
	temporal_client "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

func StartOutgoingPayment(ctx context.Context, b Backends, args openpayments.CreateOutgoingPaymentArgs) (*openpayments.OutgoingPayment, error) {
	span := trace.SpanFromContext(ctx)
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, err
	}

	// Validate the incoming, outgoing provider accounts exist
	q, err := ops.GetQuote(ctx, b, args.QuoteID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	ip, err := ops.GetIncomingPayment(ctx, b, q.IncomingPayment)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	recvPPURL := ip.PaymentPointer

	// Check that the recv payment pointer can receive
	_, err = getDefaultRecvAcc(ctx, b, recvPPURL)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Check that the sending payment pointer has the provider types
	var sendLA *linkedaccounts.LinkedAccount
	if q.FromLinkedAccount != "" {
		sendLA, err = b.LinkedAccounts().Get(ctx, q.FromLinkedAccount)
	} else {
		// Default to bank account or card if no linked account is specified
		sendLA, err = getDefaultSendAcc(ctx, b, q.PaymentPointer)
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	args.LinkedAccountTitle = sendLA.Title()

	// Check if we have already created this outgoing transaction and just return it.
	if args.IdempotencyKey != "" {
		existing, err := ops.GetOutgoingPayment(ctx, b, args.IdempotencyKey)
		if err != nil && !errors.Is(err, openpayments.ErrNotFound) {
			span.RecordError(err)
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

	id, trxID, err := ops.CreateOutgoingPayment(ctx, b, args)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	idxSlash := strings.LastIndex(id, "/")
	if idxSlash > 0 {
		id = id[idxSlash+1:]
	}

	workflowOptions := temporal_client.StartWorkflowOptions{
		ID:                       "openpayments_execute_outgoing_payment_" + id,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: time.Hour * 24 * 8, // Workflow has 8 days to complete
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	_, err = b.Temporal().ExecuteWorkflow(ctx, workflowOptions, OutgoingTransactionWorkflow, id, trxID, args.IPAddress, args.ThreeDSID)
	if err != nil {
		err = fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		span.RecordError(err)
		return nil, err
	}

	return ops.GetOutgoingPayment(ctx, b, id)
}

func OutgoingTransactionWorkflow(ctx workflow.Context, outgoingID, trxID, ipAddress, threeDSID string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("OutgoingTransactionWorkflow workflow started", "outgoingID", outgoingID, "trxID", trxID)

	var wfArgs ProviderWorkflowArgs
	err := workflow.ExecuteActivity(ctx, a.GetProviderWorkflowArgs, outgoingID).Get(ctx, &wfArgs)
	if err != nil {
		if temporal_utils.IsNonRetryableError(err) {
			innerErr := workflow.ExecuteActivity(ctx, a.FailOutgoingPayment, outgoingID).Get(ctx, nil)
			if innerErr != nil {
				return "", err
			}
		}
		logger.Error("GetProviderArgs Activity failed.", "Error", err)
		return "", err
	}
	tArgs := wfArgs.Args
	tArgs.FromTransactionID = trxID
	tArgs.IPAddress = ipAddress
	tArgs.ThreeDSID = threeDSID

	err = workflow.ExecuteActivity(ctx, a.AddContact, tArgs.FromPaymentPointer, tArgs.ToPaymentPointer).Get(ctx, nil)
	if err != nil {
		// Log but don't fail on error
		logger.Error("AddContact Activity failed.", "Error", err)
	}
	err = workflow.ExecuteActivity(ctx, a.MarkContactLastPaid, tArgs.FromPaymentPointer, tArgs.ToPaymentPointer).Get(ctx, nil)
	if err != nil {
		// Log but don't fail on error
		logger.Error("MarkContactLastPaid Activity failed.", "Error", err)
	}

	childWorkflowOptions := workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_REQUEST_CANCEL,
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)
	var resp providers.TransferResponse
	switch wfArgs.Key {
	case providers.GMTACH2ACH:
		err = workflow.ExecuteChildWorkflow(ctx, gmt_workflows.ACH2ACHTransferWorkflow, tArgs).Get(ctx, &resp)
	case providers.GMTACH2CARD:
		err = workflow.ExecuteChildWorkflow(ctx, gmt_workflows.ACH2CardTransferWorkflow, tArgs).Get(ctx, &resp)
	case providers.GMTCARD2ACH:
		err = workflow.ExecuteChildWorkflow(ctx, gmt_workflows.Card2ACHTransferWorkflow, tArgs).Get(ctx, &resp)
	case providers.GMTCARD2CARD:
		err = workflow.ExecuteChildWorkflow(ctx, gmt_workflows.Card2CardTransferWorkflow, tArgs).Get(ctx, &resp)
	default:
		err = temporal.NewNonRetryableApplicationError(fmt.Sprintf("unknown transfer worklfow (%s)", wfArgs.Key), "UNKNOWN_TRANSFER_WORKFLOW", nil)
	}
	if err != nil {
		logger.Error("OutgoingTransactionWorkflow child workflow failed.", "Error", err)
		if temporal_utils.IsNonRetryableError(err) {
			innerErr := workflow.ExecuteActivity(ctx, a.FailOutgoingPayment, outgoingID).Get(ctx, nil)
			if innerErr != nil {
				return "", err
			}
		}
		return "", err
	}
	if resp.OutgoingTransferState == transactions.StateFailed {
		err = workflow.ExecuteActivity(ctx, a.FailOutgoingPayment, outgoingID).Get(ctx, nil)
		if err != nil {
			logger.Error("FailOutgoingPayment Activity failed for failed transaction state.", "Error", err)
			return "", err
		}

		err = workflow.ExecuteActivity(ctx, a.SendFailedOutgoingPaymentMail, outgoingID).Get(ctx, nil)
		if err != nil {
			logger.Error("SendFailedOutgoingPaymentMail Activity failed.", "Error", err)
			return "", err
		}

		return "", nil
	}

	if resp.OutgoingTransferExternalID != "" {
		// Update Outgoing payment
		err = workflow.ExecuteActivity(ctx, a.CompleteOutgoingPayment, outgoingID, resp).Get(ctx, nil)
		if err != nil {
			logger.Error("GetProviderArgs Activity failed.", "Error", err)
			return "", err
		}

		err = workflow.ExecuteActivity(ctx, a.SendOutgoingPaymentReceipt, outgoingID, trxID).Get(ctx, nil)
		if err != nil {
			logger.Error("SendOutgoingPaymentReceipt Activity failed.", "Error", err)
			return "", err
		}
	}

	if resp.IncomingTransferState == transactions.StateCompleted {
		err = workflow.ExecuteActivity(ctx, a.SendIncomingPaymentReceipt, outgoingID, tArgs.ToForeignID).Get(ctx, nil)
		if err != nil {
			logger.Error("SendIncomingPaymentReceipt Activity failed.", "Error", err)
			return "", err
		}
	}

	return resp.OutgoingTransferExternalID, nil
}
