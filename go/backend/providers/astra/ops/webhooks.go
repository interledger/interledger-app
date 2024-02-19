package ops

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/env"

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

		switch hook.Type {
		case "user_intent_updated":
			handleUserIntentUpdated(r.Context(), b, hook, w)
		case "transfer_updated":
			handleTransferUpdate(r.Context(), b, hook, w)
		default:
			log.Info("Unhandled astra webhook", zap.String("type", hook.Type))
			w.WriteHeader(http.StatusOK)
			return
		}
	}
}

func handleTransferUpdate(ctx context.Context, b Backends, hook Webhook, w http.ResponseWriter) {
	var walletID string
	err := b.DB().GetContext(ctx, &walletID, "SELECT wallet_id FROM astra_user_intents WHERE user_id=$1 ", hook.UserID)
	if err != nil {
		log.Error("failed to get astra intent for webhook", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	token, err := GetToken(ctx, b, walletID)
	if err != nil {
		log.Error("failed to get astra user token for webhook", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx, err := b.External().GetTransfer(ctx, token, hook.ResourceID)
	if err != nil {
		log.Error("failed to get astra transfer for webhook", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = b.Payments().SignalAstraTransferUpdate(ctx, tx.ClientCorrelationID)
	if err != nil {
		log.Error("failed to signal payment workflow for astra webhook", zap.Error(err))
		// Return OK, we poll hourly anyway.
		w.WriteHeader(http.StatusOK)
	}
}

func handleUserIntentUpdated(ctx context.Context, b Backends, hook Webhook, w http.ResponseWriter) {
	intent, err := b.External().GetIntent(ctx, hook.ResourceID)
	if err != nil {
		log.Error("failed to retrieve astra intent for webhook", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var walletID string
	err = b.DB().GetContext(ctx, &walletID, "UPDATE astra_user_intents SET status=$1, user_id=$2, updated_at=now() WHERE intent_id=$3 RETURNING  wallet_id",
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

var (
	authHeaderRegex = regexp.MustCompile("Bearer .*")
)

// This endpoint is secured using a Bearer Token set in the Authorization header.
func GetTrustedAuthenticationInfo(b Backends) http.HandlerFunc {
	astraWebhookBearerToken := os.Getenv("ASTRA_WEBHOOK_BEARER_TOKEN")
	if astraWebhookBearerToken == "" && !env.IsLocal() {
		log.Fatal("ASTRA_WEBHOOK_BEARER_TOKEN is not set")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		parts := strings.Split(authHeader, " ")
		var bearerToken string
		if authHeaderRegex.MatchString(authHeader) && len(parts) >= 2 {
			bearerToken = parts[1]
		}

		if strings.TrimSpace(bearerToken) != astraWebhookBearerToken {
			log.Warn("Unauthenticated request to astra trusted authentication webhook", zap.String("Authorization header", authHeader))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		walletID := chi.URLParam(r, "id")
		var ret WalletInfoResponse
		ret.CustomerID = walletID
		data, err := b.KYC().GetIndividualDetails(r.Context(), walletID)
		if errors.Is(err, kyc.ErrNoKYCInfo) {
			log.Error("Failed to respond to Astra trusted authentication request", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if err != nil {
			log.Error("Failed to respond to Astra trusted authentication request", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		ret.FirstName = data.FirstName
		ret.LastName = data.LastName

		usrs, err := b.Users().ListUsers(r.Context(), walletID)
		if err != nil {
			log.Error("Failed to respond to Astra trusted authentication request", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		for _, usr := range usrs {
			ret.Phone = usr.PhoneNumber
			break
		}

		if ret.Phone != "" {
			verifications, err := b.Twilio().ListSuccessfulVerificationAttempts(r.Context(), twilio.ListSuccessfulVerificationAttemptsArgs{
				To:    ret.Phone,
				Limit: 1,
				After: time.Now().AddDate(0, 0, -30),
			})
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if len(verifications) > 0 {
				ret.LastAuthenticationTime = verifications[0].UpdatedAt.Format(time.RFC3339)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(ret)
		if err != nil {
			log.Error("Failed to respond to Astra trusted authentication request", zap.Error(err))
		}
	}
}

type WalletInfoResponse struct {
	CustomerID             string `json:"customer_id"`
	FirstName              string `json:"first_name"`
	LastName               string `json:"last_name"`
	Phone                  string `json:"phone"`
	LastAuthenticationTime string `json:"last_authentication_time"`
}
