package webhooks

import (
	"encoding/json"
	"io"
	"net/http"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/providers/unit/workflows"
	"gitlab.com/fynbos/log"
)

const SignatureHeader = "x-unit-signature"

func MakeHttpHandler(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to parse payload", 500)
			return
		}

		if err := b.Unit().VerifyWebhook(r.Context(), payload, r.Header.Get(SignatureHeader)); err != nil {
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

		_, err = b.Temporal().ExecuteWorkflow(
			r.Context(),
			client.StartWorkflowOptions{},
			workflows.UnitHandleEventsWorkflow,
			body.Data,
		)
		if err != nil {
			log.Error("Unit webhooks failed to execute workflow.", zap.Error(err))
			http.Error(w, "Internal server error", 500)
			return
		}

		w.WriteHeader(200)
	}
}
