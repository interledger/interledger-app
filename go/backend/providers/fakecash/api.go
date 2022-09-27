package fakecash

import "context"

type Client interface {
	Create(ctx context.Context, args CreateArgs) (*Account, error)
	Get(ctx context.Context, id string) (*Account, error)
}
