package fundingsources

//go:generate mockgen -destination=./mock.go -package=fundingsources -source=./service.go

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	_identity "gitlab.com/fynbos/backend/identity"
	_mx "gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/noop"
)

var (
	ErrDuplicate       = errors.New("funding source: duplicate.")
	ErrNotFound        = errors.New("funding source: not found.")
	ErrInvalidArgument = errors.New("funding source: invalid argument.")
	ErrInternal        = errors.New("funding source: internal error.")
	ErrUnauthorized    = errors.New("funding source: unauthorized.")
)

type Service interface {
	Create(ctx context.Context, tx *sqlx.Tx, args *CreateArgs) (*FundingSource, error)
	Get(ctx context.Context, id string) (*FundingSource, error)
	GetByAccountId(ctx context.Context, identityId string) ([]FundingSource, error)
	Verify(ctx context.Context, args *VerifyArgs) (*FundingSource, error)
	CreateBankAccount(ctx context.Context, args *CreateBankAccountArgs) (*FundingSource, error)
	GetMxConnectWidget(ctx context.Context, accountID string, identityID string) (string, error)
}

type service struct {
	validator *validator.Validate
	db        *sqlx.DB
	is        _identity.Service
	as        accounts.Service
	noop      noop.Service
	mx        _mx.Service
}

type ServiceArgs struct {
	Is   _identity.Service `validate:"required"`
	As   accounts.Service  `validate:"required"`
	Db   *sqlx.DB          `validate:"required"`
	Noop noop.Service      `validate:"required"`
	Mx   _mx.Service       `validate:"required"`
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
		mx:        args.Mx,
	}, nil
}

type FundingSource struct {
	ID                string
	AccountID         string `db:"account_id"`
	Name              string
	VerificationState string `db:"verification_state"`
	Mask              string
	Type              string // pivot table type
	TypeID            string `db:"type_id"` // pivot table typeID
	SubType           string `db:"subtype"`
	CreatedAt         string `db:"created_at"`
	UpdatedAt         string `db:"updated_at"`
}

type CreateArgs struct {
	IdentityID        string `validate:"required,uuid4"`
	AccountID         string `validate:"required,uuid4"`
	Name              string `validate:"required"`
	Mask              string `validate:"required"`
	VerificationState string `validate:"required"`
	Type              string `validate:"oneof=noop"`
	TypeID            string `validate:"required,uuid4"`
	SubType           string `validate:"required"`
}

func (s *service) Create(ctx context.Context, tx *sqlx.Tx, args *CreateArgs) (*FundingSource, error) {
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

	var fs FundingSource
	stmt, err := tx.PrepareNamed(`
			INSERT INTO funding_sources (
				account_id, name, mask, verification_state, type, type_id, subtype
			)
			VALUES (:account_id, :name, :mask, :verification_state, :type, :type_id, :subtype)
			RETURNING *;
		`)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	err = stmt.Stmt.Get(&fs,
		acc.ID,
		args.Name,
		args.Mask,
		args.VerificationState,
		args.Type,
		args.TypeID,
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

		// performing provider verification here for now.
		err = s.noop.VerifyBankAccount(ctx, &noop.VerifyArgs{})
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}

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

	var fundingsource *FundingSource
	err = crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		fs, err := s.Create(ctx, tx, &CreateArgs{
			IdentityID:        args.IdentityID,
			AccountID:         args.AccountID,
			Name:              args.Name,
			Mask:              args.AccountNumber[:4],
			VerificationState: "required",
			Type:              "noop",
			TypeID:            uuid.NewString(),
			SubType:           args.Type,
		})
		if err != nil {
			return err
		}

		fundingsource = fs
		return nil
	})
	if err != nil {
		return nil, err
	}

	return fundingsource, nil
}
func (s *service) GetMxConnectWidget(ctx context.Context, accountID string, identityID string) (string, error) {
	acc, err := s.as.Get(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	if acc.IdentityID != identityID {
		return "", ErrUnauthorized
	}

	mxUserGuid, err := s.mx.CreateUser(ctx)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	url, err := s.mx.GetWidgetUrl(ctx, mxUserGuid)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return url, nil
}

func IsVerified(fs *FundingSource) bool {
	return fs.VerificationState == "verified"
}
