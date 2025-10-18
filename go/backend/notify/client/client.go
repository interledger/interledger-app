package client

import (
	"context"

	"github.com/pusher/pusher-http-go/v5"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/notify/ops"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var _ notify.Client = client{}

type client struct {
	b ops.Backends
}

func New(b Backends, pusherClientURL string) notify.Client {
	if pusherClientURL == "" {
		log.Info("no pusherClientURL specified.")
		return &client{
			b: &opsBackends{
				Backends: b,
			},
		}
	}

	pc, err := pusher.ClientFromURL(pusherClientURL)
	if err != nil {
		log.Error("error creating pusher client", zap.Error(err))
	}

	ob := &opsBackends{
		Backends: b,
		pusher:   pc,
	}

	return &client{
		b: ob,
	}
}

func (c client) NotifyWallet(ctx context.Context, walletId string, event notify.NotificationType) error {
	return ops.NotifyWallet(ctx, c.b, walletId, event)
}
