package pacioli

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/tigerbeetle_go"
	tigerbeetleTypes "gitlab.com/fynbos/tigerbeetle_go/pkg/types"
)

// Models
type Tenant struct {
	ID         string
	Identifier string
	CreatedAt  string `db:"created_at"`
	UpdatedAt  string `db:"updated_at"`
}

type AccountCategory struct {
	ID          string
	TenantID    string `db:"tenant_id"`
	Name        string
	Type        string
	Description string
	Code        uint16
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

type TransactionType struct {
	ID                        string
	TenantID                  string `db:"tenant_id"`
	Name                      string
	Description               string
	CreditAccountCategoryCode uint16 `db:"credit_account_category_code"`
	DebitAccountCategoryCode  uint16 `db:"debit_account_category_code"`
	CreatedAt                 string `db:"created_at"`
	UpdatedAt                 string `db:"updated_at"`
}

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
	ID                string
	DebitAccountID    string
	CreditAccountID   string
	TransactionTypeID string
	Amount            uint64
	Flags             TransferFlags
}

type TransferFlags = tigerbeetleTypes.TransferFlags

type Service interface {
	GetTenant(id string) (*Tenant, error)
	CreateTenant(identifier string) (*Tenant, error)
	GetAccountCategoryByCode(tenantID string, code uint16) (*AccountCategory, error)
	CreateAccountCategory(tenantID string, args AccountCategoryArgs) (*AccountCategory, error)
	GetTransactionType(tenantID string, transactionTypeID string) (*TransactionType, error)
	CreateTransactionType(tenantID string, args TransactionTypeArgs) (*TransactionType, error)
	GetLedger(tenantID string, ledgerID string) (*Ledger, error)
	CreateLedger(tenantID string, name string) (*Ledger, error)
	CreateAccount(tenantID string, args CreateAccountArgs) (*Account, error)
	GetAccount(tenantID string, accountID string) (*Account, error)
	// The transfer api only allows creating a single non-two-phase transfer for now.
	// The underlying TB client supports batching, two-phase transfers and linking so this
	// API can be adjusted as we know more about the requirements.
	CreateTransfer(tenantID string, args CreateTransferArgs) (*Transfer, error)
}

type service struct {
	db *sqlx.DB
	tb tigerbeetle_go.Client
}

func NewPacioliService(db *sqlx.DB, tb tigerbeetle_go.Client) (Service, error) {
	return &service{db: db, tb: tb}, nil
}

func (s *service) CreateTenant(identifier string) (*Tenant, error) {
	if identifier == "" {
		return nil, ErrInvalidArg{Err: "Identifier is required."}
	}

	duplicate := Tenant{
		ID: "-1",
	}
	err := s.db.Get(&duplicate, "SELECT * FROM tenants WHERE identifier=$1", identifier)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			// no duplicate found.
		default:
			return nil, err
		}
	}
	if duplicate.ID != "-1" {
		return nil, ErrDuplicate{Err: "Duplicate tenant."}
	}

	var ret Tenant
	stmt, err := s.db.PrepareNamed("INSERT INTO tenants (identifier) VALUES (:identifier) RETURNING *")
	if err != nil {
		return nil, err
	}

	err = stmt.Stmt.Get(&ret, identifier)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s service) GetTenant(id string) (*Tenant, error) {
	var ret Tenant
	err := s.db.Get(&ret, "SELECT * FROM tenants WHERE id=$1 LIMIT 1", id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound{Err: "Tenant not found."}
		default:
			return nil, err
		}
	}

	return &ret, nil
}

type AccountCategoryArgs struct {
	Name        string
	Type        string
	Description string
	Code        uint16
}

// Will only return the account category if it exists and belongs to the specified tenant. Otherwise
// will return ErrNotFound.
func (s service) GetAccountCategoryByCode(tenantID string, code uint16) (*AccountCategory, error) {
	var category AccountCategory
	err := s.db.Get(
		&category,
		"SELECT * FROM account_categories WHERE code=$1 AND tenant_id=$2 LIMIT 1",
		code,
		tenantID,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound{
				Err: "Account category not found.",
			}
		default:
			return nil, err
		}
	}

	return &category, nil
}

// Creates a uniquely named account category for the specified LedgerID.
// - Type matches ^(ASSET|EQUITY|LIABILITY)$
// - Code is unique
func (s *service) CreateAccountCategory(tenantID string, args AccountCategoryArgs) (*AccountCategory, error) {
	err := validateAccountCategoryArgs(args)
	if err != nil {
		return nil, err
	}

	tenant, err := s.GetTenant(tenantID)
	if err != nil {
		return nil, err
	}

	duplicate := AccountCategory{
		ID: "-1",
	}
	err = s.db.Get(&duplicate, "SELECT * FROM account_categories WHERE code=$1 AND tenant_id=$2",
		args.Code,
		tenant.ID,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			// no duplicate found.
		default:
			return nil, err
		}
	}
	if duplicate.ID != "-1" {
		return nil, ErrDuplicate{Err: "Duplicate account category."}
	}

	var ret AccountCategory
	stmt, err := s.db.PrepareNamed(
		"INSERT INTO account_categories (name, tenant_id, description, type, code)" +
			"VALUES (:name, :tenantid, :description, :type, :code) RETURNING *;",
	)
	if err != nil {
		return nil, err
	}

	err = stmt.Stmt.Get(&ret,
		args.Name,
		tenant.ID,
		args.Description,
		args.Type,
		args.Code,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func validateAccountCategoryArgs(args AccountCategoryArgs) error {
	match, err := regexp.Match("^(ASSET|EQUITY|LIABILITY)$", []byte(args.Type))
	if err != nil || !match {
		return ErrInvalidArg{Err: "Type must be one of ASSET | LIABILITY | EQUITY."}
	}

	if args.Name == "" {
		return ErrInvalidArg{Err: "Name is required."}
	}

	return nil
}

type TransactionTypeArgs struct {
	Name                      string
	Description               string
	CreditAccountCategoryCode uint16
	DebitAccountCategoryCode  uint16
}

func (s service) GetTransactionType(tenantID string, transactionTypeID string) (*TransactionType, error) {
	var ret TransactionType
	err := s.db.Get(
		&ret,
		"SELECT * FROM transaction_types WHERE tenant_id=$1 AND id=$2 LIMIT 1",
		tenantID,
		transactionTypeID,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound{
				Err: "Transaction type not found.",
			}
		default:
			return nil, err
		}
	}

	return &ret, nil
}

func (s *service) CreateTransactionType(tenantID string, args TransactionTypeArgs) (*TransactionType, error) {
	tenant, err := s.GetTenant(tenantID)
	if err != nil {
		return nil, err
	}

	duplicate := TransactionType{
		ID: "-1",
	}
	err = s.db.Get(
		&duplicate,
		"SELECT * FROM transaction_types WHERE tenant_id=$1 AND name=$2 LIMIT 1",
		tenant.ID,
		args.Name,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			// no duplicate
		default:
			return nil, err
		}
	}
	if duplicate.ID != "-1" {
		return nil, ErrDuplicate{Err: "Duplicate transaction type."}
	}

	debitAccountCategory, err := s.GetAccountCategoryByCode(tenant.ID, args.DebitAccountCategoryCode)
	if err != nil {
		return nil, err
	}

	creditAccountCategory, err := s.GetAccountCategoryByCode(tenant.ID, args.CreditAccountCategoryCode)
	if err != nil {
		return nil, err
	}

	if creditAccountCategory.ID == debitAccountCategory.ID {
		return nil, ErrInvalidArg{Err: "Account category codes must be different."}
	}

	var ret TransactionType
	stmt, err := s.db.PrepareNamed(
		"INSERT INTO transaction_types (name, tenant_id, description, credit_account_category_code, debit_account_category_code)" +
			"VALUES (:name, :tenantid, :description, :credit_account_category_code, :debit_account_category_code) RETURNING *;",
	)
	if err != nil {
		return nil, err
	}

	err = stmt.Stmt.Get(&ret,
		args.Name,
		tenant.ID,
		args.Description,
		creditAccountCategory.Code,
		debitAccountCategory.Code,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *service) CreateLedger(tenantID string, name string) (*Ledger, error) {
	tenant, err := s.GetTenant(tenantID)
	if err != nil {
		return nil, err
	}

	var ret Ledger
	stmt, err := s.db.PrepareNamed("INSERT INTO ledgers (name, tenant_id) VALUES (:name, :tenantid) RETURNING *")
	if err != nil {
		return nil, err
	}

	err = stmt.Stmt.Get(&ret, name, tenant.ID)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

// Will only return the ledger if it exists and belongs to the specified tenant. Otherwise will
// return ErrNotFound.
func (s service) GetLedger(tenantID string, id string) (*Ledger, error) {
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

	if ledger.TenantID != tenantID {
		return nil, ErrNotFound{
			Err: "Ledger not found.",
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
func (s *service) CreateAccount(tenantID string, args CreateAccountArgs) (*Account, error) {
	ledger, err := s.GetLedger(tenantID, args.LedgerID)
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

	acc, err := s.GetAccount(tenantID, accountID)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *service) GetAccount(tenantID string, accountID string) (*Account, error) {
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

	ledger, err := s.GetLedger(tenantID, U128ToUuid(results[0].UserData))
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
func (s *service) CreateTransfer(tenantID string, args CreateTransferArgs) (*Transfer, error) {
	// TODO: function to get an array of accounts
	creditAccount, err := s.GetAccount(tenantID, args.CreditAccountID)
	if err != nil {
		return nil, err
	}

	debitAccount, err := s.GetAccount(tenantID, args.DebitAccountID)
	if err != nil {
		return nil, err
	}

	if creditAccount.LedgerID != debitAccount.LedgerID {
		return nil, ErrCrossLedger{Err: "Accounts don't belong to the same ledger."}
	}

	// this will make sure that the tenant owns the ledger.
	_, err = s.GetLedger(tenantID, creditAccount.LedgerID)
	if err != nil {
		return nil, err
	}

	transactionType, err := s.GetTransactionType(tenantID, args.TransactionTypeID)
	if err != nil {
		return nil, err
	}

	if transactionType.CreditAccountCategoryCode != creditAccount.Code {
		return nil, ErrInvalidTransfer{Err: "Incorrect credit account category for transfer."}
	}

	if transactionType.DebitAccountCategoryCode != debitAccount.Code {
		return nil, ErrInvalidTransfer{Err: "Incorrect debit account category for transfer."}
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
		ID:                transferID,
		DebitAccountID:    debitAccount.ID,
		CreditAccountID:   creditAccount.ID,
		TransactionTypeID: transactionType.ID,
		Amount:            args.Amount,
		Flags:             args.Flags,
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
