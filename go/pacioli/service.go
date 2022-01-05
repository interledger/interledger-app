package pacioli

import (
	"database/sql"
	"regexp"

	"github.com/jmoiron/sqlx"
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

type Service interface {
	GetTenant(id string) (*Tenant, error)
	CreateTenant(identifier string) (*Tenant, error)
	GetAccountCategoryByCode(tenantID string, code uint16) (*AccountCategory, error)
	CreateAccountCategory(tenantID string, args AccountCategoryArgs) (*AccountCategory, error)
	GetTransactionType(tenantID string, transactionTypeID string) (*TransactionType, error)
	CreateTransactionType(tenantID string, args TransactionTypeArgs) (*TransactionType, error)
}

type service struct {
	db *sqlx.DB
}

func NewPacioliService(db *sqlx.DB) (Service, error) {
	return &service{db: db}, nil
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

// Error set
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
