package pacioli

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
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
	TenantID  string `db:"tenant_id"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type Account struct {
	ID              string
	LedgerID        string
	Unit            uint16
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
}

type TransferFlags = tigerbeetleTypes.TransferFlags

type Service interface {
	GetLedger(ledgerID string) (*Ledger, error)
	CreateLedger(name string) (*Ledger, error)
	CreateAccount(args CreateAccountArgs) (*Account, error)
	GetAccount(accountID string) (*Account, error)
	// The transfer api only allows creating a single non-two-phase transfer for now.
	// The underlying TB client supports batching, two-phase transfers and linking so this
	// API can be adjusted as we know more about the requirements.
	CreateTransfer(args CreateTransferArgs) (*Transfer, error)
}

type service struct {
	db *sqlx.DB
	tb tigerbeetle_go.Client
}

func NewLedgerService(db *sqlx.DB, tb tigerbeetle_go.Client) (Service, error) {
	return &service{db: db, tb: tb}, nil
}

func (s *service) CreateLedger(name string) (*Ledger, error) {
	var ret Ledger
	stmt, err := s.db.PrepareNamed("INSERT INTO ledgers (name, tenant_id) VALUES (:name) RETURNING *")
	if err != nil {
		return nil, err
	}

	err = stmt.Stmt.Get(&ret, name)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

// Will only return the ledger if it exists and belongs to the specified tenant. Otherwise will
// return ErrNotFound.
func (s service) GetLedger(id string) (*Ledger, error) {
	var ledger Ledger
	err := s.db.Get(&ledger, "SELECT * FROM ledgers WHERE id=$1 LIMIT 1", id)
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
	LedgerID string
	Code     uint16
	Unit     uint16
}

// Helper function to convert uuids into u128 needed for TigerBeetle IDs.
// TODO: see if there is a better way to do this.
func UuidToU128(value string) (*tigerbeetleTypes.Uint128, error) {
	src := strings.Replace(value, "-", "", -1)
	temp, err := hex.DecodeString(src)
	if err != nil {
		return nil, err
	}
	if len(temp) > 16 {
		return nil, errors.New("String could not be converted into uint128.")
	}

	return (*tigerbeetleTypes.Uint128)(temp), nil
}

// Helper function to extract the uuid we put into the u128.
func U128ToUuid(value tigerbeetleTypes.Uint128) string {
	s := hex.EncodeToString(value[:])
	ret := s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]

	return ret
}

// This function will create an account in TigerBeetle by sending a batch of 1 CreateAccount event.
func (s *service) CreateAccount(args CreateAccountArgs) (*Account, error) {
	ledger, err := s.GetLedger(args.LedgerID)
	if err != nil {
		return nil, err
	}

	accountID := uuid.NewString()
	tbAccID, err := UuidToU128(accountID)
	tbUserData, err := UuidToU128(ledger.ID)
	if err != nil {
		return nil, err
	}

	eventErrors, err := s.tb.CreateAccounts([]tigerbeetleTypes.Account{
		{
			ID:       *tbAccID,
			UserData: *tbUserData, // We store the ledgerID so that ACL can be applied when we lookup the account.
			Code:     args.Code,
			Unit:     args.Unit,
		},
	})
	// this error will be due to connection / io buffer issues
	if err != nil {
		return nil, err
	}

	if len(eventErrors) != 0 {
		result := eventErrors[0]
		switch result.Code {
		case tigerbeetleTypes.AccountExists:
			return nil, ErrDuplicate{Err: "Account exists."}
		// TODO: exhaustive switch
		default:
			return nil, errors.New(fmt.Sprintf("Failed to create account. tigerbeetle error code: %d", result.Code))
		}
	}

	acc, err := s.GetAccount(accountID)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *service) GetAccount(accountID string) (*Account, error) {
	tbAccID, err := UuidToU128(accountID)
	if err != nil {
		return nil, err
	}

	results, err := s.tb.LookupAccounts([]tigerbeetleTypes.Uint128{*tbAccID})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrNotFound{Err: "Account not found."}
	}

	// We do a check that the ID converted back to the UUID matches that sent in. If you get
	// this error then the way we convert the string to and from [16]uint8 is incorrect.
	parsedID := U128ToUuid(results[0].ID)
	if parsedID != accountID {
		return nil, errors.New("Failed to parse account ID correctly.")
	}

	ledger, err := s.GetLedger(U128ToUuid(results[0].UserData))
	if err != nil {
		return nil, err
	}

	return &Account{
		ID:              U128ToUuid(results[0].ID),
		LedgerID:        ledger.ID,
		Unit:            results[0].Unit,
		Code:            results[0].Code,
		DebitsReserved:  results[0].DebitsReserved,
		DebitsAccepted:  results[0].DebitsAccepted,
		CreditsReserved: results[0].CreditsReserved,
		CreditsAccepted: results[0].CreditsAccepted,
	}, nil
}

type CreateTransferArgs struct {
	Amount            uint64
	DebitAccountID    string
	CreditAccountID   string
	TransactionTypeID string
	Flags             TransferFlags
}

// The transfer api only allows creating a single non-two-phase transfer for now.
// The underlying TB client supports batching, two-phase transfers and linking so this
// API can be adjusted as we know more about the requirements.
func (s *service) CreateTransfer(args CreateTransferArgs) (*Transfer, error) {
	// TODO: function to get an array of accounts
	creditAccount, err := s.GetAccount(args.CreditAccountID)
	if err != nil {
		return nil, err
	}

	debitAccount, err := s.GetAccount(args.DebitAccountID)
	if err != nil {
		return nil, err
	}

	if creditAccount.LedgerID != debitAccount.LedgerID {
		return nil, ErrCrossLedger{Err: "Accounts don't belong to the same ledger."}
	}

	// this will make sure that the tenant owns the ledger.
	_, err = s.GetLedger(creditAccount.LedgerID)
	if err != nil {
		return nil, err
	}

	transferID := uuid.NewString()
	tbTransferID, err := UuidToU128(transferID)
	if err != nil {
		return nil, err
	}
	tbDebitAccountID, err := UuidToU128(debitAccount.ID)
	if err != nil {
		return nil, err
	}
	tbCreditAccountID, err := UuidToU128(creditAccount.ID)
	if err != nil {
		return nil, err
	}

	eventErrors, err := s.tb.CreateTransfers([]tigerbeetleTypes.Transfer{
		{
			ID:              *tbTransferID,
			DebitAccountID:  *tbDebitAccountID,
			CreditAccountID: *tbCreditAccountID,
			Amount:          args.Amount,
			Flags:           args.Flags.ToUint32(),
		},
	})
	if err != nil {
		return nil, err
	}
	if len(eventErrors) != 0 {
		result := eventErrors[0]
		switch result.Code {
		case tigerbeetleTypes.TransferExists:
			return nil, ErrDuplicate{Err: "Transfer exists."}
		// TODO: exhaustive switch
		default:
			return nil, errors.New(fmt.Sprintf("Failed to create transfer. tigerbeetle error code: %d", result.Code))
		}
	}

	return &Transfer{
		ID:              transferID,
		DebitAccountID:  debitAccount.ID,
		CreditAccountID: creditAccount.ID,
		Amount:          args.Amount,
		Flags:           args.Flags,
	}, nil
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

type ErrCrossLedger struct {
	Err string
}

func (s ErrCrossLedger) Error() string {
	return s.Err
}

type ErrInvalidTransfer struct {
	Err string
}

func (s ErrInvalidTransfer) Error() string {
	return s.Err
}
