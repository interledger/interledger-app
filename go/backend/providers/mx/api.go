package mx

import (
	"context"
)

type Client interface {
	GetWidget(ctx context.Context, walletID string) (string, error)
}
