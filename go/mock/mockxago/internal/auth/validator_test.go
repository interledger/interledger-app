package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/mock/mockxago/internal/models"
	"gitlab.com/fynbos/mock/mockxago/internal/storage"
)

func TestValidateToken_MissingHeader(t *testing.T) {
	validator := NewValidator(storage.NewMemoryStorage())

	_, err := validator.ValidateToken(context.Background(), "")
	assert.Equal(t, ErrMissingToken, err)
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	validator := NewValidator(storage.NewMemoryStorage())

	_, err := validator.ValidateToken(context.Background(), "Token abc")
	assert.Equal(t, ErrInvalidFormat, err)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	validator := NewValidator(storage.NewMemoryStorage())

	_, err := validator.ValidateToken(context.Background(), "Bearer missing")
	assert.Equal(t, ErrInvalidToken, err)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	store := storage.NewMemoryStorage()
	validator := NewValidator(store)

	token := &models.AccessToken{
		ID:        "token-id",
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	assert.NoError(t, store.SaveAccessToken(context.Background(), token))

	_, err := validator.ValidateToken(context.Background(), "Bearer expired-token")
	assert.Equal(t, ErrTokenExpired, err)
}

func TestValidateToken_ValidToken(t *testing.T) {
	store := storage.NewMemoryStorage()
	validator := NewValidator(store)

	token := &models.AccessToken{
		ID:        "token-id",
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	assert.NoError(t, store.SaveAccessToken(context.Background(), token))

	validated, err := validator.ValidateToken(context.Background(), "Bearer valid-token")
	assert.NoError(t, err)
	assert.Equal(t, token.Token, validated.Token)
}

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name              string
		publicKey         string
		secret            string
		expectedPublicKey string
		expectedSecret    string
		expectedError     error
	}{
		{
			name:              "missing public key",
			publicKey:         "",
			secret:            "secret",
			expectedPublicKey: "pub",
			expectedSecret:    "secret",
			expectedError:     ErrMissingCredentials,
		},
		{
			name:              "missing secret",
			publicKey:         "pub",
			secret:            "",
			expectedPublicKey: "pub",
			expectedSecret:    "secret",
			expectedError:     ErrMissingCredentials,
		},
		{
			name:              "invalid credentials",
			publicKey:         "pub",
			secret:            "wrong",
			expectedPublicKey: "pub",
			expectedSecret:    "secret",
			expectedError:     ErrInvalidCredentials,
		},
		{
			name:              "valid credentials",
			publicKey:         "pub",
			secret:            "secret",
			expectedPublicKey: "pub",
			expectedSecret:    "secret",
			expectedError:     nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateCredentials(test.publicKey, test.secret, test.expectedPublicKey, test.expectedSecret)
			assert.Equal(t, test.expectedError, err)
		})
	}
}
