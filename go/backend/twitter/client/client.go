package client

import (
	"context"

	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/twitter/ops"
)

var _ twitter.Client = &Client{}

type (
	Client struct {
		b            ops.Backends
		clientID     string
		redirectURL  string
		authEndpoint string
	}

	NewClientArgs struct {
		ClientID     string
		RedirectURL  string
		AuthEndpoint string
	}
)

func New(args NewClientArgs, b ops.Backends) (*Client, error) {
	return &Client{
		b:            b,
		clientID:     args.ClientID,
		redirectURL:  args.RedirectURL,
		authEndpoint: args.AuthEndpoint,
	}, nil
}

func (c *Client) CreateAuthURL(ctx context.Context, b ops.Backends, scopes []string) (string, error) {
	return ops.CreateAuthURL(ctx, b, &ops.CreateAuthURLArgs{
		ClientID:     c.clientID,
		RedirectURL:  c.redirectURL,
		AuthEndpoint: c.authEndpoint,
		Scopes:       scopes,
	})
}
