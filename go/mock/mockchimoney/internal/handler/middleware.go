package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/models"
)

// APIKeyMiddleware validates the X-API-KEY header for protected endpoints.
func APIKeyMiddleware(key string, enforce bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enforce {
				next.ServeHTTP(w, r)
				return
			}

			if strings.TrimSpace(r.Header.Get("X-API-KEY")) != key {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(models.APIResponse{
					Status: "error",
					Error:  "unauthorized",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
