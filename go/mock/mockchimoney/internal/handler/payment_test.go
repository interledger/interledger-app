package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/models"
)

func TestPaymentInitiateVerifyAndPayPageFlow(t *testing.T) {
	recorder := &webhookRecorder{}
	webhookServer := httptest.NewServer(http.HandlerFunc(recorder.handler))
	defer webhookServer.Close()

	router, store, cleanup := setupFullRouter(t, webhookServer.URL)
	defer cleanup()

	_, err := store.CreateSubAccount(context.Background(), models.SubAccount{ID: "chi-sub-001", Name: "User", KYCStatus: "pending"})
	if err != nil {
		t.Fatalf("CreateSubAccount() error: %v", err)
	}

	initBody := `{"amount":"100.00","currency":"CAD","subAccount":"chi-sub-001","payerEmail":"payer@example.com","redirect_url":"https://app.test/callbacks/chimoney"}`
	initResp := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/payment/initiate", initBody)
	if initResp.Code != http.StatusOK {
		t.Fatalf("init status mismatch: got %d", initResp.Code)
	}

	var initEnvelope apiResponse
	if err := json.NewDecoder(initResp.Body).Decode(&initEnvelope); err != nil {
		t.Fatalf("decode init response failed: %v", err)
	}
	issueID, _ := initEnvelope.Data["issueID"].(string)
	if !strings.HasPrefix(issueID, "chi-sub-001_") {
		t.Fatalf("issueID format mismatch: %q", issueID)
	}

	verifyResp := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/payment/verify", `{"id":"`+issueID+`"}`)
	if verifyResp.Code != http.StatusOK {
		t.Fatalf("verify status mismatch: got %d", verifyResp.Code)
	}
	var verifyEnvelope apiResponse
	if err := json.NewDecoder(verifyResp.Body).Decode(&verifyEnvelope); err != nil {
		t.Fatalf("decode verify response failed: %v", err)
	}
	if verifyEnvelope.Data["status"] != "pending" {
		t.Fatalf("payment status mismatch before pay: got %#v", verifyEnvelope.Data["status"])
	}

	payPageResp := doJSONRequest(t, router, http.MethodGet, "/pay/"+issueID, "")
	if payPageResp.Code != http.StatusOK {
		t.Fatalf("pay page status mismatch: got %d", payPageResp.Code)
	}
	if !strings.Contains(payPageResp.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("pay page content type mismatch: %q", payPageResp.Header().Get("Content-Type"))
	}

	confirmReq := httptest.NewRequest(http.MethodPost, "/pay/"+issueID+"/confirm", nil)
	confirmResp := httptest.NewRecorder()
	router.ServeHTTP(confirmResp, confirmReq)
	if confirmResp.Code != http.StatusFound {
		t.Fatalf("confirm status mismatch: got %d", confirmResp.Code)
	}
	loc := confirmResp.Header().Get("Location")
	if !strings.HasPrefix(loc, "https://app.test/callbacks/chimoney") {
		t.Fatalf("redirect mismatch: %q", loc)
	}
	if !strings.Contains(loc, "issueID=") || !strings.Contains(loc, "status=success") {
		t.Fatalf("redirect query mismatch: %q", loc)
	}

	verifyAfter := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/payment/verify", `{"id":"`+issueID+`"}`)
	var verifyAfterEnvelope apiResponse
	_ = json.NewDecoder(verifyAfter.Body).Decode(&verifyAfterEnvelope)
	if verifyAfterEnvelope.Data["status"] != "redeemed" {
		t.Fatalf("payment status mismatch after pay: got %#v", verifyAfterEnvelope.Data["status"])
	}
	meta, ok := verifyAfterEnvelope.Data["meta"].(map[string]any)
	if !ok {
		t.Fatalf("meta missing in verify response")
	}
	procFee, ok := meta["processingFee"].(map[string]any)
	if !ok || procFee["provider"] != "interac" {
		t.Fatalf("processing fee provider mismatch: %#v", meta)
	}

	events := recorder.waitForCount(t, 2, 3*time.Second)
	if events[0].Body["eventType"] != "charge.interac.completed" || events[1].Body["eventType"] != "chimoney.redeem.completed" {
		t.Fatalf("deposit webhook order mismatch: %#v %#v", events[0].Body["eventType"], events[1].Body["eventType"])
	}
	for _, ev := range events[:2] {
		if !strings.HasPrefix(ev.Headers.Get("svix-id"), "msg_") {
			t.Fatalf("svix-id mismatch: %q", ev.Headers.Get("svix-id"))
		}
		if !strings.HasPrefix(ev.Headers.Get("svix-signature"), "v1,") {
			t.Fatalf("svix-signature mismatch: %q", ev.Headers.Get("svix-signature"))
		}
	}
}

func TestPaymentInitiateValidationErrors(t *testing.T) {
	router, store, cleanup := setupFullRouter(t, "http://localhost:1/webhooks")
	defer cleanup()
	_, _ = store.CreateSubAccount(context.Background(), models.SubAccount{ID: "chi-sub-001", Name: "User", KYCStatus: "pending"})

	tests := []struct {
		name string
		body string
	}{
		{name: "missing payerEmail", body: `{"amount":"100.00","currency":"CAD","subAccount":"chi-sub-001"}`},
		{name: "unsupported currency", body: `{"amount":"100.00","currency":"GBP","subAccount":"chi-sub-001","payerEmail":"payer@example.com"}`},
		{name: "missing amount", body: `{"currency":"CAD","subAccount":"chi-sub-001","payerEmail":"payer@example.com"}`},
		{name: "missing subaccount", body: `{"amount":"100.00","currency":"CAD","subAccount":"does-not-exist","payerEmail":"payer@example.com"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := doJSONRequest(t, router, http.MethodPost, "/v0.2.4/payment/initiate", tc.body)
			if resp.Code != http.StatusBadRequest {
				t.Fatalf("status mismatch: got %d", resp.Code)
			}
		})
	}
}
