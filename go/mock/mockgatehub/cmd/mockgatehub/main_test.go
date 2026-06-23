package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/consts"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/handler"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/storage"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/webhook"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() chi.Router {
	store := storage.NewMemoryStorage()
	_ = storage.SeedTestUsers(store)
	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	h := handler.NewHandler(store, webhookManager)
	r := chi.NewRouter()
	setupRoutes(r, h)
	return r
}

func TestUIRoutes(t *testing.T) {
	r := setupTestRouter()
	tests := []struct {
		method string
		path   string
		code   int
	}{
		{"GET", "/ui/", http.StatusOK},
		{"GET", "/ui/users/" + consts.TestUser1ID, http.StatusOK},
		{"GET", "/ui/users/nonexistent-id", http.StatusNotFound},
		{"GET", "/ui/actions/kyc", http.StatusOK},
		{"POST", "/ui/actions/kyc", http.StatusSeeOther},
		{"GET", "/ui/actions/card-transaction", http.StatusOK},
		{"POST", "/ui/actions/card-transaction", http.StatusBadRequest},
		{"GET", "/ui/actions/card-transaction/cards", http.StatusOK},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, tt.code, rr.Code, "%s %s", tt.method, tt.path)
	}
}

func TestOrderAdditionalCardRoute(t *testing.T) {
	store := storage.NewMemoryStorage()
	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	h := handler.NewHandler(store, webhookManager)

	r := chi.NewRouter()
	setupRoutes(r, h)

	req := httptest.NewRequest("POST", "/cards/v1/cards/test-account-id/card", bytes.NewBufferString(`{"nameOnCard":"Test User"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}
