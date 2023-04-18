package external

import (
	"context"

	"golang.org/x/oauth2"
)

type Client interface {
	CreateAccessToken(ctx context.Context, args CreateAccessTokenArgs) (*oauth2.Token, error)
}
