package ops

import (
	"context"
	"database/sql"
	"fmt"

	"gitlab.com/fynbos/backend/country"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
)

func Get(ctx context.Context, b Backends, id string) (*country.Country, error) {
	if id == "" {
		return nil, fmt.Errorf("%w ID is required.", country.ErrInvalidArgument)
	}

	var ret country.Country
	err := b.DB().GetContext(ctx, &ret, "SELECT * FROM countries WHERE id=$1 LIMIT 1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, country.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", country.ErrInternal, err.Error())
	}

	return &ret, nil
}

func GetByAlpha2(ctx context.Context, b Backends, code string) (*country.Country, error) {
	if code == "" {
		return nil, fmt.Errorf("%w Code is required.", country.ErrInvalidArgument)
	}

	var ret country.Country
	err := b.DB().GetContext(ctx, &ret, "SELECT * FROM countries WHERE alpha_2=$1 LIMIT 1", code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, country.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", country.ErrInternal, err.Error())
	}

	return &ret, nil
}

func GetAll(ctx context.Context, b Backends) ([]*country.Country, error) {
	countries := []country.Country{}
	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		err := tx.Select(&countries, "SELECT * FROM countries ORDER BY name ASC")
		if err != nil {
			if err == sql.ErrNoRows {
				return country.ErrNotFound
			}

			return fmt.Errorf("%w %s", country.ErrInternal, err.Error())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ret := make([]*country.Country, len(countries))
	for i, trx := range countries {
		ret[i] = &country.Country{
			ID:          trx.ID,
			Name:        trx.Name,
			Alpha_2:     trx.Alpha_2,
			Alpha_3:     trx.Alpha_3,
			NumericCode: trx.NumericCode,
		}
	}

	return ret, nil
}
