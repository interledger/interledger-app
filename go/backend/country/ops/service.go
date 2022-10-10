package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/country"
)

func Get(ctx context.Context, b Backends, id string) (*country.Country, error) {
	if id == "" {
		return nil, fmt.Errorf("%w ID is required.", country.ErrInvalidArgument)
	}

	var ret country.Country
	err := b.DB().GetContext(ctx, &ret, "SELECT * FROM countries WHERE id=$1 LIMIT 1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, country.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", country.ErrInternal, err.Error())
	}

	return &ret, nil
}

func GetAll(ctx context.Context, b Backends) ([]country.Country, error) {
	var ret []country.Country
	err := b.DB().SelectContext(ctx, &ret, "SELECT * FROM countries ORDER BY name ASC;")
	if err != nil {
		return nil, err
	}

	return ret, nil
}
