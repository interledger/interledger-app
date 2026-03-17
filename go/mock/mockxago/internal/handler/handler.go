package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"gitlab.com/fynbos/mock/mockxago/internal/auth"
	"gitlab.com/fynbos/mock/mockxago/internal/jobs"
	"gitlab.com/fynbos/mock/mockxago/internal/logger"
	"gitlab.com/fynbos/mock/mockxago/internal/models"
	"gitlab.com/fynbos/mock/mockxago/internal/storage"
	"gitlab.com/fynbos/mock/mockxago/internal/utils"
)

// Handler handles HTTP requests
type Handler struct {
	store     storage.Storage
	validator *auth.Validator
	queue     *jobs.Queue
	publicKey string
	secret    string
	testMode  bool
}

// NewHandler creates a new handler
func NewHandler(store storage.Storage, queue *jobs.Queue) *Handler {
	return &Handler{
		store:     store,
		validator: auth.NewValidator(store),
		queue:     queue,
		publicKey: os.Getenv("XAGO_API_PUBLIC_KEY"),
		secret:    os.Getenv("XAGO_API_SECRET"),
		testMode:  strings.EqualFold(os.Getenv("XAGO_MOCK_TEST_MODE"), "true"),
	}
}

// AuthMiddleware validates bearer tokens for protected routes
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := h.validator.ValidateToken(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			switch err {
			case auth.ErrMissingToken:
				h.sendError(w, http.StatusUnauthorized, "unauthorized", "missing authorization header")
				return
			case auth.ErrInvalidFormat:
				h.sendError(w, http.StatusUnauthorized, "unauthorized", "invalid authorization format")
				return
			case auth.ErrTokenExpired:
				h.sendError(w, http.StatusUnauthorized, "unauthorized", "token expired")
				return
			case auth.ErrInvalidToken:
				h.sendError(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			default:
				h.sendError(w, http.StatusUnauthorized, "unauthorized", "authentication failed")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Login handles POST /xago/v1/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if strings.TrimSpace(req.PolicyID) == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "policyId is required")
		return
	}

	// Extract credentials from fields
	var publicKey, secret string
	for _, field := range req.Fields {
		switch field.FieldName {
		case "publicKey", "apiPublicKey":
			publicKey = field.FieldValue
		case "secret", "secretKey", "apiSecretKey":
			secret = field.FieldValue
		}
	}

	if strings.TrimSpace(publicKey) == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "apiPublicKey is required")
		return
	}

	if strings.TrimSpace(secret) == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "apiSecretKey is required")
		return
	}

	// Validate credentials
	if err := auth.ValidateCredentials(publicKey, secret, h.publicKey, h.secret); err != nil {
		if err == auth.ErrMissingCredentials {
			h.sendError(w, http.StatusBadRequest, "missing_credentials", "Missing credentials")
			return
		}
		h.sendError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		logger.Infof("Failed login attempt with public_key=%s", publicKey)
		return
	}

	// Generate token
	tokenValue, err := utils.GenerateToken()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "internal_error", "Failed to generate token")
		logger.Errorf("Failed to generate token: %v", err)
		return
	}

	token := &models.AccessToken{
		ID:        utils.GenerateUUID(),
		Token:     tokenValue,
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

func bearerTokenFromHeader(authHeader string) string {
	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
