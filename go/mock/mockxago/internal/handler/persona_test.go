package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/jobs"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/storage"
)

func TestPersonaGetInquiryIframe_UsesInquiryID(t *testing.T) {
	h := setupTestHandler(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}

	repoRoot := filepath.Clean(filepath.Join(cwd, "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change dir to repo root: %v", err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	req := httptest.NewRequest(http.MethodGet, "/v1/inquiries/inq_123/iframe?user_id=wallet-abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inquiryId", "inq_123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaGetInquiryIframe(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `name="user_id" value="inq_123"`)
	assert.NotContains(t, body, `value="wallet-abc"`)
}

func TestPersonaGetInquiryIframe_DefaultsToInquiryID(t *testing.T) {
	h := setupTestHandler(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}

	repoRoot := filepath.Clean(filepath.Join(cwd, "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change dir to repo root: %v", err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	req := httptest.NewRequest(http.MethodGet, "/v1/inquiries/inq_456/iframe", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inquiryId", "inq_456")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaGetInquiryIframe(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `name="user_id" value="inq_456"`)
}

// = Persona API Handler Tests =

func TestPersonaCreateInquiry_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	reqBody := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"reference-id": "test-user-123",
				"country-code": "ZA",
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	r := authorizedRequest(http.MethodPost, "/persona/inquiries", token, body)

	w := httptest.NewRecorder()
	h.PersonaCreateInquiry(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	assert.NotNil(t, resp["data"])
}

func TestPersonaCreateInquiry_InvalidJSON(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodPost, "/persona/inquiries", token, []byte("invalid"))

	w := httptest.NewRecorder()
	h.PersonaCreateInquiry(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPersonaCreateInquiry_EmptyReferenceID(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	reqBody := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"reference-id": "",
				"country-code": "ZA",
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	r := authorizedRequest(http.MethodPost, "/persona/inquiries", token, body)

	w := httptest.NewRecorder()
	h.PersonaCreateInquiry(w, r)

	// Should still create with empty reference - returns 201 Created
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestPersonaGetInquiry_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Create an inquiry first
	createReqBody := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"reference-id": "test-inquiry-456",
				"country-code": "ZA",
			},
		},
	}

	body, _ := json.Marshal(createReqBody)
	createReq := authorizedRequest(http.MethodPost, "/persona/inquiries", token, body)
	createW := httptest.NewRecorder()
	h.PersonaCreateInquiry(createW, createReq)

	var createResp map[string]interface{}
	json.NewDecoder(createW.Body).Decode(&createResp)
	data := createResp["data"].(map[string]interface{})
	inquiryID := data["id"].(string)

	// Now get it
	r := authorizedRequest(http.MethodGet, fmt.Sprintf("/persona/inquiries/%s", inquiryID), token, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inquiryId", inquiryID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaGetInquiry(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	assert.NotNil(t, resp["data"])
}

func TestPersonaGetInquiry_NotFound(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodGet, "/persona/inquiries/nonexistent-id", token, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inquiryId", "nonexistent-id")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaGetInquiry(w, r)

	// Handler returns 200 for nonexistent IDs (returns empty or default data)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPersonaGetInquiry_MissingID(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodGet, "/persona/inquiries/", token, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inquiryId", "")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaGetInquiry(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPersonaGetAccount_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Create an account first via sub-account
	subreq := models.CreateSubAccountRequest{
		WalletID:                  "wallet_persona",
		FirstName:                 "Test",
		LastName:                  "User",
		Email:                     "persona@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Test St",
		ThirdPartyVerificationURL: "https://example.com",
	}
	subbody, _ := json.Marshal(subreq)
	subr := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, subbody)
	subw := httptest.NewRecorder()
	h.CreateSubAccount(subw, subr)

	require.Equal(t, http.StatusOK, subw.Code)
	var subresp models.CreateSubAccountResponse
	json.NewDecoder(subw.Body).Decode(&subresp)
	require.NotEmpty(t, subresp.AccountID)

	// Seed DOB via KYC submit so PersonaGetAccount can return non-hardcoded birthdate
	kycForm := url.Values{
		"user_id":    {subreq.WalletID},
		"first_name": {"Test"},
		"last_name":  {"User"},
		"dob":        {"1990-01-01"},
		"address":    {"123 Test St"},
	}
	kycReq := httptest.NewRequest(http.MethodPost, "/kyc/submit", strings.NewReader(kycForm.Encode()))
	kycReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	kycW := httptest.NewRecorder()
	h.KYCIframeSubmit(kycW, kycReq)
	require.Equal(t, http.StatusOK, kycW.Code)

	// Now get the account via Persona API
	r := authorizedRequest(http.MethodGet, fmt.Sprintf("/persona/accounts/%s", subresp.AccountID), token, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", subresp.AccountID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaGetAccount(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	assert.NotNil(t, resp["data"])
}

func TestPersonaGetAccount_NotFound(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodGet, "/persona/accounts/nonexistent-account", token, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "nonexistent-account")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaGetAccount(w, r)

	// Account doesn't exist, should return 404.
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPersonaGetAccount_MissingID(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodGet, "/persona/accounts/", token, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaGetAccount(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPersonaRemoveTag_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Create an account first
	subreq := models.CreateSubAccountRequest{
		WalletID:                  "wallet_remove_tag",
		FirstName:                 "Test",
		LastName:                  "User",
		Email:                     "removetag@example.com",
		MobileNumber:              "+27123456789",
		IdentityType:              "individual",
		IDNumber:                  "9001011234567",
		PhysicalAddress:           "123 Test St",
		ThirdPartyVerificationURL: "https://example.com",
	}
	subbody, _ := json.Marshal(subreq)
	subr := authorizedRequest(http.MethodPost, "/xago/v1/company/accounts", token, subbody)
	subw := httptest.NewRecorder()
	h.CreateSubAccount(subw, subr)

	require.Equal(t, http.StatusOK, subw.Code)
	var subresp models.CreateSubAccountResponse
	json.NewDecoder(subw.Body).Decode(&subresp)

	// Now remove a tag
	removeReq := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"tag": "test-tag",
			},
		},
	}

	removeBody, _ := json.Marshal(removeReq)
	r := authorizedRequest(http.MethodDelete, fmt.Sprintf("/persona/accounts/%s/tags", subresp.AccountID), token, removeBody)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", subresp.AccountID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaRemoveTag(w, r)

	// Should succeed even if tag doesn't exist in this mock
	assert.True(t, w.Code >= 200 && w.Code < 300)
}

func TestPersonaRemoveTag_MissingAccountID(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	removeReq := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"tag": "test-tag",
			},
		},
	}

	removeBody, _ := json.Marshal(removeReq)
	r := authorizedRequest(http.MethodDelete, "/persona/accounts//tags", token, removeBody)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaRemoveTag(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPersonaInquirySubmit_Success(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	// Create an inquiry first
	createReqBody := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"reference-id": "submit-test-inquiry",
				"country-code": "ZA",
			},
		},
	}

	body, _ := json.Marshal(createReqBody)
	createReq := authorizedRequest(http.MethodPost, "/persona/inquiries", token, body)
	createW := httptest.NewRecorder()
	h.PersonaCreateInquiry(createW, createReq)

	require.Equal(t, http.StatusCreated, createW.Code)
	var createResp map[string]interface{}
	json.NewDecoder(createW.Body).Decode(&createResp)
	data := createResp["data"].(map[string]interface{})
	inquiryID := data["id"].(string)

	// Submit the inquiry - PersonaInquirySubmit expects form-encoded submission
	submitReq := map[string]string{
		"first_name": "John",
		"last_name":  "Doe",
		"dob":        "1990-01-01",
		"address":    "123 Main Street",
	}

	submitBody := bytes.NewBufferString("")
	for k, v := range submitReq {
		submitBody.WriteString(k)
		submitBody.WriteString("=")
		submitBody.WriteString(v)
		submitBody.WriteString("&")
	}

	r := authorizedRequest(http.MethodPost, fmt.Sprintf("/persona/inquiries/%s/submit", inquiryID), token, submitBody.Bytes())
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inquiryId", inquiryID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaInquirySubmit(w, r)

	// PersonaInquirySubmit returns 200 OK with simple JSON response
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "ok", resp["status"])
}

func TestPersonaInquirySubmit_InvalidJSON(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	r := authorizedRequest(http.MethodPost, "/persona/inquiries/test-id/submit", token, []byte("invalid"))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inquiryId", "test-id")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaInquirySubmit(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPersonaInquirySubmit_MissingInquiryID(t *testing.T) {
	h := setupAuthHandler(t)
	token := issueToken(t, h)

	submitReq := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"status": "approved",
			},
		},
	}

	submitBody, _ := json.Marshal(submitReq)
	r := authorizedRequest(http.MethodPost, "/persona/inquiries//submit", token, submitBody)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inquiryId", "")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaInquirySubmit(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPersonaSDK_ServesScript(t *testing.T) {
	h := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/persona-sdk.js", nil)
	w := httptest.NewRecorder()
	h.PersonaSDK(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/javascript; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Greater(t, w.Body.Len(), 0)
}

// --- Tests for the fix: PersonaInquirySubmit must fire the Persona webhook chain ---
//
// Before the fix, POST /v1/inquiries/{id}/submit saved KYC form data but never fired
// any webhooks. The backend's accountTagAddedWebhook (triggered by account.tag-added)
// is what starts SetKYCStatusWorkflow → CreateBalanceAccountWorkflow → CreateSubAccount.
// Without it, xago_sub_accounts was never populated and all deposit tests timed out.

// submitInquiryForm is a test helper that calls PersonaInquirySubmit with the given data.
func submitInquiryForm(t *testing.T, h *Handler, inquiryID, firstName, lastName, dob, address string) {
	t.Helper()
	form := url.Values{
		"first_name": {firstName},
		"last_name":  {lastName},
		"dob":        {dob},
		"address":    {address},
	}
	r := httptest.NewRequest(http.MethodPost, "/v1/inquiries/"+inquiryID+"/submit", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inquiryId", inquiryID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	h.PersonaInquirySubmit(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}

// personaGetAccount is a test helper that calls PersonaGetAccount and returns the decoded body.
func personaGetAccount(t *testing.T, h *Handler, accountID string) map[string]interface{} {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/v1/accounts/"+accountID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", accountID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	h.PersonaGetAccount(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	return resp
}

// TestPersonaInquirySubmit_FiresInquiryApprovedWebhook verifies that a successful
// KYC form submission fires the inquiry.approved webhook to PERSONA_WEBHOOK_URL.
// This is the first event in the chain that eventually creates the Xago sub-account.
func TestPersonaInquirySubmit_FiresInquiryApprovedWebhook(t *testing.T) {
	eventNames := make(chan string, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		data, _ := payload["data"].(map[string]interface{})
		attrs, _ := data["attributes"].(map[string]interface{})
		if name, ok := attrs["name"].(string); ok {
			eventNames <- name
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := storage.NewMemoryStorage()
	cfg := testConfig()
	cfg.PersonaWebhookURL = server.URL
	h := NewHandler(store, jobs.NewQueue(store), cfg)
	submitInquiryForm(t, h, "wallet-inq-approved", "Thabo", "Mbeki", "1990-01-15", "42 Nelson Mandela Dr")

	select {
	case name := <-eventNames:
		assert.Equal(t, "inquiry.approved", name)
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for inquiry.approved webhook")
	}
}

// TestPersonaInquirySubmit_FiresFullWebhookChain verifies that both inquiry.approved
// and account.tag-added are sent after a successful KYC form submission.
// The backend needs both: inquiry.approved marks the inquiry as approved in the DB,
// and account.tag-added triggers SetKYCStatusWorkflow which creates the Xago sub-account.
func TestPersonaInquirySubmit_FiresFullWebhookChain(t *testing.T) {
	eventNames := make(chan string, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		data, _ := payload["data"].(map[string]interface{})
		attrs, _ := data["attributes"].(map[string]interface{})
		if name, ok := attrs["name"].(string); ok {
			eventNames <- name
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := storage.NewMemoryStorage()
	cfg := testConfig()
	cfg.PersonaWebhookURL = server.URL
	h := NewHandler(store, jobs.NewQueue(store), cfg)
	submitInquiryForm(t, h, "wallet-full-chain", "Thabo", "Mbeki", "1990-01-15", "42 Nelson Mandela Dr")

	// Collect both events; account.tag-added is delayed 2s after inquiry.approved
	var received []string
	deadline := time.After(5 * time.Second)
	for len(received) < 2 {
		select {
		case name := <-eventNames:
			received = append(received, name)
		case <-deadline:
			t.Fatalf("timed out waiting for both webhooks; got so far: %v", received)
		}
	}

	assert.Contains(t, received, "inquiry.approved")
	assert.Contains(t, received, "account.tag-added")
}

// TestPersonaInquirySubmit_WebhookContainsWalletID verifies that account.tag-added
// embeds the wallet ID in the payload so the backend can correlate it with the user.
func TestPersonaInquirySubmit_WebhookContainsWalletID(t *testing.T) {
	walletID := "wallet-webhook-id-check"
	payloads := make(chan map[string]interface{}, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p map[string]interface{}
		json.NewDecoder(r.Body).Decode(&p)
		payloads <- p
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := storage.NewMemoryStorage()
	cfg := testConfig()
	cfg.PersonaWebhookURL = server.URL
	h := NewHandler(store, jobs.NewQueue(store), cfg)
	submitInquiryForm(t, h, walletID, "Thabo", "Mbeki", "1990-01-15", "42 Nelson Mandela Dr")

	// Wait for account.tag-added (second event, arrives ~2s after submit)
	deadline := time.After(5 * time.Second)
	for {
		select {
		case p := <-payloads:
			data := p["data"].(map[string]interface{})
			attrs := data["attributes"].(map[string]interface{})
			if attrs["name"] != "account.tag-added" {
				continue
			}
			eventPayload := attrs["payload"].(map[string]interface{})
			accData := eventPayload["data"].(map[string]interface{})
			assert.Equal(t, walletID, accData["id"])
			accAttrs := accData["attributes"].(map[string]interface{})
			assert.Equal(t, walletID, accAttrs["reference-id"])
			return
		case <-deadline:
			t.Fatal("timed out waiting for account.tag-added webhook")
		}
	}
}

// --- Tests for PersonaGetAccount returning data required by the backend ---

// TestPersonaGetAccount_LookupByWalletIDAfterInquirySubmit verifies that GetAccount
// succeeds when called with the inquiry/wallet ID (not the sub-account's internal AccountID).
// This is the exact lookup path used by GetZAIDNumber: it reads external_id from
// kyc_persona_accounts (which equals walletID) and calls GetAccount with that value.
func TestPersonaGetAccount_LookupByWalletIDAfterInquirySubmit(t *testing.T) {
	h := setupTestHandler(t)
	walletID := "wallet-lookup-by-wallet-id"

	submitInquiryForm(t, h, walletID, "Thabo", "Mbeki", "1990-01-15", "42 Nelson Mandela Dr")

	resp := personaGetAccount(t, h, walletID)
	data := resp["data"].(map[string]interface{})
	attrs := data["attributes"].(map[string]interface{})

	assert.Equal(t, walletID, attrs["reference-id"])
	assert.Equal(t, "Thabo", attrs["name-first"])
	assert.Equal(t, "Mbeki", attrs["name-last"])
	assert.Equal(t, "1990-01-15", attrs["birthdate"])
	assert.Equal(t, "ZA", attrs["country-code"])
}

// TestPersonaGetAccount_ReturnsValidZAIDNumber verifies the account response includes
// a valid 13-digit South African ID number. The backend's isValidZAID function checks
// for exactly 13 numeric digits and a valid Luhn-style checksum. Without a valid ID,
// GetPersonaZAIDNumber returns ErrNoKYCInfo and CreateSubAccount fails non-retryably.
func TestPersonaGetAccount_ReturnsValidZAIDNumber(t *testing.T) {
	h := setupTestHandler(t)
	walletID := "wallet-za-id-valid"

	submitInquiryForm(t, h, walletID, "Thabo", "Mbeki", "1990-01-15", "42 Nelson Mandela Dr")

	resp := personaGetAccount(t, h, walletID)
	data := resp["data"].(map[string]interface{})
	attrs := data["attributes"].(map[string]interface{})
	idNums, ok := attrs["identification-numbers"].(map[string]interface{})
	require.True(t, ok, "identification-numbers must be present in account attributes")
	require.NotEmpty(t, idNums)

	var zaID string
	for _, group := range idNums {
		entries := group.([]interface{})
		for _, entry := range entries {
			e := entry.(map[string]interface{})
			if e["issuing-country"] == "ZA" {
				zaID, _ = e["identification-number"].(string)
			}
		}
	}

	require.NotEmpty(t, zaID, "must include a ZA identification-number")
	assert.Len(t, zaID, 13, "ZA ID must be exactly 13 digits")
	assert.Regexp(t, `^\d{13}$`, zaID, "ZA ID must consist only of digits")
}

// TestPersonaGetAccount_ReturnsDIRTYAndKYCLevel1Tags verifies both tags are present.
// DIRTY drives the dirty-sync path in accountTagAddedWebhook (syncing individual details).
// STATUS:KYC-LEVEL:1 is what SetKYCStatus reads to set the user's KYC level, which
// triggers CreateBalanceAccountWorkflow for Xago users.
func TestPersonaGetAccount_ReturnsDIRTYAndKYCLevel1Tags(t *testing.T) {
	h := setupTestHandler(t)
	walletID := "wallet-tags-check"

	submitInquiryForm(t, h, walletID, "Thabo", "Mbeki", "1990-01-15", "42 Nelson Mandela Dr")

	resp := personaGetAccount(t, h, walletID)
	data := resp["data"].(map[string]interface{})
	attrs := data["attributes"].(map[string]interface{})
	rawTags, ok := attrs["tags"].([]interface{})
	require.True(t, ok, "tags must be present in account attributes")

	tags := make([]string, len(rawTags))
	for i, tag := range rawTags {
		tags[i] = tag.(string)
	}

	assert.Contains(t, tags, "DIRTY", "DIRTY tag required for dirty-sync path")
	assert.Contains(t, tags, "STATUS:KYC-LEVEL:1", "KYC-LEVEL:1 tag required to set KYC status")
}
