package webhooks

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

func TestSender_SendUserAssessment_NoURL(t *testing.T) {
	s := NewSender("")
	err := s.SendUserAssessment(context.Background(), &models.Assessment{RequestID: "r1", UserID: "u1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSender_SendUserAssessment_Success(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := NewSender(srv.URL)
	assessment := &models.Assessment{
		RequestID:  "req-99",
		UserID:     "user-99",
		ClientID:   "client-1",
		Assessment: "ACCEPTED",
		Tier:       2,
		Date:       "2026-03-20T00:00:00Z",
	}

	if err := s.SendUserAssessment(context.Background(), assessment); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, token := range []string{"USER_ASSESSMENT", "req-99", "user-99", "ACCEPTED"} {
		if !strings.Contains(body, token) {
			t.Errorf("expected webhook body to contain %q, body=%s", token, body)
		}
	}
}

func TestSender_SendTransactionStatus_Success(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := NewSender(srv.URL)
	tx := &models.Transaction{
		RequestID:       "tx-123",
		Status:          "SETTLED",
		TransactionType: "DEPOSIT",
		Amount:          100.0,
		Currency:        "USD",
		UserID:          "user-7",
		ClientID:        "client-1",
		Date:            "2026-03-20T00:00:00Z",
	}

	if err := s.SendTransactionStatus(context.Background(), tx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, token := range []string{"TRANSACTION_STATUS", "tx-123", "SETTLED", "DEPOSIT"} {
		if !strings.Contains(body, token) {
			t.Errorf("expected webhook body to contain %q, body=%s", token, body)
		}
	}
}

func TestSender_ReceiverError_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := NewSender(srv.URL)
	err := s.SendUserAssessment(context.Background(), &models.Assessment{RequestID: "r1", UserID: "u1"})
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestSender_SendUserAssessment_WithJWKSignEncrypt_RoundTrip(t *testing.T) {
	ctx := context.Background()

	signingPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}
	encryptionPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate encryption key: %v", err)
	}

	signingJWKJSON := mustMarshalPrivateJWK(t, signingPriv)
	encryptionJWKJSON := mustMarshalPrivateJWK(t, encryptionPriv)

	var encryptedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		encryptedBody = b
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := NewSender(srv.URL)
	if err := s.ConfigureSecurity(signingJWKJSON, encryptionJWKJSON); err != nil {
		t.Fatalf("configure security: %v", err)
	}

	assessment := &models.Assessment{
		RequestID:  "req-e2e-1",
		UserID:     "user-e2e-1",
		ClientID:   "client-e2e-1",
		Assessment: "ACCEPTED",
		Tier:       2,
		Date:       "2026-03-20T00:00:00Z",
	}

	if err := s.SendUserAssessment(ctx, assessment); err != nil {
		t.Fatalf("send encrypted webhook: %v", err)
	}

	if len(encryptedBody) == 0 {
		t.Fatal("expected encrypted body to be delivered")
	}
	if strings.Contains(string(encryptedBody), "USER_ASSESSMENT") {
		t.Fatal("expected encrypted payload, found plaintext marker")
	}

	decrypted, err := jwe.Decrypt(encryptedBody, jwe.WithKey(jwa.RSA_OAEP_256, encryptionPriv))
	if err != nil {
		t.Fatalf("decrypt webhook payload: %v", err)
	}

	verified, err := jws.Verify(decrypted, jws.WithKey(jwa.RS512, &signingPriv.PublicKey))
	if err != nil {
		t.Fatalf("verify webhook signature: %v", err)
	}

	var got UserAssessmentPayload
	if err := json.Unmarshal(verified, &got); err != nil {
		t.Fatalf("unmarshal verified payload: %v", err)
	}

	if got.ResourceType != "USER_ASSESSMENT" {
		t.Fatalf("expected USER_ASSESSMENT resource type, got %q", got.ResourceType)
	}
	if got.RequestID != assessment.RequestID {
		t.Fatalf("expected request id %q, got %q", assessment.RequestID, got.RequestID)
	}
	if got.UserID != assessment.UserID {
		t.Fatalf("expected user id %q, got %q", assessment.UserID, got.UserID)
	}
	if got.Assessment != assessment.Assessment {
		t.Fatalf("expected assessment %q, got %q", assessment.Assessment, got.Assessment)
	}
}

func mustMarshalPrivateJWK(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()

	jwkKey, err := jwk.FromRaw(key)
	if err != nil {
		t.Fatalf("create jwk from private key: %v", err)
	}

	b, err := json.Marshal(jwkKey)
	if err != nil {
		t.Fatalf("marshal jwk: %v", err)
	}

	return string(b)
}
