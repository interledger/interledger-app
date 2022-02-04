package pacioli

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/tigerbeetle_go"
	tigerbeetleTypes "gitlab.com/fynbos/tigerbeetle_go/pkg/types"
)

// Models

type Ledger struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      uint16 `json:"code"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type Account struct {
	ID              string
	LedgerCode      uint16
	Code            uint16
	DebitsReserved  uint64
	DebitsAccepted  uint64
	CreditsReserved uint64
	CreditsAccepted uint64
}

type Transfer struct {
	ID              string
	DebitAccountID  string
	CreditAccountID string
	Amount          uint64
	Flags           TransferFlags
	Code            uint32
}

type TransferFlags = tigerbeetleTypes.TransferFlags

type Service interface {
	GetLedger(ledgerID string) (*Ledger, error)
	GetLedgerByCode(code uint16) (*Ledger, error)
	CreateLedger(name string, code uint16) (*Ledger, error)
	CreateAccounts(ledgerID string, args []CreateAccountArgs) ([]tigerbeetleTypes.EventResult, error)
	GetAccounts(ledgerID string, accountIDs []string) ([]Account, error)
	CreateTransfers(ledgerID string, args []CreateTransferArgs) ([]tigerbeetleTypes.EventResult, error)
	GetTransfers(ledgerID string, transferIDs []string) ([]Transfer, error)
}

type service struct {
	db *sqlx.DB
	tb tigerbeetle_go.Client
}

func NewLedgerService(db *sqlx.DB, tb tigerbeetle_go.Client) (Service, error) {
	return &service{db: db, tb: tb}, nil
}

func (s *service) CreateLedger(name string, code uint16) (*Ledger, error) {
	var ret Ledger
	stmt, err := s.db.PrepareNamed("INSERT INTO ledgers (name, code) VALUES (:name, :code) RETURNING *")
	if err != nil {
		return nil, err
	}

	err = stmt.Stmt.Get(&ret, name, code)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"ledgers_code_key\"") {
			return nil, ErrDuplicate{Err: "Ledger exists."}
		}
		return nil, err
	}

	return &ret, nil
}

// Will only return the ledger if it exists otherwise will return ErrNotFound.
func (s service) GetLedger(id string) (*Ledger, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidArg{Err: "Ledger ID must be a uuid."}
	}

	var ledger Ledger
	err = s.db.Get(&ledger, "SELECT * FROM ledgers WHERE id=$1 LIMIT 1", id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound{
				Err: "Ledger not found.",
			}
		default:
			return nil, err
		}
	}

	return &ledger, nil
}

// Will only return the ledger if it exists otherwise will return ErrNotFound.
func (s service) GetLedgerByCode(code uint16) (*Ledger, error) {
	var ledger Ledger
	err := s.db.Get(&ledger, "SELECT * FROM ledgers WHERE code=$1 LIMIT 1", code)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound{
				Err: "Ledger not found.",
			}
		default:
			return nil, err
		}
	}

	return &ledger, nil
}

type CreateAccountArgs struct {
	ID   string
	Code uint16
}

func validateCreateAccountArgs(account CreateAccountArgs) error {
	_, err := uuid.Parse(account.ID)
	if err != nil {
		return ErrInvalidArg{Err: "Account ID must be a valid uuid."}
	}

	return nil
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

func (s *service) CreateAccounts(ledgerID string, args []CreateAccountArgs) ([]tigerbeetleTypes.EventResult, error) {
	ledger, err := s.GetLedger(ledgerID)
	if err != nil {
		return nil, err
	}

	tbAccounts := make([]tigerbeetleTypes.Account, len(args))
	for i, acc := range args {
		err := validateCreateAccountArgs(acc)
		if err != nil {
			return nil, err
		}

		tbAccID, err := UuidToU128(acc.ID)
		if err != nil {
			return nil, err
		}
		tbAccounts[i] = tigerbeetleTypes.Account{
			ID:   *tbAccID,
			Unit: ledger.Code,
			Code: acc.Code,
		}
	}

	eventErrors, err := s.tb.CreateAccounts(tbAccounts)
	// this error will be due to connection / io buffer issues
	if err != nil {
		return nil, err
	}

	return eventErrors, nil
}

func (s *service) GetAccounts(ledgerID string, accountIDs []string) ([]Account, error) {
	// make sure ledger exists. In future, we will be able to use this with TBs query language to look
	// for accounts in the specified ledger.
	_, err := s.GetLedger(ledgerID)
	if err != nil {
		return nil, err
	}

	tbAccIDs := make([]tigerbeetleTypes.Uint128, len(accountIDs))
	for _, id := range accountIDs {
		_, err := uuid.Parse(id)
		if err != nil {
			return nil, ErrInvalidArg{Err: "Account id must be a uuid."}
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
			LedgerCode:      result.Unit,
			Code:            result.Code,
			DebitsReserved:  result.DebitsReserved,
			DebitsAccepted:  result.DebitsAccepted,
			CreditsReserved: result.CreditsReserved,
			CreditsAccepted: result.CreditsAccepted,
		}
	}

	return ret, nil
}

type CreateTransferArgs struct {
	ID              string
	Amount          uint64
	DebitAccountID  string
	CreditAccountID string
	Flags           TransferFlags
	Code            uint32
}

func validateCreateTransferArgs(transfer CreateTransferArgs) error {
	_, err := uuid.Parse(transfer.ID)
	if err != nil {
		return ErrInvalidArg{Err: "Transfer ID must be a valid uuid."}
	}

	_, err = uuid.Parse(transfer.DebitAccountID)
	if err != nil {
		return ErrInvalidArg{Err: "Transfer DebitAccountID must be a valid uuid."}
	}

	_, err = uuid.Parse(transfer.CreditAccountID)
	if err != nil {
		return ErrInvalidArg{Err: "Transfer CreditAccountID must be a valid uuid."}
	}

	return nil
}

// TODO: Assuming that IDs are uuids and are being generated by another service for now. Might be better to be
// unopinionated and accept []byte.
func (s *service) CreateTransfers(ledgerID string, args []CreateTransferArgs) ([]tigerbeetleTypes.EventResult, error) {
	// This will be used when TB supports having a ledgerCode applied to a transfer.
	_, err := s.GetLedger(ledgerID)
	if err != nil {
		return nil, err
	}

	tbTransfers := make([]tigerbeetleTypes.Transfer, len(args))
	for i, transfer := range args {
		err = validateCreateTransferArgs(transfer)
		if err != nil {
			return nil, err
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

func (s *service) GetTransfers(ledgerID string, transferIDs []string) ([]Transfer, error) {
	// make sure ledger exists. In future, we will be able to use this with TBs query language to look
	// for transfers in the specified ledger.
	_, err := s.GetLedger(ledgerID)
	if err != nil {
		return nil, err
	}

	tbTransferIDs := make([]tigerbeetleTypes.Uint128, len(transferIDs))
	for i, id := range transferIDs {
		_, err := uuid.Parse(id)
		if err != nil {
			return nil, ErrInvalidArg{Err: "Transfer id must be a uuid."}
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

	return ret, nil
}

// Error setargs.F
type ErrInvalidArg struct {
	Err string
}

func (s ErrInvalidArg) Error() string {
	return s.Err
}

type ErrDuplicate struct {
	Err string
}

func (s ErrDuplicate) Error() string {
	return s.Err
}

type ErrNotFound struct {
	Err string
}

func (s ErrNotFound) Error() string {
	return s.Err
}
