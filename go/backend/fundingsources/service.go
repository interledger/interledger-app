package fundingsources

import (
	"context"
	"database/sql"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_identity "gitlab.com/fynbos/backend/identity"
)

type Service interface {
	Create(ctx context.Context, tx *sqlx.Tx, args *CreateArgs) (*FundingSource, error)
	Get(ctx context.Context, tx *sqlx.Tx, id string) (*FundingSource, error)
}

type service struct {
	validator *validator.Validate
	identity  _identity.Service
}

type ServiceArgs struct {
	Identity _identity.Service `validate:"required"`
}

func NewService(args *ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, &ErrInvalidArgument{Err: err.Error()}
	}

	return &service{validator: validator, identity: args.Identity}, nil
}

type FundingSource struct {
	ID                string
	IdentityID        string `db:"identity_id"`
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
	Name              string `validate:"required"`
	Mask              string `validate:"required"`
	VerificationState string `validate:"required"`
	Type              string `validate:"oneof=noop"`
	TypeID            string `validate:"required,uuid4"`
	SubType           string `validate:"required"`
}

func (s *service) Create(ctx context.Context, tx *sqlx.Tx, args *CreateArgs) (*FundingSource, error) {
	err := s.validator.Struct(args)
	if err != nil {
		return nil, &ErrInvalidArgument{Err: err.Error()}
	}

	identity, err := s.identity.Get(ctx, tx, args.IdentityID)
	if err != nil {
		switch err.(type) {
		case *_identity.ErrInvalidArgument:
		case *_identity.ErrNotFound:
			return nil, &ErrInvalidArgument{Err: "Identity must exist to create funding source."}
		default:
			return nil, &ErrInternalError{Err: err.Error()}
		}
	}

	var fs FundingSource
	stmt, err := tx.PrepareNamed(`
			INSERT INTO funding_sources (
				identity_id, name, mask, verification_state, type, type_id, subtype
			)
			VALUES (:identity_id, :name, :mask, :verification_state, :type, :type_id, :subtype)
			RETURNING *;
		`)
	if err != nil {
		return nil, &ErrInternalError{Err: err.Error()}

	}

	err = stmt.Stmt.Get(&fs,
		identity.ID,
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
