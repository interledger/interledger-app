package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestAPIKeyMiddlewareEnforced(t *testing.T) {
	t.Parallel()

	const apiKey = "local-test-api-key"

	tests := []struct {
		name           string
		headerValue    string
		wantStatusCode int
	}{
		{
			name:           "valid API key is accepted",
			headerValue:    apiKey,
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "missing API key is rejected",
			headerValue:    "",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "wrong API key is rejected",
			headerValue:    "wrong-key",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := chi.NewRouter()
			r.Use(APIKeyMiddleware(apiKey, true))
			r.Post("/v0.2.4/multicurrency-wallets/create", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
			})

			req := httptest.NewRequest(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", nil)
			if tt.headerValue != "" {
				req.Header.Set("X-API-KEY", tt.headerValue)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("status mismatch: got %d want %d", rr.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusUnauthorized {
				var body map[string]any
				if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode unauthorized response: %v", err)
				}
				if body["status"] != "error" {
					t.Fatalf("status field mismatch: got %#v want %q", body["status"], "error")
				}
			}
		})
	}
}

func TestAPIKeyMiddlewareCanBeDisabled(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	r.Use(APIKeyMiddleware("local-test-api-key", false))
	r.Post("/v0.2.4/multicurrency-wallets/create", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status mismatch: got %d want %d", rr.Code, http.StatusCreated)
	}
}

func TestHealthRouteUnauthenticatedWhenAuthEnforced(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Group(func(r chi.Router) {
		r.Use(APIKeyMiddleware("local-test-api-key", true))
		r.Post("/v0.2.4/payment/initiate", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Post("/v0.2.4/payouts/interac", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRR := httptest.NewRecorder()
	r.ServeHTTP(healthRR, healthReq)
	if healthRR.Code != http.StatusOK {
		t.Fatalf("health status mismatch: got %d want %d", healthRR.Code, http.StatusOK)
	}

	paymentReq := httptest.NewRequest(http.MethodPost, "/v0.2.4/payment/initiate", nil)
	paymentRR := httptest.NewRecorder()
	r.ServeHTTP(paymentRR, paymentReq)
	if paymentRR.Code != http.StatusUnauthorized {
		t.Fatalf("payment status mismatch: got %d want %d", paymentRR.Code, http.StatusUnauthorized)
	}

	payoutReq := httptest.NewRequest(http.MethodPost, "/v0.2.4/payouts/interac", nil)
	payoutRR := httptest.NewRecorder()
	r.ServeHTTP(payoutRR, payoutReq)
	if payoutRR.Code != http.StatusUnauthorized {
		t.Fatalf("payout status mismatch: got %d want %d", payoutRR.Code, http.StatusUnauthorized)
	}
}
