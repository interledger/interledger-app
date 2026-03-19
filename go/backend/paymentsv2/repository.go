package paymentsv2

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"gitlab.com/fynbos/geo"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) repository {
	return repository{db: db}
}

func (r repository) Get(ctx context.Context, id string) (*Payment, error) {
	var res row
	err := r.db.GetContext(ctx, &res, "SELECT * FROM payments_v2 WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	senderCurrency := geo.NewCurrency(geo.NewAsset(res.SenderNetAmountAsset, "840", uint8(res.SenderNetAmountScale), nil))
	senderCurrency.SetRawAmountInt64(res.SenderNetAmount)

	receiverCurrency := geo.NewCurrency(geo.NewAsset(res.ReceiverNetAmountAsset, "840", uint8(res.ReceiverNetAmountScale), nil))
	receiverCurrency.SetRawAmountInt64(res.ReceiverNetAmount)

	payment := Payment{
		ID:                res.ID,
		SenderWalletID:    res.SenderWalletID,
		ReceiverWalletID:  res.ReceiverWalletID,
		SenderAccountID:   res.SenderAccountID,
		ReceiverAccountID: res.ReceiverAccountID,
		SenderCurrency:    senderCurrency,
		ReceiverCurrency:  receiverCurrency,
		State:             res.State,
		Transfers:         res.Transfers,
	}

	return &payment, nil
}

func (r repository) Store(ctx context.Context, payment *Payment) error {
	transfers := pq.StringArray(payment.Transfers)

	_, err := r.db.ExecContext(ctx, "INSERT INTO payments_v2 (id, sender_wallet_id, sender_account_id, receiver_wallet_id, receiver_account_id, state, transfers, sender_net_amount, sender_net_amount_asset, sender_net_amount_scale, receiver_net_amount, receiver_net_amount_asset, receiver_net_amount_scale) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
		payment.ID,
		payment.SenderWalletID,
		payment.SenderAccountID,
		payment.ReceiverWalletID,
		payment.ReceiverAccountID,
		payment.State,
		transfers,
		payment.SenderCurrency.RawAmount().Int64(),
		payment.SenderCurrency.Code(),
		payment.SenderCurrency.Scale(),
		payment.ReceiverCurrency.RawAmount().Int64(),
		payment.ReceiverCurrency.Code(),
		payment.ReceiverCurrency.Scale(),
	)
	if err != nil {
		return err
	}
	return nil
}

func (r repository) Update(ctx context.Context, payment *Payment) error {
	transfers := pq.StringArray(payment.Transfers)
	_, err := r.db.ExecContext(ctx, "UPDATE payments_v2 SET state=$1, transfers=$2 WHERE id=$3", payment.State, transfers, payment.ID)
	if err != nil {
		return err
	}
	return nil
}
