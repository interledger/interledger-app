package ops

import (
	"context"
	"fmt"
	"gitlab.com/fynbos/backend/notify"
)

func NotifyWallet(_ context.Context, b Backends, walletID string, event notify.NotificationType) error {
	if b.Pusher() == nil {
		return nil
	}

	channel := fmt.Sprintf("wallet-%s", walletID)
	err := b.Pusher().Trigger(channel, string(event), "")

	return err
}
