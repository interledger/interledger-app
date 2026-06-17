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
	"regexp"
	"strings"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/providers/chimoney"
	"github.com/interledger/interledger-app/go/backend/providers/chimoney/external"
	httplogger "github.com/interledger/interledger-app/go/backend/providers/http"
	"github.com/interledger/interledger-app/go/log"
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
		Amount      string            `json:"amount"`
		Meta        WithdrawEventMeta `json:"meta"`
		ChiWalletID string            `json:"-"` // Populated from Meta.Issuer
	}
	WithdrawEventMeta struct {
		Issuer      string                 `json:"issuer"`
		Amount      external.FlexibleFloat `json:"amount,omitempty"`
		Currency    string                 `json:"currency,omitempty"`
		PaymentType string                 `json:"paymentType,omitempty"`
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
	if len(secret) == 0 {
		log.Error("chimoney webhook: CHIMONEY_WEBHOOK_SECRET decoded to an empty secret, which is invalid")
		return nil
	}

	return secret
}

func NewWebhook(b Backends, webhookSecret, apiKey string) http.HandlerFunc {
	if webhookSecret == "" {
		log.Error("CHIMONEY_WEBHOOK_SECRET is empty")
	}
	secret := ParseWebhookSecret(webhookSecret)

	ec := external.New(
		apiKey,
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
		case "chimoney.redeem.completed":
			err = handleRedeemWebhook(r.Context(), b, body, "completed")
		case "chimoney.redeem.failed":
			err = handleRedeemWebhook(r.Context(), b, body, "failed")
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
			log.Warn("chimoney webhook. Unhandled webhook type", zap.String("event_type", wh.EventType))
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

	var wh Webhook
	if err := json.Unmarshal(payload, &wh); err != nil {
		log.Info("chimoney webhook: received webhook payload",
			zap.Int("payloadSize", len(payload)),
		)
	} else {
		log.Info("chimoney webhook: received webhook payload",
			zap.String("eventType", wh.EventType),
			zap.Int("payloadSize", len(payload)),
		)
	}

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

func handleRedeemWebhook(ctx context.Context, b Backends, raw json.RawMessage, status string) error {
	var wh PaymentEvent
	err := json.Unmarshal(raw, &wh)
	if err != nil {
		return err
	}
	if wh.IssueID == "" {
		log.Info("Webhook data not complete", zap.String("issueID", wh.IssueID), zap.String("status", wh.Status))
		return nil
	}
	var chiWalletID string
	chiWalletID, err = ExtractChiWalletIDFromIssueID(wh.IssueID)
	if err != nil {
		return err
	}

	return ExecuteFinishDeposit(ctx, b, wh.IssueID, status, chiWalletID)
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

	// Populate wh.ChiWalletID from Meta.Issuer
	wh.ChiWalletID = wh.Meta.Issuer
	if wh.ChiWalletID == "" {
		wh.ChiWalletID, err = ExtractChiWalletIDFromIssueID(wh.IssueID)
		if err != nil {
			return err
		}
	}

	var amount currency.Amount
	if wh.Meta.Currency != "" {
		amount = currency.FromFloat64(wh.Meta.Amount.Float64(), currency.ParseCurrency(wh.Meta.Currency))
	} else {
		amount = currency.FromFloat64(wh.Meta.Amount.Float64(), currency.CAD)
	}

	return ExecuteFinishWithdraw(ctx, b, ec, wh.IssueID, wh.Status, wh.ChiWalletID, amount)
}

func handleConfirmedOrCompletedCharge(ctx context.Context, b Backends, ec external.Client, raw json.RawMessage) error {
	var wh PaymentEvent
	err := json.Unmarshal(raw, &wh)
	if err != nil {
		return err
	}
	var chiWalletID string
	chiWalletID, err = ExtractChiWalletIDFromIssueID(wh.IssueID)
	if err != nil {
		return err
	}

	_, err = CreateDeposit(ctx, b, ec, wh.IssueID, chiWalletID)
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

	walletID, err := GetWalletID(ctx, b, wh.UserID)
	if err != nil {
		return fmt.Errorf("%w %s", chimoney.ErrInternal, err)
	}
	return b.KYC().SetKYCStatus(ctx, walletID, kycStatus)
}
