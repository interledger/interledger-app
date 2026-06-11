package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	kratos "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/user"
)

func TestUpdateUserPhone_StatusMapping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		updateStatus int
		expectedErr  error
	}{
		{
			name:         "maps bad request to invalid argument",
			updateStatus: http.StatusBadRequest,
			expectedErr:  user.ErrInvalidArgument,
		},
		{
			name:         "maps conflict to duplicate phone",
			updateStatus: http.StatusConflict,
			expectedErr:  user.ErrDuplicatePhone,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{
						"id": "user-id",
						"traits": map[string]any{
							"phone":         "+10000000000",
							"phoneVerified": true,
						},
					})
					return
				}

				w.WriteHeader(tc.updateStatus)
			}))
			defer srv.Close()

			cfg := kratos.NewConfiguration()
			cfg.Servers = kratos.ServerConfigurations{
				{URL: srv.URL, Description: "Public Kratos"},
				{URL: srv.URL, Description: "Admin Kratos"},
			}

			b := &emailTestBackends{
				kr:  kratos.NewAPIClient(cfg),
				val: validator.New(),
			}

			err := UpdateUserPhone(context.Background(), b, "user-id", "+12223334444")
			require.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestUserForContext(t *testing.T) {
	ctx := context.Background()

	_, err := UserForContext(ctx)
	require.ErrorIs(t, err, user.ErrNoUserFound)

	ctx = context.WithValue(ctx, user.UserCtxKey("user"), &user.User{
		ID:    "1235",
		Email: "test@interledger.test",
	})

	u, err := UserForContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, u.ID, "1235")
	assert.Equal(t, u.Email, "test@interledger.test")
}

func TestSearchTotpURL(t *testing.T) {
	cases := []struct {
		name        string
		credentials map[string]kratos.IdentityCredentials
		expectedURL string
		expectedErr error
	}{
		{
			name: "returns the TOTP URL if it is a string",
			credentials: map[string]kratos.IdentityCredentials{
				"totp": {
					Identifiers: []string{"totp"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_TOTP),
					Config: map[string]any{
						"totp_url": "totp://totp",
					},
				},
			},
			expectedURL: "totp://totp",
			expectedErr: nil,
		},
		{
			name:        "returns ErrTotpNotConfigured if TOTP URL is not present",
			credentials: map[string]kratos.IdentityCredentials{},
			expectedURL: "",
			expectedErr: user.ErrTotpNotConfigured,
		},
		{
			name: "returns ErrInvalidTotpConfig if TOTP URL is present but not a string",
			credentials: map[string]kratos.IdentityCredentials{
				"totp": {
					Identifiers: []string{"totp"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_TOTP),
					Config: map[string]any{
						"totp_url": []string{"totp://totp"},
					},
				},
			},
			expectedURL: "",
			expectedErr: user.ErrInvalidTotpConfig,
		},
		{
			name: "skips credentials with nil Type",
			credentials: map[string]kratos.IdentityCredentials{
				"totp": {
					Identifiers: []string{"totp"},
					Type:        nil,
					Config: map[string]any{
						"totp_url": "totp://totp",
					},
				},
			},
			expectedURL: "",
			expectedErr: user.ErrTotpNotConfigured,
		},
		{
			name: "ignores non-TOTP credentials even if totp_url is present",
			credentials: map[string]kratos.IdentityCredentials{
				"password": {
					Identifiers: []string{"user@example.com"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_PASSWORD),
					Config: map[string]any{
						"totp_url": "totp://should-not-be-used",
					},
				},
			},
			expectedURL: "",
			expectedErr: user.ErrTotpNotConfigured,
		},
		{
			name: "returns ErrTotpNotConfigured if TOTP credential exists but totp_url key is missing",
			credentials: map[string]kratos.IdentityCredentials{
				"totp": {
					Identifiers: []string{"totp"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_TOTP),
					Config:      map[string]any{
						// no "totp_url"
					},
				},
			},
			expectedURL: "",
			expectedErr: user.ErrTotpNotConfigured,
		},
		{
			name: "returns ErrTotpNotConfigured if TOTP credential exists but Config is nil",
			credentials: map[string]kratos.IdentityCredentials{
				"totp": {
					Identifiers: []string{"totp"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_TOTP),
					Config:      nil,
				},
			},
			expectedURL: "",
			expectedErr: user.ErrTotpNotConfigured,
		},
		{
			name: "returns the TOTP URL when multiple credential types are present",
			credentials: map[string]kratos.IdentityCredentials{
				"password": {
					Identifiers: []string{"user@example.com"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_PASSWORD),
					Config: map[string]any{
						"some": "thing",
					},
				},
				"totp": {
					Identifiers: []string{"totp"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_TOTP),
					Config: map[string]any{
						"totp_url": "totp://totp",
					},
				},
			},
			expectedURL: "totp://totp",
			expectedErr: nil,
		},
		{
			name: "returns empty string if totp_url is an empty string",
			credentials: map[string]kratos.IdentityCredentials{
				"totp": {
					Identifiers: []string{"totp"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_TOTP),
					Config: map[string]any{
						"totp_url": "",
					},
				},
			},
			expectedURL: "",
			expectedErr: nil, // empty string is still a valid string per implementation
		},
		{
			name: "returns ErrInvalidTotpConfig if totp_url is nil",
			credentials: map[string]kratos.IdentityCredentials{
				"totp": {
					Identifiers: []string{"totp"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_TOTP),
					Config: map[string]any{
						"totp_url": nil,
					},
				},
			},
			expectedURL: "",
			expectedErr: user.ErrInvalidTotpConfig,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			totpURL, err := searchTotpURL(tc.credentials)

			assert.Equal(t, tc.expectedURL, totpURL)
			if tc.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.expectedErr)
			}
		})
	}
}

func TestFindWalletIDByEmail(t *testing.T) {
	if os.Getenv("DB_URL") == "" {
		t.Setenv("DB_URL", "postgres://postgres:password@127.0.0.1:55432/%s?sslmode=disable")
	}

	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)

	walletID := uuid.NewString()
	identityID := uuid.NewString()

	// Seed a wallet and link it to an identity.
	_, err := dbc.ExecContext(ctx, `INSERT INTO wallets (id, name) VALUES ($1, $2)`, walletID, "email-search-test")
	require.NoError(t, err)
	_, err = dbc.ExecContext(ctx, `INSERT INTO user_wallets (user_id, wallet_id) VALUES ($1, $2)`, identityID, walletID)
	require.NoError(t, err)

	t.Run("resolves email to wallet ID", func(t *testing.T) {
		b := newEmailTestBackends(t, dbc, []string{identityID})
		got, err := FindWalletIDByEmail(ctx, b, "user@example.com")
		require.NoError(t, err)
		assert.Equal(t, walletID, got)
	})

	t.Run("returns empty string when Kratos has no matching identity", func(t *testing.T) {
		b := newEmailTestBackends(t, dbc, nil)
		got, err := FindWalletIDByEmail(ctx, b, "unknown@example.com")
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("returns empty string when identity has no wallet", func(t *testing.T) {
		b := newEmailTestBackends(t, dbc, []string{uuid.NewString()})
		got, err := FindWalletIDByEmail(ctx, b, "orphan@example.com")
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

// newEmailTestBackends wires up a Backends whose Kratos() returns an httptest
// server that responds to GET /admin/identities with a list of stubs — one
// minimal Identity per supplied ID.
func newEmailTestBackends(t *testing.T, dbc *sqlx.DB, identityIDs []string) Backends {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		type identityStub struct {
			ID string `json:"id"`
		}
		stubs := make([]identityStub, len(identityIDs))
		for i, id := range identityIDs {
			stubs[i] = identityStub{ID: id}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stubs); err != nil {
			http.Error(w, fmt.Sprintf("encode: %v", err), http.StatusInternalServerError)
		}
	}))
	t.Cleanup(srv.Close)

	cfg := kratos.NewConfiguration()
	cfg.Servers = kratos.ServerConfigurations{
		{URL: srv.URL, Description: "Public Kratos"},
		{URL: srv.URL, Description: "Admin Kratos"},
	}

	return &emailTestBackends{
		db:  dbc,
		kr:  kratos.NewAPIClient(cfg),
		val: validator.New(),
	}
}

type emailTestBackends struct {
	db  *sqlx.DB
	kr  *kratos.APIClient
	val *validator.Validate
}

func (b *emailTestBackends) DB() *sqlx.DB                   { return b.db }
func (b *emailTestBackends) Kratos() *kratos.APIClient      { return b.kr }
func (b *emailTestBackends) Validator() *validator.Validate { return b.val }
func (b *emailTestBackends) Analytics() analytics.Client    { return analytics_client.New(nil, "") }
func (b *emailTestBackends) Keys() keys.Client              { return nil }

// Maybe it might sense to move to Go 1.26, since the `new` function is now
// accepting expressions and it returns a pointer to the result.
// https://go.dev/doc/go1.26#language
func ptr[T any](v T) *T {
	return &v
}
