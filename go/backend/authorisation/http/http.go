package http

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/riandyrn/otelchi"
	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/authorisation/ops"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
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

		var gr authorisation.GrantRequest
		err := json.NewDecoder(req.Body).Decode(&gr)
		if err != nil {
			log.Error("failed to unmarshal grant request body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// TODO:
		// - Lookup signatures by client URI (payment pointer)
		// - Validate signature provided in req

		label := gr.AccessToken.Label
		if label == "" {
			label = fmt.Sprintf("auto_label_%d", rand.Int31n(1000))
		}

		resp := authorisation.GrantAccessTokenResp{AccessToken: authorisation.AccessToken{
			Value:     uuid.NewString(),
			Access:    gr.AccessToken.Access,
			Label:     gr.AccessToken.Label,
			ExpiresIn: 60 * 60, // Valid for 1 hour
		}}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error("failed to marshal grant request response body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
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
