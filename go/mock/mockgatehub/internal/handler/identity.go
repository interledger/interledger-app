package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gitlab.com/fynbos/mock/mockgatehub/internal/auth"
	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"
	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// GetUser retrieves user information including KYC state
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		h.sendError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	logger.Info("getting user", zap.String("user_id", userID))

	user, err := h.store.GetUser(userID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	// Build response with verifications array matching production GateHub API
	// Verification status should reflect actual KYC state: only status=1 if accepted
	verificationStatus := 0
	if user.KYCState == consts.KYCStateAccepted {
		verificationStatus = 1
	}

	response := models.GetUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Activated: user.Activated,
		Managed:   user.Managed,
		Role:      user.Role,
		Features:  user.Features,
		KYCState:  user.KYCState,
		RiskLevel: user.RiskLevel,
		CreatedAt: user.CreatedAt,
		Profile: models.UserProfile{
			UUID:               user.ID,
			BirthDay:           user.BirthDay,
			BirthMonth:         user.BirthMonth,
			BirthYear:          user.BirthYear,
			Gender:             user.Gender,
			FirstName:          user.FirstName,
			MiddleName:         user.MiddleName,
			LastName:           user.LastName,
			Citizenship:        user.Citizenship,
			AddressPostalCode:  user.AddressPostalCode,
			AddressSubdivision: user.AddressSubdivision,
			AddressCountryCode: user.AddressCountryCode,
			AddressCity:        user.AddressCity,
			AddressStreet1:     user.AddressStreet1,
			AddressStreet2:     user.AddressStreet2,
			BirthCity:          user.BirthCity,
			BirthCountryCode:   user.BirthCountryCode,
			TaxResidency:       user.TaxResidency,
			ExpectedVolume:     user.ExpectedVolume,
		},
		Verifications: []models.UserVerification{
			{
				UUID:         "mock-verification-uuid",
				Status:       verificationStatus,
				State:        1,
				ProviderType: "sumsub",
			},
		},
	}

	h.sendJSON(w, http.StatusOK, response)
}

// StartKYC initiates the KYC verification process
func (h *Handler) StartKYC(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	gatewayID := chi.URLParam(r, "gatewayID")

	if userID == "" || gatewayID == "" {
		h.sendError(w, http.StatusBadRequest, "User ID and Gateway ID are required")
		return
	}

	logger.Info("starting kyc for user", zap.String("user_id", userID), zap.String("gateway_id", gatewayID))

	user, err := h.store.GetUser(userID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	// Generate a token for the iframe
	token := fmt.Sprintf("kyc-token-%s-%s", userID, gatewayID)
	iframeURL := fmt.Sprintf("/iframe/onboarding?token=%s&user_id=%s", token, userID)

	logger.Info("kyc iframe url generated", zap.String("iframe_url", iframeURL))

	// Only move user into action_required if KYC hasn't been completed yet
	if user.KYCState != consts.KYCStateAccepted {
		user.KYCState = consts.KYCStateActionRequired
		user.RiskLevel = consts.RiskLevelLow
		if err := h.store.UpdateUser(user); err != nil {
			logger.Error("failed to update user kyc state", zap.String("user_id", userID), zap.Error(err))
		}
	}

	response := models.StartKYCResponse{
		IframeURL: iframeURL,
		Token:     token,
	}

	h.sendJSON(w, http.StatusOK, response)
}

// UpdateKYCState updates the KYC verification state for a user
func (h *Handler) UpdateKYCState(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	gatewayID := chi.URLParam(r, "gatewayID")

	if userID == "" || gatewayID == "" {
		h.sendError(w, http.StatusBadRequest, "User ID and Gateway ID are required")
		return
	}

	var req models.UpdateKYCStateRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	logger.Info("updating kyc state for user", zap.String("user_id", userID), zap.String("state", req.State), zap.String("risk_level", req.RiskLevel))

	user, err := h.store.GetUser(userID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	user.KYCState = req.State
	user.RiskLevel = req.RiskLevel

	if err := h.store.UpdateUser(user); err != nil {
		logger.Error("failed to update user kyc state", zap.String("user_id", userID), zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// Send appropriate webhook
	var eventType string
	switch req.State {
	case consts.KYCStateAccepted:
		eventType = consts.WebhookEventKYCAccepted
	case consts.KYCStateRejected:
		eventType = consts.WebhookEventKYCRejected
	case consts.KYCStateActionRequired:
		eventType = consts.WebhookEventKYCActionRequired
	}

	if eventType != "" {
		go h.webhookManager.SendAsync(eventType, userID, map[string]interface{}{
			"state":      req.State,
			"risk_level": req.RiskLevel,
		}, 0)
	}

	h.sendJSON(w, http.StatusOK, user)
}

// KYCIframe serves the KYC onboarding iframe
// OverrideRiskLevel updates the risk level for a user
func (h *Handler) OverrideRiskLevel(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	gatewayID := chi.URLParam(r, "gatewayID")

	if userID == "" || gatewayID == "" {
		h.sendError(w, http.StatusBadRequest, "User ID and Gateway ID are required")
		return
	}

	var req struct {
		RiskLevel string `json:"risk_level"`
		Reason    string `json:"reason"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	logger.Info("overriding risk level for user", zap.String("user_id", userID), zap.String("risk_level", req.RiskLevel), zap.String("reason", req.Reason))

	user, err := h.store.GetUser(userID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	user.RiskLevel = req.RiskLevel

	if err := h.store.UpdateUser(user); err != nil {
		logger.Error("failed to update user risk level", zap.String("user_id", userID), zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	logger.Info("risk level updated successfully for user", zap.String("user_id", userID))
	h.sendJSON(w, http.StatusOK, user)
}

// KYCIframe serves the KYC onboarding iframe
func (h *Handler) KYCIframe(w http.ResponseWriter, r *http.Request) {
	// Accept both 'bearer' (new) and 'token' (legacy) parameters
	token := r.URL.Query().Get("bearer")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	userID := r.URL.Query().Get("user_id")

	logger.Info("serving kyc iframe", zap.String("token", token), zap.String("user_id", userID), zap.String("url", r.URL.String()))

	// Try multiple paths to find the template
	possiblePaths := []string{
		"web/kyc-iframe.html",
		"./web/kyc-iframe.html",
		"../web/kyc-iframe.html",
		"../../web/kyc-iframe.html",
	}

	var templatePath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			templatePath = path
			break
		}
	}

	if templatePath == "" {
		logger.Error("could not find kyc iframe template", zap.Strings("paths_tried", possiblePaths))
		h.sendError(w, http.StatusInternalServerError, "Template not found")
		return
	}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		logger.Error("failed to parse kyc iframe template", zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Template error")
		return
	}

	data := map[string]string{
		"Token":  token,
		"UserID": userID,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		logger.Error("failed to execute kyc iframe template", zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Template execution error")
		return
	}
}

// KYCIframeSubmit handles KYC form submission
func (h *Handler) KYCIframeSubmit(w http.ResponseWriter, r *http.Request) {
	logger.Info("kyc iframe form submitted", zap.String("method", r.Method), zap.String("content_type", r.Header.Get("Content-Type")), zap.Int64("content_length", r.ContentLength))

	// Parse form - for multipart/form-data, ParseForm() should handle it
	// but we need to make sure we parse it correctly
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		// If multipart parsing fails, try regular form parsing
		if err := r.ParseForm(); err != nil {
			logger.Error("failed to parse form", zap.Error(err))
			h.sendError(w, http.StatusBadRequest, "Invalid form data: "+err.Error())
			return
		}
	}

	logger.Debug("form parsed successfully", zap.Int("post_form_count", len(r.PostForm)))
	if r.MultipartForm != nil {
		logger.Debug("multipart form parsed", zap.Int("multipart_value_count", len(r.MultipartForm.Value)))
	}

	userID := r.FormValue("user_id")

	// If user_id is not in form, try to extract from bearer token
	if userID == "" {
		token := r.FormValue("token")
		logger.Warn("user id missing from form, attempting to extract from token", zap.String("token_prefix", tokenPrefix(token)))
		// Try to look up user from token in our map
		if uuid, ok := h.tokenToUser.Load(token); ok {
			if u, ok := uuid.(string); ok {
				userID = u
				logger.Info("found user from token mapping", zap.String("user_id", userID))
			}
		}
	}

	if userID == "" {
		logger.Error("user id could not be determined from form or token", zap.Any("form_fields", r.PostForm))
		h.sendError(w, http.StatusBadRequest, "User ID is required (not found in form or token mapping)")
		return
	}

	logger.Info("kyc form submitted for user", zap.String("user_id", userID))

	user, err := h.store.GetUser(userID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	// Parse KYC form data
	user.FirstName = r.FormValue("first_name")
	user.LastName = r.FormValue("last_name")
	user.AddressStreet1 = r.FormValue("address")
	user.AddressCity = r.FormValue("city")
	user.AddressCountryCode = r.FormValue("country")

	// Parse date of birth (format: YYYY-MM-DD)
	dob := r.FormValue("dob")
	if dob != "" {
		parts := strings.Split(dob, "-")
		if len(parts) == 3 {
			if year, err := strconv.Atoi(parts[0]); err == nil {
				user.BirthYear = year
			}
			if month, err := strconv.Atoi(parts[1]); err == nil {
				user.BirthMonth = month
			}
			if day, err := strconv.Atoi(parts[2]); err == nil {
				user.BirthDay = day
			}
		}
	}

	// Optional 2FA TOTP verification
	trigger2FA := r.FormValue("trigger_2fa")
	if trigger2FA == "on" {
		if err := h.verify2FA(userID, r.FormValue("totp_code")); err != nil {
			h.sendError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Determine KYC outcome from form (defaults to accepted for backward compatibility)
	kycOutcome := r.FormValue("kyc_outcome")
	if kycOutcome == "" {
		kycOutcome = consts.KYCStateAccepted
	}

	riskLevel := r.FormValue("risk_level")
	if riskLevel == "" {
		riskLevel = consts.RiskLevelLow
	}
	user.RiskLevel = riskLevel

	var webhookEvent string
	var responseMessage string

	switch kycOutcome {
	case consts.KYCStateRejected:
		user.KYCState = consts.KYCStateRejected
		webhookEvent = consts.WebhookEventKYCRejected
		responseMessage = "KYC verification rejected"
	case consts.KYCStateActionRequired:
		user.KYCState = consts.KYCStateActionRequired
		webhookEvent = consts.WebhookEventKYCActionRequired
		responseMessage = "KYC verification requires additional action"
	default:
		user.KYCState = consts.KYCStateAccepted
		webhookEvent = consts.WebhookEventKYCAccepted
		responseMessage = "KYC verification completed successfully"
	}

	logger.Info("kyc outcome selected",
		zap.String("user_id", userID),
		zap.String("outcome", kycOutcome),
		zap.String("webhook_event", webhookEvent),
	)

	if err := h.store.UpdateUser(user); err != nil {
		logger.Error("failed to update user after kyc submission", zap.String("user_id", userID), zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	go h.webhookManager.SendAsync(webhookEvent, userID, map[string]interface{}{
		"message": responseMessage,
	}, 2.0)

	h.sendJSON(w, http.StatusOK, map[string]string{
		"status":  user.KYCState,
		"message": responseMessage,
	})
}

// verify2FA resolves the org callback URL, builds the 2FA endpoint for the given user,
// and calls the integrator's VERIFY action. Returns nil on success or an error describing
// the failure (suitable for sending to the client).
func (h *Handler) verify2FA(userID, totpCode string) error {
	callbackURL := h.resolve2FACallbackURL()
	if callbackURL == "" {
		logger.Warn("2FA verification requested but no organization callback URL configured",
			zap.String("user_id", userID),
		)
		return fmt.Errorf("2FA verification requested but no organization callback URL configured")
	}

	normalizedCallbackURL := strings.TrimRight(callbackURL, "/")
	endpoint := fmt.Sprintf("%s/v1/users/managed/%s/2fa", normalizedCallbackURL, userID)
	success, err := h.call2FAVerify(endpoint, totpCode)
	if err != nil {
		logger.Error("2FA verification callback failed",
			zap.String("user_id", userID),
			zap.String("endpoint", endpoint),
			zap.Error(err),
		)
		return fmt.Errorf("2FA verification failed: %w", err)
	}
	if !success {
		logger.Info("2FA verification rejected by integrator",
			zap.String("user_id", userID),
		)
		return fmt.Errorf("2FA verification rejected")
	}

	logger.Info("2FA verification succeeded",
		zap.String("user_id", userID),
	)
	return nil
}

// resolve2FACallbackURL gets the apiBaseUrl strictly from organization config.
// Returns empty string if no org-specific apiBaseUrl is configured (no fallback to WEBHOOK_URL).
func (h *Handler) resolve2FACallbackURL() string {
	if h.webhookManager == nil {
		return ""
	}
	return h.webhookManager.ResolveOrgBaseURL()
}

// call2FAVerify calls the integrator's 2FA endpoint and returns (success, error).
func (h *Handler) call2FAVerify(endpoint, code string) (bool, error) {
	payload := map[string]string{
		"action": "VERIFY",
		"code":   code,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("failed to create request for 2FA callback: %w", err)
	}

	signature := auth.GenerateGateHubWebhookSignature(string(body), h.config.WebhookSecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GH-Webhook-Signature", signature)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("callback request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("callback returned status %d", resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode callback response: %w", err)
	}

	return result.Success, nil
}
