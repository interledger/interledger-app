package ops

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/payments"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/features"
	"gitlab.com/fynbos/backend/images"
	images_mock "gitlab.com/fynbos/backend/images/client/mock"
	"gitlab.com/fynbos/backend/keys"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/notify"
	notify_mock "gitlab.com/fynbos/backend/notify/client/mock"
	payments_mock "gitlab.com/fynbos/backend/payments/client/mock"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/twitter"
	twitter_mock "gitlab.com/fynbos/backend/twitter/client/mock"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	Twitter() twitter.Client
	Validator() *validator.Validate
	DB() *sqlx.DB
	Temporal() temporal.Client
	Analytics() analytics.Client
	Keys() keys.Client
	Images() images.Client
	Notify() notify.Client
	Wallets() wallets.Client
	Payments() payments.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	an  analytics.Client
	kc  *keys_mock.MockClient
	tc  *twitter_mock.MockClient
	img images.Client
	wc  wallets.Client
	nc  notify.Client
	pc  payments.Client
}

func (t testBackends) Payments() payments.Client {
	return t.pc
}

func (t testBackends) Twilio() twilio.Service {
	return nil
}

func (t testBackends) Users() user.Client {
	return nil
}

func (t testBackends) LinkedAccounts() linkedaccounts.Client {
	panic("implement me")
}

func (t testBackends) Transactions() transactions.Client {
	panic("implement me")
}

func (t testBackends) Temporal() temporal.Client {
	panic("implement me")
}

func (t testBackends) Twitter() twitter.Client {
	return t.tc
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Analytics() analytics.Client {
	return t.an
}

func (t testBackends) Keys() keys.Client {
	return t.kc
}

func (t testBackends) Images() images.Client {
	return t.img
}

func (t testBackends) Features() features.Client {
	return nil
}

func (t testBackends) Wallets() wallets.Client {
	return t.wc
}

func (t testBackends) Notify() notify.Client {
	return t.nc
}

func NewTestBackends(t *testing.T, db *sqlx.DB) *testBackends {
	ctrl := gomock.NewController(t)
	kc := keys_mock.NewMockClient(ctrl)
	tc := twitter_mock.NewMockClient(ctrl)
	img := images_mock.NewMockClient(ctrl)
	wc := wallets_mock.NewMockClient(ctrl)
	nc := notify_mock.NewMockClient(ctrl)
	pc := payments_mock.NewMockClient(ctrl)
	nc.EXPECT().NotifyWallet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	wc.EXPECT().AddAddress(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	wc.EXPECT().Get(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id string) (*wallets.Wallet, error) {
		wa, _ := wallets.ParseAddress("https://something.com/someaddress")
		return &wallets.Wallet{
			ID:        id,
			Name:      "name",
			Addresses: []wallets.Address{wa},
		}, nil
	})
	img.EXPECT().GenerateTwitterIdentity(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	img.EXPECT().GenerateTwitterIdentityOG(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()
	kc.EXPECT().List(gomock.Any(), gomock.Any()).Return([]keys.Key{
		{
			ID:   "test",
			Type: keys.Custodial,
		},
	}, nil).AnyTimes()
	kc.EXPECT().Sign(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	pc.EXPECT().SignalIdentityCreated(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	return &testBackends{db: db, val: validator.New(), an: analytics_client.New(nil, ""), kc: kc, tc: tc, img: img, wc: wc, nc: nc, pc: pc}
}
