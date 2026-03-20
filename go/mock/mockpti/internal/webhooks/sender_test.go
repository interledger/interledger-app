package webhooks

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
