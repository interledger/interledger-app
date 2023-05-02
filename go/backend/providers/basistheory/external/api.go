package external

import (
	"context"

	"github.com/Basis-Theory/basistheory-go/v3"
)

type Client interface {
	GetToken(ctx context.Context, id string) (*basistheory.Token, error)
}
