package webhook

import (
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
)

func NewHandleInboundCard(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		var card verygoodsecurity.Card
		err = json.Unmarshal(body, &card)
		if err != nil {
			http.Error(w, "failed to parse payload", http.StatusBadRequest)
			return
		}

		_, err = b.VGS().CreateCard(r.Context(), card)
		if err != nil {
			http.Error(w, "failed to create card", http.StatusBadRequest)
			return
		}

		// TODO: Should create new LinkedAccount by newCard.ID

		w.WriteHeader(http.StatusOK)
	}
}
