package ops

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
)

type SCAAction string

type SCARequest struct {
	Action SCAAction `json:"action"`
	Code   *string   `json:"code"`
}

type SCAResponse struct {
	Success bool `json:"success"`
}

const (
	SCAActionInitiate = "INITIATE"
	SCAActionVerify   = "VERIFY"
)

func NewSCAHandler(b Backends, cfg gatehub.Config) http.HandlerFunc {
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

		// Default response for all failures. Even if the request is invalid,
		// we still want to return a 200 OK with the given JSON object.
		resp := SCAResponse{
			Success: false,
		}

		failRespJSON, err := json.Marshal(resp)
		if err != nil {
			// Not much we can do here if marshalling the response fails.
			log.Error("Failed to marshal fail response", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		body, err := Verify(r.Context(), r, key)
		if errors.Is(err, gatehub.ErrInvalidWebhook) {
			log.Error("invalid signature webhook", zap.Error(err))
			sendResponse(w, failRespJSON)
			return
		}
		if err != nil {
			log.Error("failed to verify webhook", zap.Error(err))
			sendResponse(w, failRespJSON)
			return
		}

		gatehubUserID := chi.URLParam(r, "userId")
		if gatehubUserID == "" {
			log.Warn("received empty user id", zap.String("userID", gatehubUserID))
			sendResponse(w, failRespJSON)
			return
		}

		err = uuid.Validate(gatehubUserID)
		if err != nil {
			log.Warn("received invalid user id. expecting uuidv4", zap.String("receiverUserID", gatehubUserID), zap.Error(err))
			sendResponse(w, failRespJSON)
			return
		}

		var sr SCARequest
		err = json.Unmarshal(body, &sr)
		if err != nil {
			log.Warn("received invalid json payload", zap.Error(err))
			sendResponse(w, failRespJSON)
			return
		}

		switch sr.Action {
		case SCAActionInitiate:
			log.Info("received SCA initiate request. SMS initation not implemented", zap.String("gatehub-user-id", gatehubUserID))
			sendResponse(w, failRespJSON)
		case SCAActionVerify:
			valid := HandleSCAVerification(r.Context(), b, sr, gatehubUserID, time.Now())
			resp.Success = valid

			raw, err := json.Marshal(resp)
			if err != nil {
				// Not much we can do here if marshalling the response fails.
				log.Error("Failed to marshal fail response", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			sendResponse(w, raw)
		default:
			log.Warn("received unknown SCA action", zap.String("user", gatehubUserID), zap.String("action", string(sr.Action)))
			sendResponse(w, failRespJSON)
		}
	}
}

func HandleSCAVerification(ctx context.Context, b Backends, req SCARequest, gatehubUserID string, t time.Time) bool {
	if req.Code == nil {
		return false
	}

	walletID, err := getWalletID(ctx, b, gatehubUserID)
	if err != nil {
		log.Error("failed to retrieve wallet ID", zap.Error(err))
		return false
	}

	userID, err := b.Users().GetUserIDForWallet(ctx, walletID)
	if err != nil {
		log.Error("failed to retrieve user ID", zap.Error(err))
		return false
	}

	// Theoretically, this should never happen, just a sanity check.
	if userID == "" {
		log.Error("user ID retrieved but it is empty", zap.String("walletID", walletID))
		return false
	}

	err = b.Users().ValidateTotpCode(ctx, userID, *req.Code, t)
	if err != nil {
		log.Error("failed to validate totp", zap.String("user", userID), zap.String("walletID", walletID), zap.Error(err))
		return false
	}

	return true
}

// sendResponse is a helper function to send the SCA response.
func sendResponse(w http.ResponseWriter, raw []byte) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(raw); err != nil {
		log.Warn("failed to write response", zap.Error(err))
	}
}
