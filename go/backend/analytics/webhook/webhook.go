package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/interledger/interledger-app/go/backend/analytics"
	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
)

type kratosEvent struct {
	UserId    string `json:"userId"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
}

func NewHandleSignup(backends Backends) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		var event kratosEvent
		err = json.Unmarshal(body, &event)
		if err != nil {
			log.Error("error parsing json", zap.Error(err))
			http.Error(w, "failed to parse body", http.StatusInternalServerError)
			return
		}

		backends.Analytics().Identify(analytics.IdentifyArgs{
			UserId:    event.UserId,
			Email:     event.Email,
			FirstName: event.FirstName,
			LastName:  event.LastName,
		})
		backends.Analytics().TrackUserSignup(event.UserId)

		w.WriteHeader(http.StatusOK)
	}
}

func NewHandleLogin(backends Backends) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		var event kratosEvent
		err = json.Unmarshal(body, &event)
		if err != nil {
			log.Error("error parsing json", zap.Error(err))
			http.Error(w, "failed to parse body", http.StatusInternalServerError)
			return
		}

		backends.Analytics().Identify(analytics.IdentifyArgs{
			UserId:    event.UserId,
			Email:     event.Email,
			FirstName: event.FirstName,
			LastName:  event.LastName,
		})
		backends.Analytics().TrackUserLogin(event.UserId)

		wallets, err := backends.Wallets().List(context.Background(), event.UserId)
		if err != nil {
			log.Error("error get user wallets", zap.Error(err))
		}

		for _, wallet := range wallets {
			backends.Analytics().GroupUserWallet(wallet.ID, event.UserId)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func NewHandleLogout(backends Backends) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		var event kratosEvent
		err = json.Unmarshal(body, &event)
		if err != nil {
			log.Error("error parsing json", zap.Error(err))
			http.Error(w, "failed to parse body", http.StatusInternalServerError)
			return
		}

		backends.Analytics().Identify(analytics.IdentifyArgs{
			UserId:    event.UserId,
			Email:     event.Email,
			FirstName: event.FirstName,
			LastName:  event.LastName,
		})
		backends.Analytics().TrackUserLogout(event.UserId)

		w.WriteHeader(http.StatusOK)
	}
}
