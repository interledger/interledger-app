package twitter

import (
	"context"
)

type Client interface {
	CreateAuthURL(ctx context.Context, args *CreateAuthURLArgs) (*Authorization, error)
	CreateAccessToken(ctx context.Context, args *CreateAccessTokenArgs) (*TwitterAccessToken, error)
}
