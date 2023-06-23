package client

import (
	"context"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/linkedin"
	"gitlab.com/fynbos/backend/linkedin/external"
	external_client "gitlab.com/fynbos/backend/linkedin/external/client"
	"gitlab.com/fynbos/backend/linkedin/ops"
	temporal "go.temporal.io/sdk/client"
)

var _ linkedin.Client = &Client{}

type (
	Backends interface {
		DB() *sqlx.DB
		Temporal() temporal.Client
	}

	Client struct {
		b            ops.Backends
		clientID     string
		redirectURL  string
		authEndpoint string
	}

	opsBackends struct {
		b        Backends
		external external.Client
	}

	NewClientArgs struct {
		ClientID      string
		ClientSecret  string
		RedirectURL   string
		AuthEndpoint  string
		TokenEndpoint string
	}
)

func (ob *opsBackends) DB() *sqlx.DB {
	return ob.b.DB()
}

func (ob *opsBackends) External() external.Client {
	return ob.external
}

func (ob *opsBackends) Temporal() temporal.Client {
	return ob.b.Temporal()
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

func (c *Client) CreateAuthURL(ctx context.Context, args *linkedin.CreateAuthURLArgs) (string, error) {
	return ops.CreateAuthURL(ctx, c.b, &ops.CreateAuthURLArgs{
		ClientID:     c.clientID,
		RedirectURL:  c.redirectURL,
		AuthEndpoint: c.authEndpoint,
		State:        args.State,
		Scopes:       args.Scopes,
		WalletID:     args.WalletID,
	})
}

func (c *Client) CreateConnection(ctx context.Context, args *linkedin.CreateConnectionArgs) (*linkedin.Connection, error) {
	return ops.CreateConnection(ctx, c.b, &linkedin.CreateConnectionArgs{
		AuthCode: args.AuthCode,
		State:    args.State,
	})
}

func (c *Client) GetWalletConnections(ctx context.Context, id string) ([]linkedin.Connection, error) {
	return ops.GetWalletConnections(ctx, c.b, id)
}

func (c *Client) Post(ctx context.Context, connectionID string, text string) (string, error) {
	return ops.Post(ctx, c.b, connectionID, text)
}

func (c *Client) PublishPublicProof(ctx context.Context, identityID string) error {
	return ops.PublishPublicProof(ctx, c.b, identityID)
}

func (c *Client) GetPost(ctx context.Context, url string) (*linkedin.Post, error) {
	return ops.GetPost(ctx, c.b, url)
}
