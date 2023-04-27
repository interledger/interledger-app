package client

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/twitter/external"
	external_client "gitlab.com/fynbos/backend/twitter/external/client"
	"gitlab.com/fynbos/backend/twitter/ops"
)

var _ twitter.Client = &Client{}

type (
	Backends interface {
		DB() *sqlx.DB
	}

	opsBackends struct {
		b        Backends
		external external.Client
	}

	Client struct {
		b            ops.Backends
		clientID     string
		redirectURL  string
		authEndpoint string
	}

	NewClientArgs struct {
		ClientID      string
		ClientSecret  string
		RedirectURL   string
		AuthEndpoint  string
		TokenEndpoint string
	}
)

func (ob *opsBackends) External() external.Client {
	return ob.external
}

func (ob *opsBackends) DB() *sqlx.DB {
	return ob.b.DB()
}

func New(b Backends, args *NewClientArgs) *Client {
	externalClient := external_client.New(&external_client.NewClientArgs{
		ClientID:      args.ClientID,
		RedirectURL:   args.RedirectURL,
		AuthEndpoint:  args.AuthEndpoint,
		TokenEndpoint: args.TokenEndpoint,
		ClientSecret:  args.ClientSecret,
	})

	return &Client{
		b: &opsBackends{
			b:        b,
			external: externalClient,
		},
		clientID:     args.ClientID,
		redirectURL:  args.RedirectURL,
		authEndpoint: args.AuthEndpoint,
	}
}

func (c *Client) CreateAuthURL(ctx context.Context, args *twitter.CreateAuthURLArgs) (*twitter.Authorization, error) {
	return ops.CreateAuthURL(ctx, c.b, &ops.CreateAuthURLArgs{
		ClientID:     c.clientID,
		RedirectURL:  c.redirectURL,
		AuthEndpoint: c.authEndpoint,
		Scopes:       args.Scopes,
		WalletID:     args.WalletID,
	})
}

func (c *Client) CreateAccessToken(ctx context.Context, args *twitter.CreateAccessTokenArgs) (*twitter.TwitterAccessToken, error) {
	return ops.CreateAccessToken(ctx, c.b, &twitter.CreateAccessTokenArgs{
		AuthCode: args.AuthCode,
		State:    args.State,
	})
}
