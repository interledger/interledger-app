package webhook

import (
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/fynbos/backend/providers/machnet/external"
)

func New(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
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
