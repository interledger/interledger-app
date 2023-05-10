package handler

import (
	"net/http"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type Backends interface {
	Twitter() twitter.Client
	Identities() identities.Client
}

func NewTwitterCallbackHandler(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state == "" {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("Couldn't fetch state from query params")
			return
		}
		authCode := r.URL.Query().Get("code")
		if authCode == "" {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("Couldn't fetch auth code from query params")
			return
		}

		token, err := b.Twitter().CreateToken(r.Context(), &twitter.CreateTokenArgs{
			State:    state,
			AuthCode: authCode,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error creating token", zap.Error(err))
			return
		}

		identity, err := b.Identities().Get(r.Context(), state)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error fetching identity", zap.Error(err))
			return
		}

		_, err = b.Twitter().PostTweet(r.Context(), token, identity.VerificationCode)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error posting tweet", zap.Error(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
