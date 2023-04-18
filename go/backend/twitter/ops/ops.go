package ops

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/oauth2"
)

func CreateAuthURL(ctx context.Context, b Backends, args *CreateAuthURLArgs) (string, error) {
	oauthConfig := &oauth2.Config{
		ClientID:    args.ClientID,
		RedirectURL: args.RedirectURL,
		Scopes:      args.Scopes,
		Endpoint:    oauth2.Endpoint{AuthURL: args.AuthEndpoint},
	}

	state, err := randomBytesInBase64URL(24)
	if err != nil {
		return "", fmt.Errorf("could not generate random state: %v", err)
	}

	codeVerifier, err := randomBytesInBase64URL(32)
	if err != nil {
		return "", fmt.Errorf("could not generate code verifier: %v", err)
	}
	sha := sha256.New()
	sha.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(sha.Sum(nil))

	var stateId string

	err = b.DB().GetContext(ctx, &stateId, "INSERT INTO twitter_auth_states (state, code_verifier) VALUES ($1, $2) RETURNING id", state, codeVerifier)
	if err != nil {
		return "", fmt.Errorf("could not save state: %v", err)
	}

	return oauthConfig.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	), nil
}

func randomBytesInBase64URL(count int) (string, error) {
	buf := make([]byte, count)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", fmt.Errorf("could not generate %d random bytes: %v", count, err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
