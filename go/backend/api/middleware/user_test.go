package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/user"
)

func TestUserMiddleware_BearerToken(t *testing.T) {
	t.Parallel()

	alice := &user.User{ID: "alice"}

	tests := []struct {
		name          string
		authHeader    string
		tokenUser     *user.User
		tokenErr      error
		wantStatus    int
		wantUserInCtx bool
	}{
		{
			name:          "valid token attaches user to context",
			authHeader:    "Bearer valid-token",
			tokenUser:     alice,
			wantStatus:    http.StatusOK,
			wantUserInCtx: true,
		},
		{
			name:       "unknown token passes through without user",
			authHeader: "Bearer unknown",
			tokenErr:   user.ErrNoUserFound,
			wantStatus: http.StatusOK,
		},
		{
			name:       "AAL1 required returns 401",
			authHeader: "Bearer low-aal-token",
			tokenErr:   user.ErrAAL1Required,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "AAL2 required returns 401",
			authHeader: "Bearer low-aal-token",
			tokenErr:   user.ErrAAL2Required,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unexpected token error returns 500",
			authHeader: "Bearer bad-token",
			tokenErr:   errors.New("kratos is down"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotUser *user.User
			uc := &stubUserClient{tokenUser: tt.tokenUser, tokenErr: tt.tokenErr}

			r := chi.NewRouter()
			r.Use(MakeUserMiddleware(uc))
			r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
				gotUser, _ = uc.UserForContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantUserInCtx {
				assert.NotNil(t, gotUser)
			} else {
				assert.Nil(t, gotUser)
			}
		})
	}
}

func TestUserMiddleware_Cookie(t *testing.T) {
	t.Parallel()

	alice := &user.User{ID: "alice"}

	tests := []struct {
		name          string
		cookie        *http.Cookie
		cookieUser    *user.User
		cookieErr     error
		wantStatus    int
		wantUserInCtx bool
	}{
		{
			name:          "valid cookie attaches user to context",
			cookie:        &http.Cookie{Name: "ory_kratos_session", Value: "valid-session"},
			cookieUser:    alice,
			wantStatus:    http.StatusOK,
			wantUserInCtx: true,
		},
		{
			name:       "no cookie passes through without user",
			wantStatus: http.StatusOK,
		},
		{
			name:       "AAL1 required returns 401",
			cookie:     &http.Cookie{Name: "ory_kratos_session", Value: "low-aal"},
			cookieErr:  user.ErrAAL1Required,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "AAL2 required returns 401",
			cookie:     &http.Cookie{Name: "ory_kratos_session", Value: "low-aal"},
			cookieErr:  user.ErrAAL2Required,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unexpected cookie error returns 500",
			cookie:     &http.Cookie{Name: "ory_kratos_session", Value: "bad-session"},
			cookieErr:  errors.New("kratos is down"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotUser *user.User
			uc := &stubUserClient{cookieUser: tt.cookieUser, cookieErr: tt.cookieErr}

			r := chi.NewRouter()
			r.Use(MakeUserMiddleware(uc))
			r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
				gotUser, _ = uc.UserForContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantUserInCtx {
				assert.NotNil(t, gotUser)
			} else {
				assert.Nil(t, gotUser)
			}
		})
	}
}
