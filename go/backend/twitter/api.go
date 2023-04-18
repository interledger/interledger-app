package twitter

import (
	"context"
	"gitlab.com/fynbos/backend/twitter/ops"
)

type Client interface {
	CreateAuthURL(ctx context.Context, b ops.Backends) (string, error)
}
