package ops

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
	"os"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/gatehub"
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
)

func NewWebhook(b Backends) http.HandlerFunc {
	if os.Getenv("GATEHUB_WEBHOOK_SECRET") == "" {
		log.Error("GATEHUB_WEBHOOK_SECRET is empty")
	}

	key, err := hex.DecodeString(os.Getenv("GATEHUB_WEBHOOK_SECRET"))
	if err != nil {
		log.Fatal("Failed to decode GATEHUB_WEBHOOK_SECRET", zap.Error(err))
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

		switch wh.EventType {
		case "id.verification.accepted", "id.verification.rejected":
			HandleUserVerificationWebhook(r.Context(), b, body, w)
		case "core.deposit.completed":
			HandleUserDeposit(r.Context(), b, body, w)
		case "id.document_notice.expired", "id.document_notice.warning", "id.verification.action_required":
			HandleActionRequiredWebhook(r.Context(), b, body, w)
		default:
			log.Warn("gatehub webhook. Unhandled webhook type", zap.String("event_type", wh.EventType), zap.String("payload", string(body)))
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleActionRequiredWebhook(ctx context.Context, b Backends, raw json.RawMessage, w http.ResponseWriter) {
	slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "Fynbot", fmt.Sprintf("gatehub verification action required: %s", string(raw)))

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

	walletID, err := getWalletID(ctx, b, wh.UserID)
	if err != nil {
		log.Error("Failed to find wallet for gatehub user", zap.String("external_user_uuid", wh.UserID), zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if wh.Data.Verified.Short == verificationAccepted {
		err = b.KYC().SetKYCStatus(ctx, walletID, kyc.StatusLevel1)
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
