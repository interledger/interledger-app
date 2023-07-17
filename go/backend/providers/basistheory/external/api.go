package external

import (
	"context"

	"github.com/Basis-Theory/basistheory-go/v3"
	bt "gitlab.com/fynbos/backend/providers/basistheory"
)

type Client interface {
	GetToken(ctx context.Context, id string) (*basistheory.Token, error)
	CreateCardToken(ctx context.Context, args bt.CreateCardTokenArgs) (*basistheory.CreateTokenResponse, error)
}
