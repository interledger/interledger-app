package identity

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/user"
)

// Model
type Identity struct {
	ID        string
	Email     string
	Country   string
	LegalName string `db:"legal_name"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type Service interface {
	Create(ctx context.Context, tx *sqlx.Tx, args CreateArgs) (*Identity, error)
	Get(ctx context.Context, tx *sqlx.Tx, id string) (*Identity, error)
}

type service struct{}

func NewService() (Service, error) {
	return &service{}, nil
}

type CreateArgs struct {
	Country   string
	LegalName string
	User      *user.User
}

func validateCreateArgs(args CreateArgs) error {
	if args.User == nil {
		return &ErrInvalidArgument{Err: "User is required."}
	}

	if args.User.ID == "" {
		return &ErrInvalidArgument{Err: "User ID is required."}
	}

	if args.User.Email == "" {
		return &ErrInvalidArgument{Err: "User Email is required."}
	}

	if args.LegalName == "" {
		return &ErrInvalidArgument{Err: "LegalName is required."}
	}

	if args.Country == "" {
		return &ErrInvalidArgument{Err: "Country is required."}
	}

	return nil
}

// There is a 1-1 mapping between the identity and user stored in Kratos. The
// Kratos ID is used as the identity ID.
func (self *service) Create(ctx context.Context, tx *sqlx.Tx, args CreateArgs) (*Identity, error) {
	err := validateCreateArgs(args)
	if err != nil {
		return nil, err
	}

	var ret Identity
	stmt, err := tx.PrepareNamed("INSERT INTO identities (id, legal_name, country, email) VALUES (:id, :legalname, :country, :email) RETURNING *")
	if err != nil {
		return nil, &ErrInternalError{Err: err.Error()}
	}

	err = stmt.Stmt.Get(&ret, args.User.ID, args.LegalName, args.Country, args.User.Email)
	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"primary\"") {
			return nil, &ErrDuplicate{Err: "Identity exists."}
		}
		return nil, &ErrInternalError{Err: err.Error()}
	}

	return &ret, nil
}

func (self service) Get(ctx context.Context, tx *sqlx.Tx, id string) (*Identity, error) {
	if id == "" {
		return nil, &ErrInvalidArgument{Err: "ID is required."}
	}

	var ret Identity
	err := tx.Get(&ret, "SELECT * FROM identities WHERE id=$1 LIMIT 1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{Err: "Not found."}
		}

		return nil, &ErrInternalError{Err: err.Error()}
	}

	return &ret, nil
}

// Error set
// TODO: wrapping errors instead to preserve stack.
type ErrInvalidArgument struct {
	Err string
}

func (r *ErrInvalidArgument) Error() string {
	return r.Err
}

type ErrInternalError struct {
	Err string
}

func (r *ErrInternalError) Error() string {
	return r.Err
}

type ErrNotFound struct {
	Err string
}

func (r *ErrNotFound) Error() string {
	return r.Err
}

type ErrDuplicate struct {
	Err string
}

func (r *ErrDuplicate) Error() string {
	return r.Err
}
