package ops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func CheckCredentialsHandler(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		var ccr CredentialsCheckRequest
		err = json.Unmarshal(body, &ccr)
		if err != nil {
			http.Error(w, "failed to unmarshal body", http.StatusInternalServerError)
			return
		}

		session, err := Login(ccr.Username, ccr.Password)
		if err != nil {
			http.Error(w, "failed to login", http.StatusInternalServerError)
			return
		}

		resp := CredentialsCheckResponse{
			HasMFA:           session.hasMFA,
			CredentialsValid: session.credentialsValid,
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "failed to create response", http.StatusInternalServerError)
			return
		}

		// TODO: save in vault
		if resp.CredentialsValid && !resp.HasMFA {
			err = saveWealthUser(req.Context(), b, ccr)
			if err != nil {
				http.Error(w, "failed to save wealth user", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}
}

func saveWealthUser(ctx context.Context, b Backends, ccr CredentialsCheckRequest) error {
	_, err := b.DB().ExecContext(ctx, "INSERT INTO wealth_users (external_id, easy_equities_username) VALUES ($1, $2) ON CONFLICT (external_id) DO NOTHING", ccr.WealthUserID, ccr.Username)
	return err
}
