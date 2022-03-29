package country

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound        = errors.New("country: not found.")
	ErrInvalidArgument = errors.New("country: invalid argument.")
	ErrInternal        = errors.New("country: internal error.")
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
	Get(ctx context.Context, id string) (*Country, error)
	GetByAlpha2(ctx context.Context, code string) (*Country, error)
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

func (s *service) Get(_ context.Context, id string) (*Country, error) {
	if id == "" {
		return nil, fmt.Errorf("%w ID is required.", ErrInvalidArgument)
	}

	var ret Country
	err := s.db.Get(&ret, "SELECT * FROM countries WHERE id=$1 LIMIT 1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return &ret, nil
}

func (s *service) GetByAlpha2(_ context.Context, code string) (*Country, error) {
	if code == "" {
		return nil, fmt.Errorf("%w Code is required.", ErrInvalidArgument)
	}

	var ret Country
	err := s.db.Get(&ret, "SELECT * FROM countries WHERE alpha_2=$1 LIMIT 1", code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return &ret, nil
}

func (self *service) GetAll(ctx context.Context) ([]*Country, error) {
	countries := []Country{}
	err := crdbsqlx.ExecuteTx(ctx, self.db, nil, func(tx *sqlx.Tx) error {
		err := tx.Select(&countries, "SELECT * FROM countries ORDER BY name ASC")
		if err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}

			return fmt.Errorf("%w %s", ErrInternal, err.Error())
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
