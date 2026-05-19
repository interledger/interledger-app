package webhooks

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockpti/internal/logger"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

// UserAssessmentPayload is the plain webhook body for USER_ASSESSMENT events.
type UserAssessmentPayload struct {
	ResourceType string `json:"resourceType"`
	ClientID     string `json:"clientId"`
	RequestID    string `json:"requestId"`
	UserID       string `json:"userId"`
	Date         string `json:"date"`
	Assessment   string `json:"assessment"`
	Tier         int    `json:"tier"`
}

// TransactionStatusPayload is the plain webhook body for TRANSACTION_STATUS events.
type TransactionStatusPayload struct {
	ResourceType    string  `json:"resourceType"`
	ClientID        string  `json:"clientId"`
	RequestID       string  `json:"requestId"`
	UserID          string  `json:"userId,omitempty"`
	Status          string  `json:"status"`
	TransactionType string  `json:"transactionType"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Date            string  `json:"date"`
}

// Sender delivers plain JSON webhook payloads to a configured target URL.
type Sender struct {
	webhookURL string
	client     *http.Client
	signingKey ed25519.PrivateKey
}

// NewSender creates a webhook sender targeting webhookURL.
func NewSender(webhookURL string) *Sender {
	return &Sender{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 15 * time.Second},
	}
}

// ConfigureSecurity enables Ed25519 webhook signing from a PEM-encoded private key.
// Literal \n sequences in signingPEM are treated as newlines to support single-line env var values.
// If signingPEM is empty, the sender continues with plain unsigned JSON payloads.
func (s *Sender) ConfigureSecurity(signingPEM string) error {
	if signingPEM == "" {
		return nil
	}

	signingPEM = strings.ReplaceAll(signingPEM, `\n`, "\n")

	block, _ := pem.Decode([]byte(signingPEM))
	if block == nil {
		return fmt.Errorf("failed to decode signing key PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse signing key: %w", err)
	}

	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return fmt.Errorf("signing key is not Ed25519")
	}

	s.signingKey = edKey
	return nil
}

// SendUserAssessment delivers a USER_ASSESSMENT webhook built from the given Assessment.
func (s *Sender) SendUserAssessment(ctx context.Context, assessment *models.Assessment) error {
	if s.webhookURL == "" {
		logger.Infof("Webhook URL not configured — skipping USER_ASSESSMENT webhook for requestId=%s", assessment.RequestID)
		return nil
	}

	payload := UserAssessmentPayload{
		ResourceType: "USER_ASSESSMENT",
		ClientID:     assessment.ClientID,
		RequestID:    assessment.RequestID,
		UserID:       assessment.UserID,
		Date:         assessment.Date,
		Assessment:   assessment.Assessment,
		Tier:         assessment.Tier,
	}
	return s.post(ctx, payload)
}

// SendTransactionStatus delivers a TRANSACTION_STATUS webhook for the given Transaction.
func (s *Sender) SendTransactionStatus(ctx context.Context, tx *models.Transaction) error {
	if s.webhookURL == "" {
		logger.Infof("Webhook URL not configured — skipping TRANSACTION_STATUS webhook for requestId=%s", tx.RequestID)
		return nil
	}

	payload := TransactionStatusPayload{
		ResourceType:    "TRANSACTION_STATUS",
		ClientID:        tx.ClientID,
		RequestID:       tx.RequestID,
		UserID:          tx.UserID,
		Status:          tx.Status,
		TransactionType: tx.TransactionType,
		Amount:          tx.Amount,
		Currency:        tx.Currency,
		Date:            tx.Date,
	}
	return s.post(ctx, payload)
}

func (s *Sender) post(ctx context.Context, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if s.signingKey != nil {
		sig := ed25519.Sign(s.signingKey, body)
		req.Header.Set("X-Signature", "v1="+base64.StdEncoding.EncodeToString(sig))
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook delivery failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook receiver returned status %d", resp.StatusCode)
	}

	logger.Infof("Webhook delivered to %s (status %d)", s.webhookURL, resp.StatusCode)
	return nil
}
