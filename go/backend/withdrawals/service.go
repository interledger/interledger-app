package withdrawals

import (
	"context"
	"errors"
	"fmt"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
)

var (
	ErrUnverifiedFundingSource = errors.New("withdrawal service: unverified funding source")
	ErrInvalidArgument         = errors.New("withdrawal service: invalid arguments")
	ErrInternalError           = errors.New("withdrawal service: internal error")
	ErrInsufficientBalance     = errors.New("withdrawal service: insufficient balance")
)

type Service interface {
	InitiateWithdrawal(ctx context.Context, args *InitiateWithdrawalArgs) (*transactions.AccountTransaction, error)
}

type InitiateWithdrawalArgs struct {
	IdentityID      string `validate:"required,uuid4"`
	AccountID       string `validate:"required,uuid4"`
	FundingSourceID string `validate:"required,uuid4"`
	Amount          uint64 `validate:"required,gt=0"`
}

type ServiceArgs struct {
	Db   *sqlx.DB               `validate:"required"`
	As   accounts.Service       `validate:"required"`
	Is   identity.Service       `validate:"required"`
	Fs   fundingsources.Service `validate:"required"`
	Ts   transactions.Service   `validate:"required"`
	Noop noop.Service           `validate:"required"`
}

type service struct {
	db   *sqlx.DB
	as   accounts.Service
	is   identity.Service
	fs   fundingsources.Service
	ts   transactions.Service
	noop noop.Service
}

func NewService(args *ServiceArgs) (Service, error) {
	return &service{
		db:   args.Db,
		as:   args.As,
		is:   args.Is,
		fs:   args.Fs,
		ts:   args.Ts,
		noop: args.Noop,
	}, nil
}

func (s *service) InitiateWithdrawal(ctx context.Context, args *InitiateWithdrawalArgs) (*transactions.AccountTransaction, error) {

	var transaction *transactions.AccountTransaction
	err := crdbsqlx.ExecuteTx(ctx, s.db, nil, func(sqlTx *sqlx.Tx) error {
		// get identity
		id, err := s.is.Get(ctx, sqlTx, args.IdentityID)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInternalError, err.Error())
		}

		// get acc
		acc, err := s.as.Get(ctx, sqlTx, args.AccountID)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInternalError, err.Error())
		}
		if acc.IdentityID != id.ID {
			return fmt.Errorf("account and identity dont match %w %s", ErrInternalError, err.Error())
		}

		// get funding source
		fs, err := s.fs.Get(ctx, sqlTx, args.FundingSourceID)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInternalError, err.Error())
		}
		if fs.AccountID != args.AccountID {
			return fmt.Errorf("funding source and account dont match %w", ErrInternalError)
		}
		if !fundingsources.IsVerified(fs) {
			return ErrUnverifiedFundingSource
		}

		// ask providers for plan of transfers - in future this will be get quote
		equityAccountID := s.noop.GetEquityAccountID()

		// prepare transactions and transfers
		trx, err := s.ts.Create(ctx, &transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Type:        "withdrawal",
			NetAmount:   args.Amount,
			Description: fmt.Sprintf("to %s bank account", fs.Mask), // TODO Format to come from FS
			LedgerTransfers: []transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        s.noop.GetLedgerID(),
					DebitAccountID:  equityAccountID,
					CreditAccountID: acc.LedgerAccountID,
					Amount:          args.Amount,
					// Code: "1", // TODO: define ledger transfer codes.
					Flags: transactions.LedgerTransferFlags{},
				},
			},
		})
		if err != nil {
			switch {
			case errors.Is(err, transactions.ErrExceedsDebits):
				return ErrInsufficientBalance
			default:
				return fmt.Errorf("transaction failed %w %s", ErrInternalError, err.Error())
			}
		}

		transaction = trx

		// call out to provider to execute transfers
		err = s.noop.InitiateBankWithdrawal(ctx, &noop.BankWithdrawalArgs{
			Amount: args.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
