package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/payments"
)

const cols = `id, public_id, state, required_actions, sender_id, sender_id_type, sender_amount, sender_currency, sender_account, receiver_id, receiver_id_type, receiver_amount, receiver_currency, receiver_account, created_at, updated_at`

type dbPayment struct {
	ID               string                `db:"id"`
	PublicID         string                `db:"public_id"`
	State            payments.State        `db:"state"`
	RequiredActions  pq.Int32Array         `db:"required_actions"`
	SenderID         string                `db:"sender_id"`
	SenderIDType     payments.IdentityType `db:"sender_id_type"`
	SenderAmount     uint64                `db:"sender_amount"`
	SenderCurrency   string                `db:"sender_currency"`
	SenderAccount    sql.NullString        `db:"sender_account"`
	ReceiverID       string                `db:"receiver_id"`
	ReceiverIDType   payments.IdentityType `db:"receiver_id_type"`
	ReceiverAmount   uint64                `db:"receiver_amount"`
	ReceiverCurrency string                `db:"receiver_currency"`
	ReceiverAccount  sql.NullString        `db:"receiver_account"`
	CreatedAt        time.Time             `db:"created_at"`
	UpdatedAt        time.Time             `db:"created_at"`
}

func transformPayment(db dbPayment) payments.Payment {
	var actions []payments.RequiredAction
	for _, ra := range db.RequiredActions {
		actions = append(actions, payments.RequiredAction(ra))
	}

	return payments.Payment{
		ID:       db.ID,
		PublicID: db.PublicID,
		State:    db.State,
		Sender: payments.Identity{
			Type:       db.SenderIDType,
			Identifier: db.SenderID,
		},
		Receiver: payments.Identity{
			Type:       db.ReceiverIDType,
			Identifier: db.ReceiverID,
		},
		SenderAmount:    currency.FromUInt64(db.SenderAmount, currency.ParseCurrency(db.SenderCurrency)),
		ReceiverAmount:  currency.FromUInt64(db.ReceiverAmount, currency.ParseCurrency(db.ReceiverCurrency)),
		RequiredActions: actions,
	}
}

func getPayment(ctx context.Context, b Backends, id string) (*dbPayment, error) {
	var res dbPayment
	err := b.DB().GetContext(ctx, &res, fmt.Sprintf("SELECT %s FROM payments WHERE id=$1 OR public_id=$1", cols), id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, payments.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return &res, nil
}
