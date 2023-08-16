package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/payments"
)

const cols = `id, public_id, state, sender_id, sender_id_type, sender_amount, sender_currency, sender_account, receiver_id, receiver_id_type, receiver_amount, receiver_currency, receiver_account, send_transaction_id, receive_transaction_id, action_three_ds_required, action_three_ds_id, note, created_at, updated_at`

type dbPayment struct {
	ID                   string                `db:"id"`
	PublicID             string                `db:"public_id"`
	State                payments.State        `db:"state"`
	ThreeDSRequired      bool                  `db:"action_three_ds_required"`
	ThreeDSID            sql.NullString        `db:"action_three_ds_id"`
	SenderID             string                `db:"sender_id"`
	SenderIDType         payments.IdentityType `db:"sender_id_type"`
	SenderAmount         uint64                `db:"sender_amount"`
	SenderCurrency       string                `db:"sender_currency"`
	SenderAccount        sql.NullString        `db:"sender_account"`
	ReceiverID           string                `db:"receiver_id"`
	ReceiverIDType       payments.IdentityType `db:"receiver_id_type"`
	ReceiverAmount       uint64                `db:"receiver_amount"`
	ReceiverCurrency     string                `db:"receiver_currency"`
	ReceiverAccount      sql.NullString        `db:"receiver_account"`
	SendTransactionID    sql.NullString        `db:"send_transaction_id"`
	ReceiveTransactionID sql.NullString        `db:"receive_transaction_id"`
	Note                 sql.NullString        `db:"note"`
	CreatedAt            time.Time             `db:"created_at"`
	UpdatedAt            time.Time             `db:"updated_at"`
}

func transformPayment(db dbPayment) *payments.Payment {

	var actions []payments.RequiredActionType
	if db.ThreeDSRequired {
		actions = append(actions, payments.RequiredActionTypeThreeDS)
	}

	return &payments.Payment{
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
		SenderAmount:         currency.FromUInt64(db.SenderAmount, currency.ParseCurrency(db.SenderCurrency)),
		ReceiverAmount:       currency.FromUInt64(db.ReceiverAmount, currency.ParseCurrency(db.ReceiverCurrency)),
		SenderAccount:        db.SenderAccount.String,
		ReceiverAccount:      db.ReceiverAccount.String,
		RequiredActions:      actions,
		SendTransactionID:    db.SendTransactionID.String,
		ReceiveTransactionID: db.ReceiveTransactionID.String,
		UpdatedAt:            db.UpdatedAt,
		Note:                 db.Note.String,
	}
}

func getPayment(ctx context.Context, b Backends, id string) (*dbPayment, error) {
	var res dbPayment
	err := b.DB().GetContext(ctx, &res, fmt.Sprintf("SELECT %s FROM payments WHERE id=$1 OR public_id=$2", cols), id, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, payments.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return &res, nil
}

func Lookup(ctx context.Context, b Backends, id string) (*payments.Payment, error) {
	dbp, err := getPayment(ctx, b, id)
	if err != nil {
		return nil, err
	}

	return transformPayment(*dbp), nil
}

func Create(ctx context.Context, b Backends, p payments.CreateArgs) (*payments.Payment, error) {

	// TODO Calculate more actions required
	id := uuid.NewString()
	stmt, args, err := db.NewInsert("payments").
		Value("id", id).
		Value("public_id", "fynbos_"+strconv.Itoa(rand.Int())). // TODO: Generate "pretty" soft id for display
		Value("state", payments.StateCreated).
		Value("sender_id", p.Sender.Identifier).
		Value("sender_id_type", p.Sender.Type).
		Value("sender_amount", p.SenderAmount.Value).
		Value("sender_currency", p.SenderAmount.Currency).
		Value("sender_account", sql.NullString{String: p.SenderAccount, Valid: p.SenderAccount != ""}).
		Value("receiver_id", p.Receiver.Identifier).
		Value("receiver_id_type", p.Receiver.Type).
		Value("receiver_amount", p.ReceiverAmount.Value).
		Value("receiver_currency", p.ReceiverAmount.Currency).
		Value("receiver_account", sql.NullString{String: p.ReceiverAccount, Valid: p.ReceiverAccount != ""}).
		Value("action_three_ds_required", true).
		Value("note", sql.NullString{String: p.Note, Valid: p.Note != ""}).
		GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	_, err = b.DB().ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return Lookup(ctx, b, id)
}

func SetState(ctx context.Context, b Backends, id string, state payments.State) error {
	p, err := Lookup(ctx, b, id)
	if err != nil {
		return err
	}
	if !p.State.CanTransitionTo(state) {
		return fmt.Errorf("%w id=%s current state=%s, proposed state=%s", payments.ErrInvalidStateTransition, id, p.State, state)
	}

	result, err := b.DB().ExecContext(ctx, "UPDATE payments SET state=$1 WHERE id=$2 AND state=$3;", state, id, p.State)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}
	if rows, _ := result.RowsAffected(); rows < 1 {
		return fmt.Errorf("%w Failed to update state. id=%s, proposed state=%s", payments.ErrInternal, id, state)
	}

	return nil
}

func setSendTransactionID(ctx context.Context, b Backends, paymentID, txID string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE payments SET send_transaction_id=$1 WHERE id=$2", txID, paymentID)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return nil
}

func setReceiveTransactionID(ctx context.Context, b Backends, paymentID, txID string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE payments SET receive_transaction_id=$1 WHERE id=$2", txID, paymentID)
	if err != nil {
		return fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return nil
}
