package client

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/twitter/external"
	"golang.org/x/oauth2"
)

var _ external.Client = &client{}

type (
	client struct {
		oauthConfig *oauth2.Config
	}

	NewClientArgs struct {
		ClientID      string
		ClientSecret  string
		AuthEndpoint  string
		TokenEndpoint string
		RedirectURL   string
	}
)

func New(args *NewClientArgs) *client {
	oauthConfig := &oauth2.Config{
		ClientID:     args.ClientID,
		ClientSecret: args.ClientSecret,
		RedirectURL:  args.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:   args.AuthEndpoint,
			TokenURL:  args.TokenEndpoint,
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}

	return &client{
		oauthConfig: oauthConfig,
	}
}

func (c *client) CreateAccessToken(ctx context.Context, args external.CreateAccessTokenArgs) (*oauth2.Token, error) {
	token, err := c.oauthConfig.Exchange(ctx, args.AuthCode, oauth2.SetAuthURLParam("code_verifier", args.CodeVerifier))
	if err != nil {
		return nil, fmt.Errorf("could not exchange auth code for token: %v", err)
	}

	return token, nil
}
