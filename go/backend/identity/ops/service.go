package ops

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	identity "gitlab.com/fynbos/backend/identity"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	_country "gitlab.com/fynbos/backend/country"
)

// DB Model
type identityModel struct {
	identity.Identity
	CountryID string `db:"country_id"` // primary key of country
}

// There is a 1-1 mapping between the identityModel and user stored in Kratos. The
// Kratos ID is used as the identityModel ID.
func Create(ctx context.Context, b Backends, args *identity.CreateArgs) (*identity.Identity, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s.", identity.ErrInvalidArgument, err)
	}

	c, err := b.Countries().GetByAlpha2(ctx, args.Country)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identity.ErrInternal, err)
	}

	var id identityModel
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		stmt, err := tx.PrepareNamed(`
			INSERT INTO identities (
				id, first_name, last_name, mobile_number, email, country_id
			)
			VALUES (:id, :first_name, :last_name, :mobile_number, :email, :country_id)
			RETURNING *;
		`)
		if err != nil {
			return fmt.Errorf("%w %s", identity.ErrInternal, err)
		}

		err = stmt.Stmt.Get(&id,
			args.ID,
			args.FirstName,
			args.LastName,
			args.MobileNumber,
			args.Email,
			c.ID,
		)
		if err != nil {
			if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"primary\"") {
				return identity.ErrDuplicate
			}
			return fmt.Errorf("%w %s", identity.ErrInternal, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &identity.Identity{
		ID:           id.ID,
		FirstName:    id.FirstName,
		LastName:     id.LastName,
		MobileNumber: id.MobileNumber,
		Email:        id.Email,
		Country:      c.Alpha_2,
		CreatedAt:    id.CreatedAt,
		UpdatedAt:    id.UpdatedAt,
	}, nil
}

func Get(ctx context.Context, b Backends, identityID string) (*identity.Identity, error) {
	if identityID == "" {
		return nil, fmt.Errorf("%w ID is required.", identity.ErrInvalidArgument)
	}

	var id identityModel
	err := b.DB().GetContext(ctx, &id, `SELECT * FROM identities WHERE id=$1 LIMIT 1`, identityID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, identity.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", identity.ErrInternal, err)
	}

	country, err := b.Countries().Get(ctx, id.CountryID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", identity.ErrInternal, err)
	}

	return &identity.Identity{
		ID:           id.ID,
		FirstName:    id.FirstName,
		LastName:     id.LastName,
		MobileNumber: id.MobileNumber,
		Email:        id.Email,
		Country:      country.Alpha_2,
		CreatedAt:    id.CreatedAt,
		UpdatedAt:    id.UpdatedAt,
	}, nil
}

func GetByEmail(ctx context.Context, b Backends, email string) (*identity.Identity, error) {
	if email == "" {
		return nil, fmt.Errorf("%w Email is required.", identity.ErrInvalidArgument)
	}

	var id identityModel
	var country *_country.Country
	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		err := tx.Get(&id, `SELECT * FROM identities WHERE email=$1 LIMIT 1`, email)
		if err != nil {
			if err == sql.ErrNoRows {
				return identity.ErrNotFound
			}

			return fmt.Errorf("%w %s", identity.ErrInternal, err)
		}

		country, err = b.Countries().Get(ctx, id.CountryID)
		if err != nil {
			return fmt.Errorf("%w %s", identity.ErrInternal, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &identity.Identity{
		ID:           id.ID,
		FirstName:    id.FirstName,
		LastName:     id.LastName,
		MobileNumber: id.MobileNumber,
		Email:        id.Email,
		Country:      country.Alpha_2,
		CreatedAt:    id.CreatedAt,
		UpdatedAt:    id.UpdatedAt,
	}, nil
}
