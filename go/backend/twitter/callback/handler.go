package handler

import (
	"encoding/base64"
	"gitlab.com/fynbos/backend/openpayments"
	"net/http"

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
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error adding identity", zap.Error(err))
			return
		}

		pp, err := b.OpenPayments().GetPaymentPointer(r.Context(), connection.WalletID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error getting payment pointer by walletID", zap.Error(err))
			return
		}

		base64SigHas := []byte("")
		base64.URLEncoding.Encode(base64SigHas, identity.SignatureHash)
		_, err = b.Twitter().PostTweet(r.Context(), connection.ID, "I’ve connected my fynbos wallet, to my Twitter identity so I can send and receive payments using this identity. \n\nSee the proof at "+pp.URL+"/claims/"+string(base64SigHas))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error posting tweet", zap.Error(err))
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
