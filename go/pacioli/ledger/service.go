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
}

type TransferFlags = tigerbeetleTypes.TransferFlags

type Service interface {
	GetLedger(ledgerID string) (*Ledger, error)
	CreateLedger(name string, code uint16) (*Ledger, error)
	CreateAccounts(ledgerID string, args []CreateAccountArgs) ([]tigerbeetleTypes.EventResult, error)
	GetAccounts(ledgerID string, accountIDs []string) ([]Account, error)
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

func (s *service) CreateLedger(name string, code uint16) (*Ledger, error) {
	var ret Ledger
	stmt, err := s.db.PrepareNamed("INSERT INTO ledgers (name, code) VALUES (:name, :code) RETURNING *")
	if err != nil {
		return nil, err
	}

	err = stmt.Stmt.Get(&ret, name, code)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"ledgers_code_key\"") {
			return nil, ErrInvalidArg{Err: "Ledger Code must be unique."}
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
	// creditAccount, err := s.GetAccount(args.CreditAccountID)
	// if err != nil {
	// 	return nil, err
	// }

	// debitAccount, err := s.GetAccount(args.DebitAccountID)
	// if err != nil {
	// 	return nil, err
	// }

	// if creditAccount.LedgerID != debitAccount.LedgerID {
	// 	return nil, ErrCrossLedger{Err: "Accounts don't belong to the same ledger."}
	// }

	// // this will make sure that the tenant owns the ledger.
	// _, err = s.GetLedger(creditAccount.LedgerID)
	// if err != nil {
	// 	return nil, err
	// }

	// transferID := uuid.NewString()
	// tbTransferID, err := UuidToU128(transferID)
	// if err != nil {
	// 	return nil, err
	// }
	// tbDebitAccountID, err := UuidToU128(debitAccount.ID)
	// if err != nil {
	// 	return nil, err
	// }
	// tbCreditAccountID, err := UuidToU128(creditAccount.ID)
	// if err != nil {
	// 	return nil, err
	// }

	// eventErrors, err := s.tb.CreateTransfers([]tigerbeetleTypes.Transfer{
	// 	{
	// 		ID:              *tbTransferID,
	// 		DebitAccountID:  *tbDebitAccountID,
	// 		CreditAccountID: *tbCreditAccountID,
	// 		Amount:          args.Amount,
	// 		Flags:           args.Flags.ToUint32(),
	// 	},
	// })
	// if err != nil {
	// 	return nil, err
	// }
	// if len(eventErrors) != 0 {
	// 	result := eventErrors[0]
	// 	switch result.Code {
	// 	case tigerbeetleTypes.TransferExists:
	// 		return nil, ErrDuplicate{Err: "Transfer exists."}
	// 	// TODO: exhaustive switch
	// 	default:
	// 		return nil, errors.New(fmt.Sprintf("Failed to create transfer. tigerbeetle error code: %d", result.Code))
	// 	}
	// }

	return &Transfer{
		// ID:              transferID,
		// DebitAccountID:  debitAccount.ID,
		// CreditAccountID: creditAccount.ID,
		// Amount:          args.Amount,
		// Flags:           args.Flags,
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
