package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"gitlab.com/fynbos/mock/mockchimoney/internal/config"
	"gitlab.com/fynbos/mock/mockchimoney/internal/storage"
)

type apiResponse struct {
	Status string         `json:"status"`
	Error  string         `json:"error"`
	Data   map[string]any `json:"data"`
}

func setupWalletRouter(t *testing.T) http.Handler {
	t.Helper()

	h := NewHandlerWithStore(&config.Config{Port: "41800", LogLevel: "debug"}, storage.NewMemoryStore())
	r := chi.NewRouter()
	r.Post("/v0.2.4/multicurrency-wallets/create", h.CreateWallet)
	r.Get("/v0.2.4/multicurrency-wallets/get", h.GetWallet)
	r.Post("/v0.2.4/multicurrency-wallets/transfer", h.Transfer)
	return r
}

func decodeResponse(t *testing.T, rr *httptest.ResponseRecorder) apiResponse {
	t.Helper()

	var resp apiResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

func createWallet(t *testing.T, router http.Handler, body map[string]any) apiResponse {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("create wallet status mismatch: got %d want %d", rr.Code, http.StatusCreated)
	}

	resp := decodeResponse(t, rr)
	if resp.Status != "success" {
		t.Fatalf("create wallet status field mismatch: got %q want %q", resp.Status, "success")
	}

	return resp
}

func TestCreateWalletRequiredName(t *testing.T) {
	t.Parallel()

	router := setupWalletRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", bytes.NewBufferString(`{"email":"x@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got %d want %d", rr.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, rr)
	if resp.Status != "error" {
		t.Fatalf("status field mismatch: got %q want %q", resp.Status, "error")
	}
}

func TestCreateWalletPendingVerificationAndUniqueIDs(t *testing.T) {
	t.Parallel()

	router := setupWalletRouter(t)
	first := createWallet(t, router, map[string]any{"name": "Charlie"})
	second := createWallet(t, router, map[string]any{"name": "Charlie"})

	firstID, ok := first.Data["id"].(string)
	if !ok || firstID == "" {
		t.Fatalf("first wallet id missing or invalid: %#v", first.Data["id"])
	}
	secondID, ok := second.Data["id"].(string)
	if !ok || secondID == "" {
		t.Fatalf("second wallet id missing or invalid: %#v", second.Data["id"])
	}
	if firstID == secondID {
		t.Fatalf("wallet IDs should be unique: both were %q", firstID)
	}

	subAccount, ok := first.Data["subAccount"].(bool)
	if !ok || !subAccount {
		t.Fatalf("subAccount mismatch: got %#v", first.Data["subAccount"])
	}

	verification, ok := first.Data["verification"].(map[string]any)
	if !ok {
		t.Fatalf("verification object missing: %#v", first.Data["verification"])
	}
	status, ok := verification["status"].(string)
	if !ok || status != "pending" {
		t.Fatalf("verification.status mismatch: got %#v want %q", verification["status"], "pending")
	}
}

func TestGetWalletByIDAndErrors(t *testing.T) {
	t.Parallel()

	router := setupWalletRouter(t)
	created := createWallet(t, router, map[string]any{"name": "Eve"})
	id, _ := created.Data["id"].(string)

	req := httptest.NewRequest(http.MethodGet, "/v0.2.4/multicurrency-wallets/get?id="+id, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("get wallet status mismatch: got %d want %d", rr.Code, http.StatusOK)
	}
	gotResp := decodeResponse(t, rr)
	if gotResp.Status != "success" {
		t.Fatalf("get wallet status field mismatch: got %q want %q", gotResp.Status, "success")
	}
	if gotResp.Data["name"] != "Eve" {
		t.Fatalf("get wallet name mismatch: got %#v want %q", gotResp.Data["name"], "Eve")
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/v0.2.4/multicurrency-wallets/get?id=does-not-exist", nil)
	missingRR := httptest.NewRecorder()
	router.ServeHTTP(missingRR, missingReq)
	if missingRR.Code != http.StatusNotFound {
		t.Fatalf("missing wallet status mismatch: got %d want %d", missingRR.Code, http.StatusNotFound)
	}

	noIDReq := httptest.NewRequest(http.MethodGet, "/v0.2.4/multicurrency-wallets/get", nil)
	noIDRR := httptest.NewRecorder()
	router.ServeHTTP(noIDRR, noIDReq)
	if noIDRR.Code != http.StatusBadRequest {
		t.Fatalf("missing id status mismatch: got %d want %d", noIDRR.Code, http.StatusBadRequest)
	}
}

func TestTransferBetweenExistingSubAccounts(t *testing.T) {
	t.Parallel()

	router := setupWalletRouter(t)
	sender := createWallet(t, router, map[string]any{"name": "Sender"})
	receiver := createWallet(t, router, map[string]any{"name": "Receiver"})

	payload := map[string]any{
		"subAccount":          sender.Data["id"],
		"receiver":            receiver.Data["id"],
		"amountToSend":        "50.00",
		"originCurrency":      "CAD",
		"destinationCurrency": "CAD",
		"turnOffNotification": true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rr.Code, http.StatusOK)
	}

	resp := decodeResponse(t, rr)
	if resp.Status != "success" {
		t.Fatalf("status field mismatch: got %q want %q", resp.Status, "success")
	}
}

func TestTransferRequiresFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		payload       map[string]any
		wantErrorText string
	}{
		{
			name: "requires amountToSend",
			payload: map[string]any{
				"originCurrency":      "CAD",
				"destinationCurrency": "CAD",
			},
			wantErrorText: "amountToSend",
		},
		{
			name: "requires originCurrency",
			payload: map[string]any{
				"amountToSend":        "10.00",
				"destinationCurrency": "CAD",
			},
			wantErrorText: "originCurrency",
		},
		{
			name: "requires destinationCurrency",
			payload: map[string]any{
				"amountToSend":   "10.00",
				"originCurrency": "CAD",
			},
			wantErrorText: "destinationCurrency",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := setupWalletRouter(t)
			body, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("json.Marshal() failed: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("status mismatch: got %d want %d", rr.Code, http.StatusBadRequest)
			}

			resp := decodeResponse(t, rr)
			if resp.Status != "error" {
				t.Fatalf("status field mismatch: got %q want %q", resp.Status, "error")
			}
			if resp.Error == "" || !bytes.Contains([]byte(resp.Error), []byte(tt.wantErrorText)) {
				t.Fatalf("error mismatch: got %q should mention %q", resp.Error, tt.wantErrorText)
			}
		})
	}
}

func TestTransferFromNonExistentSenderReturns400(t *testing.T) {
	t.Parallel()

	router := setupWalletRouter(t)
	receiver := createWallet(t, router, map[string]any{"name": "Receiver"})

	payload := map[string]any{
		"subAccount":          "non-existent-wallet-id",
		"receiver":            receiver.Data["id"],
		"amountToSend":        "10.00",
		"originCurrency":      "CAD",
		"destinationCurrency": "CAD",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got %d want %d", rr.Code, http.StatusBadRequest)
	}

	resp := decodeResponse(t, rr)
	if resp.Status != "error" {
		t.Fatalf("status field mismatch: got %q want %q", resp.Status, "error")
	}
}

func TestTransferAcceptsAndIgnoresSendViaInterledger(t *testing.T) {
	t.Parallel()

	router := setupWalletRouter(t)
	sender := createWallet(t, router, map[string]any{"name": "Sender"})
	receiver := createWallet(t, router, map[string]any{"name": "Receiver"})

	payload := map[string]any{
		"subAccount":          sender.Data["id"],
		"receiver":            receiver.Data["id"],
		"amountToSend":        "5.00",
		"originCurrency":      "CAD",
		"destinationCurrency": "CAD",
		"sendViaInterledger":  true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rr.Code, http.StatusOK)
	}
}
