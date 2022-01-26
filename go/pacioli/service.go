package pacioli

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Models
type Tenant struct {
	ID         string
	Identifier string
	CreatedAt  string `db:"created_at"`
	UpdatedAt  string `db:"updated_at"`
}

type Service interface {
	GetTenant(id string) (*Tenant, error)
	CreateTenant(identifier string) (*Tenant, error)
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
