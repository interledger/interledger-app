package v1

import (
	"io"
	"net/http"
)

func (h *handlers) getAccountStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if _, err := h.users.UserForContext(ctx); err != nil {
		toHTTPError(w, r, err)
		return
	}

	wallet, err := h.wallets.ForContext(ctx)
	if err != nil {
		toHTTPError(w, r, err)
		return
	}

	body, err := h.gatehub.GetAccountStatement(ctx, wallet.ID)
	if err != nil {
		toHTTPError(w, r, err)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"account-statement.pdf\"")
	w.WriteHeader(http.StatusOK)

	io.Copy(w, body)
}
