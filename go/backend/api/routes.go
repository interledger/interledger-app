package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	corev1 "gitlab.com/fynbos/backend/api/core/v1"
	api_middleware "gitlab.com/fynbos/backend/api/middleware"
)

func NewRouter(b Backends) http.Handler {
	r := chi.NewRouter()

	r.Use(api_middleware.MakeUserMiddleware(b.Users()))
	r.Use(api_middleware.MakeWalletMiddleware(b.Users(), b.Wallets()))

	r.Mount("/core/v1", corev1.NewRouter(b))

	return r
}
