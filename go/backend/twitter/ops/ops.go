package ops

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"gitlab.com/fynbos/backend/twitter/external"
	"golang.org/x/oauth2"
)

func CreateAuthURL(ctx context.Context, b Backends, args *CreateAuthURLArgs) (*Authorization, error) {
	oauthConfig := &oauth2.Config{
		ClientID:    args.ClientID,
		RedirectURL: args.RedirectURL,
		Scopes:      args.Scopes,
		Endpoint:    oauth2.Endpoint{AuthURL: args.AuthEndpoint},
	}

	state, err := randomBytesInBase64URL(24)
	if err != nil {
		return nil, fmt.Errorf("could not generate random state: %v", err)
	}

	codeVerifier, err := randomBytesInBase64URL(32)
	if err != nil {
		return nil, fmt.Errorf("could not generate code verifier: %v", err)
	}
	sha := sha256.New()
	sha.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(sha.Sum(nil))

	var authId string

	err = b.DB().GetContext(ctx, &authId, "INSERT INTO twitter_authorizations (client_id, state, code_verifier, wallet_id, redirect_url) VALUES ($1, $2, $3, $4, $5) RETURNING id", args.ClientID, state, codeVerifier, args.WalletID, args.RedirectURL)
	if err != nil {
		return nil, fmt.Errorf("could not save state: %v", err)
	}

	url := oauthConfig.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	return &Authorization{
		URL:   url,
		State: state,
	}, nil
}

func CreateAccessToken(ctx context.Context, b Backends, args *CreateAccessTokenArgs) (*TwitterAccessToken, error) {
	var authorization TwitterAuth

	err := b.DB().GetContext(ctx, &authorization, "SELECT * FROM twitter_authorizations WHERE state = $1", args.State)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no authorization found for state %q", args.State)
		}
		return nil, fmt.Errorf("could not get auth token: %v", err)
	}
	if args.AuthCode == "" {
		return nil, fmt.Errorf("no authorization code found")
	}

	accessToken, err := b.External().CreateAccessToken(ctx, external.CreateAccessTokenArgs{
		AuthCode:     args.AuthCode,
		CodeVerifier: authorization.CodeVerifier,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create access token: %v", err)
	}

	var token TwitterAccessToken

	err = b.DB().GetContext(ctx, &token,
		"INSERT INTO twitter_access_tokens (client_id, wallet_id, access_token, refresh_token, token_type, expiry) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *",
		authorization.ClientID, authorization.WalletID, accessToken.AccessToken, accessToken.RefreshToken, accessToken.TokenType, accessToken.Expiry)
	if err != nil {
		return nil, fmt.Errorf("could not save access token: %v", err)
	}

	return &token, nil
}

func randomBytesInBase64URL(count int) (string, error) {
	buf := make([]byte, count)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", fmt.Errorf("could not generate %d random bytes: %v", count, err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
