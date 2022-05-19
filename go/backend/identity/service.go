package identity

//go:generate mockgen -destination=./mock.go -package=identity -source=./service.go

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_country "gitlab.com/fynbos/backend/country"
)

var (
	ErrInternal        = errors.New("identity: internal error.")
	ErrInvalidArgument = errors.New("identity: invalid argument.")
	ErrNotFound        = errors.New("identity: not found.")
	ErrDuplicate       = errors.New("identity: duplicate.")
)

// DB Model
type identity struct {
	ID           string
	Email        string
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	MobileNumber string `db:"mobile_number"`
	CountryID    string `db:"country_id"` // primary key of country
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}

// Application Model
type Identity struct {
	ID           string
	FirstName    string
	LastName     string
	MobileNumber string
	Email        string
	DateOfBirth  string
	Country      string
	CreatedAt    string
	UpdatedAt    string
}

type Service interface {
	Create(ctx context.Context, args *CreateArgs) (*Identity, error)
	Get(ctx context.Context, id string) (*Identity, error)
	GetByEmail(ctx context.Context, email string) (*Identity, error)
}

type service struct {
	country   _country.Service
	validator *validator.Validate
	db        *sqlx.DB
}

type ServiceArgs struct {
	CountryService _country.Service `validate:"required"`
	Db             *sqlx.DB         `validate:"required"`
}

func NewService(args ServiceArgs) (Service, error) {
	v := validator.New()
	err := v.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	return &service{
		country:   args.CountryService,
		validator: v,
		db:        args.Db,
	}, nil
}

type CreateArgs struct {
	ID           string `validate:"required,uuid"`
	FirstName    string `validate:"required"`
	LastName     string `validate:"required"`
	MobileNumber string `validate:"required"` // TODO: decide on format
	Email        string `validate:"required,email"`
	Country      string `validate:"required,iso3166_1_alpha2"`
}

// We can control what information is allowed to be logged from here.
// TODO: decide what is sensitive information. Might be better to implement at Zap level?
func (args CreateArgs) String() string {
	return fmt.Sprintf("id=%s,firstName=%s,lastName=%s,mobileNum=%s,email=%s,country=%s",
		args.ID,
		args.FirstName,
		args.LastName,
		args.MobileNumber,
		args.Email,
		args.Country,
	)
}

// There is a 1-1 mapping between the identity and user stored in Kratos. The
// Kratos ID is used as the identity ID.
func (s *service) Create(ctx context.Context, args *CreateArgs) (*Identity, error) {
	err := s.validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s.", ErrInvalidArgument, err)
	}

	c, err := s.country.GetByAlpha2(ctx, args.Country)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var identity identity
	err = crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		stmt, err := tx.PrepareNamed(`
			INSERT INTO identities (
				id, first_name, last_name, mobile_number, email, country_id
			)
			VALUES (:id, :first_name, :last_name, :mobile_number, :email, :country_id)
			RETURNING *;
		`)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}

		err = stmt.Stmt.Get(&identity,
			args.ID,
			args.FirstName,
			args.LastName,
			args.MobileNumber,
			args.Email,
			c.ID,
		)
		if err != nil {
			if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"primary\"") {
				return ErrDuplicate
			}
			return fmt.Errorf("%w %s", ErrInternal, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Identity{
		ID:           identity.ID,
		FirstName:    identity.FirstName,
		LastName:     identity.LastName,
		MobileNumber: identity.MobileNumber,
		Email:        identity.Email,
		Country:      c.Alpha_2,
		CreatedAt:    identity.CreatedAt,
		UpdatedAt:    identity.UpdatedAt,
	}, nil
}

func (s service) Get(ctx context.Context, id string) (*Identity, error) {
	if id == "" {
		return nil, fmt.Errorf("%w ID is required.", ErrInvalidArgument)
	}

	var identity identity
	err := s.db.Get(&identity, `SELECT * FROM identities WHERE id=$1 LIMIT 1`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	country, err := s.country.Get(ctx, identity.CountryID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &Identity{
		ID:           identity.ID,
		FirstName:    identity.FirstName,
		LastName:     identity.LastName,
		MobileNumber: identity.MobileNumber,
		Email:        identity.Email,
		Country:      country.Alpha_2,
		CreatedAt:    identity.CreatedAt,
		UpdatedAt:    identity.UpdatedAt,
	}, nil
}

func (self service) GetByEmail(ctx context.Context, email string) (*Identity, error) {
	if email == "" {
		return nil, fmt.Errorf("%w Email is required.", ErrInvalidArgument)
	}

	var identity identity
	var country *_country.Country
	err := crdbsqlx.ExecuteTx(ctx, self.db, nil, func(tx *sqlx.Tx) error {
		err := tx.Get(&identity, `SELECT * FROM identities WHERE email=$1 LIMIT 1`, email)
		if err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}

			return fmt.Errorf("%w %s", ErrInternal, err)
		}

		country, err = self.country.Get(ctx, identity.CountryID)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Identity{
		ID:           identity.ID,
		FirstName:    identity.FirstName,
		LastName:     identity.LastName,
		MobileNumber: identity.MobileNumber,
		Email:        identity.Email,
		Country:      country.Alpha_2,
		CreatedAt:    identity.CreatedAt,
		UpdatedAt:    identity.UpdatedAt,
	}, nil
}
