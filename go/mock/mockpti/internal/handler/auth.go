package handler

import (
	"encoding/json"
	"net/http"

	"gitlab.com/fynbos/mock/mockpti/internal/logger"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
	"gitlab.com/fynbos/mock/mockpti/internal/utils"
)

// CreateJWT handles POST /auth/jwt.
func (h *Handler) CreateJWT(w http.ResponseWriter, r *http.Request) {
	var req models.TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	token, err := utils.GenerateToken()
	if err != nil {
		logger.Errorf("failed to generate token: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to generate token")
		return
	}

	expiresAt := utils.GenerateTokenExpiresAt()

	logger.Infof("Generated JWT token for url=%s method=%s", req.URL, req.Method)

	h.sendJSON(w, http.StatusOK, models.TokenResponse{
		AccessToken: token,
		ExpiresAt:   float64(expiresAt.Unix()),
		TokenType:   "Bearer",
	})
}
