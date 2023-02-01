package ops

import (
	"context"
	"fmt"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/log"
)

func NotifyWallet(_ context.Context, b Backends, walletId string, event notify.NotificationType) error {

	log.Info("Notifying wallet from function")

	// noop if client does not exist
	if b.Pusher() == nil {
		return nil
	}

	log.Info("Pusher exists to notify wallet")

	channel := fmt.Sprintf("wallet-%s", walletId)
	err := b.Pusher().Trigger(channel, string(event), "")

	return err
}
