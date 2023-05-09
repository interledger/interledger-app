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

	"github.com/lib/pq"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/twitter/external"
	"golang.org/x/oauth2"
)

func CreateAuthURL(ctx context.Context, b Backends, args *CreateAuthURLArgs) (*twitter.Authorization, error) {
	oauthConfig := &oauth2.Config{
		ClientID:    args.ClientID,
		RedirectURL: args.RedirectURL,
		Scopes:      args.Scopes,
		Endpoint:    oauth2.Endpoint{AuthURL: args.AuthEndpoint},
	}

	state, err := randomBytesInBase64URL(24)
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	codeVerifier, err := randomBytesInBase64URL(32)
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}
	sha := sha256.New()
	sha.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(sha.Sum(nil))

	var authId string

	err = b.DB().GetContext(ctx, &authId, "INSERT INTO twitter_authorizations (client_id, state, code_verifier, scopes, wallet_id, redirect_url) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", args.ClientID, state, codeVerifier, pq.Array(args.Scopes), args.WalletID, args.RedirectURL)
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	url := oauthConfig.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	return &twitter.Authorization{
		URL:   url,
		State: state,
	}, nil
}

func CreateToken(ctx context.Context, b Backends, args *twitter.CreateTokenArgs) (*twitter.Token, error) {
	var authorization TwitterAuth

	err := b.DB().GetContext(ctx, &authorization, "SELECT * FROM twitter_authorizations WHERE state = $1", args.State)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w no authorization found for state %q", twitter.ErrNotFound, args.State)
		}
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}
	if args.AuthCode == "" {
		return nil, fmt.Errorf("%w couldn't find an auth code for this grant", twitter.ErrInternal)
	}

	token, err := b.External().CreateToken(ctx, &external.CreateTokenArgs{
		AuthCode:     args.AuthCode,
		CodeVerifier: authorization.CodeVerifier,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	user, err := b.External().GetAuthorizedUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	var dbToken twitter.Token

	err = b.DB().GetContext(ctx, &dbToken,
		"INSERT INTO twitter_access_tokens (client_id, wallet_id, access_token, refresh_token, token_type, expiry, scopes, user_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *",
		authorization.ClientID, authorization.WalletID, token.AccessToken, token.RefreshToken, token.TokenType, token.Expiry, pq.Array(authorization.Scopes), user.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	return &dbToken, nil
}

func GetTokensByUserID(ctx context.Context, b Backends, args *twitter.GetTokensByUserIdArgs) ([]twitter.Token, error) {
	var tokens []twitter.Token

	err := b.DB().SelectContext(ctx, &tokens, "SELECT * FROM twitter_access_tokens WHERE scopes @> $1 AND user_id = $2 AND wallet_id = $3", pq.Array(args.Scopes), args.UserID, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	return tokens, nil
}

func PostTweet(ctx context.Context, b Backends, token *twitter.Token, text string) (*twitter.Tweet, error) {
	oauthToken := oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}

	tweet, err := b.External().PostTweet(ctx, &oauthToken, text)
	if err != nil {
		return nil, fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	return tweet, nil
}

func randomBytesInBase64URL(count int) (string, error) {
	buf := make([]byte, count)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", fmt.Errorf("%w %s", twitter.ErrInternal, err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
