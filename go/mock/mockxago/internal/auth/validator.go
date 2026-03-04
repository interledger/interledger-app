package auth

import (
	"context"
	"strings"

	"gitlab.com/fynbos/mock/mockxago/internal/models"
	"gitlab.com/fynbos/mock/mockxago/internal/storage"
)

// Validator validates authentication tokens
type Validator struct {
	store storage.Storage
}

// NewValidator creates a new token validator
func NewValidator(store storage.Storage) *Validator {
	return &Validator{store: store}
}

// ValidateToken validates a bearer token
func (v *Validator) ValidateToken(ctx context.Context, authHeader string) (*models.AccessToken, error) {
	if authHeader == "" {
		return nil, ErrMissingToken
	}

	// Parse "Bearer <token>" format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, ErrInvalidFormat
	}

	tokenValue := parts[1]
	token, err := v.store.GetAccessToken(ctx, tokenValue)
	if err != nil {
		if err == storage.ErrTokenNotFound {
			return nil, ErrInvalidToken
		}
		if err == storage.ErrTokenExpired {
			return nil, ErrTokenExpired
		}
		return nil, err
	}

	return token, nil
}

// ValidateCredentials validates API credentials
func ValidateCredentials(publicKey, secret, expectedPublicKey, expectedSecret string) error {
	if publicKey == "" || secret == "" {
		return ErrMissingCredentials
	}

	if publicKey != expectedPublicKey || secret != expectedSecret {
		return ErrInvalidCredentials
	}

	return nil
}
