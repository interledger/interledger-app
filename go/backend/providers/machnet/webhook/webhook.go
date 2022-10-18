package webhook

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
)

const SignatureHeader = "x-raas-webhook-signature"

func New(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		err = b.Machnet().ValidateWebhook(r.Context(), body, r.Header.Get(SignatureHeader))
		if errors.Is(err, machnet.ErrInvalidSignature) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		var event external.Event
		err = json.Unmarshal(body, &event)
		if err != nil {
			http.Error(w, "failed to parse payload", http.StatusBadRequest)
			return
		}

		// TODO: validate payload
		err = b.Machnet().HandleEvent(r.Context(), event)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
