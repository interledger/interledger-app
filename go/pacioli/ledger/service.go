package pacioli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/tigerbeetle_go"
	tigerbeetleTypes "gitlab.com/fynbos/tigerbeetle_go/pkg/types"
)

var (
	ErrInvalidArg = errors.New("pacioli: invalid argument.")
	ErrNotFound   = errors.New("pacioli: not found.")
	ErrDuplicate  = errors.New("pacioli: duplicate.")
	ErrInternal   = errors.New("pacioli: internal error.")
)

// Models
type Ledger struct {
	ID        uint16 `json:"id"` // maps to TigerBeetle's `unit` (soon to be ledger) field on an account.
	Name      string `json:"name"`
	Asset     string `json:"string"`
	Scale     uint8  `json:"scale"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type Account struct {
	ID              string
	LedgerID        uint16
	Flags           AccountFlags
	Code            uint16
	DebitsReserved  uint64
	DebitsAccepted  uint64
	CreditsReserved uint64
	CreditsAccepted uint64
}

type Transfer struct {
	ID              string
	LedgerID        uint16 // this field is coming soon to a TigerBeetle near you.
	DebitAccountID  string
	CreditAccountID string
	Amount          uint64
	Flags           TransferFlags
	Code            uint32
}

type TransferFlags = tigerbeetleTypes.TransferFlags
type AccountFlags = tigerbeetleTypes.AccountFlags
type EventResult = tigerbeetleTypes.EventResult

type Service interface {
	CreateTenant(name string) error
	CreateLedgers(tenant string, args []CreateLedgerArgs) ([]EventResult, error)
	GetLedgers(tenant string, ledgerIDs []uint16) ([]Ledger, error)
	CreateAccounts(tenant string, args []CreateAccountArgs) ([]EventResult, error)
	GetAccounts(tenant string, accountIDs []string) ([]Account, error)
	CreateTransfers(tenant string, args []CreateTransferArgs) ([]EventResult, error)
	GetTransfers(tenant string, transferIDs []string) ([]Transfer, error)
}

type service struct {
	db        *sqlx.DB
	tb        tigerbeetle_go.Client
	validator *validator.Validate
}

func NewLedgerService(db *sqlx.DB, tb tigerbeetle_go.Client) (Service, error) {
	return &service{db: db, tb: tb, validator: validator.New()}, nil
}

func (s *service) CreateTenant(name string) error {
	return nil
}

type CreateLedgerArgs struct {
	ID    uint16 `validate:"required,uuid4"`
	Name  string
	Asset string
	Scale uint8
}

func (s *service) CreateLedgers(tenant string, args []CreateLedgerArgs) ([]EventResult, error) {
	// var ret Ledger
	// stmt, err := s.db.PrepareNamed(
	// 	`INSERT INTO ledgers (id, name, code, asset, scale)
	// 	VALUES (:id, :name, :asset, :scale) RETURNING *`,
	// )
	// if err != nil {
	// 	return nil, err
	// }

	// err = stmt.Stmt.Get(&ret, args.ID, args.Name, args.Asset, args.Scale)
	// if err != nil {
	// 	if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"ledgers_code_key\"") {
	// 		return nil, fmt.Errorf("Ledger exists. %w", ErrDuplicate)
	// 	}
	// 	return nil, err
	// }

	// return &ret, nil

	return nil, nil
}

// Will only return the ledger if it exists otherwise will return ErrNotFound.
func (s service) GetLedgers(tenant string, ids []uint16) ([]Ledger, error) {
	// ACL to be done here.

	// var ledger Ledger
	// err := s.db.Get(&ledger, "SELECT * FROM ledgers WHERE id=$1 LIMIT 1", id)
	// if err != nil {
	// 	switch err {
	// 	case sql.ErrNoRows:
	// 		return nil, ErrNotFound
	// 	default:
	// 		return nil, err
	// 	}
	// }

	// return &ledger, nil

	return nil, nil
}

type CreateAccountArgs struct {
	ID       string `validate:"required,uuid4"`
	LedgerID uint16
	Code     uint16
	Flags    AccountFlags
}

// Helper function to convert uuids into u128 needed for TigerBeetle IDs.
// TODO: see if there is a better way to do this.
func UuidToU128(value string) (*tigerbeetleTypes.Uint128, error) {
	src := strings.Replace(value, "-", "", -1)
	ret := new(tigerbeetleTypes.Uint128)
	bytesWritten, err := hex.Decode(ret[:], []byte(src))
	if err != nil {
		return nil, err
	}
	if bytesWritten > 16 {
		return nil, errors.New("String could not be converted into uint128.")
	}

	return ret, nil
}

// Helper function to extract the uuid we put into the u128.
func U128ToUuid(value tigerbeetleTypes.Uint128) string {
	s := hex.EncodeToString(value[:])
	ret := s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]

	return ret
}

func (s *service) CreateAccounts(tenant string, args []CreateAccountArgs) ([]EventResult, error) {
	// TODO: collect ledgers and perform ACL
	tbAccounts := make([]tigerbeetleTypes.Account, len(args))
	for i, acc := range args {
		err := s.validator.Struct(args)
		if err != nil {
			return nil, fmt.Errorf("%s %w", err.Error(), ErrInvalidArg)
		}

		tbAccID, err := UuidToU128(acc.ID)
		if err != nil {
			return nil, err
		}
		tbAccounts[i] = tigerbeetleTypes.Account{
			ID:   *tbAccID,
			Unit: acc.LedgerID,
			Code: acc.Code,
		}
	}

	eventErrors, err := s.tb.CreateAccounts(tbAccounts)
	// this error will be due to connection / io buffer issues
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
	}

	return eventErrors, nil
}

func (s *service) GetAccounts(tenant string, accountIDs []string) ([]Account, error) {
	tbAccIDs := make([]tigerbeetleTypes.Uint128, len(accountIDs))
	for _, id := range accountIDs {
		_, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("Account ID must be a uuid. %w", ErrInvalidArg)
		}
		accID, err := UuidToU128(id)
		if err != nil {
			return nil, err
		}

		tbAccIDs = append(tbAccIDs, *accID)
	}

	results, err := s.tb.LookupAccounts(tbAccIDs)
	if err != nil {
		return nil, err
	}

	ret := make([]Account, len(results))
	for i, result := range results {
		ret[i] = Account{
			ID:              U128ToUuid(result.ID),
			LedgerID:        result.Unit,
			Code:            result.Code,
			DebitsReserved:  result.DebitsReserved,
			DebitsAccepted:  result.DebitsAccepted,
			CreditsReserved: result.CreditsReserved,
			CreditsAccepted: result.CreditsAccepted,
		}
	}

	// TODO: ACL on accounts

	return ret, nil
}

type CreateTransferArgs struct {
	ID              string `validate:"required,uuid4"`
	Amount          uint64 `validate:"gt=0"`
	DebitAccountID  string `validate:"required,uuid4"`
	CreditAccountID string `validate:"required,uuid4"`
	Flags           TransferFlags
	Code            uint32
}

// TODO: Assuming that IDs are uuids and are being generated by another service for now. Might be better to be
// unopinionated and accept []byte.
func (s *service) CreateTransfers(tenant string, args []CreateTransferArgs) ([]EventResult, error) {
	// TODO: collect ledgers and perform ACL
	tbTransfers := make([]tigerbeetleTypes.Transfer, len(args))
	for i, transfer := range args {
		err := s.validator.Struct(transfer)
		if err != nil {
			return nil, fmt.Errorf("%s %w", err.Error(), ErrInvalidArg)
		}
		transferID, err := UuidToU128(transfer.ID)
		if err != nil {
			return nil, err
		}

		debitAccountID, err := UuidToU128(transfer.DebitAccountID)
		if err != nil {
			return nil, err
		}

		creditAccountID, err := UuidToU128(transfer.CreditAccountID)
		if err != nil {
			return nil, err
		}

		tbTransfers[i] = tigerbeetleTypes.Transfer{
			ID:              *transferID,
			DebitAccountID:  *debitAccountID,
			CreditAccountID: *creditAccountID,
			Amount:          transfer.Amount,
			Code:            transfer.Code,
		}
	}

	eventErrors, err := s.tb.CreateTransfers(tbTransfers)
	if err != nil {
		return nil, err
	}

	return eventErrors, nil
}

func (s *service) GetTransfers(tenant string, transferIDs []string) ([]Transfer, error) {
	tbTransferIDs := make([]tigerbeetleTypes.Uint128, len(transferIDs))
	for i, id := range transferIDs {
		_, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("Transfer ID must be a uuid. %w", ErrInvalidArg)
		}
		transferID, err := UuidToU128(id)
		if err != nil {
			return nil, err
		}

		tbTransferIDs[i] = *transferID
	}

	results, err := s.tb.LookupTransfers(tbTransferIDs)
	if err != nil {
		return nil, err
	}

	ret := make([]Transfer, len(results))
	for i, transfer := range results {
		ret[i] = Transfer{
			ID:              U128ToUuid(transfer.ID),
			DebitAccountID:  U128ToUuid(transfer.DebitAccountID),
			CreditAccountID: U128ToUuid(transfer.CreditAccountID),
			Amount:          transfer.Amount,
			Code:            transfer.Code,
		}
	}

	// TODO: ACL on ledgers involved

	return ret, nil
}
