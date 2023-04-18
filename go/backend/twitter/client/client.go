package client

import (
	"context"

	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/twitter/ops"
	"golang.org/x/oauth2"
)

var _ twitter.Client = &Client{}

type (
	Client struct {
		b           ops.Backends
		oauthConfig *oauth2.Config
	}

	NewClientArgs struct {
		ClientID     string
		RedirectURL  string
		AuthEndpoint string
	}
)

func New(args NewClientArgs, b ops.Backends) (*Client, error) {
	oauthConfig := &oauth2.Config{
		ClientID:    args.ClientID,
		RedirectURL: args.RedirectURL,
		Endpoint:    oauth2.Endpoint{AuthURL: args.AuthEndpoint},
	}

	return &Client{
		b:           b,
		oauthConfig: oauthConfig,
	}, nil
}

func (c *Client) CreateAuthURL(ctx context.Context, b ops.Backends) (string, error) {
	return ops.CreateAuthURL(ctx, b)
}
