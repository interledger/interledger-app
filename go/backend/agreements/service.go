package agreements

//go:generate mockgen -destination=./mock.go -package=agreements -source=./service.go

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
		Sign(ctx context.Context, args *SignArgs) error
		GetSignatures(ctx context.Context, identityID string) ([]Signature, error)
		Get(ctx context.Context, id string) (*Agreement, error)
	}

	ServiceArgs struct {
		Db *sqlx.DB `validate:"required"`
	}

	service struct {
		validator *validator.Validate
		db        *sqlx.DB
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

	return &service{
		validator: v,
		db:        args.Db,
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

func (s *service) Sign(ctx context.Context, args *SignArgs) error {
	if err := s.validator.Struct(args); err != nil {
		return fmt.Errorf("%w %s", ErrInvalidArgument, err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
	}()
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	txStmt, err := tx.PrepareContext(ctx, "INSERT INTO agreement_signatures (agreement_id, identity_id, ip_address) VALUES ($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err.Error())
	}
	defer txStmt.Close()

	for _, id := range args.AgreementIDs {
		_, err := txStmt.ExecContext(ctx, id, args.IdentityID, args.IPAddress)
		if err != nil {
			if pgErr, isPGErr := err.(pq.Error); isPGErr {
				if pgErr.Code != "23503" {
					return fmt.Errorf("%w %s", ErrNotFound, err.Error())
				}
			}
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return nil
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

type MigrateArgs struct {
	DirectoryPath string
	Db            *sqlx.DB
}

func MigrateFromMarkdowns(ctx context.Context, args *MigrateArgs) error {
	agreementFiles, err := ioutil.ReadDir(args.DirectoryPath)
	if err != nil {
		return fmt.Errorf("%w %s", ErrNotFound, err.Error())
	}

	tx, err := args.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
	}()
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	regex, err := regexp.Compile(`^[a-zA-Z0-9_]+-[0-9]+\.[0-9]+\.[0-9]+\.md$`)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	txStmt, err := tx.PrepareContext(ctx, `INSERT INTO agreements (id, name, version, content) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err.Error())
	}
	defer txStmt.Close()

	var agreements []string
	err = args.Db.SelectContext(ctx, &agreements, "SELECT id FROM agreements")
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	// convert to map for faster lookup
	agreementsMap := make(map[string]bool)
	for _, agreement := range agreements {
		agreementsMap[agreement] = true
	}

	for _, agreementFile := range agreementFiles {
		if !regex.MatchString(agreementFile.Name()) {
			return fmt.Errorf("%w %s", ErrInternal, "invalid agreement file name format")
		}

		agreementID := agreementFile.Name()[:len(agreementFile.Name())-3]
		if _, ok := agreementsMap[agreementID]; ok {
			continue
		}

		agreementName := agreementID[:strings.Index(agreementID, "-")]
		agreementVersion := agreementID[strings.Index(agreementID, "-")+1:]

		agreementContent, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", args.DirectoryPath, agreementFile.Name()))
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}

		_, err = txStmt.Exec(agreementID, agreementName, agreementVersion, string(agreementContent))
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return nil
}
