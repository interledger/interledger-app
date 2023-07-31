package ops

import (
	"gitlab.com/fynbos/backend/wallets"
	"testing"

	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/notify"
	notify_client "gitlab.com/fynbos/backend/notify/client/mock"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/user"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Users() user.Client
	Wallets() wallets.Client
	Notify() notify.Client
	Analytics() analytics.Client
}

type testBackends struct {
	db     *sqlx.DB
	val    *validator.Validate
	notify notify.Client
	ac     analytics.Client
}

func (t testBackends) Users() user.Client {
	return nil
}

func (t testBackends) Wallets() wallets.Client {
	return nil
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Notify() notify.Client {
	return t.notify
}

func (t testBackends) Analytics() analytics.Client {
	return t.ac
}

func NewTestBackends(t *testing.T, db *sqlx.DB) Backends {
	ctrl := gomock.NewController(t)
	nc := notify_client.NewMockClient(ctrl)
	nc.EXPECT().NotifyWallet(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return &testBackends{db: db, val: validator.New(), notify: nc, ac: analytics_client.New(nil, "")}
}
