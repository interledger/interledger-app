package ops

import (
	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/keys"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	"testing"

	temporal "go.temporal.io/sdk/client"

	analytics_client "gitlab.com/fynbos/backend/analytics/client"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Analytics() analytics.Client
	Temporal() temporal.Client
	Keys() keys.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	an  analytics.Client
	kc  keys.Client
}

func (t testBackends) Temporal() temporal.Client {
	//TODO implement me
	panic("implement me")
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

func NewTestBackends(t *testing.T, db *sqlx.DB) Backends {
	ctrl := gomock.NewController(t)
	kc := keys_mock.NewMockClient(ctrl)
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()
	return &testBackends{db: db, val: validator.New(), an: analytics_client.New(nil, ""), kc: kc}
}
