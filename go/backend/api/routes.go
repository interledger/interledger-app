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
	r := chi.NewRouter()

	r.Use(api_middleware.MakeUserMiddleware(uc))
	r.Use(api_middleware.MakeWalletMiddleware(uc, wc))

	r.Mount("/core/v1", corev1.NewRouter(uc, wc, gc))

	return r
}
