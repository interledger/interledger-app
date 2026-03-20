package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
)

func TestPayoutInteracAndStatus(t *testing.T) {
	recorder := &webhookRecorder{}
	webhookServer := httptest.NewServer(http.HandlerFunc(recorder.handler))
	defer webhookServer.Close()

	router, store, cleanup := setupFullRouter(t, webhookServer.URL)
	defer cleanup()
	_, _ = store.CreateSubAccount(context.Background(), models.SubAccount{ID: "chi-sub-002", Name: "User", KYCStatus: "pending"})

	body := `{"subAccount":"chi-sub-002","debitCurrency":"CAD","interacs":[{"name":"Alice","email":"alice@example.com","amount":95.00}]}`
	resp := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/payouts/interac", body)
	if resp.Code != http.StatusOK {
		t.Fatalf("payout init status mismatch: got %d", resp.Code)
	}
	var env apiResponse
	_ = json.NewDecoder(resp.Body).Decode(&env)
	dataObj := env.Data["data"].([]any)
	if len(dataObj) != 1 {
		t.Fatalf("payout array count mismatch: got %d", len(dataObj))
	}
	payout := dataObj[0].(map[string]any)
	issueID := payout["issueID"].(string)
	if !strings.HasPrefix(issueID, "chi-sub-002_") {
		t.Fatalf("issueID format mismatch: %q", issueID)
	}
	chiRef := payout["chiref"].(string)

	statusResp := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/payouts/status", `{"chiRef":"`+chiRef+`"}`)
	if statusResp.Code != http.StatusOK {
		t.Fatalf("status check code mismatch: got %d", statusResp.Code)
	}
	var statusEnv apiResponse
	_ = json.NewDecoder(statusResp.Body).Decode(&statusEnv)
	if statusEnv.Data["status"] != "pending" {
		t.Fatalf("initial payout status mismatch: %#v", statusEnv.Data["status"])
	}

	events := recorder.waitForCount(t, 1, 3*time.Second)
	if events[0].Body["eventType"] != "payout.interac.completed" {
		t.Fatalf("event type mismatch: %#v", events[0].Body["eventType"])
	}
	meta := events[0].Body["meta"].(map[string]any)
	if meta["issuer"] != "chi-sub-002" || meta["currency"] != "CAD" {
		t.Fatalf("meta mismatch: %#v", meta)
	}

	time.Sleep(50 * time.Millisecond)
	statusAfter := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/payouts/status", `{"chiRef":"`+chiRef+`"}`)
	var statusAfterEnv apiResponse
	_ = json.NewDecoder(statusAfter.Body).Decode(&statusAfterEnv)
	if statusAfterEnv.Data["status"] != "completed" {
		t.Fatalf("completed payout status mismatch: %#v", statusAfterEnv.Data["status"])
	}
}

func TestPayoutValidationAndMultipleEntries(t *testing.T) {
	router, store, cleanup := setupFullRouter(t, "http://localhost:1/webhooks")
	defer cleanup()
	_, _ = store.CreateSubAccount(context.Background(), models.SubAccount{ID: "chi-sub-002", Name: "User", KYCStatus: "pending"})

	badCases := []string{
		`{"subAccount":"chi-sub-002","interacs":[{"name":"A","email":"a@b.com","amount":10}]}`,
		`{"subAccount":"chi-sub-002","debitCurrency":"CAD"}`,
		`{"subAccount":"does-not-exist","debitCurrency":"CAD","interacs":[{"name":"A","email":"a@b.com","amount":10}]}`,
	}
	for _, body := range badCases {
		resp := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/payouts/interac", body)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d for body %s", resp.Code, body)
		}
	}

	multi := `{"subAccount":"chi-sub-002","debitCurrency":"CAD","interacs":[{"name":"Alice","email":"alice@example.com","amount":50},{"name":"Bob","email":"bob@example.com","amount":30}]}`
	resp := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/payouts/interac", multi)
	if resp.Code != http.StatusOK {
		t.Fatalf("multi payout status mismatch: got %d", resp.Code)
	}
	var env apiResponse
	_ = json.NewDecoder(resp.Body).Decode(&env)
	items := env.Data["data"].([]any)
	if len(items) != 2 {
		t.Fatalf("expected 2 payouts got %d", len(items))
	}

	unknown := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/payouts/status", `{"chiRef":"missing"}`)
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("unknown chiRef expected 404 got %d", unknown.Code)
	}
	missing := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/payouts/status", `{}`)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing chiRef expected 400 got %d", missing.Code)
	}
}
