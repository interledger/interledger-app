package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
)

func NewHandleInboundCard(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		var card verygoodsecurity.Card
		err = json.Unmarshal(body, &card)
		if err != nil {
			http.Error(w, "failed to parse payload", http.StatusBadRequest)
			return
		}

		_, err = b.VGS().CreateCard(r.Context(), card)
		if err != nil {
			// TODO if ErrUserHasExistingCard check if matching linked account is marked deleted and undelete it.
			http.Error(w, "failed to create card", http.StatusInternalServerError)
			return
		}

		_, err = b.Tabapay().CreateCard(r.Context(), tabapay.CreateCardArgs{
			IdempotencyKey: card.Token,
			WalletID:       card.WalletID,
			Name:           fmt.Sprintf("%s %s", card.Type, card.Last4),
			CardNumber:     card.Token,
			CVV:            card.CVV,
			ExpirationDate: card.Expiry,
		})
		if err != nil {
			http.Error(w, "failed to create card", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
