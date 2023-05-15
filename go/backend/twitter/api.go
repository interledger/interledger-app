package twitter

import (
	"context"
)

type Client interface {
	CreateAuthURL(ctx context.Context, args *CreateAuthURLArgs) (string, error)
	CreateConnection(ctx context.Context, args *CreateConnectionArgs) (*Connection, error)
	GetWalletConnections(ctx context.Context, id string) ([]Connection, error)
	PostTweet(ctx context.Context, id string, text string) (*Tweet, error)

	// TODO check connection status/refresh
}
