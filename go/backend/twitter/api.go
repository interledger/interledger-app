package twitter

import (
	"context"
	"gitlab.com/fynbos/backend/identities"
)

type Client interface {
	CreateAuthURL(ctx context.Context, args *CreateAuthURLArgs) (string, error)
	CreateConnection(ctx context.Context, args *CreateConnectionArgs) (*Connection, error)
	GetWalletConnections(ctx context.Context, id string) ([]Connection, error)
	PostTweet(ctx context.Context, id string, text string) (*Tweet, error)
	PublishTweetProof(ctx context.Context, identity *identities.Identity, connection *Connection) (string, error)

	// TODO check connection status/refresh
}
