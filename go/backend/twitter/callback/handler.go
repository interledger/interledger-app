package handler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"gitlab.com/fynbos/backend/openpayments"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type Backends interface {
	Twitter() twitter.Client
	Identities() identities.Client
	OpenPayments() openpayments.Client
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

		connection, err := b.Twitter().CreateConnection(r.Context(), &twitter.CreateConnectionArgs{
			State:    state,
			AuthCode: authCode,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error creating token", zap.Error(err))
			return
		}

		identity, err := b.Identities().Add(r.Context(), identities.AddArgs{
			WalletID:   connection.WalletID,
			Platform:   identities.PlatformTwitter,
			Identifier: connection.Username,
		})
		if err != nil {
			if errors.Is(err, identities.ErrAlreadyExists) {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error adding identity", zap.Error(err))
			return
		}

		pp, err := b.OpenPayments().GetWalletPaymentPointer(r.Context(), identity.WalletID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error getting payment pointer by walletID", zap.Error(err))
			return
		}

		base64SigHas := base64.URLEncoding.EncodeToString(identity.SignatureHash)

		tweet, err := b.Twitter().PostTweet(r.Context(), connection.ID, "I’ve connected my fynbos wallet, to my Twitter identity so I can send and receive payments using this identity. \n\nSee the proof at "+pp.URL+"/claims/"+string(base64SigHas))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error posting tweet", zap.Error(err))
			return
		}

		proofUrl := fmt.Sprintf("https://twitter.com/%s/status/%s", connection.Username, tweet.ID)
		// Verification
		_, err = b.Identities().StartVerification(r.Context(), identity.ID, proofUrl)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error starting verification", zap.Error(err))
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
