package onboarding

//go:generate mockgen -destination=./mock.go -package=onboarding -source=./service.go

import (
	"context"
	"errors"
	"fmt"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
)

var (
	ErrInternal        = errors.New("onboarding: internal error.")
	ErrInvalidArgument = errors.New("onboarding: invalid argument.")
)

type Service interface {
	CreateAccount(ctx context.Context, args *CreateAccountArgs) (*accounts.Account, error)
	VerifyAccount(ctx context.Context, args *VerifyAccountArgs) (*accounts.Account, error)
	InitiateUnitCustomerOnboarding(ctx context.Context, args *InitiateUnitCustomerOnboardingArgs) error
}

type service struct {
	validator *validator.Validate
	db        *sqlx.DB
	as        accounts.Service
	is        identity.Service
	noop      noop.Service
}

type ServiceArgs struct {
	Db   *sqlx.DB         `validate:"required"`
	As   accounts.Service `validate:"required"`
	Is   identity.Service `validate:"required"`
	Noop noop.Service     `validate:"required"`
}

func NewService(args *ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, err
	}

	return &service{
		validator: validator,
		db:        args.Db,
		as:        args.As,
		is:        args.Is,
		noop:      args.Noop,
	}, nil
}

type CreateAccountArgs struct {
	IdentityID string `validate:"required,uuid"`
	Country    string `validate:"required,iso3166_1_alpha2"`
}

// Creating an account with Fynbos means that your identity is stored in our system and you get
// a Fynbos account. The account is not yet backed by any provider and as such the user won't be
// able to do anything with it.
func (s *service) CreateAccount(
	ctx context.Context,
	args *CreateAccountArgs,
) (*accounts.Account, error) {
	if err := s.validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	id, err := s.is.Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	account, err := s.as.Create(ctx, &accounts.CreateAccountArgs{
		IdentityID:                 id.ID,
		Country:                    args.Country,
		CreditsMustNotExceedDebits: true,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	if err != nil {
		return nil, err
	}

	return account, err
}

type VerifyAccountArgs struct {
	IdentityID  string   `validate:"required,uuid"`
	AccountID   string   `validate:"required,uuid"`
	DateOfBirth string   `validate:"datetime=2006-01-02"`
	Address     []string `validate:"min=1,dive,required"`
	State       string   `validate:"required"`
	City        string   `validate:"required"`
	PostalCode  string   `validate:"required"`
	TaxIDNumber string   `validate:"required"`
}

func (s *service) VerifyAccount(ctx context.Context, args *VerifyAccountArgs) (*accounts.Account, error) {
	if err := s.validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	var verifiedAccount *accounts.Account
	err := crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		id, err := s.is.Get(ctx, args.IdentityID)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}
		acc, err := s.as.GetByIdentityIDWithTrx(ctx, tx, id.ID)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}

		customer, err := s.noop.CreateCustomer(&noop.CreateCustomerArgs{
			FirstName:   id.FirstName,
			LastName:    id.LastName,
			Email:       id.Email,
			Address1:    args.Address[0],
			State:       args.State,
			City:        args.City,
			PostalCode:  args.PostalCode,
			DateOfBirth: args.DateOfBirth,
			Ssn:         args.TaxIDNumber,
		})
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}

		verifiedAccount, err = s.as.VerifyWithTx(ctx, tx, &accounts.VerifyArgs{
			AccountID:  acc.ID,
			Provider:   "noop",
			ProviderID: customer.ID,
		})
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return verifiedAccount, nil
}

type InitiateUnitCustomerOnboardingArgs struct {
	IdentityID string `validate:"required"`
	Country    string `validate:"oneof=US"`
}

func (s *service) InitiateUnitCustomerOnboarding(ctx context.Context, args *InitiateUnitCustomerOnboardingArgs) error {
	fmt.Println(fmt.Sprintf("initiating unit customer onboarding %+v", args))
	return nil
}
