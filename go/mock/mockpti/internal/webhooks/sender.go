package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
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
	signingKey jwk.Key
	encryptKey jwk.Key
}

// NewSender creates a webhook sender targeting webhookURL.
func NewSender(webhookURL string) *Sender {
	return &Sender{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 15 * time.Second},
	}
}

// ConfigureSecurity enables PTI-like webhook signing (JWS) and encryption (JWE).
// If keys are not provided, sender continues with plain JSON payloads.
func (s *Sender) ConfigureSecurity(signingJWK, encryptionJWK string) error {
	if signingJWK == "" || encryptionJWK == "" {
		return nil
	}

	signingKey, err := jwk.ParseKey([]byte(signingJWK))
	if err != nil {
		return fmt.Errorf("invalid webhook signing jwk: %w", err)
	}

	encryptionKey, err := jwk.ParseKey([]byte(encryptionJWK))
	if err != nil {
		return fmt.Errorf("invalid webhook encryption jwk: %w", err)
	}

	publicKey, err := jwk.PublicKeyOf(encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to derive encryption public key: %w", err)
	}

	s.signingKey = signingKey
	s.encryptKey = publicKey
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

	if s.signingKey != nil && s.encryptKey != nil {
		signed, err := jws.Sign(body, jws.WithKey(jwa.RS512, s.signingKey))
		if err != nil {
			return fmt.Errorf("failed to sign webhook payload: %w", err)
		}

		encrypted, err := jwe.Encrypt(
			signed,
			jwe.WithJSON(),
			jwe.WithContentEncryption(jwa.A256CBC_HS512),
			jwe.WithKey(jwa.RSA_OAEP_256, s.encryptKey),
		)
		if err != nil {
			return fmt.Errorf("failed to encrypt webhook payload: %w", err)
		}

		body = encrypted
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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
