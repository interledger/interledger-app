package workflows

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_workflows "gitlab.com/fynbos/backend/providers/machnet/workflows"
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
		return nil, fmt.Errorf("%w %s", machnet.ErrInvalidArgument, err)
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
	_, err = getProviderLinkedAccount(ctx, b, recvPPURL, machnet.ProviderName, machnet.TypeWallet)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Check that the sending payment pointer has the provider types
	var sendLA *linkedaccounts.LinkedAccount
	if q.FromLinkedAccount != "" {
		sendLA, err = b.LinkedAccounts().Get(ctx, q.FromLinkedAccount)
	} else {
		// Default to send card if no linked account is specified
		sendLA, err = getProviderLinkedAccount(ctx, b, q.PaymentPointer, machnet.ProviderName, machnet.TypeSendCard)
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	if sendLA.Provider != machnet.ProviderName ||
		(sendLA.Type != machnet.TypeSendCard && sendLA.Type != machnet.TypeWallet) {
		return nil, fmt.Errorf("%w send linked account (%s) not send enabled", openpayments.ErrInternal, q.FromLinkedAccount)
	}

	limits, err := b.Machnet().GetUserLimits(ctx, sendLA.WalletID)
	if err != nil {
		return nil, err
	}

	fromCard := sendLA.Type == machnet.TypeSendCard

	exceeds, exceedsType := limits.Transfer.Exceeds(q.SendAmount, false)
	if exceeds {
		switch exceedsType {
		case "ANNUAL":
			return nil, machnet.ErrUserAnnualLimit
		case "MONTHLY":
			return nil, machnet.ErrUserMonthlyLimit
		case "DAILY":
			return nil, machnet.ErrUserDailyLimit
		}
	}
	exceeds, exceedsType = limits.FundWallet.Exceeds(q.SendAmount, true)
	if fromCard && exceeds {
		switch exceedsType {
		case "ANNUAL":
			return nil, machnet.ErrUserAnnualLimit
		case "MONTHLY":
			return nil, machnet.ErrUserMonthlyLimit
		case "DAILY":
			return nil, machnet.ErrUserDailyLimit
		case "HOLD":
			return nil, machnet.ErrUserHoldLimit
		}
	}

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

	_, err = b.Temporal().ExecuteWorkflow(ctx, workflowOptions, OutgoingTransactionWorkflow, id, trxID, args.IPAddress)
	if err != nil {
		err = fmt.Errorf("%w %s", openpayments.ErrInternal, err)
		span.RecordError(err)
		return nil, err
	}

	return ops.GetOutgoingPayment(ctx, b, id)
}

func OutgoingTransactionWorkflow(ctx workflow.Context, outgoingID, trxID, ipAddress string) (string, error) {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("OutgoingTransactionWorkflow workflow started", "outgoingID", outgoingID, "trxID", trxID)

	var tArgs machnet.CreateTransactionArgs
	err := workflow.ExecuteActivity(ctx, a.GetProviderArgs, outgoingID).Get(ctx, &tArgs)
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
	tArgs.IPAddress = ipAddress

	childWorkflowOptions := workflow.ChildWorkflowOptions{
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_REQUEST_CANCEL,
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	var resp machnet.CreateTransactionResponse
	err = workflow.ExecuteChildWorkflow(ctx, machnet_workflows.CreateTransactionWorkflow, tArgs, trxID).Get(ctx, &resp)
	if err != nil {
		logger.Error("CreateTransactionWorkflow child workflow failed.", "Error", err)
		if temporal_utils.IsNonRetryableError(err) {
			innerErr := workflow.ExecuteActivity(ctx, a.FailOutgoingPayment, outgoingID).Get(ctx, nil)
			if innerErr != nil {
				return "", err
			}
		}
		return "", err
	}
	if resp.TransactionState == transactions.StateFailed {
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

	// Update Outgoing payment
	err = workflow.ExecuteActivity(ctx, a.CompleteOutgoingPayment, outgoingID, resp.ExternalID).Get(ctx, nil)
	if err != nil {
		logger.Error("GetProviderArgs Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.SendOutgoingPaymentReceipt, outgoingID, resp.ExternalID).Get(ctx, nil)
	if err != nil {
		logger.Error("SendOutgoingPaymentReceipt Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, a.SendIncomingPaymentReceipt, outgoingID).Get(ctx, nil)
	if err != nil {
		logger.Error("SendIncomingPaymentReceipt Activity failed.", "Error", err)
		return "", err
	}

	return resp.ExternalID, nil
}
