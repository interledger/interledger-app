package external

import (
	"context"
	"gitlab.com/fynbos/backend/linkedin"
	"golang.org/x/oauth2"
)

type Client interface {
	CreateToken(ctx context.Context, authCode string) (*oauth2.Token, error)
	GetAuthorizedUser(ctx context.Context, token *oauth2.Token) (*linkedin.User, error)
	Share(ctx context.Context, connection *linkedin.Connection, text string) (string, error) // returns post id
}
