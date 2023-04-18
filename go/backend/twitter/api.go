package twitter

import (
	"context"
	"gitlab.com/fynbos/backend/twitter/ops"
)

type Client interface {
	CreateAuthURL(ctx context.Context, args *CreateAuthURLArgs) (*ops.Authorization, error)
	CreateAccessToken(ctx context.Context, args *ops.CreateAccessTokenArgs) (*ops.TwitterAccessToken, error)
}
