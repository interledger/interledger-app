package webhooks

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/providers/unit/external"
)

const SignatureHeader = "x-unit-signature"

func MakeHttpHandler(client unit.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to parse payload", 500)
			return
		}

		if err := client.VerifyWebhook(r.Context(), payload, r.Header.Get(SignatureHeader)); err != nil {
			http.Error(w, "Signature didn't match.", 401)
			return
		}

		body := struct {
			Data []json.RawMessage `json:"data"`
		}{}
		if err := json.Unmarshal(payload, &body); err != nil {
			http.Error(w, "Failed to parse payload", 500)
			return
		}

		didFail := false
		for _, rawEvent := range body.Data {
			var event external.Event
			if err := json.Unmarshal(rawEvent, &event); err != nil {
				didFail = true
				continue
			}

			// TODO: this should not fail. Event must be logged.
			err = client.HandleEvent(context.Background(), event, rawEvent)
			if err != nil {
				http.Error(w, "Failed to handle event", 500)
				return
			}
		}

		// Handling event must not fail. See TODO above.
		// We therefore know it was an unmarshalling error.
		if didFail {
			http.Error(w, "Failed to parse payload", 500)
			return
		}

		w.WriteHeader(200)
	}
}
