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

func (c *Client) CreateAuthURL(ctx context.Context, args *twitter.CreateAuthURLArgs) (string, error) {
	return ops.CreateAuthURL(ctx, c.b, &ops.CreateAuthURLArgs{
		ClientID:     c.clientID,
		RedirectURL:  c.redirectURL,
		AuthEndpoint: c.authEndpoint,
		Scopes:       args.Scopes,
		WalletID:     args.WalletID,
		State:        args.State,
	})
}

func (c *Client) CreateToken(ctx context.Context, args *twitter.CreateTokenArgs) (*twitter.Token, error) {
	return ops.CreateToken(ctx, c.b, &twitter.CreateTokenArgs{
		AuthCode: args.AuthCode,
		State:    args.State,
	})
}

func (c *Client) GetTokensByUserID(ctx context.Context, args *twitter.GetTokensByUserIdArgs) ([]twitter.Token, error) {
	return ops.GetTokensByUserID(ctx, c.b, &twitter.GetTokensByUserIdArgs{
		Scopes:   args.Scopes,
		WalletID: args.WalletID,
		UserID:   args.UserID,
	})
}

func (c *Client) PostTweet(ctx context.Context, token *twitter.Token, text string) (*twitter.Tweet, error) {
	return ops.PostTweet(ctx, c.b, token, text)
}
