package external

import (
	"context"

	"gitlab.com/fynbos/backend/twitter"
	"golang.org/x/oauth2"
)

type Client interface {
	CreateToken(ctx context.Context, args *CreateTokenArgs) (*oauth2.Token, error)
	GetAuthorizedUser(ctx context.Context, token *oauth2.Token) (*twitter.TwitterUser, error)
}
