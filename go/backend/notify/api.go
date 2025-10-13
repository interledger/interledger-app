package notify

import (
	"context"
)

type Client interface {
	NotifyWallet(ctx context.Context, walletID string, event NotificationType) error
}
