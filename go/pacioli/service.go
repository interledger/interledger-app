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

type Service interface {
	GetTenant(id string) (*Tenant, error)
	CreateTenant(identifier string) (*Tenant, error)
	GetAccountCategory(tenantID string, categoryID string) (*AccountCategory, error)
	CreateAccountCategory(tenantID string, args AccountCategoryArgs) (*AccountCategory, error)
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
func (s service) GetAccountCategory(tenantID string, id string) (*AccountCategory, error) {
	var category AccountCategory
	err := s.db.Get(
		&category,
		"SELECT * FROM account_categories WHERE id=$1 AND tenant_id=$2 LIMIT 1",
		id,
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
// - Name is required
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
	err = s.db.Get(&duplicate, "SELECT * FROM account_categories WHERE name=$1 AND tenant_id=$2",
		args.Name,
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
