package country

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
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
	GetAll(ctx context.Context) ([]*Country, error)
}

type service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) Service {
	return &service{
		db: db,
	}
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

func (self *service) GetAll(ctx context.Context) ([]*Country, error) {
	countries := []Country{}
	err := crdbsqlx.ExecuteTx(ctx, self.db, nil, func(tx *sqlx.Tx) error {
		err := tx.Select(&countries, "SELECT * FROM countries ORDER BY name ASC")
		if err != nil {
			if err == sql.ErrNoRows {
				return &ErrNotFound{Err: "No countries found."}
			}

			return &ErrInternalError{Err: err.Error()}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ret := make([]*Country, len(countries))
	for i, trx := range countries {
		ret[i] = &Country{
			ID:          trx.ID,
			Name:        trx.Name,
			Alpha_2:     trx.Alpha_2,
			Alpha_3:     trx.Alpha_3,
			NumericCode: trx.NumericCode,
		}
	}

	return ret, nil
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
