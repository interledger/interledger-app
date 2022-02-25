package noop

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/proto/pacioli/v1"
	tb_types "gitlab.com/fynbos/tigerbeetle_go/pkg/types"

	// TODO: this needs to be refactored to be part of this provider
	_noop "gitlab.com/fynbos/backend/identity/noop"
)

type Service interface {
	GetUserFundingSources(ctx context.Context, tx *sqlx.Tx, identityId string) ([]*fundingsources.FundingSource, error)
	LinkBankAccount(ctx context.Context, args *LinkBankAccountArgs) (*fundingsources.FundingSource, error)
	VerifyBankAccount(ctx context.Context, args *VerifyArgs) (*fundingsources.FundingSource, error)
	InitiateBankDeposit(ctx context.Context, args *BankDepositArgs) (*transactions.AccountTransaction, error)
	InitiateBankWithdrawal(ctx context.Context, args *BankWithdrawalArgs) (*transactions.AccountTransaction, error)
	InitiateOutgoingPayment(ctx context.Context, args *OutgoingPaymentArgs) (*transactions.AccountTransaction, error)

	GetEquityAccountID() string
	GetLedgerID() string // This shouldn't be necessary - here till pacioli is refactored
}

type service struct {
	validator *validator.Validate
	fs        fundingsources.Service
	as        accounts.Service
	ts        transactions.Service
	is        identity.Service
	db        *sqlx.DB
	cache     map[string]*NoopBankAccount
	// TODO: configuring and management of ledgers and ledger accounts that this provider
	// acts upon.
	ledgerID        string
	equityAccountID string
}

type ServiceArgs struct {
	Db            *sqlx.DB               `validate:"required"`
	FundingSource fundingsources.Service `validate:"required"`
	Transaction   transactions.Service   `validate:"required"`
	Account       accounts.Service       `validate:"required"`
	Identity      identity.Service       `validate:"required"`

	// TODO: configuring and management of ledgers and ledger accounts
	LedgerID    string `validate:"required"`
	EquityAccID string `validate:"required"`
}

func NewService(args ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, err
	}

	return &service{
		validator:       validator,
		db:              args.Db,
		fs:              args.FundingSource,
		ts:              args.Transaction,
		as:              args.Account,
		is:              args.Identity,
		ledgerID:        args.LedgerID,
		equityAccountID: args.EquityAccID,
	}, nil
}

func (s *service) GetUserFundingSources(ctx context.Context, tx *sqlx.Tx, identityId string) ([]*fundingsources.FundingSource, error) {
	fs, err := s.fs.GetByIdentityId(ctx, tx, identityId)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

type NoopBankAccount struct {
	ID                 string
	Name               string
	Mask               string
	VerificationStatus string
	Type               string
}

type LinkBankAccountArgs struct {
	IdentityID    string `validate:"required,uuid4"`
	Name          string `validate:"required"`
	AccountNumber string `validate:"required"`
	RoutingNumber string `validate:"required"`
	Institution   string `validate:"required"`
	Type          string `validate:"required"`
}

func (s *service) LinkBankAccount(ctx context.Context, args *LinkBankAccountArgs) (*fundingsources.FundingSource, error) {
	var fundingsource *fundingsources.FundingSource
	err := crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		// TODO: decide how to store this. In jsonb column on fundingsources or different table per provider.
		providerAcc := NoopBankAccount{
			ID:                 uuid.NewString(),
			Name:               args.Name,
			Mask:               "****" + args.AccountNumber[:4],
			VerificationStatus: "verified",
			Type:               args.Type,
		}

		fs, err := s.fs.Create(ctx, tx, &fundingsources.CreateArgs{
			IdentityID:        args.IdentityID,
			Name:              args.Name,
			Mask:              providerAcc.Mask,
			VerificationState: "pending",
			Type:              "noop",
			TypeID:            providerAcc.ID,
			SubType:           args.Type,
		})
		if err != nil {
			return err
		}

		fundingsource = fs
		return nil
	})
	if err != nil {
		switch err.(type) {
		case *fundingsources.ErrInvalidArgument:
			return nil, &ErrInvalidArgument{Err: err.Error()}
		default:
			return nil, &ErrInternalError{Err: err.Error()}
		}
	}

	return fundingsource, nil
}

type VerifyArgs struct {
	IdentityID      string `validate:"required,uuid4"`
	FundingSourceID string `validate:"required,uuid4"`
}

func (s *service) VerifyBankAccount(ctx context.Context, args *VerifyArgs) (*fundingsources.FundingSource, error) {
	err := s.validator.Struct(args)
	if err != nil {
		return nil, &ErrInvalidArgument{Err: err.Error()}
	}

	var fundingsource *fundingsources.FundingSource
	err = crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		fs, err := s.fs.Verify(ctx, tx, &fundingsources.VerifyArgs{
			IdentityID:      args.IdentityID,
			FundingSourceID: args.FundingSourceID,
		})
		if err != nil {
			return err
		}

		fundingsource = fs
		return nil
	})
	if err != nil {
		switch err.(type) {
		case *fundingsources.ErrInvalidArgument:
			return nil, &ErrInvalidArgument{Err: err.Error()}
		default:
			return nil, &ErrInternalError{Err: err.Error()}
		}
	}

	return fundingsource, nil
}

type BankDepositArgs struct {
	IdentityID      string `validate:"required,uuid4"`
	FundingSourceID string `validate:"required,uuid4"`
	Amount          uint64 `validate:"gt=0"`
}

func (s *service) InitiateBankDeposit(ctx context.Context, args *BankDepositArgs) (*transactions.AccountTransaction, error) {
	var transaction *transactions.AccountTransaction
	err := crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		fundingSource, err := s.fs.Get(ctx, tx, args.FundingSourceID)
		if err != nil {
			switch err.(type) {
			case *fundingsources.ErrInvalidArgument:
				return &ErrInvalidArgument{Err: "Noop service:" + err.Error()}
			case *fundingsources.ErrNotFound:
				return &ErrNotFound{Err: "Noop service:" + err.Error()}
			default:
				return &ErrInternalError{Err: "Noop service:" + err.Error()}
			}
		}
		if fundingSource.IdentityID != args.IdentityID {
			return &ErrNotFound{Err: "Noop service: Funding source not found."}
		}
		if fundingSource.VerificationState != "verified" {
			return &ErrUnverifiedFundingSource{Err: "Noop service: Funding source is not verified."}
		}

		acc, err := s.as.GetByIdentityID(ctx, tx, args.IdentityID)
		if err != nil {
			switch err.(type) {
			case *accounts.ErrInvalidArgument:
				return &ErrInvalidArgument{Err: "Noop service:" + err.Error()}
			case *accounts.ErrNotFound:
				return &ErrNotFound{Err: "Noop service:" + err.Error()}
			default:
				return &ErrInternalError{Err: "Noop service:" + err.Error()}
			}
		}
		trx, err := s.ts.Create(ctx, tx, &transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Type:        "deposit",
			NetAmount:   args.Amount,
			Description: "Deposit from " + fundingSource.Name,
			State:       "completed", // TODO: define states
			LedgerTransfers: []transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        s.ledgerID,
					DebitAccountID:  acc.ID,
					CreditAccountID: s.equityAccountID,
					Amount:          args.Amount,
					// Code: "1", // TODO: define ledger transfer codes.
					Flags: transactions.LedgerTransferFlags{},
				},
			},
		})
		if err != nil {
			switch err.(type) {
			case *transactions.ErrInvalidArgument:
				return &ErrInvalidArgument{Err: "Noop service:" + err.Error()}
			case *transactions.ErrNotFound:
				return &ErrNotFound{Err: "Noop service:" + err.Error()}
			default:
				return &ErrInternalError{Err: "Noop service:" + err.Error()}
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

type BankWithdrawalArgs struct {
	IdentityID      string `validate:"required,uuid4"`
	FundingSourceID string `valdiate:"required,uuid4"`
	Amount          uint64 `validate:"required,gt=0"` // Min amount can be changed according to provider
}

func (s *service) InitiateBankWithdrawal(ctx context.Context, args *BankWithdrawalArgs) (*transactions.AccountTransaction, error) {
	var transaction *transactions.AccountTransaction
	err := crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		fundingSource, err := s.fs.Get(ctx, tx, args.FundingSourceID)
		if err != nil {
			switch err.(type) {
			case *fundingsources.ErrInvalidArgument:
				return &ErrInvalidArgument{Err: "Noop service:" + err.Error()}
			case *fundingsources.ErrNotFound:
				return &ErrNotFound{Err: "Noop service:" + err.Error()}
			default:
				return &ErrInternalError{Err: "Noop service:" + err.Error()}
			}
		}
		if fundingSource.IdentityID != args.IdentityID {
			return &ErrNotFound{Err: "Noop service: Funding source not found."}
		}
		if fundingSource.VerificationState != "verified" {
			return &ErrUnverifiedFundingSource{Err: "Noop service: Funding source is not verified."}
		}

		acc, err := s.as.GetByIdentityID(ctx, tx, args.IdentityID)
		if err != nil {
			switch err.(type) {
			case *accounts.ErrInvalidArgument:
				return &ErrInvalidArgument{Err: "Noop service:" + err.Error()}
			case *accounts.ErrNotFound:
				return &ErrNotFound{Err: "Noop service:" + err.Error()}
			default:
				return &ErrInternalError{Err: "Noop service:" + err.Error()}
			}
		}
		trx, err := s.ts.Create(ctx, tx, &transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Type:        "withdrawal",
			NetAmount:   args.Amount,
			Description: "Withdrawal to " + fundingSource.Name,
			State:       "completed", // TODO: define states
			LedgerTransfers: []transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        s.ledgerID,
					DebitAccountID:  s.equityAccountID,
					CreditAccountID: acc.ID,
					Amount:          args.Amount,
					// Code: "1", // TODO: define ledger transfer codes.
					Flags: transactions.LedgerTransferFlags{},
				},
			},
		})
		if err != nil {
			switch err.(type) {
			case *transactions.ErrInvalidArgument:
				return &ErrInvalidArgument{Err: "Noop service:" + err.Error()}
			case *transactions.ErrNotFound:
				return &ErrNotFound{Err: "Noop service:" + err.Error()}
			case *transactions.ErrInvalidTransfers:
				return parseInvalidBankWithdrawalTransferErrors(
					err.(*transactions.ErrInvalidTransfers).TransferErrors,
				)
			default:
				return &ErrInternalError{Err: "Noop service:" + err.Error()}
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

// Filter out errors related to user account such as insufficient balance.
// Otherwise, this returns an internal error if the error is not related to the user's account e.g. fx.
func parseInvalidBankWithdrawalTransferErrors(transferErrors []*pacioli.EventError) error {
	if len(transferErrors) != 1 {
		panic("Noop service: There should be one ledger transfer error.")
	}

	err := transferErrors[0]
	switch err.Code {
	case tb_types.TransferExceedsCredits:
		return &ErrInsufficientBalance{Err: "Noop service: Insufficient balance to perform bank withdrawal."}
	default:
		return &ErrInternalError{Err: fmt.Sprintf("Noop service: Unable to perform bank withdrawal due to TigerBeetle error: %d ", err.Code)}
	}
}

type OutgoingPaymentArgs struct {
	IdentityID string `validate:"required,uuid4"`
	Amount     uint64 `validate:"required,gt=0"`
	To         string `validate:"required"`
}

func (s *service) InitiateOutgoingPayment(ctx context.Context, args *OutgoingPaymentArgs) (*transactions.AccountTransaction, error) {
	err := s.validator.Struct(args)
	if err != nil || !isValidIlpAddress(args.To) {
		return nil, &ErrInvalidArgument{Err: "Noop service: " + err.Error()}
	}

	var outgoingPayment *transactions.AccountTransaction
	err = crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		identity, err := s.is.Get(ctx, tx, args.IdentityID)
		if err != nil {
			return &ErrInvalidArgument{Err: "Noop service: Identity not found."}
		}
		// TODO: the status consts need to be part of this provider
		if identity.VerificationState != _noop.Verified {
			return &ErrUnverifiedIdentity{Err: "Noop service: Identity is unverified."}
		}
		acc, err := s.as.GetByIdentityID(ctx, tx, args.IdentityID)
		if err != nil {
			switch err.(type) {
			case *accounts.ErrInvalidArgument:
				return &ErrInvalidArgument{Err: "Noop service:" + err.Error()}
			case *accounts.ErrNotFound:
				return &ErrNotFound{Err: "Noop service:" + err.Error()}
			default:
				return &ErrInternalError{Err: "Noop service:" + err.Error()}
			}
		}

		transaction, err := s.ts.Create(ctx, tx, &transactions.CreateTransactionArgs{
			AccountID:   acc.ID,
			Description: "Sent to " + args.To,
			Type:        "outgoingPayment",
			NetAmount:   args.Amount,
			State:       "completed", // noop provider provides instant payments :)
			LedgerTransfers: []transactions.CreateLedgerTransferArgs{
				{
					LedgerID:        s.ledgerID,
					DebitAccountID:  s.equityAccountID,
					CreditAccountID: acc.ID,
					Amount:          args.Amount,
					// Code: uint16,
					Flags: transactions.LedgerTransferFlags{},
				},
			},
		})
		if err != nil {
			switch err.(type) {
			case *transactions.ErrInvalidArgument:
				return &ErrInvalidArgument{Err: "Noop service:" + err.Error()}
			case *transactions.ErrNotFound:
				return &ErrNotFound{Err: "Noop service:" + err.Error()}
			case *transactions.ErrInvalidTransfers:
				return parseInvalidOutgoingPaymentErrors(err.(*transactions.ErrInvalidTransfers).TransferErrors)
			default:
				return &ErrInternalError{Err: "Noop service: " + err.Error()}
			}
		}
		outgoingPayment = transaction

		return nil
	})
	if err != nil {
		return nil, err
	}

	return outgoingPayment, nil
}

// Filter out errors related to user account such as insufficient balance.
// Otherwise, this returns an internal error if the error is not related to the user's account e.g. fx.
func parseInvalidOutgoingPaymentErrors(transferErrors []*pacioli.EventError) error {
	if len(transferErrors) != 1 {
		panic("Noop service: There should be one ledger transfer error for an outgoing payment.")
	}

	err := transferErrors[0]
	switch err.Code {
	case tb_types.TransferExceedsCredits:
		return &ErrInsufficientBalance{Err: "Noop service: Insufficient balance to perform outgoing payment."}
	default:
		return &ErrInternalError{Err: fmt.Sprintf("Noop service: Unable to perform outgoing payment due to TigerBeetle error: %d ", err.Code)}
	}
}

func (s *service) GetEquityAccountID() string {
	return s.equityAccountID
}

func (s *service) GetLedgerID() string {
	return s.ledgerID
}

func isValidIlpAddress(address string) bool {
	return strings.HasPrefix(address, "$")
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

type ErrInsufficientBalance struct {
	Err string
}

func (e ErrInsufficientBalance) Error() string {
	return e.Err
}

type ErrUnverifiedIdentity struct {
	Err string
}

func (e ErrUnverifiedIdentity) Error() string {
	return e.Err
}
