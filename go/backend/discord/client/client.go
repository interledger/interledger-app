package client

import (
	"context"

	temporal "go.temporal.io/sdk/client"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/discord"
	"gitlab.com/fynbos/backend/discord/external"
	external_client "gitlab.com/fynbos/backend/discord/external/client"
	"gitlab.com/fynbos/backend/discord/ops"
)

var _ discord.Client = &Client{}

type (
	Backends interface {
		DB() *sqlx.DB
		Temporal() temporal.Client
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
		BearerToken   string
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
		BearerToken:   args.BearerToken,
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

func (c *Client) CreateAuthURL(ctx context.Context, args discord.CreateAuthURLArgs) (string, error) {
	return ops.CreateAuthURL(ctx, c.b, ops.CreateAuthURLArgs{
		ClientID:     c.clientID,
		RedirectURL:  c.redirectURL,
		AuthEndpoint: c.authEndpoint,
		Scopes:       args.Scopes,
		WalletID:     args.WalletID,
		State:        uuid.NewString(),
	})
}

func (c *Client) CreateConnection(ctx context.Context, args discord.CreateConnectionArgs) (*discord.Connection, error) {
	return ops.CreateConnection(ctx, c.b, discord.CreateConnectionArgs{
		AuthCode: args.AuthCode,
		State:    args.State,
	})
}

func (c *Client) GetWalletConnections(ctx context.Context, id string) ([]discord.Connection, error) {
	return ops.GetWalletConnections(ctx, c.b, id)
}
