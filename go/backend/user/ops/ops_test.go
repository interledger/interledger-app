package ops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kratos "github.com/ory/kratos-client-go"
	"gitlab.com/fynbos/backend/user"
)

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
		userID      string
		expected    string
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
			userID:   "1234",
			expected: "totp://totp",
		},
		{
			name:        "returns empty string if TOTP URL is not present",
			credentials: map[string]kratos.IdentityCredentials{},
			userID:      "1234",
			expected:    "",
		},
		{
			name: "returns empty string if TOTP URL is present but not a string",
			credentials: map[string]kratos.IdentityCredentials{
				"totp": {
					Identifiers: []string{"totp"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_TOTP),
					Config: map[string]any{
						"totp_url": []string{"totp://totp"},
					},
				},
			},
			userID:   "1234",
			expected: "",
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
			userID:   "1234",
			expected: "",
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
			userID:   "1234",
			expected: "",
		},
		{
			name: "returns empty string if TOTP credential exists but totp_url key is missing",
			credentials: map[string]kratos.IdentityCredentials{
				"totp": {
					Identifiers: []string{"totp"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_TOTP),
					Config:      map[string]any{
						// no "totp_url"
					},
				},
			},
			userID:   "1234",
			expected: "",
		},
		{
			name: "returns empty string if TOTP credential exists but Config is nil",
			credentials: map[string]kratos.IdentityCredentials{
				"totp": {
					Identifiers: []string{"totp"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_TOTP),
					Config:      nil,
				},
			},
			userID:   "1234",
			expected: "",
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
			userID:   "1234",
			expected: "totp://totp",
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
			userID:   "1234",
			expected: "",
		},
		{
			name: "returns empty string if totp_url is nil",
			credentials: map[string]kratos.IdentityCredentials{
				"totp": {
					Identifiers: []string{"totp"},
					Type:        ptr(kratos.IDENTITYCREDENTIALSTYPE_TOTP),
					Config: map[string]any{
						"totp_url": nil,
					},
				},
			},
			userID:   "1234",
			expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			totpURL := searchTotpURL(tc.credentials, tc.userID)
			assert.Equal(t, tc.expected, totpURL)
		})
	}
}

// Maybe it might sense to move to Go 1.26, since the `new` function is now
// accepting expressions and it returns a pointer to the result.
// https://go.dev/doc/go1.26#language
func ptr[T any](v T) *T {
	return &v
}
