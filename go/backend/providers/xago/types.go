package xago

import "context"

type Await func(ctx context.Context, result interface{}) error
