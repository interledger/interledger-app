package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/notify"
	"io"
	"net/http"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"

	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func NewHandleInbound(b Backends) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}
		fmt.Println("REQ body", string(body))
		// {
		//  "card-number":"tok_sandbox_kBQ8yZhZCui16Tb6EGrh1p",
		//	"exp-date":"tok_sandbox_u3SGiAqrsDR2NCff1QqhcH",
		//	"card-security-code":"tok_sandbox_wQ5wmJDCKC1ZdVgcCjJizU",
		//	"walletId":"7b0327f0-9114-48b4-bae0-d30bddf56362",
		//	"last4":"1111",
		//	"cardType":"visa"
		// }

		//w.WriteHeader(200)
		//body, err := io.ReadAll(r.Body)
		//if err != nil {
		//	http.Error(w, "failed to read body", http.StatusInternalServerError)
		//	return
		//}

		//err = ValidateWebhook(body, webhookSecret, r.Header.Get(SignatureHeader))
		//if errors.Is(err, machnet.ErrInvalidSignature) {
		//	http.Error(w, "bad request", http.StatusBadRequest)
		//	return
		//}
		//if err != nil {
		//	http.Error(w, "internal server error", http.StatusInternalServerError)
		//	return
		//}

		//var event external.Event
		//err = json.Unmarshal(body, &event)
		//if err != nil {
		//	http.Error(w, "failed to parse payload", http.StatusBadRequest)
		//	return
		//}
		//
		//err = SaveWebhook(r.Context(), b, event)
		//if err != nil {
		//	http.Error(w, "internal server error", http.StatusInternalServerError)
		//	return
		//}
		//
		//// TODO: validate payload
		//err = HandleEvent(r.Context(), b, event)
		//if err != nil {
		//	http.Error(w, "internal server error", http.StatusInternalServerError)
		//	return
		//}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleEvent(ctx context.Context, b ops.Backends, event external.Event) error {
	var err error
	switch event.EventName {
	case external.UserCardAdded:
		err = HandleUserCardAddedEvent(ctx, b, event)
	case external.UserBankAdded:
		err = HandleBankAccountAddedEvent(ctx, b, event)
	case external.UserKYCInProgressEvent, external.UserKYCSuspendedEvent, external.UserKYCRetryEvent, external.UserKYCVerifiedEvent, external.UserKYCReviewPendingEvent:
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
	mu, err := ops.GetUserByID(ctx, b, event.UserID)
	if err != nil {
		return err
	}

	// KYC event occurred, check latest status against API
	u, err := b.External().GetUserByID(ctx, event.UserID)
	if err != nil {
		return err
	}

	var newStatus machnet.KYCStatus
	switch u.Status {
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

	fUserId := getWalletUserID(ctx, b, mu.WalletID)
	b.Analytics().TrackWalletMachnetKYCStatus(analytics.MachnetKYCArgs{
		UserID:   fUserId,
		WalletID: mu.WalletID,
		Status:   newStatus,
	})

	refs, err := ops.ListActiveUserWorkflowRefs(ctx, b, event.UserID)
	if err != nil {
		return err
	}

	for _, ref := range refs {
		err = b.Temporal().SignalWorkflow(ctx, ref.WorkflowID, ref.WorkflowRunID, ops.UserEventsChannel, u)
		if err != nil {
			return fmt.Errorf("%w %s", machnet.ErrInternal, err)
		}
	}

	err = b.Notify().NotifyWallet(ctx, mu.WalletID, notify.NotificationTypeKyc)
	if err != nil {
		log.Error("error notifying wallet", zap.Error(err))
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

	fUserId := getWalletUserID(ctx, b, user.WalletID)
	b.Analytics().TrackWalletMachnetCardAdded(analytics.MachnetCardAddedArgs{
		UserID:   fUserId,
		WalletID: user.WalletID,
		Scheme:   card.InstitutionName,
	})

	return nil
}

func HandleBankAccountAddedEvent(ctx context.Context, b ops.Backends, event external.Event) error {
	user, err := ops.GetUserByID(ctx, b, event.UserID)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	// TODO: find out if these details are in the event payload
	bankAcc, err := b.External().GetUserFundingsource(ctx, user.ID, event.ResourceID)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	_, err = b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:   user.WalletID,
		Name:       bankAcc.FundingsourceName,
		Mask:       bankAcc.AccountNumber,
		Provider:   machnet.ProviderName,
		ProviderID: bankAcc.ID,
		Type:       machnet.TypeBankAccount,
	})
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	fUserId := getWalletUserID(ctx, b, user.WalletID)
	b.Analytics().TrackWalletMachnetBankAdded(analytics.MachnetBankAddedArgs{
		UserID:      fUserId,
		WalletID:    user.WalletID,
		Institution: bankAcc.InstitutionName,
	})

	return nil
}

func HandleTransactionEvent(ctx context.Context, b ops.Backends, event external.Event) error {
	trx, err := ops.GetTransactionWorkflowRef(ctx, b, event.UserID, event.ResourceID)
	if err != nil {
		return err
	}

	exTrx, err := b.External().GetUserTransaction(ctx, event.UserID, event.ResourceID)
	if err != nil {
		return err
	}

	err = b.Temporal().SignalWorkflow(ctx, trx.WorkflowID, trx.WorkflowRunID, ops.TransactionEventsChannel, exTrx)
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

func SaveWebhook(ctx context.Context, b ops.Backends, event external.Event) error {
	ib := db.NewInsert("machnet_webhook").
		Value("user_id", event.UserID).
		Value("event_name", event.EventName).
		Value("resource_id", event.ResourceID).
		Value("subscription_id", event.SubscriptionID)
	if event.ID != "" {
		ib.Value("id", event.ID)
	}
	if len(event.Payload) > 0 {
		ib.Value("payload", event.Payload)
	}

	q, args, err := ib.GetStatement()
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, q, args...)
	if err != nil && !db.IsErrorCode(err, db.UniqueViolationError) {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return nil
}

func getWalletUserID(ctx context.Context, b Backends, walletID string) string {
	if b.Users() == nil {
		return ""
	}

	users, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return ""
	}

	// if there are more than 1 user or no users don't return anything
	if len(users) != 1 {
		return ""
	}

	firstUser := users[0]
	return firstUser.ID
}
