package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/authorisation/ops"
	"gitlab.com/fynbos/httpmessagesignatures"
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
		rawBody, err := io.ReadAll(req.Body)
		if err != nil {
			log.Error("failed to read grant request body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err = httpmessagesignatures.VerifyContentDigest(req.Context(), req.Header.Get("Content-Digest"), rawBody); err != nil {
			log.Error("grant request does not match Content-Digest header.", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var gr authorisation.GrantRequest
		err = json.Unmarshal(rawBody, &gr)
		if err != nil {
			log.Error("failed to unmarshal grant request body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if !ops.VerifyRequest(req.Context(), b, req, gr.Client) {
			log.Error(
				"grant request failed signature validation",
				zap.String("clientURI", gr.Client),
			)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		grant, err := ops.CreateGrant(req.Context(), b, gr)
		if err != nil {
			log.Error("failed to create grant", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		resp := authorisation.GrantAccessTokenResp{AccessTokens: grant.Tokens}

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
