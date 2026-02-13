package ops

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/log"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

var (
	verificationAccepted = "accepted"
	verificationRejected = "rejected"
)

type (
	Webhook struct {
		ID        string `json:"uuid"`
		EventType string `json:"event_type"`
		UserID    string `json:"user_uuid"`
	}

	UserVerificationWebhook struct {
		ID          string                      `json:"uuid"`
		EventType   string                      `json:"event_type"`
		Timestamp   string                      `json:"timestamp"`
		UserID      string                      `json:"user_uuid"`
		Environment string                      `json:"environment"`
		Data        UserVerificationWebhookData `json:"data"`
	}

	UserVerificationWebhookData struct {
		Gateway  string         `json:"gateway"`
		Verified VerifiedStatus `json:"verified"`
	}

	VerifiedStatus struct {
		Status int    `json:"status"`
		Short  string `json:"short"`
	}

	DepositWebhook struct {
		ID          string             `json:"uuid"`
		EventType   string             `json:"event_type"`
		Timestamp   string             `json:"timestamp"`
		UserID      string             `json:"user_uuid"`
		Environment string             `json:"environment"`
		Data        DepositWebhookData `json:"data"`
	}

	DepositWebhookData struct {
		Amount      string `json:"amount"`
		Currency    string `json:"currency"`
		Address     string `json:"address"`
		DepositType string `json:"deposit_type"` // hosted or external
		TrxID       string `json:"tx_uuid"`
	}

	CardCreatedWebhook struct {
		UUID        string                 `json:"uuid"`
		Timestamp   string                 `json:"timestamp"`
		EventType   string                 `json:"event_type"`
		UserUUID    string                 `json:"user_uuid"`
		Environment string                 `json:"environment"`
		Data        CardCreatedWebhookData `json:"data"`
	}

	CardCreatedWebhookData struct {
		CardID           string  `json:"cardId"`
		CardSourceID     string  `json:"cardSourceId"`
		NameOnCard       string  `json:"nameOnCard"`
		ProductCode      string  `json:"productCode"`
		MaskedPan        string  `json:"maskedPan"`
		AccountID        string  `json:"accountId"`
		AccountSourceID  string  `json:"accountSourceId"`
		LockLevel        *string `json:"lockLevel"` // pointer so it can be null
		CustomerID       string  `json:"customerId"`
		CustomerSourceID string  `json:"customerSourceId"`
	}

	Card3DSConfirmationWebhookData struct {
		Type    string                              `json:"type"`
		Payload external.PendingThreeDSConfirmation `json:"payload"`
	}

	Card3DSConfirmationWebhook struct {
		UUID        string                         `json:"uuid"`
		Timestamp   string                         `json:"timestamp"`
		EventType   string                         `json:"event_type"`
		UserUUID    string                         `json:"user_uuid"`
		Environment string                         `json:"environment"`
		Data        Card3DSConfirmationWebhookData `json:"data"`
	}

	CardTransactionEventWebhook struct {
		ID          string                          `json:"uuid"`
		EventType   string                          `json:"event_type"`
		Timestamp   string                          `json:"timestamp"`
		UserID      string                          `json:"user_uuid"`
		Environment string                          `json:"environment"`
		Data        CardTransactionEventWebhookData `json:"data"`
	}

	CardTransactionEventWebhookData struct {
		Title         string `json:"title"`
		Body          string `json:"body"`
		TransactionID string `json:"transactionId"`
		CardID        string `json:"cardId"`
	}
)

func NewWebhook(b Backends, cfg gatehub.Config) http.HandlerFunc {
	if cfg.WebhookSecret == "" {
		log.Error("WebhookSecret is empty in Gatehub configuration")
	}

	key, err := hex.DecodeString(cfg.WebhookSecret)
	if err != nil {
		log.Fatal("Failed to decode WebhookSecret", zap.Error(err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Gatehub webhook", zap.String("method", r.Method))
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		body, err := Verify(r.Context(), r, key)
		if errors.Is(err, gatehub.ErrInvalidWebhook) {
			log.Warn("Webhook failed validation", zap.Error(err))
			w.WriteHeader(http.StatusOK) // Gatehub sends empty POST to test the webhook when registering
			return
		}
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var wh Webhook
		err = json.Unmarshal(body, &wh)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if _, err := getWalletID(r.Context(), b, wh.UserID); err != nil {
			log.Info("Wallet not found for Gatehub user; attempting cards fallback",
				zap.String("external_user_uuid", wh.UserID),
				zap.Error(err),
			)

			if err := forwardWebhookToFallback(r.Context(), body, r.Header, cfg.FallbackWebhookURL); err != nil {
				log.Error("failed to forward webhook to cards fallback",
					zap.Error(err),
					zap.String("external_user_uuid", wh.UserID),
				)
				http.Error(w, "Failed to forward webhook", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		switch wh.EventType {
		case "id.verification.accepted", "id.verification.rejected":
			HandleUserVerificationWebhook(r.Context(), b, body, w)
		case "core.deposit.completed":
			HandleUserDeposit(r.Context(), b, body, w)
		case "id.document_notice.expired", "id.document_notice.warning", "id.verification.action_required":
			HandleActionRequiredWebhook(r.Context(), b, body, w)
		case "cards.card.created":
			HandleCardCreatedWebhook(r.Context(), b, body, w)
		case "cards.3ds.auth_3ds_confirmation":
			HandleCardThreeDSConfirmation(r.Context(), b, body, w)
		case "cards.transaction.event":
			HandleCardTransactionEvent(r.Context(), b, body, w)
		default:
			log.Warn("gatehub webhook. Unhandled webhook type", zap.String("event_type", wh.EventType), zap.String("payload", string(body)))
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleCardCreatedWebhook(ctx context.Context, b Backends, raw json.RawMessage, w http.ResponseWriter) {
	var wh CardCreatedWebhook

	err := json.Unmarshal(raw, &wh)
	if err != nil {
		log.Error("gatehub webhook: Failed to unmarshal card created webhook", zap.String("customer source id", wh.Data.CardSourceID), zap.Error(err))
		return
	}

	wo := client.StartWorkflowOptions{
		ID:                    "gatehub_card_created_webhook_" + wh.UUID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	var workflowStatus enums.WorkflowExecutionStatus

	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:

		log.Warn("failed to handle gatehub card created webhook", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		_, err = b.Temporal().ExecuteWorkflow(ctx, wo, ProcessCardCreationWorkflow, wh)
		if err != nil {
			log.Warn("failed to handle gatehub card created webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func HandleActionRequiredWebhook(ctx context.Context, b Backends, raw json.RawMessage, w http.ResponseWriter) {
	slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "wallet-info-bot", fmt.Sprintf("gatehub verification action required: %s", string(raw)))
	w.WriteHeader(http.StatusOK)
}

func HandleUserVerificationWebhook(ctx context.Context, b Backends, raw json.RawMessage, w http.ResponseWriter) {
	var wh UserVerificationWebhook
	err := json.Unmarshal(raw, &wh)
	if err != nil {
		log.Error("gatehub webhook: Failed to unmarshal user verification webhook", zap.String("external_user_uuid", wh.UserID), zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if !strings.Contains(strings.ToLower(wh.Data.Gateway), "paywiser") {
		log.Warn("received user verification webhook for another gateway", zap.String("webhook-id", wh.ID), zap.String("user_uuid", wh.UserID), zap.String("gateway", wh.Data.Gateway))
		w.WriteHeader(http.StatusOK)
		return
	}

	walletID, err := getWalletID(ctx, b, wh.UserID)
	if err != nil {
		log.Error("Failed to find wallet for gatehub user", zap.String("external_user_uuid", wh.UserID), zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if wh.Data.Verified.Short == verificationAccepted {
		err = BackfillAccountAndSetKYC(ctx, b, walletID, wh.ID)
		// err = b.KYC().SetKYCStatus(ctx, walletID, kyc.StatusLevel1)
	} else if wh.Data.Verified.Short == verificationRejected {
		err = b.KYC().SetKYCStatus(ctx, walletID, kyc.StatusDenied)
	} else {
		log.Error("Unknown gatehub user verification status", zap.String("external_user_uuid", wh.UserID), zap.String("short", wh.Data.Verified.Short), zap.Int("status", wh.Data.Verified.Status))
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HandleUserDeposit(ctx context.Context, b Backends, raw json.RawMessage, w http.ResponseWriter) {
	var wh DepositWebhook
	err := json.Unmarshal(raw, &wh)
	if err != nil {
		log.Error("gatehub webhook: Failed to unmarshal deposit webhook", zap.String("external_user_uuid", wh.UserID), zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// `hosted` deposit type is for wallet-to-wallet transfers. Here we signal the payments engine
	if wh.Data.DepositType == "hosted" {
		err = b.Payments().SignalGatehubTransferComplete(ctx, wh.Data.TrxID)
		if err != nil {
			log.Error("gatehub webhook: Failed to signal payments workflow about wallet transfer", zap.String("external_user_uuid", wh.UserID), zap.String("external_transaction_id", wh.Data.TrxID), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		return
	}

	// `external` deposit type is for a deposit done through the ramp widget. We start a workflow
	// to handle this.
	wo := client.StartWorkflowOptions{
		ID:                    "gatehub_deposit_webhook" + wh.ID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:

		log.Error("Failed to handle gatehub deposit webhook", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// execute workflow if it's not running
	if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		_, err = b.Temporal().ExecuteWorkflow(ctx, wo, CreateGatehubDeposit, wh)
		if err != nil {
			log.Error("Failed to handle gatehub deposit webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func HandleCardTransactionEvent(ctx context.Context, b Backends, raw json.RawMessage, w http.ResponseWriter) {
	var wh CardTransactionEventWebhook
	err := json.Unmarshal(raw, &wh)
	if err != nil {
		log.Error("gatehub webhook: Failed to unmarshal card transaction event webhook", zap.String("webhook", string(raw)), zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	wo := client.StartWorkflowOptions{
		ID:                    "gatehub_card_transaction_event_" + wh.ID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:

		log.Error("Failed to handle gatehub card transaction event webhook", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// execute workflow if it's not running
	if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		_, err = b.Temporal().ExecuteWorkflow(ctx, wo, CreateCardTransaction, wh)
		if err != nil {
			log.Error("Failed to handle gatehub card transaction event webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func HandleCardThreeDSConfirmation(ctx context.Context, b Backends, raw json.RawMessage, w http.ResponseWriter) {
	var wh Card3DSConfirmationWebhook
	err := json.Unmarshal(raw, &wh)
	if err != nil {
		log.Error("gatehub webhook: Failed to unmarshal card 3DS confirmation webhook", zap.String("wh", string(raw)), zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	walletID, err := getWalletID(ctx, b, wh.UserUUID)
	if err != nil {
		log.Error("Failed to find wallet for gatehub user", zap.String("external_user_uuid", wh.UserUUID), zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	b.Notify().NotifyPending3DSConfirmation(ctx, walletID, wh.Data.Payload)
	b.Email().SendPending3DSConfirmation(ctx, walletID, wh.Data.Payload.TransactionID)

	w.WriteHeader(http.StatusOK)
}

func Verify(ctx context.Context, r *http.Request, key []byte) ([]byte, error) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("gatehub webhook: Failed to get request body.", zap.Error(err))
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	log.Info("Gatehub webhook: ", zap.String("body", string(payload)))
	hmac := hmac.New(sha256.New, key)
	_, err = hmac.Write(payload)
	if err != nil {
		log.Error("gatehub webhook: Failed to compute webhook signature hash.", zap.Error(err))
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if hex.EncodeToString(hmac.Sum(nil)) != r.Header.Get("x-gh-webhook-signature") {
		return nil, gatehub.ErrInvalidWebhook
	}

	return payload, nil
}

func forwardWebhookToFallback(ctx context.Context, body []byte, headers http.Header, fallbackURL string) error {
	if fallbackURL == "" {
		log.Error("FallbackWebhookURL is not set in Gatehub configuration")
		return fmt.Errorf("FallbackWebhookURL is not set in configuration")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fallbackURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating forward request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-gh-webhook-signature", headers.Get("x-gh-webhook-signature"))
	req.Header.Set("x-gh-webhook-timestamp", headers.Get("x-gh-webhook-timestamp"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending forward request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook forward failed with status %d to %s", resp.StatusCode, fallbackURL)
	}

	log.Info("Webhook successfully forwarded", zap.String("url", fallbackURL), zap.Int("status", resp.StatusCode))
	return nil
}
