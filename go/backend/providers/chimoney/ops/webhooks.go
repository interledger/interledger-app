package ops

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/chimoney/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type (
	Webhook struct {
		EventType string `json:"eventType"`
	}

	PaymentEvent struct {
		EventType string `json:"eventType"`
		IssueID   string `json:"issueID"`
		Status    string `json:"status"`
	}
	WithdrawEvent struct {
		EventType   string            `json:"eventType"`
		IssueID     string            `json:"issueID"`
		Status      string            `json:"status"`
		Meta        WithdrawEventMeta `json:"meta"`
		ChiWalletID string            `json:"-"` // Populated from Meta.Issuer
	}
	WithdrawEventMeta struct {
		Issuer string `json:"issuer"`
	}
	KYCEvent struct {
		EventType string `json:"eventType"`
		UserID    string `json:"userID"`
	}
)

func ParseWebhookSecret(input string) []byte {
	secretParts := strings.Split(input, "_")
	var secret []byte
	var err error
	if len(secretParts) < 2 {
		log.Error("chimoney webhook: CHIMONEY_WEBHOOK_SECRET has incorrect format.")
		return nil
	}

	secret, err = base64.StdEncoding.DecodeString(secretParts[1])

	if err != nil {
		log.Error("chimoney webhook: error parsing CHIMONEY_WEBHOOK_SECRET", zap.Error(err))
		return nil
	}

	return secret
}

func NewWebhook(b Backends) http.HandlerFunc {
	if os.Getenv("CHIMONEY_WEBHOOK_SECRET") == "" {
		log.Error("CHIMONEY_WEBHOOK_SECRET is empty")
	}
	secret := ParseWebhookSecret(os.Getenv("CHIMONEY_WEBHOOK_SECRET"))

	ec := external.New(
		&http.Client{
			Transport: otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, b, external.Redact),
			),
		},
	)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Chimoney webhook", zap.String("method", r.Method))
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		body, err := Verify(r.Context(), r, secret)
		if errors.Is(err, chimoney.ErrInvalidWebhook) {
			log.Warn("Webhook failed validation", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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
		case "payout.interac.expired",
			"payout.interac.cancelled",
			"payout.interac.completed":
			err = handleWithdrawal(r.Context(), b, ec, body)
		case "chimoney.redeem.completed", "chimoney.redeem.failed":
			err = handleRedeemWebhook(r.Context(), b, body)
		case "charge.card.completed",
			"charge.chimoney-wallet.completed",
			"charge.interac.completed",
			"charge.crypto.xrpl.confirmed",
			"charge.crypto.celo.confirmed":
			err = handleConfirmedOrCompletedCharge(r.Context(), b, ec, body)
		case "user.kyc.declined",
			"user.kyc.completed":
			err = handleKYC(r.Context(), b, body)
		default:
			log.Warn("chimoney webhook. Unhandled webhook type", zap.String("event_type", wh.EventType), zap.String("payload", string(body)))
		}
		if err != nil {
			log.Error("chimoney webhook: failed to handle webhook", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

var signatureRegex = regexp.MustCompile("^(.+?),")

func Verify(ctx context.Context, r *http.Request, key []byte) ([]byte, error) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("chimoney webhook: Failed to get request body.", zap.Error(err))
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}
	log.Info("chimoney webhook: ", zap.String("body", string(payload)))

	signedContent := fmt.Sprintf("%s.%s.%s", r.Header.Get("svix-id"), r.Header.Get("svix-timestamp"), string(payload))
	hmac := hmac.New(sha256.New, key)
	_, err = hmac.Write([]byte(signedContent))
	if err != nil {
		log.Error("chimoney webhook: Failed to compute webhook signature hash.", zap.Error(err))
		return nil, fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}

	signatures := strings.Split(r.Header.Get("svix-signature"), " ")
	if len(signatures) < 1 {
		return nil, fmt.Errorf("%w No signatures found in headers", chimoney.ErrInvalidWebhook)
	}

	expectedSignature := signatureRegex.ReplaceAllString(signatures[0], "")
	computedSignature := base64.StdEncoding.EncodeToString(hmac.Sum(nil))
	if computedSignature != expectedSignature {
		log.Warn("chimoney webhook: invalid webhook signature", zap.String("computed", computedSignature), zap.String("expectedSignature", expectedSignature))
		return nil, chimoney.ErrInvalidWebhook
	}

	return payload, nil
}

func handleRedeemWebhook(ctx context.Context, b Backends, raw json.RawMessage) error {
	var wh PaymentEvent
	err := json.Unmarshal(raw, &wh)
	if err != nil {
		return err
	}
	if wh.IssueID == "" {
		log.Info("Webhook data not complete", zap.String("issueID", wh.IssueID), zap.String("status", wh.Status))
		return nil
	}
	var ChiWalletID string
	ChiWalletID, err = ExtractChiWalletIDFromIssueID(wh.IssueID)
	if err != nil {
		return err
	}

	return ExecuteFinishDeposit(ctx, b, wh.IssueID, wh.Status, ChiWalletID)
}

func handleWithdrawal(ctx context.Context, b Backends, ec external.Client, raw json.RawMessage) error {
	var wh WithdrawEvent
	err := json.Unmarshal(raw, &wh)
	if err != nil {
		return err
	}
	if wh.Status == "" || wh.IssueID == "" {
		log.Info("Webhook data not complete", zap.String("issueID", wh.IssueID), zap.String("status", wh.Status))
		return nil
	}

	// Populate ChiWalletID from Meta.Issuer
	wh.ChiWalletID = wh.Meta.Issuer
	if wh.ChiWalletID == "" {
		wh.ChiWalletID, err = ExtractChiWalletIDFromIssueID(wh.IssueID)
		if err != nil {
			return err
		}
	}

	return ExecuteFinishWithdraw(ctx, b, ec, wh.IssueID, wh.Status, wh.ChiWalletID)
}

func handleConfirmedOrCompletedCharge(ctx context.Context, b Backends, ec external.Client, raw json.RawMessage) error {
	var wh PaymentEvent
	err := json.Unmarshal(raw, &wh)
	if err != nil {
		return err
	}

	_, err = CreateDeposit(ctx, b, ec, wh.IssueID)
	return err
}

func handleKYC(ctx context.Context, b Backends, raw json.RawMessage) error {
	var wh KYCEvent
	err := json.Unmarshal(raw, &wh)
	if err != nil {
		return err
	}
	if wh.UserID == "" {
		log.Info("Webhook data not complete", zap.String("userID", wh.UserID))
		return nil
	}

	var kycStatus kyc.Status
	switch wh.EventType {
	case "user.kyc.completed":
		kycStatus = kyc.StatusLevel1
	case "user.kyc.declined":
		kycStatus = kyc.StatusDenied
	default:
		return fmt.Errorf("unknown KYC status: %s", wh.EventType)
	}

	return ExecuteCompleteKYCWorkflow(ctx, b, wh.UserID, kycStatus)
}
