package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/appcontext"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		headerValue string
		wantID      string
	}{
		{
			name:        "passthrough client-sent ID",
			headerValue: "request-id",
		},
		{
			name: "generates ID when header is absent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotID string

			r := chi.NewRouter()
			r.Use(MakeRequestIDMiddleware())
			r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
				gotID = appcontext.RequestIDFromContext(r.Context())
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.headerValue != "" {
				req.Header.Set("X-Request-ID", tt.headerValue)
			}
			r.ServeHTTP(httptest.NewRecorder(), req)

			if tt.headerValue != "" {
				assert.Equal(t, tt.headerValue, gotID)
			} else {
				assert.NotEmpty(t, gotID)
			}
		})
	}
}
