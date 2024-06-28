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
	"time"

	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

const (
	EventPaymentCompleted    = "chimoney.payment.completed"
	EventPaymentFailed       = "chimoney.payment.failed"
	EventPaymentInitiated    = "chimoney.payment.initiated"
	EventPayoutBankCompleted = "chimoney.bank.completed"
	EventPayoutBankDelivered = "chimoney.bank.delivered"
	EventPayoutBankInitiated = "chimoney.bank.initiated"
)

type (
	Webhook struct {
		EventType string `json:"eventType"`
	}

	ChargeCardEvent struct {
		EventType string `json:"eventType"`
		Amount    string `json:"amount"`
		IssueID   string `json:"issueID"`
	}

	PaymentEvent struct {
		EventType string `json:"eventType"`
		IssueID   string `json:"issueID"`
	}

	BankMeta struct {
		ChiRef     string  `json:"chiRef"`
		Country    string  `json:"country"`
		Currency   string  `json:"currency"`
		Type       string  `json:"type"`
		ValueInUSD float64 `json:"valueInUSD"`
	}

	BankData struct {
		AccountNumber    string    `json:"account_number"`
		Amount           int       `json:"amount"`
		BankCode         string    `json:"bank_code"`
		BankName         string    `json:"bank_name"`
		CompleteMessage  string    `json:"complete_message"`
		CreatedAt        time.Time `json:"created_at"`
		Currency         string    `json:"currency"`
		DebitCurrency    string    `json:"debit_currency"`
		EventType        string    `json:"eventType"`
		Fee              int       `json:"fee"`
		FullName         string    `json:"full_name"`
		ID               int       `json:"id"`
		IsApproved       int       `json:"is_approved"`
		Meta             BankMeta  `json:"meta"`
		Narration        string    `json:"narration"`
		Reference        string    `json:"reference"`
		RequiresApproval int       `json:"requires_approval"`
		Status           string    `json:"status"`
	}

	BankEvent struct {
		Data      BankData `json:"data"`
		EventType string   `json:"eventType"`
		Status    string   `json:"status"`
	}
)

func ParseWebhookSecret(input string) []byte {
	secretParts := strings.Split(input, "_")
	var secret []byte
	var err error
	if len(secretParts) < 1 {
		log.Error("chimoney webhook: CHIMONEY_WEBHOOK_SECRET has incorrect format.")
	} else {
		secret, err = base64.StdEncoding.DecodeString(secretParts[1])
	}
	if err != nil {
		log.Error("chimoney webhook: error parsing CHIMONEY_WEBHOOK_SECRET", zap.Error(err))
	}

	return secret
}

func NewWebhook(b Backends) http.HandlerFunc {
	if os.Getenv("CHIMONEY_WEBHOOK_SECRET") == "" {
		log.Error("CHIMONEY_WEBHOOK_SECRET is empty")
	}
	secret := ParseWebhookSecret(os.Getenv("CHIMONEY_WEBHOOK_SECRET"))

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
		if errors.Is(err, gatehub.ErrInvalidWebhook) {
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
		default:
			log.Warn("gatehub webhook. Unhandled webhook type", zap.String("event_type", wh.EventType), zap.String("payload", string(body)))
		}

		w.WriteHeader(http.StatusOK)
	}
}

var signatureRegex = regexp.MustCompile("^(.+?),")

func Verify(ctx context.Context, r *http.Request, key []byte) ([]byte, error) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("chimoney webhook: Failed to get request body.", zap.Error(err))
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	log.Info("chimoney webhook: ", zap.String("body", string(payload)))

	signedContent := fmt.Sprintf("%s.%s.%s", r.Header.Get("svix-id"), r.Header.Get("svix-timestamp"), string(payload))
	hmac := hmac.New(sha256.New, key)
	_, err = hmac.Write([]byte(signedContent))
	if err != nil {
		log.Error("chimoney webhook: Failed to compute webhook signature hash.", zap.Error(err))
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	signatures := strings.Split(r.Header.Get("svix-signature"), " ")
	if len(signatures) < 1 {
		return nil, fmt.Errorf("%w No signatures found in headers", gatehub.ErrInvalidWebhook)
	}

	expectedSignature := signatureRegex.ReplaceAllString(signatures[0], "")
	computedSignature := base64.StdEncoding.EncodeToString(hmac.Sum(nil))
	if computedSignature != expectedSignature {
		log.Warn("chimoney webhook: invalid webhook signature", zap.String("computed", computedSignature), zap.String("expectedSignature", expectedSignature))
		return nil, gatehub.ErrInvalidWebhook
	}

	return payload, nil
}
