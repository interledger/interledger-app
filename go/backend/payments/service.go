package payments

import (
	"context"
	"errors"
	"fmt"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
)

var (
	ErrInvalidArgument     = errors.New("payments service: invalid argument.")
	ErrInternal            = errors.New("payments service: internal error.")
	ErrUnverifiedAccount   = errors.New("payments service: unverified account.")
	ErrInsufficientBalance = errors.New("payments service: insufficient balance.")
	ErrUnauthorized        = errors.New("payments service: unauthorized.")
)

const (
	Verified  string = "verified"
	Retry     string = "retry"
	Kba       string = "kba"
	Document  string = "document"
	Suspended string = "suspended"
)

type Service interface {
	InitiateOutgoingPayment(ctx context.Context, args *InitiateOutgoingPaymentArgs) (*account_transactions.AccountTransaction, error)
}

type service struct {
	validator *validator.Validate
	db        *sqlx.DB
	as        accounts.Service
	is        identity.Service
	ts        account_transactions.Service
	noop      noop.Service
}

type ServiceArgs struct {
	Db   *sqlx.DB                     `validate:"required"`
	As   accounts.Service             `validate:"required"`
	Is   identity.Service             `validate:"required"`
	Ts   account_transactions.Service `validate:"required"`
	Noop noop.Service                 `validate:"required"`
}

func NewService(args *ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, err
	}

	return &service{
		validator: validator,
		db:        args.Db,
		as:        args.As,
		is:        args.Is,
		ts:        args.Ts,
		noop:      args.Noop,
	}, nil
}

type InitiateOutgoingPaymentArgs struct {
	IdentityID string `validate:"required"`
	AccountID  string `validate:"required"`
	Amount     uint64 `validate:"gt=0"`
	To         string `validate:"required"`
}

func (s *service) InitiateOutgoingPayment(
	ctx context.Context,
	args *InitiateOutgoingPaymentArgs,
) (*account_transactions.AccountTransaction, error) {
	err := s.validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInvalidArgument)
	}

	var outgoingPayment *account_transactions.AccountTransaction
	err = crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		id, err := s.is.Get(ctx, args.IdentityID)
		if err != nil {
			return fmt.Errorf("%s %w", err.Error(), ErrInternal)
		}
		acc, err := s.as.Get(ctx, tx, args.AccountID)
		if err != nil {
			return fmt.Errorf("%s %w", err.Error(), ErrInternal)
		}
		if !s.as.CanMakeOutgoingPayment(acc, id.ID) {
			return fmt.Errorf("%w", ErrUnauthorized)
		}
		if !acc.IsVerified() {
			return fmt.Errorf("%w", ErrUnverifiedAccount)
		}

		// We would typically create pending transfers here then ask the provider to initiate the
		// payment. But for the noop case we will create single phase transfers.
		transaction, err := s.ts.Create(ctx, &account_transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Description: "Sent to " + args.To,
			Type:        "outgoingPayment",
			NetAmount:   args.Amount,
			LedgerTransfers: []account_transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        s.noop.GetLedgerID(),
					DebitAccountID:  s.noop.GetEquityAccountID(),
					CreditAccountID: acc.LedgerAccountID,
					Amount:          args.Amount,
					// Code: uint16,
					Flags: account_transactions.LedgerTransferFlags{},
				},
			},
		})
		if err != nil {
			switch {
			case errors.Is(err, account_transactions.ErrExceedsDebits):
				return ErrInsufficientBalance
			default:
				return fmt.Errorf("%s %w", err.Error(), ErrInternal)
			}
		}
		outgoingPayment = transaction

		err = s.noop.InitiateOutgoingPayment(ctx, &noop.OutgoingPaymentArgs{
			Amount: args.Amount,
			To:     args.To,
		})
		if err != nil {
			return fmt.Errorf("Provider failed to initiate payment: %s %w", err.Error(), ErrInternal)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return outgoingPayment, nil
}
