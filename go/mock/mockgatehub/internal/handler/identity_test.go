package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/storage"
	"gitlab.com/fynbos/mock/mockgatehub/internal/webhook"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setup2FATestHandler creates a handler with a managed user and optional organization config.
// It returns the handler, the store, and the user ID.
func setup2FATestHandler(t *testing.T, orgAPIBaseURL string) (*Handler, *storage.MemoryStorage, string) {
	t.Helper()
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "test-secret", nil, store, "default-org")
	h := NewHandler(store, wm)

	// Create a test user
	user := &models.User{
		ID:        "test-user-001",
		Email:     "kycuser@example.com",
		Managed:   true,
		Activated: true,
		KYCState:  consts.KYCStateActionRequired,
		RiskLevel: consts.RiskLevelLow,
		CreatedAt: time.Now(),
	}
	require.NoError(t, store.CreateUser(user))

	// Map a bearer token to this user
	h.tokenToUser.Store("kyc-token-test", "test-user-001")

	// Create default organization if apiBaseUrl is provided
	if orgAPIBaseURL != "" {
		org := &models.Organization{
			ID:         "default-org",
			APIBaseURL: orgAPIBaseURL,
			TwoFAType:  "totp",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		require.NoError(t, store.CreateOrganization(org))
	}

	return h, store, user.ID
}

// postKYCForm sends a form POST to KYCIframeSubmit with given form values.
func postKYCForm(h *Handler, formValues url.Values) *httptest.ResponseRecorder {
	body := formValues.Encode()
	req := httptest.NewRequest(http.MethodPost, "/iframe/submit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)
	return w
}

func baseKYCForm(userID string) url.Values {
	return url.Values{
		"user_id":    {userID},
		"first_name": {"Jane"},
		"last_name":  {"Doe"},
		"dob":        {"1990-05-15"},
		"address":    {"123 Main St"},
		"city":       {"Testville"},
		"country":    {"US"},
		"risk_level": {"low"},
	}
}

func TestKYCIframeSubmit_Without2FA_NormalFlow(t *testing.T) {
	h, store, userID := setup2FATestHandler(t, "")

	form := baseKYCForm(userID)
	// No trigger_2fa field — normal flow
	w := postKYCForm(h, form)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify user KYC state is accepted
	user, err := store.GetUser(userID)
	require.NoError(t, err)
	assert.Equal(t, consts.KYCStateAccepted, user.KYCState)
}

func TestKYCIframeSubmit_2FA_Success(t *testing.T) {
	// Create a mock integrator 2FA endpoint that returns success
	var receivedBody map[string]interface{}
	var receivedPath string
	mock2FA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		bodyBytes, _ := io.ReadAll(r.Body)
		json.Unmarshal(bodyBytes, &receivedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer mock2FA.Close()

	h, store, userID := setup2FATestHandler(t, mock2FA.URL)

	form := baseKYCForm(userID)
	form.Set("trigger_2fa", "on")
	form.Set("totp_code", "123456")

	w := postKYCForm(h, form)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify 2FA callback was made with correct payload
	assert.Equal(t, fmt.Sprintf("/v1/users/managed/%s/2fa", userID), receivedPath)
	assert.Equal(t, "VERIFY", receivedBody["action"])
	assert.Equal(t, "123456", receivedBody["code"])

	// Verify user KYC state is accepted
	user, err := store.GetUser(userID)
	require.NoError(t, err)
	assert.Equal(t, consts.KYCStateAccepted, user.KYCState)
}

func TestKYCIframeSubmit_2FA_Rejected(t *testing.T) {
	// Mock integrator returns success: false
	mock2FA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	}))
	defer mock2FA.Close()

	h, store, userID := setup2FATestHandler(t, mock2FA.URL)

	form := baseKYCForm(userID)
	form.Set("trigger_2fa", "on")
	form.Set("totp_code", "wrong-code")

	w := postKYCForm(h, form)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify KYC state is NOT accepted (still action_required)
	user, err := store.GetUser(userID)
	require.NoError(t, err)
	assert.Equal(t, consts.KYCStateActionRequired, user.KYCState)
}

func TestKYCIframeSubmit_2FA_IntegratorNon2xx(t *testing.T) {
	// Mock integrator returns 500
	mock2FA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer mock2FA.Close()

	h, store, userID := setup2FATestHandler(t, mock2FA.URL)

	form := baseKYCForm(userID)
	form.Set("trigger_2fa", "on")
	form.Set("totp_code", "123456")

	w := postKYCForm(h, form)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify KYC state is NOT accepted
	user, err := store.GetUser(userID)
	require.NoError(t, err)
	assert.Equal(t, consts.KYCStateActionRequired, user.KYCState)
}

func TestKYCIframeSubmit_2FA_NoOrgCallbackURL(t *testing.T) {
	// No organization configured (empty orgAPIBaseURL)
	h, store, userID := setup2FATestHandler(t, "")

	form := baseKYCForm(userID)
	form.Set("trigger_2fa", "on")
	form.Set("totp_code", "123456")

	w := postKYCForm(h, form)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify error message mentions callback URL
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	msg, _ := resp["message"].(string)
	assert.Contains(t, msg, "no organization callback URL configured")

	// Verify KYC state is NOT accepted
	user, err := store.GetUser(userID)
	require.NoError(t, err)
	assert.Equal(t, consts.KYCStateActionRequired, user.KYCState)
}

func TestKYCIframeSubmit_2FA_EmptyCode(t *testing.T) {
	// 2FA checkbox checked, empty TOTP code — callback should still be made (no client-side validation)
	var receivedBody map[string]interface{}
	mock2FA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		json.Unmarshal(bodyBytes, &receivedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer mock2FA.Close()

	h, store, userID := setup2FATestHandler(t, mock2FA.URL)

	form := baseKYCForm(userID)
	form.Set("trigger_2fa", "on")
	form.Set("totp_code", "") // empty code

	w := postKYCForm(h, form)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify callback was still made with empty code
	assert.Equal(t, "VERIFY", receivedBody["action"])
	assert.Equal(t, "", receivedBody["code"])

	// Verify KYC accepted
	user, err := store.GetUser(userID)
	require.NoError(t, err)
	assert.Equal(t, consts.KYCStateAccepted, user.KYCState)
}

func TestKYCIframeSubmit_2FA_NetworkError(t *testing.T) {
	// Use a URL that will fail to connect
	h, store, userID := setup2FATestHandler(t, "https://127.0.0.1:1") // unroutable port

	form := baseKYCForm(userID)
	form.Set("trigger_2fa", "on")
	form.Set("totp_code", "123456")

	w := postKYCForm(h, form)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify KYC state is NOT accepted
	user, err := store.GetUser(userID)
	require.NoError(t, err)
	assert.Equal(t, consts.KYCStateActionRequired, user.KYCState)
}

func TestKYCIframeSubmit_2FA_TokenResolvedUser(t *testing.T) {
	// 2FA with user resolved from token (no user_id in form)
	mock2FA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer mock2FA.Close()

	h, store, userID := setup2FATestHandler(t, mock2FA.URL)

	form := baseKYCForm("")
	form.Del("user_id")
	form.Set("token", "kyc-token-test") // maps to test-user-001
	form.Set("trigger_2fa", "on")
	form.Set("totp_code", "654321")

	w := postKYCForm(h, form)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify user KYC state is accepted
	user, err := store.GetUser(userID)
	require.NoError(t, err)
	assert.Equal(t, consts.KYCStateAccepted, user.KYCState)
}

// ── Helper for chi URL params ──

func idWithURLParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func idTestHandler(t *testing.T) (*Handler, *storage.MemoryStorage) {
	t.Helper()
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	wm := webhook.NewManager("", "mock-secret", nil, store, "default-org")
	h := NewHandler(store, wm)
	return h, store
}

// ── GetUser ──

func TestGetUser_Success(t *testing.T) {
	h, _ := idTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/id/v1/users/"+consts.TestUser1ID, nil)
	req = idWithURLParams(req, map[string]string{"userID": consts.TestUser1ID})
	w := httptest.NewRecorder()
	h.GetUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.GetUserResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, consts.TestUser1ID, resp.ID)
	assert.Equal(t, consts.TestUser1Email, resp.Email)
	assert.True(t, resp.Activated)
	assert.Equal(t, 0, resp.Verifications[0].Status)
}

func TestGetUser_KYCAccepted(t *testing.T) {
	h, store := idTestHandler(t)

	user, _ := store.GetUser(consts.TestUser1ID)
	user.KYCState = consts.KYCStateAccepted
	store.UpdateUser(user)

	req := httptest.NewRequest(http.MethodGet, "/id/v1/users/"+consts.TestUser1ID, nil)
	req = idWithURLParams(req, map[string]string{"userID": consts.TestUser1ID})
	w := httptest.NewRecorder()
	h.GetUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.GetUserResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 1, resp.Verifications[0].Status)
}

func TestGetUser_NotFound(t *testing.T) {
	h, _ := idTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/id/v1/users/nonexistent", nil)
	req = idWithURLParams(req, map[string]string{"userID": "nonexistent"})
	w := httptest.NewRecorder()
	h.GetUser(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUser_MissingUserID(t *testing.T) {
	h, _ := idTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/id/v1/users/", nil)
	w := httptest.NewRecorder()
	h.GetUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── StartKYC ──

func TestStartKYC_Success(t *testing.T) {
	h, store := idTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/id/v1/users/"+consts.TestUser1ID+"/hubs/gateway1", nil)
	req = idWithURLParams(req, map[string]string{"userID": consts.TestUser1ID, "gatewayID": "gateway1"})
	w := httptest.NewRecorder()
	h.StartKYC(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.StartKYCResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.IframeURL)
	assert.NotEmpty(t, resp.Token)
	assert.Contains(t, resp.IframeURL, "onboarding")

	user, _ := store.GetUser(consts.TestUser1ID)
	assert.Equal(t, consts.KYCStateActionRequired, user.KYCState)
}

func TestStartKYC_UserNotFound(t *testing.T) {
	h, _ := idTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/id/v1/users/nonexistent/hubs/gw1", nil)
	req = idWithURLParams(req, map[string]string{"userID": "nonexistent", "gatewayID": "gw1"})
	w := httptest.NewRecorder()
	h.StartKYC(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStartKYC_MissingParams(t *testing.T) {
	h, _ := idTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/id/v1/users//hubs/", nil)
	w := httptest.NewRecorder()
	h.StartKYC(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── UpdateKYCState ──

func TestUpdateKYCState_Accepted(t *testing.T) {
	h, store := idTestHandler(t)

	body, _ := json.Marshal(models.UpdateKYCStateRequest{
		State:     consts.KYCStateAccepted,
		RiskLevel: consts.RiskLevelLow,
	})
	req := httptest.NewRequest(http.MethodPut, "/id/v1/users/"+consts.TestUser1ID+"/hubs/gw1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = idWithURLParams(req, map[string]string{"userID": consts.TestUser1ID, "gatewayID": "gw1"})
	w := httptest.NewRecorder()
	h.UpdateKYCState(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	user, _ := store.GetUser(consts.TestUser1ID)
	assert.Equal(t, consts.KYCStateAccepted, user.KYCState)
	assert.Equal(t, consts.RiskLevelLow, user.RiskLevel)
}

func TestUpdateKYCState_Rejected(t *testing.T) {
	h, store := idTestHandler(t)

	body, _ := json.Marshal(models.UpdateKYCStateRequest{
		State:     consts.KYCStateRejected,
		RiskLevel: "high",
	})
	req := httptest.NewRequest(http.MethodPut, "/id/v1/users/"+consts.TestUser1ID+"/hubs/gw1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = idWithURLParams(req, map[string]string{"userID": consts.TestUser1ID, "gatewayID": "gw1"})
	w := httptest.NewRecorder()
	h.UpdateKYCState(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	user, _ := store.GetUser(consts.TestUser1ID)
	assert.Equal(t, consts.KYCStateRejected, user.KYCState)
}

func TestUpdateKYCState_UserNotFound(t *testing.T) {
	h, _ := idTestHandler(t)

	body, _ := json.Marshal(models.UpdateKYCStateRequest{State: consts.KYCStateAccepted})
	req := httptest.NewRequest(http.MethodPut, "/id/v1/users/nonexistent/hubs/gw1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = idWithURLParams(req, map[string]string{"userID": "nonexistent", "gatewayID": "gw1"})
	w := httptest.NewRecorder()
	h.UpdateKYCState(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateKYCState_MissingParams(t *testing.T) {
	h, _ := idTestHandler(t)

	body, _ := json.Marshal(models.UpdateKYCStateRequest{State: consts.KYCStateAccepted})
	req := httptest.NewRequest(http.MethodPut, "/id/v1/users//hubs/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.UpdateKYCState(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateKYCState_InvalidBody(t *testing.T) {
	h, _ := idTestHandler(t)

	req := httptest.NewRequest(http.MethodPut, "/id/v1/users/"+consts.TestUser1ID+"/hubs/gw1", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	req = idWithURLParams(req, map[string]string{"userID": consts.TestUser1ID, "gatewayID": "gw1"})
	w := httptest.NewRecorder()
	h.UpdateKYCState(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── OverrideRiskLevel ──

func TestOverrideRiskLevel_Success(t *testing.T) {
	h, store := idTestHandler(t)

	body, _ := json.Marshal(map[string]string{"risk_level": "high", "reason": "manual override"})
	req := httptest.NewRequest(http.MethodPut, "/id/v1/users/"+consts.TestUser1ID+"/hubs/gw1/risk-level", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = idWithURLParams(req, map[string]string{"userID": consts.TestUser1ID, "gatewayID": "gw1"})
	w := httptest.NewRecorder()
	h.OverrideRiskLevel(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	user, _ := store.GetUser(consts.TestUser1ID)
	assert.Equal(t, "high", user.RiskLevel)
}

func TestOverrideRiskLevel_UserNotFound(t *testing.T) {
	h, _ := idTestHandler(t)

	body, _ := json.Marshal(map[string]string{"risk_level": "high"})
	req := httptest.NewRequest(http.MethodPut, "/id/v1/users/nonexistent/hubs/gw1/risk-level", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = idWithURLParams(req, map[string]string{"userID": "nonexistent", "gatewayID": "gw1"})
	w := httptest.NewRecorder()
	h.OverrideRiskLevel(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOverrideRiskLevel_MissingParams(t *testing.T) {
	h, _ := idTestHandler(t)

	body, _ := json.Marshal(map[string]string{"risk_level": "high"})
	req := httptest.NewRequest(http.MethodPut, "/id/v1/users//hubs//risk-level", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.OverrideRiskLevel(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOverrideRiskLevel_InvalidBody(t *testing.T) {
	h, _ := idTestHandler(t)

	req := httptest.NewRequest(http.MethodPut, "/id/v1/users/"+consts.TestUser1ID+"/hubs/gw1/risk-level", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	req = idWithURLParams(req, map[string]string{"userID": consts.TestUser1ID, "gatewayID": "gw1"})
	w := httptest.NewRecorder()
	h.OverrideRiskLevel(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── KYCIframeSubmit (additional functional cases) ──

func TestKYCIframeSubmit_FormDataSuccess(t *testing.T) {
	h, store := idTestHandler(t)

	form := "user_id=" + consts.TestUser1ID + "&first_name=John&last_name=Doe&dob=1990-05-15&address=123+Main+St&city=NYC&country=US"
	req := httptest.NewRequest(http.MethodPost, "/iframe/submit", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	user, _ := store.GetUser(consts.TestUser1ID)
	assert.Equal(t, consts.KYCStateAccepted, user.KYCState)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.Equal(t, 1990, user.BirthYear)
	assert.Equal(t, 5, user.BirthMonth)
	assert.Equal(t, 15, user.BirthDay)
}

func TestKYCIframeSubmit_TokenMappingResolvesUser(t *testing.T) {
	h, store := idTestHandler(t)

	h.tokenToUser.Store("test-tok", consts.TestUser1ID)

	form := "token=test-tok&first_name=Jane"
	req := httptest.NewRequest(http.MethodPost, "/iframe/submit", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	user, _ := store.GetUser(consts.TestUser1ID)
	assert.Equal(t, consts.KYCStateAccepted, user.KYCState)
}

func TestKYCIframeSubmit_NoUserIDOrToken(t *testing.T) {
	h, _ := idTestHandler(t)

	form := "first_name=Jane"
	req := httptest.NewRequest(http.MethodPost, "/iframe/submit", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestKYCIframeSubmit_UserNotFoundInStore(t *testing.T) {
	h, _ := idTestHandler(t)

	form := "user_id=nonexistent"
	req := httptest.NewRequest(http.MethodPost, "/iframe/submit", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestKYCIframeSubmit_CustomRiskLevel(t *testing.T) {
	h, store := idTestHandler(t)

	form := "user_id=" + consts.TestUser1ID + "&risk_level=medium"
	req := httptest.NewRequest(http.MethodPost, "/iframe/submit", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.KYCIframeSubmit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	user, _ := store.GetUser(consts.TestUser1ID)
	assert.Equal(t, "medium", user.RiskLevel)
}
