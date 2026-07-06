package external

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/twitter"
	"golang.org/x/oauth2"
)

type Client interface {
	CreateToken(ctx context.Context, args *CreateTokenArgs) (*oauth2.Token, error)
	GetAuthorizedUser(ctx context.Context, token *oauth2.Token) (*twitter.TwitterUser, error)
	PostTweet(ctx context.Context, token *oauth2.Token, text string) (*twitter.Tweet, error)
	GetTweet(ctx context.Context, id string) (*twitter.Tweet, error)
}
