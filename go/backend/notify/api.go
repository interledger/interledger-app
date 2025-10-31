package notify

import (
	"context"
)

type Client interface {
	NotifyWallet(ctx context.Context, walletID string, event NotificationType) error
	NotifyPending3DSConfirmation(ctx context.Context, walletID string, data any) error
}
