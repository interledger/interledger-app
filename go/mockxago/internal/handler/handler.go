package handler

import (
	"encoding/json"
	"net/http"
	"os"

	"gitlab.com/fynbos/mockxago/internal/auth"
	"gitlab.com/fynbos/mockxago/internal/logger"
	"gitlab.com/fynbos/mockxago/internal/models"
	"gitlab.com/fynbos/mockxago/internal/storage"
	"gitlab.com/fynbos/mockxago/internal/utils"
)

// Handler handles HTTP requests
type Handler struct {
	store      storage.Storage
	validator  *auth.Validator
	publicKey  string
	secret     string
}

// NewHandler creates a new handler
func NewHandler(store storage.Storage) *Handler {
	return &Handler{
		store:     store,
		validator: auth.NewValidator(store),
		publicKey: os.Getenv("XAGO_API_PUBLIC_KEY"),
		secret:    os.Getenv("XAGO_API_SECRET"),
	}
}

// Login handles POST /xago/v1/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Extract credentials from fields
	var publicKey, secret string
	for _, field := range req.Fields {
		switch field.FieldName {
		case "publicKey":
			publicKey = field.FieldValue
		case "secret":
			secret = field.FieldValue
		}
	}

	// Validate credentials
	if err := auth.ValidateCredentials(publicKey, secret, h.publicKey, h.secret); err != nil {
		if err == auth.ErrMissingCredentials {
			h.sendError(w, http.StatusBadRequest, "missing_credentials", "Missing credentials")
			return
		}
		h.sendError(w, http.StatusUnauthorized, "unauthorized", "Invalid credentials")
		logger.Infof("Failed login attempt with public_key=%s", publicKey)
		return
	}

	// Generate token
	token := &models.AccessToken{
		ID:        utils.GenerateUUID(),
		Token:     utils.GenerateToken(),
		ExpiresAt: utils.GenerateTokenExpiresAt(),
	}

	if err := h.store.SaveAccessToken(r.Context(), token); err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "Failed to create token")
		logger.Errorf("Failed to save token: %v", err)
		return
	}

	logger.Infof("Successful login with public_key=%s, issued token=%s", publicKey, token.ID)

	h.sendJSON(w, http.StatusOK, models.LoginResponse{
		TokenValue: token.Token,
	})
}

// sendJSON sends a JSON response
func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// sendError sends an error response
func (h *Handler) sendError(w http.ResponseWriter, status int, error, message string) {
	h.sendJSON(w, status, models.ErrorResponse{
		Error:   error,
		Message: message,
	})
}
