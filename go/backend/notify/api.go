package notify

import (
	"context"
)

type Client interface {
	NotifyWallet(ctx context.Context, walletId string, event NotificationType) error
}
