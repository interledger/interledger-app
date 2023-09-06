package external

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/fynbos/env"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type client struct {
	conf     *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func getEnvDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	return val
}

func New() (Client, error) {
	provider, err := oidc.NewProvider(context.Background(), "https://slack.com")
	if err != nil {
		return nil, err
	}

	oidcConfig := &oidc.Config{
		ClientID: getEnvDefault("SLACK_CLIENT_ID", "2317468772181.5841878200565"),
	}
	verifier := provider.Verifier(oidcConfig)

	conf := &oauth2.Config{
		ClientID:     getEnvDefault("SLACK_CLIENT_ID", "2317468772181.5841878200565"),
		ClientSecret: getEnvDefault("SLACK_CLIENT_SECRET", "e0705d863bc2726505cd175b65cc12d9"),
		Endpoint:     provider.Endpoint(),
		RedirectURL:  getEnvDefault("SLACK_REDIRECT_URL", env.GetUrl()+"/connect/slack"),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return client{
		conf:     conf,
		verifier: verifier,
	}, nil
}

func (c client) GetConfig() *oauth2.Config {
	return c.conf
}

func (c client) CreateUserToken(ctx context.Context, authCode string) (*oauth2.Token, *User, error) {
	oauth2Token, err := c.conf.Exchange(ctx, authCode)
	if err != nil {
		return nil, nil, err
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, nil, fmt.Errorf("no id_token filed in outh2 token")
	}
	idToken, err := c.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, nil, err
	}

	var user User
	if err = idToken.Claims(&user); err != nil {
		return nil, nil, err
	}

	return oauth2Token, &user, nil
}
