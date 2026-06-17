package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/consts"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/logger"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// CreateToken generates an access token (stub - always succeeds)
func (h *Handler) CreateToken(w http.ResponseWriter, r *http.Request) {
	logger.Info("creating token")

	// Check for managedUserUuid header (used for iframe tokens)
	managedUserUuid := r.Header.Get("x-gatehub-managed-user-uuid")
	if managedUserUuid == "" {
		managedUserUuid = r.Header.Get("managedUserUuid")
	}

	logger.Debug("create token: checking managed user uuid", zap.String("managed_user_uuid", managedUserUuid))

	var token string
	if managedUserUuid != "" {
		// Generate a unique token for this user's iframe session
		token = "iframe-token-" + utils.GenerateUUID()

		// Store the mapping of token -> user UUID
		h.tokenToUser.Store(token, managedUserUuid)

		logger.Info("created iframe token for user", zap.String("user_uuid", managedUserUuid), zap.String("token_prefix", token[:20]))
	} else {
		// Regular access token (backward compatibility)
		token = "mock-access-token-" + consts.TestUser1ID
		logger.Warn("no managed user uuid header found, using default token")
	}

	// In sandbox mode, always return a valid token
	response := models.TokenResponse{
		AccessToken: token,
		Token:       token,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	h.sendJSON(w, http.StatusOK, response)
}

// CreateManagedUser creates a new managed user
func (h *Handler) CreateManagedUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateManagedUserRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		h.sendError(w, http.StatusBadRequest, "Email is required")
		return
	}

	logger.Info("creating managed user", zap.String("email", req.Email))

	// Check if user already exists
	existing, _ := h.store.GetUserByEmail(req.Email)
	if existing != nil {
		logger.Info("user already exists", zap.String("email", req.Email))
		h.sendJSON(w, http.StatusOK, *existing)
		return
	}

	// Create new user
	user := &models.User{
		ID:        utils.GenerateUUID(),
		Email:     req.Email,
		Activated: true,
		Managed:   true,
		Role:      "user",
		Features:  []string{"wallet"},
		KYCState:  "", // Not verified yet
		RiskLevel: "",
	}

	if err := h.store.CreateUser(user); err != nil {
		logger.Error("failed to create user", zap.String("email", req.Email), zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	logger.Info("user created", zap.String("email", user.Email), zap.String("user_id", user.ID))
	h.sendJSON(w, http.StatusCreated, *user)
}

// GetManagedUser retrieves a managed user by email
func (h *Handler) GetManagedUser(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		h.sendError(w, http.StatusBadRequest, "Email parameter is required")
		return
	}

	logger.Info("getting managed user", zap.String("email", email))

	user, err := h.store.GetUserByEmail(email)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	h.sendJSON(w, http.StatusOK, models.GetManagedUserResponse{User: *user})
}

// UpdateManagedUserEmail updates a user's email address
func (h *Handler) UpdateManagedUserEmail(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateEmailRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.NewEmail == "" {
		h.sendError(w, http.StatusBadRequest, "Both email and new_email are required")
		return
	}

	logger.Info("updating user email", zap.String("old_email", req.Email), zap.String("new_email", req.NewEmail))

	user, err := h.store.GetUserByEmail(req.Email)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "User not found")
		return
	}

	user.Email = req.NewEmail
	if err := h.store.UpdateUser(user); err != nil {
		logger.Error("failed to update user", zap.String("user_id", user.ID), zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	logger.Info("user email updated", zap.String("user_id", user.ID))
	h.sendJSON(w, http.StatusOK, models.GetManagedUserResponse{User: *user})
}

// UpdateOrganizationConfiguration handles PATCH /auth/v1/users/organization/{organizationID}
func (h *Handler) UpdateOrganizationConfiguration(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "organizationID")
	if orgID == "" {
		h.sendError(w, http.StatusBadRequest, "organization ID is required")
		return
	}

	var req models.UpdateOrganizationConfigurationRequest
	if err := h.decodeJSON(r, &req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate API Base URL
	parsed, err := url.Parse(req.APIBaseURL)
	if err != nil || parsed.Host == "" {
		h.sendError(w, http.StatusBadRequest, "invalid API base URL")
		return
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		h.sendError(w, http.StatusBadRequest, "API base URL must use HTTPS")
		return
	}

	// Validate 2FA type
	if req.TwoFAType != "sms" && req.TwoFAType != "totp" {
		h.sendError(w, http.StatusBadRequest, "2FA type must be 'sms' or 'totp'")
		return
	}

	// Get existing organization
	org, err := h.store.GetOrganization(orgID)
	if err != nil {
		logger.Error("organization not found", zap.String("org_id", orgID), zap.Error(err))
		h.sendError(w, http.StatusNotFound, "organization not found")
		return
	}

	// Update organization (normalize trailing slash to prevent double-slash in constructed URLs)
	org.APIBaseURL = strings.TrimRight(req.APIBaseURL, "/")
	org.TwoFAType = req.TwoFAType

	if err := h.store.UpdateOrganization(org); err != nil {
		logger.Error("failed to update organization", zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "failed to update organization")
		return
	}

	logger.Info("organization configuration updated",
		zap.String("org_id", orgID),
		zap.String("api_base_url", org.APIBaseURL),
		zap.String("2fa_type", org.TwoFAType),
	)

	// Enqueue 2FA callback test job (trigger 3 seconds later)
	h.enqueue2FACallbackTestJob(orgID, 3)

	// Build response - GateHub uses "2006-01-02 15:04:05.999999" format
	timeFormat := "2006-01-02 15:04:05.999999"
	resp := models.UpdateOrganizationConfigurationResponse{
		ID:         org.ID,
		APIBaseURL: org.APIBaseURL,
		TwoFAType:  org.TwoFAType,
		CreatedAt:  org.CreatedAt.Format(timeFormat),
		UpdatedAt:  org.UpdatedAt.Format(timeFormat),
	}

	h.sendJSON(w, http.StatusOK, resp)
}

// enqueue2FACallbackTestJob enqueues a job to test 2FA callbacks after organization config is updated
func (h *Handler) enqueue2FACallbackTestJob(orgID string, offsetDelaySeconds int) {
	if h.webhookManager == nil {
		logger.Warn("webhook manager not available, skipping 2FA callback test job")
		return
	}

	// Using first test user for callback simulation
	testUserID := consts.TestUser1ID

	data := map[string]interface{}{
		"organization_id": orgID,
		"test_user_id":    testUserID,
	}

	h.webhookManager.SendAsync(
		"internal.2fa_callback_test",
		testUserID,
		data,
		float64(offsetDelaySeconds),
	)

	logger.Info("2FA callback test job enqueued",
		zap.String("org_id", orgID),
		zap.String("test_user_id", testUserID),
		zap.Int("delay_seconds", offsetDelaySeconds),
	)
}
