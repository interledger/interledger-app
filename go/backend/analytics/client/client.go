package client

import (
	"github.com/interledger/interledger-app/go/backend/analytics"
	"github.com/interledger/interledger-app/go/backend/analytics/ops"
	"github.com/interledger/interledger-app/go/log"
	segment "github.com/segmentio/analytics-go/v3"
	"go.uber.org/zap"
)

var _ analytics.Client = client{}

type client struct {
	b       ops.Backends
	enabled bool
}

func New(b Backends, key string) analytics.Client {

	segmentClient := segment.New(key)

	enabled := false
	if key != "" {
		enabled = true
	}

	ob := &opsBackends{
		Backends: b,
		segment:  segmentClient,
	}

	return &client{
		b:       ob,
		enabled: enabled,
	}
}

func (c client) Close() {
	if c.enabled {
		err := c.b.Segment().Close()
		if err != nil {
			log.Error("error closing segment", zap.Error(err))
		}
	}
}

func (c client) Identify(args analytics.IdentifyArgs) {
	if c.enabled {
		ops.Identify(c.b, args)
	}
}

func (c client) TrackUserSignup(userID string) {
	if c.enabled {
		ops.TrackUserSignup(c.b, userID)
	}
}

func (c client) TrackUserLogin(userID string) {
	if c.enabled {
		ops.TrackUserLogin(c.b, userID)
	}
}

func (c client) TrackUserLogout(userID string) {
	if c.enabled {
		ops.TrackUserLogout(c.b, userID)
	}
}

func (c client) TrackWalletCreated(walletID, userID string) {
	if c.enabled {
		ops.TrackWalletCreated(c.b, walletID, userID)
	}
}

func (c client) TrackWalletPaymentPointerCreated(walletID string) {
	if c.enabled {
		ops.TrackWalletPaymentPointerCreated(c.b, walletID)
	}
}

func (c client) GroupUserWallet(walletID, userID string) {
	if c.enabled {
		ops.GroupWallet(c.b, walletID, userID)
	}
}

func (c client) TrackWalletTransactionCreated(walletID string, args analytics.WalletTransactionArgs) {
	if c.enabled {
		ops.TrackWalletTransactionCreated(c.b, walletID, args)
	}
}

func (c client) TrackWalletTransactionCompleted(walletID string, args analytics.WalletTransactionArgs) {
	if c.enabled {
		ops.TrackWalletTransactionCompleted(c.b, walletID, args)
	}
}

func (c client) TrackWalletTransactionFailed(walletID string, args analytics.WalletTransactionArgs) {
	if c.enabled {
		ops.TrackWalletTransactionFailed(c.b, walletID, args)
	}
}
