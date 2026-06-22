package ops

import (
	"github.com/interledger/interledger-app/go/backend/keys"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/chimoney"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/pti"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/interledger/interledger-app/go/backend/rafiki"
	"github.com/interledger/interledger-app/go/backend/rafiki/external"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/jmoiron/sqlx"

	temporal "go.temporal.io/sdk/client"
)

type TestActivityBackends struct {
	db             *sqlx.DB
	payments       payments.Client
	temporal       temporal.Client
	linkedAccounts linkedaccounts.Client
	kyc            kyc.Client
	wallets        wallets.Client
	transactions   transactions.Client
	pacioli        pacioli.Client
	gatehub        gatehub.Client
	rafiki         rafiki.Client
}

func (t *TestActivityBackends) DB() *sqlx.DB                          { return t.db }
func (t *TestActivityBackends) Payments() payments.Client             { return t.payments }
func (t *TestActivityBackends) Temporal() temporal.Client             { return t.temporal }
func (t *TestActivityBackends) LinkedAccounts() linkedaccounts.Client { return t.linkedAccounts }
func (t *TestActivityBackends) KYC() kyc.Client                       { return t.kyc }
func (t *TestActivityBackends) Wallets() wallets.Client               { return t.wallets }
func (t *TestActivityBackends) Transactions() transactions.Client     { return t.transactions }
func (t *TestActivityBackends) Pacioli() pacioli.Client               { return t.pacioli }
func (t *TestActivityBackends) Gatehub() gatehub.Client               { return t.gatehub }
func (t *TestActivityBackends) Rafiki() rafiki.Client                 { return t.rafiki }

func (t *TestActivityBackends) SetDB(db *sqlx.DB)                          { t.db = db }
func (t *TestActivityBackends) SetPayments(p payments.Client)              { t.payments = p }
func (t *TestActivityBackends) SetTemporal(tp temporal.Client)             { t.temporal = tp }
func (t *TestActivityBackends) SetLinkedAccounts(la linkedaccounts.Client) { t.linkedAccounts = la }
func (t *TestActivityBackends) SetKYC(k kyc.Client)                        { t.kyc = k }
func (t *TestActivityBackends) SetWallets(w wallets.Client)                { t.wallets = w }
func (t *TestActivityBackends) SetTransactions(tx transactions.Client)     { t.transactions = tx }
func (t *TestActivityBackends) SetPacioli(p pacioli.Client)                { t.pacioli = p }
func (t *TestActivityBackends) SetGatehub(g gatehub.Client)                { t.gatehub = g }
func (t *TestActivityBackends) SetRafiki(r rafiki.Client)                  { t.rafiki = r }

func NewTestActivityBackends() *TestActivityBackends {
	return &TestActivityBackends{}
}

type Backends interface {
	DB() *sqlx.DB
	External() external.Client
	Payments() payments.Client
	Temporal() temporal.Client
	LinkedAccounts() linkedaccounts.Client
	Wallets() wallets.Client
	Keys() keys.Client
	PTI() pti.Client
	Gatehub() gatehub.Client
	Xago() xago.Client
	Chimoney() chimoney.Client
	KYC() kyc.Client
}

type ActivityBackends interface {
	DB() *sqlx.DB
	Payments() payments.Client
	Temporal() temporal.Client
	LinkedAccounts() linkedaccounts.Client
	KYC() kyc.Client
	Wallets() wallets.Client
	Transactions() transactions.Client
	Pacioli() pacioli.Client
	Gatehub() gatehub.Client
	Rafiki() rafiki.Client
}

type TestBackends struct {
	db             *sqlx.DB
	external       external.Client
	payments       payments.Client
	temporal       temporal.Client
	linkedAccounts linkedaccounts.Client
	wallets        wallets.Client
	keys           keys.Client
	pti            pti.Client
	gatehub        gatehub.Client
	xago           xago.Client
	chimoney       chimoney.Client
	kyc            kyc.Client
}

func (t *TestBackends) DB() *sqlx.DB                          { return t.db }
func (t *TestBackends) External() external.Client             { return t.external }
func (t *TestBackends) Payments() payments.Client             { return t.payments }
func (t *TestBackends) Temporal() temporal.Client             { return t.temporal }
func (t *TestBackends) LinkedAccounts() linkedaccounts.Client { return t.linkedAccounts }
func (t *TestBackends) Wallets() wallets.Client               { return t.wallets }
func (t *TestBackends) Keys() keys.Client                     { return t.keys }
func (t *TestBackends) PTI() pti.Client                       { return t.pti }
func (t *TestBackends) Gatehub() gatehub.Client               { return t.gatehub }
func (t *TestBackends) Xago() xago.Client                     { return t.xago }
func (t *TestBackends) Chimoney() chimoney.Client             { return t.chimoney }
func (t *TestBackends) KYC() kyc.Client                       { return t.kyc }

func (t *TestBackends) SetDB(db *sqlx.DB)                          { t.db = db }
func (t *TestBackends) SetExternal(ext external.Client)            { t.external = ext }
func (t *TestBackends) SetPayments(p payments.Client)              { t.payments = p }
func (t *TestBackends) SetTemporal(tp temporal.Client)             { t.temporal = tp }
func (t *TestBackends) SetLinkedAccounts(la linkedaccounts.Client) { t.linkedAccounts = la }
func (t *TestBackends) SetWallets(w wallets.Client)                { t.wallets = w }
func (t *TestBackends) SetKeys(k keys.Client)                      { t.keys = k }
func (t *TestBackends) SetPTI(p pti.Client)                        { t.pti = p }
func (t *TestBackends) SetGatehub(g gatehub.Client)                { t.gatehub = g }
func (t *TestBackends) SetXago(x xago.Client)                      { t.xago = x }
func (t *TestBackends) SetChimoney(c chimoney.Client)              { t.chimoney = c }
func (t *TestBackends) SetKYC(k kyc.Client)                        { t.kyc = k }

func NewTestBackends(opts ...func(*TestBackends)) *TestBackends {
	tb := &TestBackends{}
	for _, opt := range opts {
		opt(tb)
	}
	return tb
}
