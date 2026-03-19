package paymentsv2

import (
	"github.com/lib/pq"
	"gitlab.com/fynbos/geo"
)

type row struct {
	ID                     string         `db:"id"`
	SenderWalletID         string         `db:"sender_wallet_id"`
	SenderAccountID        string         `db:"sender_account_id"`
	ReceiverWalletID       string         `db:"receiver_wallet_id"`
	ReceiverAccountID      string         `db:"receiver_account_id"`
	State                  string         `db:"state"`
	Transfers              pq.StringArray `db:"transfers"`
	SenderNetAmount        int64          `db:"sender_net_amount"`
	SenderNetAmountAsset   string         `db:"sender_net_amount_asset"`
	SenderNetAmountScale   int            `db:"sender_net_amount_scale"`
	ReceiverNetAmount      int64          `db:"receiver_net_amount"`
	ReceiverNetAmountAsset string         `db:"receiver_net_amount_asset"`
	ReceiverNetAmountScale int            `db:"receiver_net_amount_scale"`
}

type Payment struct {
	ID string

	SenderWalletID   string
	ReceiverWalletID string

	SenderAccountID   string
	ReceiverAccountID string

	SenderCurrency   *geo.Currency
	ReceiverCurrency *geo.Currency

	State     string
	Transfers []string
}

func (p *Payment) SetState(state string) {
	p.State = state
}

func (p *Payment) AddTrasnsfer(transferID string) {
	p.Transfers = append(p.Transfers, transferID)
}
