package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"gitlab.com/fynbos/mock/mockpti/internal/config"
	"gitlab.com/fynbos/mock/mockpti/internal/handler"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
)

func TestHealthEndpoint(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := &config.Config{ClientID: "test"}
	h := handler.NewHandler(store, cfg)

	router := chi.NewRouter()
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})
	setupRoutes(router, h)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %s", resp["status"])
	}
}

func TestHealthEndpoint_NoAuth(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := &config.Config{ClientID: "test"}
	h := handler.NewHandler(store, cfg)

	router := chi.NewRouter()
	setupRoutes(router, h)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 without auth, got %d", rr.Code)
	}
}
