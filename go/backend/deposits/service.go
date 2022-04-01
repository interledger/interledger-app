package deposits

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
)

var (
	ErrUnauthorized            = errors.New("deposit: unauthorized.")
	ErrInternal                = errors.New("deposit: internal error.")
	ErrInvalidArgument         = errors.New("deposit: invalid argument.")
	ErrNotFound                = errors.New("deposit: not found.")
	ErrUnverifiedFundingSource = errors.New("deposit: unverified funding source.")
)

type Service interface {
	InitiateDeposit(ctx context.Context, args *InitiateDepositArgs) (*Deposit, error)
}

type ServiceArgs struct {
	Db *sqlx.DB               `validate:"required"`
	As accounts.Service       `validate:"required"`
	Is identity.Service       `validate:"required"`
	Fs fundingsources.Service `validate:"required"`
	Tp client.Client          `validate:"required"`
}

type service struct {
	validator *validator.Validate
	db        *sqlx.DB
	as        accounts.Service
	is        identity.Service
	fs        fundingsources.Service
	tp        client.Client
}

type Deposit struct {
	ID              string
	AccountID       string `db:"account_id"`
	FundingSourceId string `db:"funding_source_id"`
	Amount          int64
	CreatedAt       string `db:"created_at"`
	UpdatedAt       string `db:"updated_at"`
}

func NewService(args *ServiceArgs) (Service, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	return &service{
		validator: v,
		db:        args.Db,
		as:        args.As,
		is:        args.Is,
		fs:        args.Fs,
		tp:        args.Tp,
	}, nil
}

type InitiateDepositArgs struct {
	IdentityID      string `validate:"required,uuid4"`
	AccountID       string `validate:"required,uuid4"`
	FundingSourceID string `validate:"required,uuid4"`
	Amount          uint64 `validate:"required,gt=0"`
}

func (s *service) InitiateDeposit(ctx context.Context, args *InitiateDepositArgs) (*Deposit, error) {
	if err := s.validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}
	/*
		The flow should be as follows
		* Check if existing deposit already
		* Checks on Identity, Account, FS
		* Create Deposit Object (accountId, funding source etc)
		* Kickoff workflow
	*/
	id, err := s.is.Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}
	acc, err := s.as.Get(ctx, args.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}
	fundingSource, err := s.fs.Get(ctx, args.FundingSourceID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}
	if fundingSource.AccountID != acc.ID {
		return nil, ErrNotFound
	}
	if fundingSource.VerificationState != "verified" {
		return nil, ErrUnverifiedFundingSource
	}

	if !s.as.CanMakeDeposit(acc, id.ID) {
		return nil, ErrUnauthorized
	}

	// Create the deposit
	// TODO should this be an idempotent key?
	depositId := uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID: "deposit_" + depositId.String(),
	}
	_, err = s.tp.ExecuteWorkflow(context.Background(), workflowOptions, DepositWorkflow)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return nil, nil
}
