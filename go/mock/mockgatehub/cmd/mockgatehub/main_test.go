package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/fynbos/mock/mockgatehub/internal/handler"
	"gitlab.com/fynbos/mock/mockgatehub/internal/storage"
	"gitlab.com/fynbos/mock/mockgatehub/internal/webhook"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

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
