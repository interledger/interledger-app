package external

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/providers/xago/external/domain/dto"
)

type tokenProvider struct {
	dbc *sqlx.DB

	identityURL string
	publicKey   string
	secret      string
	policyID    string

	mu sync.RWMutex

	cachedAccess dto.Access

	loginHTTP *http.Client
}

func newTokenProvider(dbc *sqlx.DB, cfg Config, transport http.RoundTripper) *tokenProvider {
	return &tokenProvider{
		dbc:         dbc,
		identityURL: cfg.IdentityBaseURL,
		publicKey:   cfg.PublicKey,
		secret:      cfg.Secret,
		policyID:    cfg.PolicyID,
		loginHTTP:   &http.Client{Transport: transport},
	}
}

// getCached returns the current token if it hasn't expired, otherwise empty string.
// No network call is made.
func (tokenProvider *tokenProvider) getCached() string {
	tokenProvider.mu.RLock()
	defer tokenProvider.mu.RUnlock()
	if !tokenProvider.cachedAccess.IsExpired() {
		return tokenProvider.cachedAccess.Token
	}
	return ""
}

// get returns a valid token, refreshing from the identity service if needed.
// forceRefresh skips the expiry check and always fetches a new token.
func (tokenProvider *tokenProvider) get(ctx context.Context, forceRefresh bool) (string, error) {
	if !forceRefresh {
		tokenProvider.mu.RLock()
		token := tokenProvider.cachedAccess
		tokenProvider.mu.RUnlock()
		if !token.IsExpired() {
			return token.Token, nil
		}
	}
	tokenProvider.mu.Lock()
	defer tokenProvider.mu.Unlock()
	if !tokenProvider.cachedAccess.IsExpired() && !forceRefresh {
		return tokenProvider.cachedAccess.Token, nil
	}
	if err := tokenProvider.refresh(ctx); err != nil {
		return "", err
	}
	return tokenProvider.cachedAccess.Token, nil
}

func (tokenProvider *tokenProvider) refresh(ctx context.Context) error {
	tx, err := tokenProvider.dbc.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var access dto.Access
	err = tx.GetContext(ctx, &access, "SELECT token, expires_at FROM xago_access_token WHERE id=$1 FOR UPDATE", accessTokenID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Another process already refreshed; adopt its token.
	if !access.IsExpired() && access.IsDifferent(tokenProvider.cachedAccess) {
		tokenProvider.cachedAccess = access
		return tx.Commit()
	}

	reqURL := tokenProvider.identityURL + "login"

	fmt.Println("reqURL", reqURL)

	login := dto.NewLoginRequest(tokenProvider.policyID, tokenProvider.publicKey, tokenProvider.secret)
	body, err := login.Body()
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, body)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := tokenProvider.loginHTTP.Do(request)
	if err != nil {
		return err
	}

	loginResponse, err := consumeResponse[dto.LoginResponse](response, http.StatusOK)
	if err != nil {
		return err
	}

	tokenProvider.cachedAccess = dto.NewAccess(loginResponse.Token)

	if _, err = tx.ExecContext(ctx, "INSERT INTO xago_access_token (id, token, expires_at) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET token = excluded.token, expires_at = excluded.expires_at", accessTokenID, tokenProvider.cachedAccess.Token, tokenProvider.cachedAccess.ExpiresAt); err != nil {
		return err
	}

	return tx.Commit()
}
