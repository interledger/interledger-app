package v1

import (
	"net/http"
	"strings"
)

/*
note(bradu): these hooks are only for sandbox testing purposes
they allow us to trigger settlement and return of ACH transactions
in a sandbox environment
*/
func (ctrl *Controller) ReturnTransactionHook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		requestID := parts[len(parts)-1]

		if err := ctrl.Transactions.SandboxAction(r.Context(), requestID, ReturnAch); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (ctrl *Controller) SettleTransactionHook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		requestID := parts[len(parts)-1]

		if err := ctrl.Transactions.SandboxAction(r.Context(), requestID, SettleAch); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
