package ops

import (
	"context"
	"fmt"
	"gitlab.com/fynbos/backend/notify"
)

func NotifyWallet(_ context.Context, b Backends, walletId string, event notify.NotificationType) error {

	// noop if client does not exist
	if b.Pusher() == nil {
		return nil
	}

	channel := fmt.Sprintf("wallet-%s", walletId)
	err := b.Pusher().Trigger(channel, string(event), "")

	return err
}
