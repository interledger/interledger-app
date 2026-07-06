package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/config"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/jobs"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/storage"
)

func TestUserAssessmentWebhookJobHandler_SendsWebhook(t *testing.T) {
	store := storage.NewMemoryStorage()
	user := &models.User{ID: "user-1", Type: "PERSON"}
	if err := store.SaveUser(context.Background(), user); err != nil {
		t.Fatalf("save user: %v", err)
	}
	assessment := &models.Assessment{
		RequestID:  "req-1",
		UserID:     "user-1",
		ClientID:   "client-1",
		Date:       time.Now().Format(time.RFC3339),
		Assessment: "ACCEPTED",
		Tier:       1,
	}
	if err := store.SaveAssessment(context.Background(), assessment); err != nil {
		t.Fatalf("save assessment: %v", err)
	}

	var payload map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	h := NewHandler(store, &config.Config{ClientID: "client-1", WebhookURL: srv.URL})
	handler := h.NewUserAssessmentWebhookJobHandler()

	job := &models.Job{JobType: jobs.JobTypeUserAssessmentWebhook, Data: map[string]interface{}{"user_id": "user-1", "request_id": "req-1"}}
	if err := handler(context.Background(), job); err != nil {
		t.Fatalf("job handler failed: %v", err)
	}

	if payload["resourceType"] != "USER_ASSESSMENT" {
		t.Fatalf("expected USER_ASSESSMENT payload, got %v", payload["resourceType"])
	}
}

func TestTransactionStatusWebhookJobHandler_SendsWebhook(t *testing.T) {
	store := storage.NewMemoryStorage()
	tx := &models.Transaction{
		RequestID:       "tx-1",
		Status:          "SETTLED",
		TransactionType: "DEPOSIT",
		Amount:          100,
		Currency:        "USD",
		Date:            time.Now().Format(time.RFC3339),
		ClientID:        "client-1",
	}
	if err := store.SaveTransaction(context.Background(), tx); err != nil {
		t.Fatalf("save tx: %v", err)
	}

	var payload map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	h := NewHandler(store, &config.Config{ClientID: "client-1", WebhookURL: srv.URL})
	handler := h.NewTransactionStatusWebhookJobHandler()

	job := &models.Job{JobType: jobs.JobTypeTransactionStatusWebhook, Data: map[string]interface{}{"request_id": "tx-1"}}
	if err := handler(context.Background(), job); err != nil {
		t.Fatalf("job handler failed: %v", err)
	}

	if payload["resourceType"] != "TRANSACTION_STATUS" {
		t.Fatalf("expected TRANSACTION_STATUS payload, got %v", payload["resourceType"])
	}
}

func TestUserAssessmentWebhookJobHandler_MissingUserID(t *testing.T) {
	h := NewHandler(storage.NewMemoryStorage(), &config.Config{ClientID: "client-1"})
	handler := h.NewUserAssessmentWebhookJobHandler()

	err := handler(context.Background(), &models.Job{JobType: jobs.JobTypeUserAssessmentWebhook, Data: map[string]interface{}{}})
	if err == nil {
		t.Fatal("expected error for missing user_id")
	}
}

func TestUserAssessmentWebhookJobHandler_RequestIDMismatch(t *testing.T) {
	store := storage.NewMemoryStorage()
	_ = store.SaveUser(context.Background(), &models.User{ID: "user-1", Type: "PERSON"})
	_ = store.SaveAssessment(context.Background(), &models.Assessment{RequestID: "req-real", UserID: "user-1", Assessment: "ACCEPTED"})

	h := NewHandler(store, &config.Config{ClientID: "client-1"})
	handler := h.NewUserAssessmentWebhookJobHandler()

	err := handler(context.Background(), &models.Job{JobType: jobs.JobTypeUserAssessmentWebhook, Data: map[string]interface{}{"user_id": "user-1", "request_id": "req-other"}})
	if err == nil {
		t.Fatal("expected request id mismatch error")
	}
}

func TestTransactionStatusWebhookJobHandler_MissingRequestID(t *testing.T) {
	h := NewHandler(storage.NewMemoryStorage(), &config.Config{ClientID: "client-1"})
	handler := h.NewTransactionStatusWebhookJobHandler()

	err := handler(context.Background(), &models.Job{JobType: jobs.JobTypeTransactionStatusWebhook, Data: map[string]interface{}{}})
	if err == nil {
		t.Fatal("expected error for missing request_id")
	}
}
