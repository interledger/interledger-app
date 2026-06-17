package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
)

type handlers struct {
	users   user.Client
	wallets wallets.Client
	gatehub gatehub.Client
}

func NewRouter(uc user.Client, wc wallets.Client, gc gatehub.Client) http.Handler {
	r := chi.NewRouter()
	h := &handlers{users: uc, wallets: wc, gatehub: gc}

	r.Route("/statements", func(r chi.Router) {
		r.Get("/account-confirmation", h.getAccountConfirmation)
		r.Get("/monthly/{year}/{month}", h.getAccountStatement)
		r.Get("/transaction/{id}", h.getTransactionStatement)
	})

	return r
}
