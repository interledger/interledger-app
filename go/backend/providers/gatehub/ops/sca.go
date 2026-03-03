package ops

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type SCAAction string

type SCARequest struct {
	Action SCAAction `json:"action"`
	Code   *string   `json:"code"`
}

const (
	SCAActionInitiate = "INITIATE"
	SCAActionVerify   = "VERIFY"
)

func NewSCA(b Backends, cfg gatehub.Config) http.HandlerFunc {
	if cfg.WebhookSecret == "" {
		log.Error("WebhookSecret is empty in Gatehub configuration")
	}

	key, err := hex.DecodeString(cfg.WebhookSecret)
	if err != nil {
		log.Fatal("Failed to decode WebhookSecret", zap.Error(err))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		log.Info("Received GateHub SCA request")

		body, err := Verify(r.Context(), r, key)
		if errors.Is(err, gatehub.ErrInvalidWebhook) {
			log.Warn("Webhook failed validation", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		gatehubUserID := chi.URLParam(r, "userId")
		if gatehubUserID == "" {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = uuid.Validate(gatehubUserID)
		if err != nil {
			log.Warn("received invalid user id. expecting uuidv4", zap.String("receiverUserID", gatehubUserID), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var sr SCARequest
		err = json.Unmarshal(body, &sr)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		switch sr.Action {
		case SCAActionInitiate:
			log.Error("received SCA initiate request. SMS initation not implemented", zap.String("user", gatehubUserID))
			w.WriteHeader(http.StatusOK)
			return
		case SCAActionVerify:
			walletID, err := getWalletID(r.Context(), b, gatehubUserID)
			if err != nil {
				log.Error("failed to retrieve wallet ID", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

			userID, err := b.Users().GetUserIDForWallet(r.Context(), walletID)
			if err != nil {
				log.Error("failed to retrieve user ID", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			// Theoretically, this should never happen, just a sanity check.
			if userID == "" {
				log.Error("user ID retrieved but it is empty", zap.String("walletID", walletID))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			// func getWalletID(ctx context.Context, b Backends, externalUserID string) (string, error) {
			// 	var walletID string
			// 	err := b.DB().GetContext(ctx, &walletID, "SELECT wallet_id FROM gatehub_users WHERE external_id=$1;", externalUserID)
			// 	if errors.Is(err, sql.ErrNoRows) {
			// 		return "", fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
			// 	} else if err != nil {
			// 		return "", fmt.Errorf("%w %s", gatehub.ErrInternal, err)
			// 	}
			//
			// 	return walletID, nil
			// }
		default:
			log.Error("received unknown SCA action", zap.String("user", gatehubUserID), zap.String("action", string(sr.Action)))
			w.WriteHeader(http.StatusOK)
			return
		}

	}
}
