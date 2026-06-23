package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/logger"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/storage"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/utils"
)

// CreateUser handles POST /users.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	userID := req.ID
	if userID == "" {
		userID = utils.GenerateUUID()
	}

	user := &models.User{
		ID:               userID,
		Type:             req.Type,
		Status:           "active",
		SourceOfFunds:    req.SourceOfFunds,
		UserCreationDate: time.Now().Format(time.RFC3339),
		Addresses:        req.Addresses,
		Emails:           req.Emails,
		Phones:           req.Phones,
		Name:             &req.Name,
		DateOfBirth:      req.DateOfBirth,
	}

	if err := h.store.SaveUser(r.Context(), user); err != nil {
		logger.Errorf("failed to save user: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to create user")
		return
	}

	logger.Infof("Created PTI user %s", userID)

	h.sendJSON(w, http.StatusOK, models.IDResponse{
		ID:   userID,
		Link: "/users/" + userID,
	})
}

// GetUser handles GET /users/{id}.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "user id is required")
		return
	}

	user, err := h.store.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			h.sendError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to get user")
		return
	}

	h.sendJSON(w, http.StatusOK, user)
}

// PatchUser handles PATCH /users (merge user info).
func (h *Handler) PatchUser(w http.ResponseWriter, r *http.Request) {
	var req models.PatchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.ID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	user, err := h.store.GetUser(r.Context(), req.ID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			h.sendError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to get user")
		return
	}

	// Merge fields (only overwrite if non-zero in request)
	if req.Type != "" {
		user.Type = req.Type
	}
	if req.DateOfBirth != "" {
		user.DateOfBirth = req.DateOfBirth
	}
	if req.Name != nil {
		user.Name = req.Name
	}
	if len(req.Emails) > 0 {
		user.Emails = req.Emails
	}
	if len(req.Addresses) > 0 {
		user.Addresses = req.Addresses
	}
	if len(req.Phones) > 0 {
		user.Phones = req.Phones
	}
	if req.SourceOfFunds != "" {
		user.SourceOfFunds = req.SourceOfFunds
	}

	if err := h.store.UpdateUser(r.Context(), user); err != nil {
		logger.Errorf("failed to update user: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to update user")
		return
	}

	logger.Infof("Patched PTI user %s", req.ID)

	h.sendJSON(w, http.StatusOK, models.IDResponse{
		ID:   req.ID,
		Link: "/users/" + req.ID,
	})
}

// PutUser handles PUT /users (replace user).
func (h *Handler) PutUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.ID == "" {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	existing, err := h.store.GetUser(r.Context(), req.ID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			h.sendError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to get user")
		return
	}

	// Replace entire user but preserve creation date and status
	user := &models.User{
		ID:               req.ID,
		Type:             req.Type,
		Status:           existing.Status,
		SourceOfFunds:    req.SourceOfFunds,
		UserCreationDate: existing.UserCreationDate,
		Addresses:        req.Addresses,
		Emails:           req.Emails,
		Phones:           req.Phones,
		Name:             &req.Name,
		DateOfBirth:      req.DateOfBirth,
	}

	if err := h.store.UpdateUser(r.Context(), user); err != nil {
		logger.Errorf("failed to update user: %v", err)
		h.sendError(w, http.StatusInternalServerError, "internal_error", "failed to update user")
		return
	}

	logger.Infof("Put PTI user %s", req.ID)

	h.sendJSON(w, http.StatusOK, models.IDResponse{
		ID:   req.ID,
		Link: "/users/" + req.ID,
	})
}
