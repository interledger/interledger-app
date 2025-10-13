package ops

import (
	"testing"

	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/rafiki"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/wallets"

	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/user"

	temporal "go.temporal.io/sdk/client"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Signup() signup.Client
	Temporal() temporal.Client
	Users() user.Client
	Notify() notify.Client
	Email() email.Client
	Wallets() wallets.Client
	Xago() xago.Client
	Astra() astra.Client
	PTI() pti.Client
	Chimoney() chimoney.Client
	Rafiki() rafiki.Client
}

type testBackends struct {
	db  *sqlx.DB
	val *validator.Validate
	tp  temporal.Client
	uc  user.Client
	sc  signup.Client
	nc  notify.Client
	em  email.Client
	wc  wallets.Client
	xg  xago.Client
	rf  rafiki.Client
}

func (b testBackends) Chimoney() chimoney.Client {
	return nil
}

func (t testBackends) PTI() pti.Client {
	return nil
}

func (t testBackends) Astra() astra.Client {
	return nil
}

func (t testBackends) Users() user.Client {
	return t.uc
}

func (t testBackends) Validator() *validator.Validate {
	return t.val
}

func (t testBackends) DB() *sqlx.DB {
	return t.db
}

func (t testBackends) Temporal() temporal.Client {
	return t.tp
}

func (t testBackends) Signup() signup.Client {
	return t.sc
}

func (t testBackends) Notify() notify.Client {
	return t.nc
}

func (t testBackends) Email() email.Client {
	return t.em
}

func (t testBackends) Wallets() wallets.Client {
	return t.wc
}

func (t testBackends) Xago() xago.Client {
	return t.xg
}
func (t testBackends) Rafiki() rafiki.Client {
	return t.rf
}

func NewTestBackends(_ *testing.T, db *sqlx.DB, tp temporal.Client, uc user.Client, sc signup.Client, nc notify.Client, em email.Client, wc wallets.Client) Backends {
	return &testBackends{db: db, val: validator.New(), tp: tp, uc: uc, sc: sc, nc: nc, em: em, wc: wc}
}
