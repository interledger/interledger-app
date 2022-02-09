package country

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Model
type Country struct {
	ID          string
	Name        string
	Alpha_2     string
	Alpha_3     string
	NumericCode uint16 `db:"numeric_code"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

type Service interface {
	Get(ctx context.Context, tx *sqlx.Tx, id string) (*Country, error)
	GetByAlpha2(ctx context.Context, tx *sqlx.Tx, code string) (*Country, error)
}

type service struct {
	db *sqlx.DB
}

func NewService() Service {
	return &service{}
}

func (self *service) Get(ctx context.Context, tx *sqlx.Tx, id string) (*Country, error) {
	if id == "" {
		return nil, &ErrInvalidArgument{Err: "ID is required."}
	}

	var ret Country
	err := tx.Get(&ret, "SELECT * FROM countries WHERE id=$1 LIMIT 1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{Err: "Not found."}
		}

		return nil, &ErrInternalError{Err: err.Error()}
	}

	return &ret, nil
}

func (self *service) GetByAlpha2(ctx context.Context, tx *sqlx.Tx, code string) (*Country, error) {
	if code == "" {
		return nil, &ErrInvalidArgument{Err: "code is required."}
	}

	var ret Country
	err := tx.Get(&ret, "SELECT * FROM countries WHERE alpha_2=$1 LIMIT 1", code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{Err: "Not found."}
		}

		return nil, &ErrInternalError{Err: err.Error()}
	}

	return &ret, nil
}

// Error set
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
