package deposits

//go:generate mockgen -destination=./mock.go -package=deposits -source=./service.go

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

var (
	ErrUnauthorized            = errors.New("deposit: unauthorized")
	ErrInternal                = errors.New("deposit: internal error")
	ErrInvalidArgument         = errors.New("deposit: invalid argument")
	ErrNotFound                = errors.New("deposit: not found")
	ErrUnverifiedFundingSource = errors.New("deposit: unverified funding source")
)

type Service interface {
	InitiateDeposit(ctx context.Context, args *InitiateDepositArgs) (*Deposit, error)
	Get(ctx context.Context, id string) (*Deposit, error)
	SetState(ctx context.Context, id string, state State) error
}

type ServiceArgs struct {
	Db *sqlx.DB              `validate:"required"`
	As accounts.Client       `validate:"required"`
	Is identity.Service      `validate:"required"`
	Fs fundingsources.Client `validate:"required"`
	Tp client.Client         `validate:"required"`
}

type service struct {
	validator *validator.Validate
	db        *sqlx.DB
	as        accounts.Client
	is        identity.Service
	fs        fundingsources.Client
	tp        client.Client
}

type Deposit struct {
	ID              string
	AccountID       string `db:"account_id"`
	FundingSourceId string `db:"funding_source_id"`
	Amount          uint64
	State           State
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

	// TODO should this be an idempotent key?
	var deposit Deposit
	err = s.db.Get(&deposit, `INSERT INTO deposits
		(account_id, funding_source_id, amount, state) VALUES ($1, $2, $3, $4)
		RETURNING *;
		`, acc.ID, fundingSource.ID, args.Amount, Created)
	if err != nil {
		return nil, fmt.Errorf("failed to insert into db %s %w", err.Error(), ErrInternal)
	}

	// TODO add mechanism to handle if deposit is created but workflow is not
	workflowOptions := client.StartWorkflowOptions{
		ID:                    "deposit_" + deposit.ID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}
	_, err = s.tp.ExecuteWorkflow(context.Background(), workflowOptions, DepositWorkflow, deposit.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert execute workflow %s %w", err.Error(), ErrInternal)
	}

	return &deposit, nil
}

func (s *service) Get(ctx context.Context, id string) (*Deposit, error) {
	var deposit Deposit
	err := s.db.GetContext(ctx, &deposit, `select * from deposits where id = $1 LIMIT 1`, id)
	if err != nil {
		return nil, err
	}

	return &deposit, nil
}

func (s *service) SetState(ctx context.Context, id string, state State) error {

	deposit, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// TODO add checks for legitimate state changes

	_, err = s.db.ExecContext(ctx, "update deposits set state = $1 where id = $2", state.String(), deposit.ID)
	if err != nil {
		return err
	}

	return nil
}

type State string

const (
	Created    = State("CREATED")
	Processing = State("PROCESSING")
	Complete   = State("COMPLETE")
	Failed     = State("FAILED")
)

func (s State) String() string {
	return string(s)
}

func (s State) IsValid() bool {
	switch s {
	case Created, Processing, Complete, Failed:
		return true
	}
	return false
}

func (s *State) Unmarshall(v string) error {
	state := State(v)
	if !state.IsValid() {
		return fmt.Errorf("%s is not a valid State", v)
	}
	*s = state
	return nil
}
