package agreements

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrInternal        = errors.New("agreements: internal error")
	ErrInvalidArgument = errors.New("agreements: invalid argument")
	ErrNotFound        = errors.New("agreements: not found")
)

type (
	Service interface {
		Sign(ctx context.Context, args *SignArgs) ([]Signature, error)
		GetSignatures(ctx context.Context, identityID string) ([]Signature, error)
		Get(ctx context.Context, id string) (*Agreement, error)
	}

	ServiceArgs struct {
		Db            *sqlx.DB `validate:"required"`
		AgreementsDir string   `validate:"required"`
	}

	service struct {
		validator     *validator.Validate
		db            *sqlx.DB
		agreementsDir string
	}

	Signature struct {
		ID          string `db:"id"`
		AgreementID string `db:"agreement_id"`
		IdentityID  string `db:"identity_id"`
		IPAddress   string `db:"ip_address"`
		CreatedAt   string `db:"created_at"`
		UpdatedAt   string `db:"updated_at"`
	}

	Agreement struct {
		ID        string `db:"id"`
		Name      string `db:"name"`
		Version   string `db:"version"`
		Content   string `db:"content"`
		CreatedAt string `db:"created_at"`
		UpdatedAt string `db:"updated_at"`
	}
)

func NewService(args *ServiceArgs) (Service, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	err := StoreAgreements(context.Background(), &StoreAgreementsArgs{
		db:  args.Db,
		dir: args.AgreementsDir,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return &service{
		validator:     v,
		db:            args.Db,
		agreementsDir: args.AgreementsDir,
	}, nil
}

func (s *service) Get(ctx context.Context, id string) (*Agreement, error) {
	var agreement Agreement

	err := s.db.GetContext(ctx, &agreement, "SELECT * FROM agreements WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w %s", ErrNotFound, err.Error())
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return &agreement, nil
}

type SignArgs struct {
	AgreementIDs []string `validate:"required"`
	IdentityID   string   `validate:"required"`
	IPAddress    string   `validate:"required,ip_addr"`
}

func (s *service) Sign(ctx context.Context, args *SignArgs) ([]Signature, error) {
	if err := s.validator.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	signatures := make([]Signature, 0, len(args.AgreementIDs))

	for _, id := range args.AgreementIDs {
		var signRecord Signature

		err := s.db.GetContext(ctx, &signRecord, "INSERT INTO agreement_signatures (agreement_id, identity_id, ip_address) VALUES ($1, $2, $3) RETURNING *", id, args.IdentityID, args.IPAddress)
		if err != nil {
			if pgErr, isPGErr := err.(pq.Error); isPGErr {
				if pgErr.Code != "23503" {
					return nil, fmt.Errorf("%w %s", ErrNotFound, err.Error())
				}
			}
			return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
		}

		signatures = append(signatures, signRecord)
	}

	return signatures, nil
}

func (s *service) GetSignatures(ctx context.Context, identityID string) ([]Signature, error) {
	var agreementSigns []Signature

	err := s.db.SelectContext(ctx, &agreementSigns, "SELECT * FROM agreement_signatures WHERE identity_id = $1", identityID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	var signatures []Signature
	for _, sign := range agreementSigns {
		signatures = append(signatures, Signature{
			ID:          sign.ID,
			AgreementID: sign.AgreementID,
			IdentityID:  sign.IdentityID,
			IPAddress:   sign.IPAddress,
			CreatedAt:   sign.CreatedAt,
			UpdatedAt:   sign.UpdatedAt,
		})
	}

	return signatures, nil
}

type StoreAgreementsArgs struct {
	dir string
	db  *sqlx.DB
}

func StoreAgreements(ctx context.Context, args *StoreAgreementsArgs) error {
	agreementFiles, err := ioutil.ReadDir(args.dir)
	if err != nil {
		return fmt.Errorf("%w %s", ErrNotFound, err.Error())
	}

	regex, err := regexp.Compile(`^[a-zA-Z0-9_]+-[0-9]+\.[0-9]+\.[0-9]+\.md$`)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	for _, agreementFile := range agreementFiles {
		if !regex.MatchString(agreementFile.Name()) {
			return fmt.Errorf("%w %s", ErrInternal, "invalid agreement file name format")
		}

		agreementID := agreementFile.Name()[:len(agreementFile.Name())-3]
		agreementName := agreementID[:strings.Index(agreementID, "-")]
		agreementVersion := agreementID[strings.Index(agreementID, "-")+1:]

		agreementContent, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", args.dir, agreementFile.Name()))
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}

		_, err = args.db.ExecContext(ctx, `INSERT INTO agreements (id, name, version, content) VALUES ($1, $2, $3, $4)`, agreementID, agreementName, agreementVersion, string(agreementContent))
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}
	}

	return nil
}
