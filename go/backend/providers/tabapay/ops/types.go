package ops

import (
	"database/sql"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
)

type PullFromCardArgs struct {
	WalletID            string
	ProviderID          string
	ReferenceID         string
	SettlementAccountID string
	Amount              currency.Amount
	ThreeDSID           string
	SoftDescriptor      *external.SoftDescriptor
}

type dbThreeDSSession struct {
	ID                     string       `db:"id"`
	CardID                 string       `db:"card_id"`
	OrderID                string       `db:"order_id"`
	Revision               int          `db:"revision"`
	Amount                 uint64       `db:"amount"`
	Currency               string       `db:"currency"`
	Version                string       `db:"version"`
	Enrolled               string       `db:"enrolled"`
	ProcessorTransactionID string       `db:"processor_transaction_id"`
	DsTransactionID        string       `db:"ds_transaction_id"`
	Status                 string       `db:"status"`
	ECI                    string       `db:"eci"`
	UCAF                   string       `db:"ucaf"`
	XID                    string       `db:"xid"`
	ChallengeURL           string       `db:"challenge_url"`
	Payload                string       `db:"payload"`
	InitAt                 time.Time    `db:"init_at"`
	LookupAt               sql.NullTime `db:"lookup_at"`
	AuthenticatedAt        sql.NullTime `db:"authenticated_at"`
}

var dbThreeDSSessionFields = "id, card_id, order_id, revision, amount, currency, version, enrolled, processor_transaction_id, ds_transaction_id, status, eci, ucaf, xid, challenge_url, payload, init_at, lookup_at, authenticated_at"

type dbFXRate struct {
	ID          string            `db:"id"`
	Currency    currency.Currency `db:"currency_code"`
	BuyRate     float64           `db:"buy_rate"`
	BuyRateInv  float64           `db:"buy_rate_inverted"`
	SellRate    float64           `db:"sell_rate"`
	SellRateInv float64           `db:"sell_rate_inverted"`
}
