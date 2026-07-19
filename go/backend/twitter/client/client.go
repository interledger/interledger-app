package client

import (
	"context"
	temporal "go.temporal.io/sdk/client"

	"github.com/interledger/interledger-app/go/backend/twitter"
	"github.com/interledger/interledger-app/go/backend/twitter/external"
	external_client "github.com/interledger/interledger-app/go/backend/twitter/external/client"
	"github.com/interledger/interledger-app/go/backend/twitter/ops"
	"github.com/jmoiron/sqlx"
)

var _ twitter.Client = &Client{}

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

func (ob *opsBackends) Temporal() temporal.Client {
	return ob.b.Temporal()
}

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

func (c *Client) CreateConnection(ctx context.Context, args *twitter.CreateConnectionArgs) (*twitter.Connection, error) {
	return ops.CreateConnection(ctx, c.b, &twitter.CreateConnectionArgs{
		AuthCode: args.AuthCode,
		State:    args.State,
	})
}

func (c *Client) GetWalletConnections(ctx context.Context, id string) ([]twitter.Connection, error) {
	return ops.GetWalletConnections(ctx, c.b, id)
}

func (c *Client) PostTweet(ctx context.Context, id string, text string) (*twitter.Tweet, error) {
	return ops.PostTweet(ctx, c.b, id, text)
}

func (c *Client) PublishTweetProof(ctx context.Context, identityID string) (string, error) {
	return ops.PublishTweetProof(ctx, c.b, identityID)
}

func (c *Client) GetTweet(ctx context.Context, id string) (*twitter.Tweet, error) {
	return ops.GetTweet(ctx, c.b, id)
}
