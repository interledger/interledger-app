package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	corev1 "gitlab.com/fynbos/backend/api/core/v1"
	api_middleware "gitlab.com/fynbos/backend/api/middleware"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
)

func NewRouter(uc user.Client, wc wallets.Client, gc gatehub.Client) http.Handler {
	if uc == nil {
		panic("api: user.Client is nil")
	}
	if wc == nil {
		panic("api: wallets.Client is nil")
	}
	if gc == nil {
		panic("api: gatehub.Client is nil")
	}

	r := chi.NewRouter()

	// TODO: discuss adding a recovery middleware
	r.Use(api_middleware.MakeRequestIDMiddleware())
	r.Use(api_middleware.MakeUserMiddleware(uc))
	r.Use(api_middleware.MakeWalletMiddleware(uc, wc))

	r.Mount("/core/v1", corev1.NewRouter(uc, wc, gc))

	return r
}
