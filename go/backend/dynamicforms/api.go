package dynamicforms

import (
	"context"
)

type Client interface {
	Create(ctx context.Context, args *CreateFormArgs) (*Form, error)
}
