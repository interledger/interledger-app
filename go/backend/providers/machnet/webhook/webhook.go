package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/fynbos/backend/providers/machnet/ops"

	"gitlab.com/fynbos/log"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
)

const SignatureHeader = "x-raas-webhook-signature"

func New(globalBackends Backends, webhookSecret string) http.HandlerFunc {
	b := opsBackends{
		Backends: globalBackends,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		err = ValidateWebhook(body, webhookSecret, r.Header.Get(SignatureHeader))
		if errors.Is(err, machnet.ErrInvalidSignature) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		var event external.Event
		err = json.Unmarshal(body, &event)
		if err != nil {
			http.Error(w, "failed to parse payload", http.StatusBadRequest)
			return
		}

		// TODO: validate payload
		err = HandleEvent(r.Context(), b, event)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleEvent(ctx context.Context, b ops.Backends, event external.Event) error {
	var err error
	switch event.EventName {
	case external.UserCardAdded:
		err = HandleUserCardAddedEvent(ctx, b, event)
	case external.UserKYCInProgress, external.UserKYCSuspended, external.UserKYCRetry, external.UserKYCVerified, external.UserKYCReviewPending:
		err = HandleUserKYCEvent(ctx, b, event)
	case external.TransactionPendingEvent, external.TransactionProcessingEvent, external.TransactionHoldEvent,
		external.TransactionProcessedEvent, external.TransactionCancelledEvent, external.TransactionFailedEvent,
		external.TransactionReturnedEvent:
		err = HandleTransactionEvent(ctx, b, event)
	case external.TransactionDeliveryHoldEvent, external.TransactionDeliveryPendingEvent, external.TransactionDeliveryRequestedEvent,
		external.TransactionDeliveredEvent, external.TransactionDeliveryFailedEvent, external.TransactionDeliveryAuthorizedEvent,
		external.TransactionDeliveryPayoutReadyEvent:
		err = HandleTransactionDeliveryEvent(ctx, b, event)
	default:
		log.Warn(
			"Unhandled machnet event",
			zap.String("eventName", event.EventName),
			zap.String("externalUserID", event.UserID),
			zap.String("externalResourceID", event.ResourceID),
			zap.String("body", string(event.Payload)),
		)
	}
	if err != nil {
		return err
	}

	return nil
}

func ValidateWebhook(payload []byte, secret, signature string) error {
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write(payload); err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	if signature != hex.EncodeToString(mac.Sum(nil)) {
		return machnet.ErrInvalidSignature
	}

	return nil
}

func HandleUserKYCEvent(ctx context.Context, b ops.Backends, event external.Event) error {
	_, err := ops.GetUserByID(ctx, b, event.UserID)
	if err != nil {
		return err
	}

	var newStatus machnet.KYCStatus
	switch event.EventName {
	case external.UserKYCInProgress:
		newStatus = machnet.KYCStatusInProgress
	case external.UserKYCSuspended:
		newStatus = machnet.KYCStatusSuspended
	case external.UserKYCRetry:
		newStatus = machnet.KYCStatusRetry
	case external.UserKYCVerified:
		newStatus = machnet.KYCStatusVerified
	case external.UserKYCReviewPending:
		newStatus = machnet.KYCStatusReviewPending
	}

	_, err = b.DB().ExecContext(ctx, "UPDATE machnet_users SET updated_at=now(), kyc_status=$1 WHERE id=$2", newStatus, event.UserID)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	refs, err := ops.ListActiveUserWorkflowRefs(ctx, b, event.UserID)
	if err != nil {
		return err
	}

	for _, ref := range refs {
		err = b.Temporal().SignalWorkflow(ctx, ref.WorkflowID, ref.WorkflowRunID, ops.UserEventsChannel, event)
		if err != nil {
			return fmt.Errorf("%w %s", machnet.ErrInternal, err)
		}
	}

	return nil
}

func HandleUserCardAddedEvent(ctx context.Context, b ops.Backends, event external.Event) error {
	user, err := ops.GetUserByID(ctx, b, event.UserID)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	// TODO: find out if these details are in the event payload
	card, err := b.External().GetUserFundingsource(ctx, user.ID, event.ResourceID)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	_, err = b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:   user.WalletID,
		Name:       card.FundingsourceName,
		Mask:       card.AccountNumber,
		Provider:   machnet.ProviderName,
		ProviderID: card.ID,
		Type:       machnet.TypeSendCard,
	})
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return nil
}

func HandleTransactionEvent(ctx context.Context, b ops.Backends, event external.Event) error {
	trx, err := ops.GetTransactionWorkflowRef(ctx, b, event.UserID, event.ResourceID)
	if err != nil {
		return err
	}

	err = b.Temporal().SignalWorkflow(ctx, trx.WorkflowID, trx.WorkflowRunID, ops.TransactionEventsChannel, event)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return nil
}

func HandleTransactionDeliveryEvent(ctx context.Context, b ops.Backends, event external.Event) error {
	trx, err := ops.GetTransactionWorkflowRef(ctx, b, event.UserID, event.ResourceID)
	if err != nil {
		return err
	}

	err = b.Temporal().SignalWorkflow(ctx, trx.WorkflowID, trx.WorkflowRunID, ops.TransactionDeliveryEventsChannel, event)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return nil
}
