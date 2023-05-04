package external

import (
	"context"

	"golang.org/x/oauth2"
)

type Client interface {
	CreateToken(ctx context.Context, args *CreateTokenArgs) (*oauth2.Token, error)
}
