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
	"go.temporal.io/api/enums/v1"
	temporal_client "go.temporal.io/sdk/client"
)

const cols = `id, public_id, state, sender_id, sender_id_type, sender_amount, sender_currency, sender_account, receiver_id, receiver_id_type, receiver_amount, receiver_currency, receiver_account, send_transaction_id, receive_transaction_id, action_three_ds_required, action_three_ds_id, created_at, updated_at`

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

/*
	This checks that the payment has the following:

1) Sender and send account
2) Receiver identifier
3) Send amount
4) Receive amount
*/
func GetRequiredActions(ctx context.Context, b Backends, id string) ([]payments.RequiredActionType, error) {
	payment, err := Lookup(ctx, b, id)
	if err != nil {
		return nil, err
	}

	var requiredActions []payments.RequiredActionType
	if payment.PublicID == "" {
		requiredActions = append(requiredActions, payments.RequiredActionTypePublicID)
	}

	if payment.Sender.Identifier == "" {
		requiredActions = append(requiredActions, payments.RequiredActionTypeSenderIdentifier)
	}

	if payment.SenderAccount == "" {
		requiredActions = append(requiredActions, payments.RequiredActionTypeSenderAccount)
	}

	if payment.Receiver.Identifier == "" || payment.Receiver.Type == payments.IdentityTypeUnknown {
		requiredActions = append(requiredActions, payments.RequiredActionTypeReceiverIdentifier)
	}

	if payment.SenderAmount.Value < 1 || !payment.SenderAmount.Currency.Valid() {
		requiredActions = append(requiredActions, payments.RequiredActionTypeSenderAmount)
	}

	if payment.ReceiverAmount.Value < 1 || !payment.ReceiverAmount.Currency.Valid() {
		requiredActions = append(requiredActions, payments.RequiredActionTypeReceiverAmount)
	}

	return requiredActions, nil
}

func Confirm(ctx context.Context, b Backends, id string) (*payments.Payment, []payments.RequiredActionType, error) {
	requiredActions, err := GetRequiredActions(ctx, b, id)
	if err != nil {
		return nil, nil, err
	}
	if len(requiredActions) > 0 {
		return nil, requiredActions, payments.ErrRequiredActions
	}

	err = SetState(ctx, b, id, payments.StateConfirmed)
	if err != nil {
		return nil, nil, err
	}

	workflowOptions := temporal_client.StartWorkflowOptions{
		ID:                       "payments_" + id,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: time.Hour * 24 * 8, // Workflow has 8 days to complete
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	_, err = b.Temporal().ExecuteWorkflow(ctx, workflowOptions, PaymentWorkflow, id)
	if err != nil {
		return nil, nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	payment, err := Lookup(ctx, b, id)
	if err != nil {
		return nil, nil, err
	}

	return payment, nil, nil
}

func Update(ctx context.Context, b Backends, args payments.UpdateArgs) (*payments.Payment, error) {
	payment, err := getPayment(ctx, b, args.ID)
	if err != nil {
		return nil, err
	}
	if !args.SenderAmount.IsEmpty() && !args.SenderAmount.Currency.Valid() {
		return nil, fmt.Errorf("%w Sender amount currency is invalid.", payments.ErrInvalidAmount)
	}
	if !args.ReceiverAmount.IsEmpty() && !args.ReceiverAmount.Currency.Valid() {
		return nil, fmt.Errorf("%w Receiver amount currency is invalid.", payments.ErrInvalidAmount)
	}
	if !args.Receiver.IsEmpty() && !args.Receiver.Type.Valid() {
		return nil, fmt.Errorf("%w Receiver is invalid.", payments.ErrInvalidIdentifier)
	}

	noop := true
	receiver := payments.Identity{Identifier: payment.ReceiverID, Type: payment.ReceiverIDType}
	if !args.Receiver.IsEmpty() && !args.Receiver.IsEqual(receiver) {
		payment.ReceiverID = args.Receiver.Identifier
		payment.ReceiverIDType = args.Receiver.Type
		noop = false
	}

	receiveAmount := currency.FromUInt64(payment.ReceiverAmount, currency.Currency(payment.ReceiverCurrency))
	if !args.ReceiverAmount.IsEmpty() && !args.ReceiverAmount.IsEqual(receiveAmount) {
		payment.ReceiverAmount = args.ReceiverAmount.Value
		payment.ReceiverCurrency = args.ReceiverAmount.Currency.String()
		noop = false
	}
	if args.ReceiverAccount != "" && args.ReceiverAccount != payment.ReceiverAccount.String {
		payment.ReceiverAccount.String = args.ReceiverAccount
		payment.ReceiverAccount.Valid = true
		noop = false
	}
	if args.SenderAccount != "" && args.SenderAccount != payment.SenderAccount.String {
		payment.SenderAccount.String = args.SenderAccount
		payment.SenderAccount.Valid = true
		noop = false
	}

	sendAmount := currency.FromUInt64(payment.SenderAmount, currency.Currency(payment.SenderCurrency))
	if !args.SenderAmount.IsEmpty() && !args.SenderAmount.IsEqual(sendAmount) {
		payment.SenderAmount = args.SenderAmount.Value
		payment.SenderCurrency = args.SenderAmount.Currency.String()
		noop = false
	}
	if args.ThreeDSID != "" && args.ThreeDSID != payment.ThreeDSID.String {
		payment.ThreeDSID.String = args.ThreeDSID
		payment.ThreeDSID.Valid = true
		noop = false
	}
	if noop {
		return transformPayment(*payment), nil
	}

	payment.UpdatedAt = time.Now()
	stmt, values, err := db.NewUpdate("payments").ID(args.ID).
		Value("sender_amount", payment.SenderAmount).
		Value("sender_currency", payment.SenderCurrency).
		Value("sender_account", payment.SenderAccount).
		Value("receiver_id", payment.ReceiverID).
		Value("receiver_id_type", payment.ReceiverIDType).
		Value("receiver_amount", payment.ReceiverAmount).
		Value("receiver_currency", payment.ReceiverCurrency).
		Value("receiver_account", payment.ReceiverAccount).
		Value("updated_at", payment.UpdatedAt).
		Value("action_three_ds_id", payment.ThreeDSID).Returning(cols).GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	var ret dbPayment
	err = b.DB().GetContext(ctx, &ret, stmt, values...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", payments.ErrInternal, err)
	}

	return transformPayment(ret), nil
}
