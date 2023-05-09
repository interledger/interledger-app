package ops

import (
	"testing"

	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/keys"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"
	"gitlab.com/fynbos/backend/twitter"
	twitter_mock "gitlab.com/fynbos/backend/twitter/client/mock"

	temporal "go.temporal.io/sdk/client"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Twitter() twitter.Client
	Validator() *validator.Validate
	DB() *sqlx.DB
	Temporal() temporal.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	an  analytics.Client
	kc  *keys_mock.MockClient
	tc  *twitter_mock.MockClient
}

func (t testBackends) Temporal() temporal.Client {
	//TODO implement me
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

func NewTestBackends(t *testing.T, db *sqlx.DB) *testBackends {
	ctrl := gomock.NewController(t)
	kc := keys_mock.NewMockClient(ctrl)
	tc := twitter_mock.NewMockClient(ctrl)
	kc.EXPECT().ProvisionPrivateKey(gomock.Any(), gomock.Any()).AnyTimes()
	return &testBackends{db: db, val: validator.New(), an: analytics_client.New(nil, ""), kc: kc, tc: tc}
}
