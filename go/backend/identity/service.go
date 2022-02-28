package identity

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_country "gitlab.com/fynbos/backend/country"
)

// DB Model
type identity struct {
	ID           string
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	MobileNumber string `db:"mobile_number"`
	Email        string
	DateOfBirth  string `db:"date_of_birth"`
	// Array of addresses as forms will generally have multiple lines to
	// store the address. e.g. line1= street, line2=apartment building name and unit number.
	Address    pq.StringArray
	State      string
	City       string
	PostalCode string `db:"postal_code"`

	// primary key of country
	CountryID string `db:"country_id"`

	TaxIDNumber       string `db:"tax_id_number"`
	Provider          string
	ProviderID        string `db:"provider_id"`
	VerificationState string `db:"verification_state"`
	CreatedAt         string `db:"created_at"`
	UpdatedAt         string `db:"updated_at"`
}

// Application Model
type Identity struct {
	ID           string
	FirstName    string
	LastName     string
	MobileNumber string
	Email        string
	DateOfBirth  string
	// Array of addresses as forms will generally have multiple lines to
	// store the address. e.g. line1= street, line2=apartment building name and unit number.
	Address           []string
	State             string
	City              string
	PostalCode        string
	Country           string
	TaxIDNumber       string
	Provider          string
	ProviderID        string
	VerificationState string
	CreatedAt         string
	UpdatedAt         string
}

type Service interface {
	Create(ctx context.Context, tx *sqlx.Tx, args CreateArgs) (*Identity, error)
	Get(ctx context.Context, tx *sqlx.Tx, id string) (*Identity, error)
	Verify(ctx context.Context, tx *sqlx.Tx, args *VerifyArgs) (*Identity, error)
}

type service struct {
	country   _country.Service
	validator *validator.Validate
}

type ServiceArgs struct {
	CountryService _country.Service `validate:"required"`
}

func NewService(args ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, &ErrInvalidArgument{Err: err.Error()}
	}

	return &service{
		country:   args.CountryService,
		validator: validator,
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
func (self *service) Create(ctx context.Context, tx *sqlx.Tx, args CreateArgs) (*Identity, error) {
	err := self.validator.Struct(args)
	if err != nil {
		return nil, err
	}

	country, err := self.country.GetByAlpha2(ctx, tx, args.Country)
	if err != nil {
		switch err.(type) {
		case *_country.ErrNotFound:
			return nil, &ErrInvalidArgument{Err: err.Error()}
		default:
			return nil, &ErrInternalError{Err: err.Error()}
		}
	}

	provider := self.determineProvider(country.Alpha_2)
	var identity identity
	stmt, err := tx.PrepareNamed(`
			INSERT INTO identities (
				id, first_name, last_name, mobile_number, email, country_id, provider
			)
			VALUES (:id, :first_name, :last_name, :mobile_number, :email, :country_id, :provider)
			RETURNING *;
		`)
	if err != nil {
		return nil, &ErrInternalError{Err: err.Error()}

	}

	err = stmt.Stmt.Get(&identity,
		args.ID,
		args.FirstName,
		args.LastName,
		args.MobileNumber,
		args.Email,
		country.ID,
		provider,
	)
	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"primary\"") {
			return nil, &ErrDuplicate{Err: "Identity exists."}
		}
		return nil, &ErrInternalError{Err: err.Error()}
	}

	return self.Get(ctx, tx, args.ID)
}

func (self service) determineProvider(country string) string {
	return "noop"
}

func (self service) Get(ctx context.Context, tx *sqlx.Tx, id string) (*Identity, error) {
	if id == "" {
		return nil, &ErrInvalidArgument{Err: "ID is required."}
	}

	var identity identity
	err := tx.Get(&identity, `SELECT * FROM identities WHERE id=$1 LIMIT 1`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{Err: "Not found."}
		}

		return nil, &ErrInternalError{Err: err.Error()}
	}

	country, err := self.country.Get(ctx, tx, identity.CountryID)
	if err != nil {
		return nil, &ErrInternalError{Err: err.Error()}
	}

	// crdb < v21.2.5 retrieves date as a timestamp.
	// cut off the hour, day etc
	parts := strings.Split(identity.DateOfBirth, "T")
	dob := ""
	if parts[0] != "0000-01-01" {
		dob = parts[0]
	}
	return &Identity{
		ID:                identity.ID,
		FirstName:         identity.FirstName,
		LastName:          identity.LastName,
		MobileNumber:      identity.MobileNumber,
		Email:             identity.Email,
		DateOfBirth:       dob,
		Address:           identity.Address,
		State:             identity.State,
		City:              identity.City,
		PostalCode:        identity.PostalCode,
		Country:           country.Alpha_2,
		TaxIDNumber:       identity.TaxIDNumber,
		Provider:          identity.Provider,
		ProviderID:        identity.ProviderID,
		VerificationState: identity.VerificationState,
		CreatedAt:         identity.CreatedAt,
		UpdatedAt:         identity.UpdatedAt,
	}, nil
}

// TODO: we may not be able to store the user's personal information and will just pass the
// information through to the provider.
type VerifyArgs struct {
	IdentityID        string   `validate:"required,uuid"`
	DateOfBirth       string   `validate:"datetime=2006-01-02"`
	Address           []string `validate:"min=1,dive,required"`
	State             string   `validate:"required"`
	City              string   `validate:"required"`
	PostalCode        string   `validate:"required"`
	TaxIDNumber       string   `validate:"required"`
	ProviderID        string   `valdiate:"required"`
	VerificationState string   `validate:"required"`
}

func (self *service) Verify(ctx context.Context, tx *sqlx.Tx, args *VerifyArgs) (*Identity, error) {
	err := self.validator.Struct(args)
	if err != nil {
		return nil, &ErrInvalidArgument{Err: err.Error()}
	}

	id, err := self.Get(ctx, tx, args.IdentityID)
	if err != nil {
		return nil, err
	}

	stmt, err := tx.PrepareNamed(`
			UPDATE identities SET
				date_of_birth=:dob,
				address=:address,
				state=:state,
				city=:city,
				postal_code=:postal,
				tax_id_number=:tax_number,
				verification_state=:status,
				provider_id=:providerId
			WHERE id=:id
			RETURNING *;
		`)
	if err != nil {
		return nil, &ErrInternalError{Err: err.Error()}
	}

	var identity identity
	err = stmt.Stmt.Get(&identity,
		args.DateOfBirth,
		pq.Array(args.Address),
		args.State,
		args.City,
		args.PostalCode,
		args.TaxIDNumber,
		args.VerificationState,
		args.ProviderID,
		id.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "pq: not found") {
			return nil, &ErrNotFound{Err: "Identity not found."}
		}
		return nil, &ErrInternalError{Err: err.Error()}
	}

	return self.Get(ctx, tx, identity.ID)
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
