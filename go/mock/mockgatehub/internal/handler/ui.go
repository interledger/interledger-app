package handler

import (
	"fmt"
	"net/http"
)

// UIDashboard serves the admin UI dashboard (stub — replaced in Phase 3).
func (h *Handler) UIDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<html><body><h1>MockGatehub Admin UI — coming soon</h1></body></html>")
}

// UIUserDetail serves the per-user detail page (stub — replaced in Phase 3).
func (h *Handler) UIUserDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<html><body><h1>User detail — coming soon</h1></body></html>")
}

// UIKYCForm serves the KYC trigger form (stub — replaced in Phase 4).
func (h *Handler) UIKYCForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<html><body><h1>KYC action — coming soon</h1></body></html>")
}

// UIKYCAction handles KYC trigger form POST (stub — replaced in Phase 4).
func (h *Handler) UIKYCAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<html><body><h1>KYC action — coming soon</h1></body></html>")
}

// UICardTxForm serves the card transaction form (stub — replaced in Phase 5).
func (h *Handler) UICardTxForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<html><body><h1>Card transaction — coming soon</h1></body></html>")
}

// UICardTxAction handles card transaction form POST (stub — replaced in Phase 5).
func (h *Handler) UICardTxAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<html><body><h1>Card transaction — coming soon</h1></body></html>")
}

// UICardTxCards returns active cards for a user as JSON (stub — replaced in Phase 5).
func (h *Handler) UICardTxCards(w http.ResponseWriter, r *http.Request) {
	h.sendJSON(w, http.StatusOK, []interface{}{})
}
