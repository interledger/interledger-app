package country

import "context"

type Client interface {
	Get(ctx context.Context, id string) (*Country, error)
	GetByAlpha2(ctx context.Context, code string) (*Country, error)
	GetAll(ctx context.Context) ([]*Country, error)
}
