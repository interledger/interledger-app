package twitter

import (
	"context"
)

type Client interface {
	CreateAuthURL(ctx context.Context, args *CreateAuthURLArgs) (*Authorization, error)
	CreateToken(ctx context.Context, args *CreateTokenArgs) (*Token, error)
	GetTokensByUserID(ctx context.Context, args *GetTokensByUserIdArgs) ([]Token, error)
	PostTweet(ctx context.Context, token *Token, text string) (*Tweet, error)
}
