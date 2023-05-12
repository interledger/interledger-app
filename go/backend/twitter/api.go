package twitter

import (
	"context"
)

type Client interface {
	CreateAuthURL(ctx context.Context, args *CreateAuthURLArgs) (string, error)
	CreateConnection(ctx context.Context, args *CreateConnectionArgs) (*Connection, error)
	GetTokensByUserID(ctx context.Context, args *GetTokensByUserIdArgs) ([]Connection, error)
	PostTweet(ctx context.Context, token *Connection, text string) (*Tweet, error)
}
