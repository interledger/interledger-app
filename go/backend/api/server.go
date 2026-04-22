package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	api_middleware "gitlab.com/fynbos/backend/api/middleware"
)

func NewRouter(b Backends) http.Handler {
	r := chi.NewRouter()
	r.Use(api_middleware.MakeUserMiddleware(b.Users()))
	r.Use(api_middleware.MakeWalletMiddleware(b.Users(), b.Wallets()))
	return r
}
