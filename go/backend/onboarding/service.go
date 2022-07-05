package onboarding

//go:generate mockgen -destination=./mock.go -package=onboarding -source=./service.go

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/unit"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

var (
	ErrInternal        = errors.New("onboarding: internal error.")
	ErrInvalidArgument = errors.New("onboarding: invalid argument.")
	ErrNotFound        = errors.New("onboarding: not found.")
)

type Onboarding struct {
	ID               string
	FirstName        string `db:"first_name"`
	LastName         string `db:"last_name"`
	Country          string `db:"country_of_residence"`
	Email            string `db:"email"`
	Phone            string `db:"phone"`
	PhoneVerified    bool   `db:"phone_verified"`
	ServiceAgreement bool   `db:"service_agreement"`
	CreatedAt        string `db:"created_at"`
	UpdatedAt        string `db:"updated_at"`
}

type Service interface {
	GetOnboarding(ctx context.Context, args *GetOnboardingArgs) (*Onboarding, error)
	UpdateOnboarding(ctx context.Context, args *UpdateOnboardingArgs) (*Onboarding, error)
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
	tp        client.Client
}

type ServiceArgs struct {
	Db   *sqlx.DB         `validate:"required"`
	As   accounts.Service `validate:"required"`
	Is   identity.Service `validate:"required"`
	Noop noop.Service     `validate:"required"`
	Tp   client.Client    `validate:"required"`
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
		tp:        args.Tp,
	}, nil
}

type GetOnboardingArgs struct {
	Id string `validate:"required,uuid"`
}

// Fetch a users onboarding data
func (s *service) GetOnboarding(
	ctx context.Context,
	args *GetOnboardingArgs,
) (*Onboarding, error) {
	if err := s.validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	var onboarding Onboarding
	err := s.db.GetContext(ctx, &onboarding, "SELECT * FROM  onboarding WHERE id = $1 LIMIT 1;", args.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return &onboarding, err
}

type UpdateOnboardingArgs struct {
	Id               string `validate:"omitempty,uuid"`
	Country          string `validate:"omitempty,iso3166_1_alpha2"`
	FirstName        string `validate:"omitempty"`
	LastName         string `validate:"omitempty"`
	Email            string `validate:"omitempty,email"`
	Phone            string `validate:"omitempty,e164"`
	PhoneVerified    bool   `validate:"omitempty"`
	ServiceAgreement bool   `validate:"omitempty"`
}

func (s *service) UpdateOnboarding(
	ctx context.Context,
	args *UpdateOnboardingArgs,
) (*Onboarding, error) {
	if err := s.validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	var onboarding Onboarding
	if args.Id == "" {
		err := s.db.GetContext(ctx, &onboarding,
			`INSERT INTO onboarding (first_name,last_name,country_of_residence,email,phone,phone_verified,service_agreement) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING *;`,
			args.FirstName, args.LastName, args.Country, args.Email, args.Phone, args.PhoneVerified, args.ServiceAgreement,
		)
		if err != nil {
			return nil, err
		}
	} else {
		currentOnboarding, err := s.GetOnboarding(ctx, &GetOnboardingArgs{
			Id: args.Id,
		})
		if err != nil {
			return nil, err
		}
		// Manually replace currentOnboarding vals with args if arg not empty
		if args.Country == "" {
			args.Country = currentOnboarding.Country
		}
		if args.FirstName == "" {
			args.FirstName = currentOnboarding.FirstName
		}
		if args.LastName == "" {
			args.LastName = currentOnboarding.LastName
		}
		if args.Email == "" {
			args.Email = currentOnboarding.Email
		}
		if args.Phone == "" {
			args.Phone = currentOnboarding.Phone
		}
		if args.PhoneVerified != currentOnboarding.PhoneVerified {
			args.PhoneVerified = currentOnboarding.PhoneVerified
		}
		if args.ServiceAgreement != currentOnboarding.ServiceAgreement {
			args.ServiceAgreement = currentOnboarding.ServiceAgreement
		}
		err = s.db.GetContext(ctx, &onboarding,
			`UPDATE onboarding SET (first_name,last_name,country_of_residence,email,phone,phone_verified,service_agreement) = ($2,$3,$4,$5,$6,$7,$8) WHERE id = $1 RETURNING *;`,
			args.Id, args.FirstName, args.LastName, args.Country, args.Email, args.Phone, args.PhoneVerified, args.ServiceAgreement,
		)
		if err != nil {
			return nil, err
		}
	}

	return &onboarding, nil
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
	IdentityID         string   `validate:"required"`
	Ssn                string   `validate:"required"`
	DateOfBirth        string   `validate:"required"`
	Street             string   `validate:"required"`
	Street2            string   `validate:""`
	City               string   `validate:"required"`
	State              string   `validate:"required"`
	PostalCode         string   `validate:"required"`
	IpAddress          string   `validate:"required"`
	DeviceFingerprints []string `validate:"required"`
}

func (s *service) InitiateUnitCustomerOnboarding(ctx context.Context, args *InitiateUnitCustomerOnboardingArgs) error {
	if err := s.validator.Struct(args); err != nil {
		return fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	// TODO: store these args in vault and just pass the key to the workflow.

	_, err := s.tp.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:                    "unit_onboarding_" + args.IdentityID,
			TaskQueue:             "backend",
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		},
		unit.UnitOnboardCustomerWorkflow, unit.UnitOnboardCustomerState{
			CustomerID: "",
			Type:       "",
			IdentityID: args.IdentityID,
			AccountID:  "",
			ApplicationArgs: unit.CreateApplicationArgs{
				Ssn:                args.Ssn,
				DateOfBirth:        args.DateOfBirth,
				Street:             args.Street,
				Street2:            args.Street2,
				City:               args.City,
				State:              args.State,
				PostalCode:         args.PostalCode,
				IpAddress:          args.IpAddress,
				UserID:             args.IdentityID,
				DeviceFingerprints: args.DeviceFingerprints,
			},
		})
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return nil
}
