package handler

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"go.uber.org/zap"
)

//go:embed web/ui/dashboard.html web/ui/user.html web/ui/kyc_action.html web/ui/card_tx_action.html
var uiTemplateFS embed.FS

func (h *Handler) UIDashboard(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.ListUsers()
	if err != nil {
		logger.Error("ui: failed to list users", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFS(uiTemplateFS, "web/ui/dashboard.html")
	if err != nil {
		logger.Error("ui: failed to parse dashboard template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, map[string]interface{}{"Users": users}); err != nil {
		logger.Error("ui: failed to render dashboard", zap.Error(err))
	}
}

func (h *Handler) UIUserDetail(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	user, err := h.store.GetUser(userID)
	if err != nil || user == nil {
		http.NotFound(w, r)
		return
	}

	balances, err := h.store.GetAllBalances(userID)
	if err != nil {
		logger.Error("ui: failed to get balances", zap.String("user_id", userID), zap.Error(err))
		balances = map[string]float64{}
	}

	txns, err := h.store.ListTransactionsByUser(userID)
	if err != nil {
		logger.Error("ui: failed to list transactions", zap.String("user_id", userID), zap.Error(err))
		txns = nil
	}

	tmpl, err := template.ParseFS(uiTemplateFS, "web/ui/user.html")
	if err != nil {
		logger.Error("ui: failed to parse user template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, map[string]interface{}{
		"User":         user,
		"Balances":     balances,
		"Transactions": txns,
	}); err != nil {
		logger.Error("ui: failed to render user detail", zap.Error(err))
	}
}

// UIKYCForm serves the KYC trigger form (implemented in Phase 4).
func (h *Handler) UIKYCForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// UIKYCAction handles KYC trigger form POST (implemented in Phase 4).
func (h *Handler) UIKYCAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// UICardTxForm serves the card transaction form (implemented in Phase 5).
func (h *Handler) UICardTxForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// UICardTxAction handles card transaction form POST (implemented in Phase 5).
func (h *Handler) UICardTxAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// UICardTxCards returns active cards for a user as JSON (implemented in Phase 5).
func (h *Handler) UICardTxCards(w http.ResponseWriter, r *http.Request) {
	h.sendJSON(w, http.StatusOK, []interface{}{})
}
