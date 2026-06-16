package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"gitlab.com/fynbos/mock/mockplaid/internal/config"
	"gitlab.com/fynbos/mock/mockplaid/internal/handler"
	"gitlab.com/fynbos/mock/mockplaid/internal/storage"
)

func TestHealthEndpoint(t *testing.T) {
	router := chi.NewRouter()
	h := handler.NewHandler(storage.NewMemoryStorage(), &config.Config{})
	setupRoutes(router, h)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %s", resp["status"])
	}
}
