package fundingsources

//go:generate mockgen -destination=./mock.go -package=fundingsources -source=./service.go

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/accounts/ops"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
	_unit "gitlab.com/fynbos/backend/providers/unit"
	"go.temporal.io/sdk/client"
)

var (
	ErrDuplicate       = errors.New("funding source: duplicate.")
	ErrNotFound        = errors.New("funding source: not found.")
	ErrInvalidArgument = errors.New("funding source: invalid argument.")
	ErrInternal        = errors.New("funding source: internal error.")
	ErrUnauthorized    = errors.New("funding source: unauthorized.")
)

type Service interface {
	Create(ctx context.Context, args *CreateArgs) (*FundingSource, error)
	Get(ctx context.Context, id string) (*FundingSource, error)
	GetByAccountId(ctx context.Context, identityId string) ([]FundingSource, error)
	Verify(ctx context.Context, args *VerifyArgs) (*FundingSource, error)
	CreateBankAccount(ctx context.Context, args *CreateBankAccountArgs) (*FundingSource, error)
}

type service struct {
	validator *validator.Validate
	db        *sqlx.DB
	is        _identity.Service
	as        ops.Service
	noop      noop.Service
	tp        client.Client
	unit      _unit.Service
}

type ServiceArgs struct {
	Is   _identity.Service `validate:"required"`
	As   ops.Service       `validate:"required"`
	Db   *sqlx.DB          `validate:"required"`
	Noop noop.Service      `validate:"required"`
	Tp   client.Client     `validate:"required"`
	Unit _unit.Service     `validate:"required"`
}

func NewService(args *ServiceArgs) (Service, error) {
	v := validator.New()
	err := v.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	return &service{
		validator: v,
		is:        args.Is,
		as:        args.As,
		db:        args.Db,
		noop:      args.Noop,
		tp:        args.Tp,
		unit:      args.Unit,
	}, nil
}

type VerificationState string

const (
	REQUIRED   = VerificationState("required")
	PROCESSING = VerificationState("processing")
	VERIFIED   = VerificationState("verified")
)

type FundingSource struct {
	ID                string
	AccountID         string `db:"account_id"`
	Name              string
	VerificationState string `db:"verification_state"`
	Mask              string
	Type              string
	SubType           string `db:"subtype"`
	CreatedAt         string `db:"created_at"`
	UpdatedAt         string `db:"updated_at"`
}

type UnitCounterParty struct {
	ID                 string
	UnitCounterpartyID string `db:"unit_counterparty_id"`
	CreatedAt          string `db:"created_at"`
	UpdatedAt          string `db:"updated_at"`
}

type CreateArgs struct {
	ID                string `validate:"omitempty,uuid4"`
	IdentityID        string `validate:"required,uuid4"`
	AccountID         string `validate:"required,uuid4"`
	Name              string `validate:"required"`
	Mask              string
	VerificationState string `validate:"required"`
	Type              string `validate:"oneof=noop mx"`
	SubType           string `validate:"required"`
}

func (s *service) Create(ctx context.Context, args *CreateArgs) (*FundingSource, error) {
	// TODO: refactor errors
	err := s.validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	identity, err := s.is.Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}
	acc, err := s.as.Get(ctx, args.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}
	if !s.as.CanCreateFundingSource(acc, identity.ID) {
		return nil, ErrUnauthorized
	}

	fundingsourceID := args.ID
	if fundingsourceID == "" {
		fundingsourceID = uuid.NewString()
	}
	var fs FundingSource
	err = s.db.GetContext(
		ctx,
		&fs,
		`
			INSERT INTO funding_sources (
				id, account_id, name, mask, verification_state, type, subtype
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING *;
		`,
		fundingsourceID,
		acc.ID,
		args.Name,
		args.Mask,
		args.VerificationState,
		args.Type,
		args.SubType,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return &fs, nil
}

func (s service) Get(ctx context.Context, id string) (*FundingSource, error) {
	if id == "" {
		return nil, fmt.Errorf("%w ID is required.", ErrInvalidArgument)
	}

	var fundingsource FundingSource
	err := s.db.GetContext(ctx, &fundingsource, "SELECT * FROM funding_sources where id=$1 LIMIT 1;", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return &fundingsource, nil
}

func (s service) GetByAccountId(ctx context.Context, identityId string) ([]FundingSource, error) {
	if identityId == "" {
		return nil, fmt.Errorf("%w IdentityID is required.", ErrInvalidArgument)
	}

	fundingSources := []FundingSource{}
	err := s.db.SelectContext(ctx, &fundingSources, "SELECT * FROM funding_sources WHERE account_id=$1;", identityId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return fundingSources, nil
}

type VerifyArgs struct {
	IdentityID      string `validate:"required,uuid4"`
	FundingSourceID string `validate:"required,uuid4"`
}

func (s *service) Verify(ctx context.Context, args *VerifyArgs) (*FundingSource, error) {
	// TODO: refactor errors
	err := s.validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	id, err := s.is.Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}
	fs, err := s.Get(ctx, args.FundingSourceID)
	if err != nil {
		return nil, err
	}
	acc, err := s.as.Get(ctx, fs.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}
	if !s.as.CanVerifyFundingSource(acc, id.ID) {
		return nil, ErrUnauthorized
	}

	var verifiedFs FundingSource
	err = crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		stmt, err := tx.PrepareNamed(`
			UPDATE funding_sources SET verification_state=$1 where id=$2 RETURNING *;
		`)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}

		err = stmt.Stmt.Get(&verifiedFs,
			"verified",
			args.FundingSourceID,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}

			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &verifiedFs, nil
}

type CreateBankAccountArgs struct {
	IdentityID    string `validate:"required,uuid4"`
	AccountID     string `validate:"required,uuid4"`
	Name          string `validate:"required"`
	AccountNumber string `validate:"required"`
	RoutingNumber string `validate:"required"`
	Institution   string `validate:"required"`
	Type          string `validate:"required"`
}

func (s *service) CreateBankAccount(
	ctx context.Context,
	args *CreateBankAccountArgs,
) (*FundingSource, error) {
	err := s.validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	fundingsource, err := s.Create(ctx, &CreateArgs{
		IdentityID:        args.IdentityID,
		AccountID:         args.AccountID,
		Name:              args.Name,
		Mask:              args.AccountNumber[:4],
		VerificationState: "required",
		Type:              "noop",
		SubType:           args.Type,
	})
	if err != nil {
		return nil, err
	}

	return fundingsource, nil
}

func IsVerified(fs *FundingSource) bool {
	return fs.VerificationState == "verified"
}
