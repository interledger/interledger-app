package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
)

func TestKYCPageAndApprovalDeclineFlows(t *testing.T) {
	recorder := &webhookRecorder{}
	webhookServer := httptest.NewServer(http.HandlerFunc(recorder.handler))
	defer webhookServer.Close()

	router, store, cleanup := setupFullRouter(t, webhookServer.URL)
	defer cleanup()

	for _, id := range []string{"kyc-sub-001", "kyc-sub-003", "kyc-sub-004", "kyc-sub-006", "kyc-sub-007", "kyc-sub-009"} {
		_, _ = store.CreateSubAccount(context.Background(), models.SubAccount{ID: id, Name: id, KYCStatus: "pending"})
	}

	page := doJSONRequest(t, router, http.MethodGet, "/verify/kyc/kyc-sub-001?redirect=https://app.test/callbacks/chimoney%3Fkyc", "")
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "Approve KYC") {
		t.Fatalf("kyc page response mismatch: status=%d body=%s", page.Code, page.Body.String())
	}

	missing := doJSONRequest(t, router, http.MethodGet, "/verify/kyc/does-not-exist?redirect=https://app.test/callbacks/chimoney", "")
	if missing.Code != http.StatusNotFound {
		t.Fatalf("unknown kyc page expected 404 got %d", missing.Code)
	}
	missingRedirect := doJSONRequest(t, router, http.MethodGet, "/verify/kyc/kyc-sub-001", "")
	if missingRedirect.Code != http.StatusBadRequest {
		t.Fatalf("missing redirect expected 400 got %d", missingRedirect.Code)
	}

	approveReq := httptest.NewRequest(http.MethodPost, "/verify/kyc/kyc-sub-003/approve?redirect=https://app.test/callbacks/chimoney?kyc", nil)
	approveResp := httptest.NewRecorder()
	router.ServeHTTP(approveResp, approveReq)
	if approveResp.Code != http.StatusFound || !strings.HasPrefix(approveResp.Header().Get("Location"), "https://app.test/callbacks/chimoney?kyc") {
		t.Fatalf("approve redirect mismatch: code=%d loc=%q", approveResp.Code, approveResp.Header().Get("Location"))
	}

	_ = doJSONRequest(t, router, http.MethodPost, "/verify/kyc/kyc-sub-004/approve?redirect=https://app.test/callbacks/chimoney?kyc", "")
	acct, _ := store.GetSubAccount(context.Background(), "kyc-sub-004")
	if acct.KYCStatus != "completed" {
		t.Fatalf("kyc status expected completed got %q", acct.KYCStatus)
	}

	_ = doJSONRequest(t, router, http.MethodPost, "/verify/kyc/kyc-sub-007/decline?redirect=https://app.test/callbacks/chimoney?kyc", "")
	acct2, _ := store.GetSubAccount(context.Background(), "kyc-sub-007")
	if acct2.KYCStatus != "declined" {
		t.Fatalf("kyc status expected declined got %q", acct2.KYCStatus)
	}

	declineReq := httptest.NewRequest(http.MethodPost, "/verify/kyc/kyc-sub-006/decline?redirect=https://app.test/callbacks/chimoney?kyc", nil)
	declineResp := httptest.NewRecorder()
	router.ServeHTTP(declineResp, declineReq)
	if declineResp.Code != http.StatusFound || !strings.Contains(declineResp.Header().Get("Location"), "status=failed") {
		t.Fatalf("decline redirect mismatch: %q", declineResp.Header().Get("Location"))
	}

	_ = doJSONRequest(t, router, http.MethodPost, "/verify/kyc/kyc-sub-009/approve?redirect=https://app.test/callbacks/chimoney?kyc", "")
	again := doJSONRequest(t, router, http.MethodPost, "/verify/kyc/kyc-sub-009/approve?redirect=https://app.test/callbacks/chimoney?kyc", "")
	if again.Code != http.StatusConflict {
		t.Fatalf("re-approve expected 409 got %d", again.Code)
	}

	events := recorder.waitForCount(t, 4, 3*time.Second)
	if events[0].Body["eventType"] != "user.kyc.completed" {
		t.Fatalf("first kyc event mismatch: %#v", events[0].Body)
	}
	if !strings.HasPrefix(events[0].Headers.Get("svix-signature"), "v1,") {
		t.Fatalf("signature header missing on kyc webhook")
	}
}
