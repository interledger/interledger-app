package fundingsources

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
)

type Service interface {
	Create(ctx context.Context, tx *sqlx.Tx, args *CreateArgs) (*FundingSource, error)
	Get(ctx context.Context, tx *sqlx.Tx, id string) (*FundingSource, error)
	GetByAccountId(ctx context.Context, tx *sqlx.Tx, identityId string) ([]FundingSource, error)
	Verify(ctx context.Context, args *VerifyArgs) (*FundingSource, error)
	CreateBankAccount(ctx context.Context, args *CreateBankAccountArgs) (*FundingSource, error)
}

type service struct {
	validator *validator.Validate
	db        *sqlx.DB
	is        _identity.Service
	as        accounts.Service
	noop      noop.Service
}

type ServiceArgs struct {
	Is   _identity.Service `validate:"required"`
	As   accounts.Service  `validate:"required"`
	Db   *sqlx.DB          `validate:"required"`
	Noop noop.Service      `validate:"required"`
}

func NewService(args *ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, &ErrInvalidArgument{Err: err.Error()}
	}

	return &service{
		validator: validator,
		is:        args.Is,
		as:        args.As,
		db:        args.Db,
		noop:      args.Noop,
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
		return nil, &ErrInvalidArgument{Err: err.Error()}
	}

	identity, err := s.is.Get(ctx, tx, args.IdentityID)
	if err != nil {
		switch err.(type) {
		case *_identity.ErrInvalidArgument:
		case *_identity.ErrNotFound:
			return nil, &ErrInvalidArgument{Err: "Identity must exist to create funding source."}
		default:
			return nil, &ErrInternalError{Err: err.Error()}
		}
	}
	acc, err := s.as.Get(ctx, tx, args.AccountID)
	if err != nil {
		switch err.(type) {
		default:
			return nil, &ErrInternalError{Err: err.Error()}
		}
	}
	if !s.as.CanCreateFundingSource(acc, identity.ID) {
		return nil, &ErrInternalError{Err: "unauthorized."}
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
		return nil, &ErrInternalError{Err: err.Error()}

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
		return nil, &ErrInternalError{Err: err.Error()}
	}

	return &fs, nil
}

func (s service) Get(ctx context.Context, tx *sqlx.Tx, id string) (*FundingSource, error) {
	if id == "" {
		return nil, &ErrInvalidArgument{Err: "ID is required to look up funding source."}
	}

	var fundingsource FundingSource
	err := tx.Get(&fundingsource, "SELECT * FROM funding_sources where id=$1 LIMIT 1;", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{Err: "Funding source not found."}
		}

		return nil, &ErrInternalError{Err: err.Error()}
	}

	return &fundingsource, nil
}

func (s service) GetByAccountId(ctx context.Context, tx *sqlx.Tx, identityId string) ([]FundingSource, error) {
	if identityId == "" {
		return nil, &ErrInvalidArgument{Err: "Identity ID is required to look up funding source."}
	}

	fundingSources := []FundingSource{}
	err := tx.SelectContext(ctx, &fundingSources, "SELECT * FROM funding_sources WHERE account_id=$1;", identityId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{Err: "Funding sources not found."}
		}

		return nil, &ErrInternalError{Err: err.Error()}
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
		return nil, &ErrInvalidArgument{Err: err.Error()}
	}

	var verifiedFs FundingSource
	err = crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		id, err := s.is.Get(ctx, tx, args.IdentityID)
		if err != nil {
			switch err.(type) {
			case *_identity.ErrInvalidArgument:
			case *_identity.ErrNotFound:
				return &ErrInvalidArgument{Err: "Identity must exist to verify funding source."}
			default:
				return &ErrInternalError{Err: err.Error()}
			}
		}
		fs, err := s.Get(ctx, tx, args.FundingSourceID)
		if err != nil {
			return err
		}
		acc, err := s.as.Get(ctx, tx, fs.AccountID)
		if err != nil {
			return &ErrInternalError{Err: err.Error()}
		}
		if !s.as.CanVerifyFundingSource(acc, id.ID) {
			return &ErrInternalError{Err: "Funding source not found."}
		}

		// performing provider verification here for now.
		err = s.noop.VerifyBankAccount(ctx, &noop.VerifyArgs{})
		if err != nil {
			return &ErrInternalError{Err: "Provider failed to verify funding source."}
		}

		stmt, err := tx.PrepareNamed(`
			UPDATE funding_sources SET verification_state=$1 where id=$2 RETURNING *;
		`)
		if err != nil {
			return &ErrInternalError{Err: err.Error()}

		}

		err = stmt.Stmt.Get(&verifiedFs,
			"verified",
			args.FundingSourceID,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return &ErrNotFound{Err: "Funding source not found."}
			}

			return &ErrInternalError{Err: err.Error()}
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
		return nil, &ErrInvalidArgument{Err: err.Error()}
	}

	var fundingsource *FundingSource
	err = crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		fs, err := s.Create(ctx, tx, &CreateArgs{
			IdentityID:        args.IdentityID,
			AccountID:         args.AccountID,
			Name:              args.Name,
			Mask:              "****" + args.AccountNumber[:4],
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

func IsVerified(fs *FundingSource) bool {
	return fs.VerificationState == "verified"
}

type ErrInternalError struct {
	Err string
}

func (e ErrInternalError) Error() string {
	return e.Err
}

type ErrDuplicate struct {
	Err string
}

func (e ErrDuplicate) Error() string {
	return e.Err
}

type ErrInvalidArgument struct {
	Err string
}

func (e ErrInvalidArgument) Error() string {
	return e.Err
}

type ErrNotFound struct {
	Err string
}

func (e ErrNotFound) Error() string {
	return e.Err
}
