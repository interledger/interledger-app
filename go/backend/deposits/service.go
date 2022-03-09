package deposits

import (
	"context"
	"errors"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
)

var (
	ErrUnauthorized = errors.New("deposit service: unauthorized.")
)

type Service interface {
	InitiateDeposit(ctx context.Context, args *InitiateDepositArgs) (*account_transactions.AccountTransaction, error)
}

type ServiceArgs struct {
	Db   *sqlx.DB                     `validate:"required"`
	As   accounts.Service             `validate:"required"`
	Is   identity.Service             `validate:"required"`
	Fs   fundingsources.Service       `validate:"required"`
	Ts   account_transactions.Service `validate:"required"`
	Noop noop.Service                 `validate:"required"`
}

type service struct {
	validator *validator.Validate
	db        *sqlx.DB
	as        accounts.Service
	is        identity.Service
	fs        fundingsources.Service
	ts        account_transactions.Service
	noop      noop.Service
}

func NewService(args *ServiceArgs) (Service, error) {
	validator := validator.New()
	if err := validator.Struct(args); err != nil {
		return nil, &ErrInvalidArgument{Err: "Deposit service: " + err.Error()}
	}

	return &service{
		validator: validator,
		db:        args.Db,
		as:        args.As,
		is:        args.Is,
		fs:        args.Fs,
		ts:        args.Ts,
		noop:      args.Noop,
	}, nil
}

type InitiateDepositArgs struct {
	IdentityID      string `validate:"required,uuid4"`
	AccountID       string `validate:"required,uuid4"`
	FundingSourceID string `validate:"required,uuid4"`
	Amount          uint64 `validate:"required,gt=0"`
}

func (s *service) InitiateDeposit(ctx context.Context, args *InitiateDepositArgs) (*account_transactions.AccountTransaction, error) {
	if err := s.validator.Struct(args); err != nil {
		return nil, &ErrInvalidArgument{Err: "Deposit service: " + err.Error()}
	}

	var transaction *transactions.AccountTransaction
	err := crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		id, err := s.is.Get(ctx, tx, args.IdentityID)
		if err != nil {
			switch err.(type) {
			case *identity.ErrNotFound:
				return &ErrNotFound{Err: "Deposit service:" + err.Error()}
			case *identity.ErrInvalidArgument:
				return &ErrInvalidArgument{Err: "Deposit service:" + err.Error()}
			default:
				return &ErrInternalError{Err: "Deposit service:" + err.Error()}
			}
		}
		acc, err := s.as.Get(ctx, tx, args.AccountID)
		if err != nil {
			return &ErrInternalError{Err: "Deposit service:" + err.Error()}
		}
		if !s.as.CanMakeDeposit(acc, id.ID) {
			return ErrUnauthorized
		}
		fundingSource, err := s.fs.Get(ctx, tx, args.FundingSourceID)
		if err != nil {
			switch err.(type) {
			case *fundingsources.ErrInvalidArgument:
				return &ErrInvalidArgument{Err: "Deposit service:" + err.Error()}
			case *fundingsources.ErrNotFound:
				return &ErrNotFound{Err: "Deposit service:" + err.Error()}
			default:
				return &ErrInternalError{Err: "Deposit service:" + err.Error()}
			}
		}
		if fundingSource.AccountID != acc.ID {
			return &ErrNotFound{Err: "Deposit service: Funding source not found."}
		}
		if fundingSource.VerificationState != "verified" {
			return &ErrUnverifiedFundingSource{Err: "Deposit service: Funding source is not verified."}
		}

		trx, err := s.ts.Create(ctx, tx, &transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Type:        "deposit",
			NetAmount:   args.Amount,
			Description: "Deposit from " + fundingSource.Name,
			State:       "completed", // TODO: define states
			LedgerTransfers: []transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        s.noop.GetLedgerID(),
					DebitAccountID:  acc.ID,
					CreditAccountID: s.noop.GetEquityAccountID(),
					Amount:          args.Amount,
					// Code: "1", // TODO: define ledger transfer codes.
					Flags: transactions.LedgerTransferFlags{},
				},
			},
		})
		if err != nil {
			switch err.(type) {
			case *transactions.ErrInvalidArgument:
				return &ErrInvalidArgument{Err: "Deposit service:" + err.Error()}
			case *transactions.ErrNotFound:
				return &ErrNotFound{Err: "Deposit service:" + err.Error()}
			default:
				return &ErrInternalError{Err: "Deposit service:" + err.Error()}
			}
		}
		transaction = trx
		return nil
	})
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

type ErrInternalError struct {
	Err string
}

func (e ErrInternalError) Error() string {
	return e.Err
}

type ErrInvalidArgument struct {
	Err string
}

func (e ErrInvalidArgument) Error() string {
	return e.Err
}

type ErrNotFound struct {
	Err string
}

func (e ErrNotFound) Error() string {
	return e.Err
}

type ErrUnverifiedFundingSource struct {
	Err string
}

func (e ErrUnverifiedFundingSource) Error() string {
	return e.Err
}
