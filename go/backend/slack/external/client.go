package external

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/env"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Config struct {
	ClientID       string
	ClientSecret   string
	RedirectURL    string
	BotRedirectURL string
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
	clientSecret := cfg.ClientSecret
	redirectURL := cfg.RedirectURL
	botRedirectURL := cfg.BotRedirectURL

	slackConfigured := clientID != "" || clientSecret != "" || redirectURL != "" || botRedirectURL != ""
	if slackConfigured && (clientID == "" || clientSecret == "") {
		return nil, fmt.Errorf("SLACK_CLIENT_ID and SLACK_CLIENT_SECRET must both be set when Slack OAuth is configured")
	}

	if redirectURL == "" {
		redirectURL = env.GetUrl() + "/connect/slack"
	}
	if botRedirectURL == "" {
		botRedirectURL = env.GetUrl() + "/webhooks/slack/bot/install"
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
		RedirectURL:  botRedirectURL,
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
