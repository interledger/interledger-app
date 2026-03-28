package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"gitlab.com/fynbos/mock/mockchimoney/internal/config"
	"gitlab.com/fynbos/mock/mockchimoney/internal/jobs"
	"gitlab.com/fynbos/mock/mockchimoney/internal/storage"
	"gitlab.com/fynbos/mock/mockchimoney/internal/webhook"
)

type webhookEvent struct {
	Headers http.Header
	Body    map[string]any
}

type webhookRecorder struct {
	mu     sync.Mutex
	events []webhookEvent
}

func (r *webhookRecorder) handler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(req.Body).Decode(&body)

	r.mu.Lock()
	r.events = append(r.events, webhookEvent{Headers: req.Header.Clone(), Body: body})
	r.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (r *webhookRecorder) waitForCount(t *testing.T, count int, timeout time.Duration) []webhookEvent {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		r.mu.Lock()
		if len(r.events) >= count {
			out := make([]webhookEvent, len(r.events))
			copy(out, r.events)
			r.mu.Unlock()
			return out
		}
		r.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	t.Fatalf("timed out waiting for %d webhooks; got %d", count, len(r.events))
	return nil
}

func setupFullRouter(t *testing.T, webhookURL string) (http.Handler, storage.Store, func()) {
	t.Helper()

	cfg := &config.Config{
		Port:                  "41800",
		LogLevel:              "debug",
		APIKey:                "local-test-api-key",
		EnforceAuthentication: false,
		WebhookURL:            webhookURL,
		WebhookSecret:         "local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA==",
		WebhookMinDelaySec:    0.01,
		InteracFeeFlat:        1.50,
		CADToUSDRate:          0.735,
		PublicBaseURL:         "https://mockchimoney.interledger.test",
	}

	store := storage.NewMemoryStore()
	queue := jobs.NewInMemoryQueue(32)
	sender := webhook.NewSender(&http.Client{Timeout: 2 * time.Second})
	h := NewHandlerWithDeps(cfg, store, queue, sender)

	ctx, cancel := context.WithCancel(context.Background())
	go jobs.StartWorker(ctx, queue)

	r := chi.NewRouter()
	r.Post("/v0.2.4/multicurrency-wallets/create", h.CreateWallet)
	r.Get("/v0.2.4/multicurrency-wallets/get", h.GetWallet)
	r.Post("/v0.2.4/multicurrency-wallets/transfer", h.Transfer)
	r.Post("/v0.2.4/payment/initiate", h.InitiatePayment)
	r.Post("/v0.2.4/payment/verify", h.VerifyPayment)
	r.Get("/pay/{issueID}", h.PayPage)
	r.Post("/pay/{issueID}/confirm", h.ConfirmPayPage)
	r.Post("/v0.2.4/payouts/interac", h.PayoutInterac)
	r.Post("/v0.2.4/payouts/status", h.PayoutStatus)
	r.Post("/v0.2.4/info/fee-estimate", h.FeeEstimate)
	r.Get("/v0.2.4/info/convert/local-amount-to-usd", h.ConvertLocalAmountToUSD)
	r.Get("/verify/kyc/{externalID}", h.KYCPage)
	r.Post("/verify/kyc/{externalID}/approve", h.KYCApprove)
	r.Post("/verify/kyc/{externalID}/decline", h.KYCDecline)

	cleanup := func() {
		cancel()
		queue.Close()
	}

	return r, store, cleanup
}

func doJSONRequest(t *testing.T, router http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}
