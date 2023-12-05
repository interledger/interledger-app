package ops

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func GetTrustedAuthenticationInfo(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: authenticate astra

		walletID := chi.URLParam(r, "id")
		var ret WalletInfoResponse
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
