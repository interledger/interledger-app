package ops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type Webhook struct {
	Type       string `json:"webhook_type"`
	ID         string `json:"webhook_id"`
	UserID     string `json:"user_id"`
	ResourceID string `json:"resource_id"`
}

func EventWebhook(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("failed to read astra webhook body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var hook Webhook
		err = json.Unmarshal(raw, &hook)
		if err != nil {
			log.Error("failed to unmarshal astra webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if hook.Type != "user_intent_updated" {
			w.WriteHeader(http.StatusOK)
			return
		}

		intent, err := b.External().GetIntent(r.Context(), hook.ResourceID)
		if err != nil {
			log.Error("failed to retrieve astra intent for webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var walletID string
		err = b.DB().GetContext(r.Context(), &walletID, "UPDATE astra_user_intents SET status=$1, user_id=$2, updated_at=now() WHERE intent_id=$3 RETURNING  wallet_id",
			intent.UserID, intent.Status, hook.ResourceID)
		if err != nil {
			log.Error("failed to update astra intent for webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if intent.Status == "approved" || intent.Status == "converted_to_user" {
			// Start polling for their access token as a best effort, as soon as it's used it will do it anyway.
			// No need to hold up the webhook.
			go func(b Backends, walletID string) {
				_, err := CreateOrRefreshToken(context.Background(), b, walletID)
				if err != nil {
					log.Warn("failed to start best effort polling for user token", zap.Error(err))
				}
			}(b, walletID)
		}
	}
}
