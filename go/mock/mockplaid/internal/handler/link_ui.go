package handler

import (
	"net/http"

	"gitlab.com/fynbos/mock/mockplaid/web"
)

// LinkInitializeJS serves the window.Plaid shim at the Plaid CDN path
// (GET /link/v2/stable/link-initialize.js).
func (h *Handler) LinkInitializeJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(web.LinkInitializeJS)
}

// LinkPage serves the mock Link UI (GET /link) — the bank/account dropdown that
// posts the public token back to the parent app via postMessage.
func (h *Handler) LinkPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(web.LinkHTML)
}
