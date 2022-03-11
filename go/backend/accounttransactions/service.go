package account_transactions

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/proto/pacioli/v1"
	tb_types "gitlab.com/fynbos/tigerbeetle_go/pkg/types"
)

var (
	ErrInternal              = errors.New("account transactions: internal error.")
	ErrInvalidArgument       = errors.New("account transactions: invalid argument.")
	ErrNotFound              = errors.New("account transactions: not found.")
	ErrDuplicate             = errors.New("account transactions: duplicate.")
	ErrInvalidLedgerTransfer = errors.New("account transactions: ledger transfer failed.")
	ErrExceedsDebits         = errors.New("account transactions: exceeds debits.")
	ErrExceedsCredits        = errors.New("account transactions: exceeds credits.")
)

// This represents the transaction that a user has performed on their account.
// e.g. deposit, withdrawal, send etc. These HAVE to be backed by one or more ledger transfers
// that are stored in TigerBeetle.
type AccountTransaction struct {
	ID          string
	Type        string
	AccountID   string `db:"account_id"`
	Description string
	State       string

	// This is the net position change on the account from the user's point of view.
	NetAmount   int64    `db:"net_amount"`
	TransferIDs []string `db:"transfer_ids"` // The ids of the transfers that are stored in TigerBeetle.
	CreatedAt   string   `db:"created_at"`
	UpdatedAt   string   `db:"updated_at"`
}

type accountTransaction struct {
	ID          string
	Type        string
	AccountID   string `db:"account_id"`
	Description string
	State       string
	NetAmount   int64          `db:"net_amount"`
	TransferIDs pq.StringArray `db:"transfer_ids"`
	CreatedAt   string         `db:"created_at"`
	UpdatedAt   string         `db:"updated_at"`
}

type Service interface {
	Create(ctx context.Context, tx *sqlx.Tx, args *CreateTransactionArgs) (*AccountTransaction, error) // TODO: this should return an error array
	GetByAccount(ctx context.Context, tx *sqlx.Tx, args *GetByAccountArgs) ([]*AccountTransaction, error)
}

type service struct {
	validator *validator.Validate
	as        accounts.Service
	pacioli   pacioli.PacioliServiceClient
}

type ServiceArgs struct {
	AccountService accounts.Service             `validate:"required"`
	PacioliClient  pacioli.PacioliServiceClient `validate:"required"`
}

func NewService(args *ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	return &service{
		validator: validator,
		as:        args.AccountService,
		pacioli:   args.PacioliClient,
	}, nil
}

type LedgerTransferFlags struct { // duplicate of Pacioli.TransferFlags
	Linked         bool
	TwoPhaseCommit bool // TODO: update to latest TigerBeetle interface
	Condition      bool
}

// Arguments to create a transfer in TigerBeetle
type CreateLedgerTransferArgs struct {
	LedgerID        uint16
	DebitAccountID  string `validate:"required,uuid4"`
	CreditAccountID string `validate:"required,uuid4"`
	Amount          uint64 `validate:"required"`
	Code            uint16
	Flags           LedgerTransferFlags
}

type CreateTransactionArgs struct {
	AccountID       string `validate:"required,uuid4"`
	Description     string
	Type            string                     `validate:"oneof=deposit withdrawal outgoingPayment"`
	NetAmount       uint64                     `validate:"gt=0"`      // a uint64 as you can't have a negative deposit/withdrawal etc.
	State           string                     `validate:"required"`  // TODO: decide on transaction states
	LedgerTransfers []CreateLedgerTransferArgs `validate:"dive,gt=0"` // We assume an account transaction has to backed by at least one ledger transfer
}

// Calls out to Pacioli first and then inserts an account transaction into CRDB.
// TODO: this should return an error array
func (s *service) Create(ctx context.Context, tx *sqlx.Tx, args *CreateTransactionArgs) (*AccountTransaction, error) {
	err := s.validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	acc, err := s.as.Get(ctx, tx, args.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	ledgerTransfers := make([]*pacioli.Transfer, len(args.LedgerTransfers))
	transferIDs := make([]string, len(args.LedgerTransfers))
	for i, transfer := range args.LedgerTransfers {
		id := uuid.NewString()
		transferIDs[i] = id
		ledgerTransfers[i] = &pacioli.Transfer{
			Id:              id,
			DebitAccountId:  transfer.DebitAccountID,
			CreditAccountId: transfer.CreditAccountID,
			Amount:          transfer.Amount,
			Code:            uint32(transfer.Code),
			Flags: &pacioli.TransferFlags{
				Linked:         transfer.Flags.Linked,
				TwoPhaseCommit: transfer.Flags.TwoPhaseCommit,
				Condition:      transfer.Flags.Condition,
			},
			// TODO: add ledger id once TB supports it
		}
	}

	response, err := s.pacioli.CreateTransfers(ctx, &pacioli.CreateTransfersRequest{
		Transfers: ledgerTransfers,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	transferErrors := response.GetErrors()
	if len(transferErrors) > 0 {
		for _, err := range transferErrors {
			switch err.Code {
			case tb_types.TransferExceedsCredits:
				return nil, fmt.Errorf("%w %+v", ErrExceedsCredits, err)
			case tb_types.TransferExceedsDebits:
				return nil, fmt.Errorf("%w %+v", ErrExceedsDebits, err)
			default:
				return nil, fmt.Errorf("%w %+v", ErrInvalidLedgerTransfer, err)
			}
		}
	}

	stmt, err := tx.PrepareNamedContext(ctx, `INSERT INTO account_transactions
		(account_id, type, description, net_amount, state, transfer_ids) VALUES
		(:accountid, :type, :description, :netamount, :state, :transfer_ids)
		RETURNING *;
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
	}

	var transaction accountTransaction
	err = stmt.Stmt.Get(
		&transaction,
		acc.ID,
		args.Type,
		args.Description,
		args.NetAmount,
		args.State,
		pq.StringArray(transferIDs),
	)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
	}

	return &AccountTransaction{
		ID:          transaction.ID,
		Type:        transaction.Type,
		AccountID:   transaction.AccountID,
		Description: transaction.Description,
		State:       transaction.State,
		NetAmount:   transaction.NetAmount,
		TransferIDs: []string(transaction.TransferIDs),
		CreatedAt:   transaction.CreatedAt,
		UpdatedAt:   transaction.UpdatedAt,
	}, nil
}

type GetByAccountArgs struct {
	AccountID string `validate:"required,uuid4"`
	Limit     uint32 `validate:"gt=1"`
	OrderBy   string `validate:"oneof=ASC DESC"`
}

func (s *service) GetByAccount(ctx context.Context, tx *sqlx.Tx, args *GetByAccountArgs) ([]*AccountTransaction, error) {
	err := s.validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	transactions := []accountTransaction{}
	err = tx.SelectContext(ctx, &transactions,
		fmt.Sprintf("SELECT * FROM account_transactions WHERE account_id=$1 ORDER BY created_at %s LIMIT $2;", args.OrderBy),
		args.AccountID,
		args.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	ret := make([]*AccountTransaction, len(transactions))
	for i, trx := range transactions {
		ret[i] = &AccountTransaction{
			ID:          trx.ID,
			Type:        trx.Type,
			AccountID:   trx.AccountID,
			Description: trx.Description,
			State:       trx.State,
			NetAmount:   trx.NetAmount,
			TransferIDs: trx.TransferIDs,
			CreatedAt:   trx.CreatedAt,
			UpdatedAt:   trx.UpdatedAt,
		}
	}

	return ret, nil
}
