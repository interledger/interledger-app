// note(bradu): this is just a stub
// nothing here is tested or guaranteed to be correct, it's just a starting point for development
// needs more polishing and testing, but this is a starting point for development

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
func (client *Client) ReturnTransactionHook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		requestID := parts[len(parts)-1]

		if err := client.TransactionsService.SandboxAction(r.Context(), requestID, RETURN_ACH); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (client *Client) SettleTransactionHook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		requestID := parts[len(parts)-1]

		if err := client.TransactionsService.SandboxAction(r.Context(), requestID, SETTLE_ACH); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
