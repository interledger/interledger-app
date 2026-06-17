package webhooks

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/models"
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

func TestSender_SendUserAssessment_WithEd25519Sign_RoundTrip(t *testing.T) {
	ctx := context.Background()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	privPEM := mustMarshalPrivateKeyPEM(t, priv)

	var capturedBody []byte
	var capturedSig string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		capturedSig = r.Header.Get("X-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := NewSender(srv.URL)
	if err := s.ConfigureSecurity(privPEM); err != nil {
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
		t.Fatalf("send webhook: %v", err)
	}

	if len(capturedBody) == 0 {
		t.Fatal("expected body to be delivered")
	}
	if !strings.Contains(string(capturedBody), "USER_ASSESSMENT") {
		t.Fatalf("expected plaintext payload, got: %s", string(capturedBody))
	}

	if !strings.HasPrefix(capturedSig, "v1=") {
		t.Fatalf("expected X-Signature header with v1= prefix, got: %q", capturedSig)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(capturedSig, "v1="))
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}

	if !ed25519.Verify(pub, capturedBody, sigBytes) {
		t.Fatal("signature verification failed")
	}
}

func mustMarshalPrivateKeyPEM(t *testing.T, key ed25519.PrivateKey) string {
	t.Helper()

	b, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b}))
}
