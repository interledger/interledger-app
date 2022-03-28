package deposits

import (
	"context"
	"errors"
	"fmt"

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
	ErrUnauthorized            = errors.New("deposit: unauthorized.")
	ErrInternal                = errors.New("deposit: internal error.")
	ErrInvalidArgument         = errors.New("deposit: invalid argument.")
	ErrNotFound                = errors.New("deposit: not found.")
	ErrUnverifiedFundingSource = errors.New("deposit: unverified funding source.")
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
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
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
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	var transaction *transactions.AccountTransaction
	err := crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		id, err := s.is.Get(ctx, tx, args.IdentityID)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}
		acc, err := s.as.Get(ctx, tx, args.AccountID)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}
		if !s.as.CanMakeDeposit(acc, id.ID) {
			return ErrUnauthorized
		}
		fundingSource, err := s.fs.Get(ctx, tx, args.FundingSourceID)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}
		if fundingSource.AccountID != acc.ID {
			return ErrNotFound
		}
		if fundingSource.VerificationState != "verified" {
			return ErrUnverifiedFundingSource
		}

		trx, err := s.ts.Create(ctx, &transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Type:        "deposit",
			NetAmount:   args.Amount,
			Description: fmt.Sprintf("from %s bank account", fundingSource.Mask), // TODO Format to come from FS
			LedgerTransfers: []transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        s.noop.GetLedgerID(),
					DebitAccountID:  acc.LedgerAccountID,
					CreditAccountID: s.noop.GetEquityAccountID(),
					Amount:          args.Amount,
					// Code: "1", // TODO: define ledger transfer codes.
					Flags: transactions.LedgerTransferFlags{},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}
		transaction = trx
		return nil
	})
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
