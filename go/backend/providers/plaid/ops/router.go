package ops

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	api_middleware "gitlab.com/fynbos/backend/api/middleware"
	"gitlab.com/fynbos/backend/providers/plaid"
	"gitlab.com/fynbos/backend/user"
)

func NewRouter(client plaid.Client, store plaid.TokenStore, uc user.Client, linker FiantLinker, processor string) http.Handler {
	if client == nil {
		panic("plaid: client is nil")
	}
	if store == nil {
		panic("plaid: store is nil")
	}
	if uc == nil {
		panic("plaid: user.Client is nil")
	}

	h := New(client, store, linker, processor)

	r := chi.NewRouter()
	r.Use(api_middleware.MakeRequestIDMiddleware())
	r.Use(api_middleware.MakeUserMiddleware(uc))

	r.Post("/link-token", h.CreateLinkToken)
	r.Post("/exchange", h.ExchangePublicToken)
	r.Get("/state", h.GetState)
	r.Get("/accounts", h.GetAccounts)
	r.Get("/auth", h.GetAuth)
	r.Get("/balance", h.GetBalance)
	r.Get("/identity", h.GetIdentity)
	r.Get("/transactions", h.GetTransactions)
	r.Delete("/disconnect", h.Disconnect)

	if linker != nil {
		r.Get("/registered", h.GetRegistered)
		if processor != "" {
			r.Post("/link-to-fiant", h.LinkToFiant)
		}
	}

	return r
}
