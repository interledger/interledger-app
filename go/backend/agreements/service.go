package agreements

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrInternal        = errors.New("agreements: internal error")
	ErrInvalidArgument = errors.New("agreements: invalid argument")
)

type (
	Service interface {
		SignAgreement(ctx context.Context, args *SignAgreementArgs) (*SignedAgreements, error)
	}

	ServiceArgs struct {
		Db *sqlx.DB `validate:"required"`
	}

	service struct {
		validator *validator.Validate
		db        *sqlx.DB
	}

	SignedAgreements struct {
		ID           string         `db:"id"`
		AgreementIDs pq.StringArray `db:"agreement_ids"`
		IdentityID   string         `db:"identity_id"`
		IPAddress    string         `db:"ip_address"`
		CreatedAt    string         `db:"created_at"`
		UpdatedAt    string         `db:"updated_at"`
	}
)

func NewService(args *ServiceArgs) (Service, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	return &service{
		validator: v,
		db:        args.Db,
	}, nil
}

type SignAgreementArgs struct {
	AgreementIDs []string `validate:"required"`
	IdentityID   string   `validate:"required"`
	IPAddress    string   `validate:"required,ip_addr"`
}

func (s *service) SignAgreement(ctx context.Context, args *SignAgreementArgs) (*SignedAgreements, error) {
	if err := s.validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	var signedAgreements SignedAgreements

	err := s.db.GetContext(ctx, &signedAgreements, "INSERT INTO signed_agreements (agreement_ids, identity_id, ip_address) VALUES ($1, $2, $3) RETURNING *", pq.StringArray(args.AgreementIDs), args.IdentityID, args.IPAddress)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return &signedAgreements, nil
}
