package http

import (
	"net/http"

	"gitlab.com/fynbos/backend/authorisation/ops"

	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
)

func AuthorisationHTTPHandler(b ops.Backends) http.Handler {
	router := chi.NewRouter()
	router.Use(otelchi.Middleware("authorisation", otelchi.WithChiRoutes(router)))

	router.Post("/grant", grantHandler(b))
	router.Post("/continue", continueHandler(b))
	router.Post("/refresh", refreshHandler(b))

	return router
}

func grantHandler(b ops.Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}

func continueHandler(b ops.Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}
func refreshHandler(b ops.Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}
