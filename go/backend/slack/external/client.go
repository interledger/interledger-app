package external

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/env"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type client struct {
	conf     *oauth2.Config
	botConf  *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func New(cfg Config) (Client, error) {
	provider, err := oidc.NewProvider(context.Background(), "https://slack.com")
	if err != nil {
		return nil, err
	}

	clientID := cfg.ClientID
	if clientID == "" {
		clientID = "2317468772181.5841878200565"
	}
	clientSecret := cfg.ClientSecret
	if clientSecret == "" {
		clientSecret = "e0705d863bc2726505cd175b65cc12d9"
	}
	redirectURL := cfg.RedirectURL
	if redirectURL == "" {
		redirectURL = env.GetUrl() + "/connect/slack"
	}

	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	botConf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     oauth2.Endpoint{TokenURL: "https://slack.com/api/oauth.v2.access", AuthStyle: oauth2.AuthStyleInParams},
		RedirectURL:  env.GetUrl() + "/webhooks/slack/bot/install",
	}

	return client{
		conf:     conf,
		botConf:  botConf,
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

func (c client) CreateBotToken(ctx context.Context, authCode string) (*oauth2.Token, error) {
	oauth2Token, err := c.botConf.Exchange(ctx, authCode)
	if err != nil {
		return nil, err
	}

	return oauth2Token, err
}
