//go:build e2e
// +build e2e

package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cucumber/godog"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/config"
	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/handler"
	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/jobs"
	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/storage"
	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/webhook"
)

type webhookEvent struct {
	Body    map[string]any
	RawBody []byte
	Header  http.Header
}

type TestContext struct {
	authEnforced bool
	apiKey       string
	useAPIKey    bool
	overrideKey  *string

	interacFee float64
	usdRate    float64
	secret     string

	store  storage.Store
	closer io.Closer
	queue  *jobs.InMemoryQueue
	sender *webhook.Sender

	mockServer    *httptest.Server
	workerCancel  context.CancelFunc
	webhookServer *httptest.Server
	webhooks      []webhookEvent

	lastResponse *http.Response
	lastBody     []byte
	lastJSON     map[string]any
	lastErr      error

	walletID      string
	senderID      string
	receiverID    string
	issueID       string
	paymentLink   string
	chiRef        string
	withdrawIssue string
	withdrawAmt   float64
	redirectURL   string

	firstTotalFee float64
	firstUSD      float64
	captured      *webhookEvent
	manualSig     string
	sigValid      bool
	preserveStore bool
	lastKYCSubID  string

	webhookMu sync.Mutex
}

func newTestContext() *TestContext {
	return &TestContext{}
}

func (tc *TestContext) resetState() {
	tc.authEnforced = true
	tc.apiKey = "local-test-api-key"
	tc.useAPIKey = false
	tc.overrideKey = nil
	tc.interacFee = 1.5
	tc.usdRate = 0.735
	tc.secret = "local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA=="

	tc.lastResponse = nil
	tc.lastBody = nil
	tc.lastJSON = nil
	tc.lastErr = nil
	tc.walletID = ""
	tc.senderID = ""
	tc.receiverID = ""
	tc.issueID = ""
	tc.paymentLink = ""
	tc.chiRef = ""
	tc.withdrawIssue = ""
	tc.withdrawAmt = 0
	tc.redirectURL = ""
	tc.firstTotalFee = 0
	tc.firstUSD = 0
	tc.captured = nil
	tc.manualSig = ""
	tc.sigValid = false
	tc.webhooks = nil
	tc.preserveStore = false
	tc.lastKYCSubID = ""
}

func (tc *TestContext) closeServers() {
	if tc.workerCancel != nil {
		tc.workerCancel()
		tc.workerCancel = nil
	}
	if tc.queue != nil {
		tc.queue.Close()
		tc.queue = nil
	}
	if tc.closer != nil {
		_ = tc.closer.Close()
		tc.closer = nil
	}
	if tc.mockServer != nil {
		tc.mockServer.Close()
		tc.mockServer = nil
	}
	if tc.webhookServer != nil {
		tc.webhookServer.Close()
		tc.webhookServer = nil
	}
}

func (tc *TestContext) ensureWebhookServer() {
	if tc.webhookServer != nil {
		return
	}

	tc.webhookServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, _ := io.ReadAll(r.Body)
		parsed := map[string]any{}
		_ = json.Unmarshal(body, &parsed)
		tc.webhookMu.Lock()
		tc.webhooks = append(tc.webhooks, webhookEvent{Body: parsed, RawBody: body, Header: r.Header.Clone()})
		tc.webhookMu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
}

func (tc *TestContext) ensureMockServer() error {
	if tc.mockServer != nil {
		return nil
	}

	tc.ensureWebhookServer()
	if redisURL := os.Getenv("MOCKCHIMONEY_REDIS_URL"); redisURL != "" {
		redisDB := 5
		if dbStr := os.Getenv("MOCKCHIMONEY_REDIS_DB"); dbStr != "" {
			parsed, err := strconv.Atoi(dbStr)
			if err != nil {
				return fmt.Errorf("invalid MOCKCHIMONEY_REDIS_DB: %w", err)
			}
			redisDB = parsed
		}

		redisStore, err := storage.NewRedisStore(redisURL, redisDB)
		if err != nil {
			return err
		}
		if !tc.preserveStore {
			if err := redisStore.FlushAll(context.Background()); err != nil {
				_ = redisStore.Close()
				return err
			}
		}
		tc.store = redisStore
		tc.closer = redisStore
	} else {
		tc.store = storage.NewMemoryStore()
		tc.closer = nil
	}
	tc.queue = jobs.NewInMemoryQueue(128)
	tc.sender = webhook.NewSender(&http.Client{Timeout: 5 * time.Second})

	cfg := &config.Config{
		Port:                  "41800",
		LogLevel:              "debug",
		APIKey:                tc.apiKey,
		EnforceAuthentication: tc.authEnforced,
		WebhookURL:            tc.webhookServer.URL,
		WebhookSecret:         tc.secret,
		WebhookMinDelaySec:    0.01,
		InteracFeeFlat:        tc.interacFee,
		CADToUSDRate:          tc.usdRate,
		PublicBaseURL:         "https://mockchimoney.interledger.test",
	}

	h := handler.NewHandlerWithDeps(cfg, tc.store, tc.queue, tc.sender)
	r := chi.NewRouter()
	r.Get("/health", h.Health)
	r.Get("/pay/{issueID}", h.PayPage)
	r.Post("/pay/{issueID}/confirm", h.ConfirmPayPage)
	r.Get("/verify/kyc/{externalID}", h.KYCPage)
	r.Post("/verify/kyc/{externalID}/approve", h.KYCApprove)
	r.Post("/verify/kyc/{externalID}/decline", h.KYCDecline)
	r.Group(func(r chi.Router) {
		r.Use(handler.APIKeyMiddleware(cfg.APIKey, cfg.EnforceAuthentication))
		r.Post("/v0.2.4/multicurrency-wallets/create", h.CreateWallet)
		r.Get("/v0.2.4/multicurrency-wallets/get", h.GetWallet)
		r.Post("/v0.2.4/multicurrency-wallets/transfer", h.Transfer)
		r.Post("/v0.2.4/payment/initiate", h.InitiatePayment)
		r.Post("/v0.2.4/payment/verify", h.VerifyPayment)
		r.Post("/v0.2.4/payouts/interac", h.PayoutInterac)
		r.Post("/v0.2.4/payouts/status", h.PayoutStatus)
		r.Post("/v0.2.4/info/fee-estimate", h.FeeEstimate)
		r.Get("/v0.2.4/info/convert/local-amount-to-usd", h.ConvertLocalAmountToUSD)
	})

	ctx, cancel := context.WithCancel(context.Background())
	tc.workerCancel = cancel
	go jobs.StartWorker(ctx, tc.queue)
	tc.mockServer = httptest.NewServer(r)
	tc.preserveStore = false
	return nil
}

func (tc *TestContext) restartMockServer() error {
	tc.preserveStore = true
	if tc.workerCancel != nil {
		tc.workerCancel()
		tc.workerCancel = nil
	}
	if tc.queue != nil {
		tc.queue.Close()
	}
	if tc.closer != nil {
		_ = tc.closer.Close()
		tc.closer = nil
	}
	if tc.mockServer != nil {
		tc.mockServer.Close()
		tc.mockServer = nil
	}
	return tc.ensureMockServer()
}

func (tc *TestContext) request(method string, path string, payload any) error {
	if err := tc.ensureMockServer(); err != nil {
		return err
	}
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, tc.mockServer.URL+path, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if tc.overrideKey != nil {
		if *tc.overrideKey != "" {
			req.Header.Set("X-API-KEY", *tc.overrideKey)
		}
	} else if tc.useAPIKey {
		req.Header.Set("X-API-KEY", tc.apiKey)
	}

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		tc.lastErr = err
		return err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	tc.lastResponse = resp
	tc.lastBody = rb
	tc.lastJSON = map[string]any{}
	_ = json.Unmarshal(rb, &tc.lastJSON)
	return nil
}

func (tc *TestContext) tableToMap(table *godog.Table) map[string]any {
	m := map[string]any{}
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		k := strings.TrimSpace(row.Cells[0].Value)
		v := strings.TrimSpace(row.Cells[1].Value)
		m[k] = parseScalar(v)
	}
	return buildNested(m)
}

func parseScalar(v string) any {
	lower := strings.ToLower(v)
	if lower == "true" {
		return true
	}
	if lower == "false" {
		return false
	}
	return v
}

func buildNested(flat map[string]any) map[string]any {
	root := map[string]any{}
	for k, v := range flat {
		setNested(root, k, v)
	}
	return root
}

func setNested(root map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	cur := root
	for i, p := range parts {
		if strings.Contains(p, "[") {
			name := p[:strings.Index(p, "[")]
			idxStr := p[strings.Index(p, "[")+1 : len(p)-1]
			idx, _ := strconv.Atoi(idxStr)
			arr, _ := cur[name].([]any)
			for len(arr) <= idx {
				arr = append(arr, map[string]any{})
			}
			if i == len(parts)-1 {
				arr[idx] = value
				cur[name] = arr
				return
			}
			child, _ := arr[idx].(map[string]any)
			cur[name] = arr
			cur = child
			continue
		}
		if i == len(parts)-1 {
			cur[p] = value
			return
		}
		next, ok := cur[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[p] = next
		}
		cur = next
	}
}

func getPath(m map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var cur any = m
	for _, p := range parts {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = mm[p]
	}
	return cur
}

func asFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	default:
		return 0
	}
}

func (tc *TestContext) decodeWebhookKey(secret string) ([]byte, error) {
	idx := strings.LastIndex(secret, "_")
	if idx < 0 || idx == len(secret)-1 {
		return nil, fmt.Errorf("invalid secret")
	}
	return base64.StdEncoding.DecodeString(secret[idx+1:])
}

func (tc *TestContext) validateWebhookSig(ev webhookEvent, key []byte) bool {
	sid := ev.Header.Get("svix-id")
	ts := ev.Header.Get("svix-timestamp")
	sig := strings.TrimPrefix(ev.Header.Get("svix-signature"), "v1,")
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(sid + "." + ts + "." + string(ev.RawBody)))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}

func (tc *TestContext) createSubAccount(id string) error {
	acct := models.SubAccount{ID: id, ParentID: "mockchimoney-root", UID: uuid.NewString(), Name: id, SubAccount: true, KYCStatus: "pending", CreatedAt: time.Now().UTC()}
	_, err := tc.store.CreateSubAccount(context.Background(), acct)
	return err
}
