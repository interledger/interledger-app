package payments

import (
	"context"
	"errors"
	"fmt"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/identity"
)

var (
	ErrInvalidArgument     = errors.New("payments service: invalid argument.")
	ErrInternal            = errors.New("payments service: internal error.")
	ErrUnverifiedAccount   = errors.New("payments service: unverified account.")
	ErrInsufficientBalance = errors.New("payments service: insufficient balance.")
	ErrUnauthorized        = errors.New("payments service: unauthorized.")
)

const (
	Verified  string = "verified"
	Retry     string = "retry"
	Kba       string = "kba"
	Document  string = "document"
	Suspended string = "suspended"
)

type Service interface {
	InitiateOutgoingPayment(ctx context.Context, args *InitiateOutgoingPaymentArgs) (*OutgoingPayment, error)
	Get(ctx context.Context, id string) (*OutgoingPayment, error)
	SetState(ctx context.Context, id string, state State) error
}

type service struct {
	validator *validator.Validate
	db        *sqlx.DB
	as        accounts.Service
	is        identity.Service
	tp        client.Client
}

type OutgoingPayment struct {
	ID                   string
	AccountID            string `db:"account_id"`
	Destination          string `db:"destination"`
	Amount               uint64
	State                State
	CreatedAt            string `db:"created_at"`
	UpdatedAt            string `db:"updated_at"`
}

type ServiceArgs struct {
	Db   *sqlx.DB                     `validate:"required"`
	As   accounts.Service             `validate:"required"`
	Is   identity.Service             `validate:"required"`
	Tp   client.Client                `validate:"required"`
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
		tp:        args.Tp,
	}, nil
}

type InitiateOutgoingPaymentArgs struct {
	IdentityID string `validate:"required"`
	AccountID  string `validate:"required"`
	Amount     uint64 `validate:"gt=0"`
	To         string `validate:"required"`
}

func (s *service) InitiateOutgoingPayment(
	ctx context.Context,
	args *InitiateOutgoingPaymentArgs,
) (*OutgoingPayment, error) {
	err := s.validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInvalidArgument)
	}

	id, err := s.is.Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
	}
	acc, err := s.as.Get(ctx, args.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
	}
	if !s.as.CanMakeOutgoingPayment(acc, id.ID) {
		return nil, fmt.Errorf("%w", ErrUnauthorized)
	}
	if !acc.IsVerified() {
		return nil, fmt.Errorf("%w", ErrUnverifiedAccount)
	}

	var outgoingPayment OutgoingPayment
	err = s.db.Get(&outgoingPayment, `INSERT INTO outgoing_payments
	(account_id, amount, destination, state) VALUES ($1, $2, $3, $4) RETURNING *`, acc.ID, args.Amount, args.To, Created)

	if err != nil {
		return nil, fmt.Errorf("failed to insert into db %s %w", err.Error(), ErrInternal)
	}

	// TODO add mechanism to handle if deposit is created but workflow is not
	workflowOptions := client.StartWorkflowOptions{
		ID:                    "outgoingPayment_" + outgoingPayment.ID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}
	_, err = s.tp.ExecuteWorkflow(context.Background(), workflowOptions, OutgoingPaymentWorkflow, outgoingPayment.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to insert execute workflow %s %w", err.Error(), ErrInternal)
	}

	return &outgoingPayment, nil
}

func (s *service) Get(ctx context.Context, id string) (*OutgoingPayment, error) {
	var outgoingPayment OutgoingPayment
	err := s.db.GetContext(ctx, &outgoingPayment, `select * from outgoing_payments where id = $1 LIMIT 1`, id)
	if err != nil {
		return nil, err
	}

	return &outgoingPayment, nil
}

func (s *service) SetState(ctx context.Context, id string, state State) error {
	outgoingPayment, err := s.Get(ctx, id)

	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, "update outgoing_payments set state = $1 where id = $2", state.String(), outgoingPayment.ID)
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
