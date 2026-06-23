package ops

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	api_middleware "gitlab.com/fynbos/backend/api/middleware"
	"gitlab.com/fynbos/backend/providers/plaid"
	"gitlab.com/fynbos/backend/user"
)

func NewRouter(client plaid.Client, uc user.Client, linker FiantLinker, processor string) http.Handler {
	if client == nil {
		panic("plaid: client is nil")
	}
	if uc == nil {
		panic("plaid: user.Client is nil")
	}
	if linker == nil {
		panic("plaid: linker is nil")
	}
	if processor == "" {
		panic("plaid: processor is empty")
	}

	h := New(client, linker, processor)

	r := chi.NewRouter()
	r.Use(api_middleware.MakeRequestIDMiddleware())
	r.Use(api_middleware.MakeUserMiddleware(uc))

	r.Post("/link-token", h.CreateLinkToken)
	r.Get("/registered", h.GetRegistered)
	r.Post("/link-to-fiant", h.LinkToFiant)

	return r
}
