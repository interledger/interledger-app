package fundingsources

//go:generate mockgen -destination=./mock.go -package=fundingsources -source=./service.go

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	_identity "gitlab.com/fynbos/backend/identity"
	_mx "gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/noop"
	_unit "gitlab.com/fynbos/backend/providers/unit"
	"go.temporal.io/api/enums/v1"
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
	CreateMxBankAccount(ctx context.Context, args *CreateMxBankAccountArgs) (*FundingSource, error)
	GetMxConnectWidget(ctx context.Context, accountID string, identityID string) (string, error)
	VerifyMxBankAccount(ctx context.Context, identityID string, fundingsourceID string) (*FundingSource, error)
	CreateUnitCounterPartyFromMxAccount(ctx context.Context, fundingsourceID string) (*UnitCounterParty, error)
	GetUnitCounterParty(ctx context.Context, fundingsourceID string) (*UnitCounterParty, error)
	CreateUnitCounterParty(ctx context.Context, fundingsourceID string, unitCounterPartyID string) (*UnitCounterParty, error)
}

type service struct {
	validator *validator.Validate
	db        *sqlx.DB
	is        _identity.Service
	as        accounts.Service
	noop      noop.Service
	mx        _mx.Service
	tp        client.Client
	unit      _unit.Service
}

type ServiceArgs struct {
	Is   _identity.Service `validate:"required"`
	As   accounts.Service  `validate:"required"`
	Db   *sqlx.DB          `validate:"required"`
	Noop noop.Service      `validate:"required"`
	Mx   _mx.Service       `validate:"required"`
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
		mx:        args.Mx,
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

	var fs FundingSource
	err = s.db.GetContext(
		ctx,
		&fs,
		`
			INSERT INTO funding_sources (
				account_id, name, mask, verification_state, type, subtype
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING *;
		`,
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

type CreateMxBankAccountArgs struct {
	IdentityID   string `validate:"required"`
	AccountID    string `validate:"required"`
	MxUserGuid   string `validate:"required"`
	MxMemberGuid string `validate:"required"`
	Name         string `validate:"required"`
}

func (s *service) CreateMxBankAccount(
	ctx context.Context,
	args *CreateMxBankAccountArgs,
) (*FundingSource, error) {
	acc, err := s.as.Get(ctx, args.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	if acc.IdentityID != args.IdentityID {
		return nil, ErrUnauthorized
	}

	fundingSource, err := s.Create(ctx, &CreateArgs{
		IdentityID:        args.IdentityID,
		AccountID:         args.AccountID,
		Name:              args.Name,
		VerificationState: "processing",
		Type:              "mx",
		SubType:           "bank",
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	_, err = s.tp.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:                    "create_mx_bank_account_" + fundingSource.ID,
			TaskQueue:             "backend",
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		},
		CreateMxBankAccountWorkflow,
		&CreateMxBankAccountWorkflowArgs{
			IdentityID:      args.IdentityID,
			FundingSourceID: fundingSource.ID,
			MxUserGuid:      args.MxUserGuid,
			MxMemberGuid:    args.MxMemberGuid,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return fundingSource, nil

}

func (s *service) VerifyMxBankAccount(
	ctx context.Context,
	identityID string,
	fundingsourceID string,
) (*FundingSource, error) {
	if identityID == "" {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, "IdentityID is required.")
	}

	if fundingsourceID == "" {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, "FundingSourceID is required.")
	}

	_, err := s.Get(ctx, fundingsourceID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	// TODO: authz on identityID
	mxFs, err := s.mx.GetMxFundingSource(ctx, fundingsourceID) // we map 1-1 between funding source and mx account
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	ownerDetails, err := s.mx.GetAccountOwner(ctx, mxFs.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	acc, err := s.as.Get(ctx, mxFs.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	user, err := s.is.Get(ctx, acc.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	// TODO: how do we verify ownership
	var ret *FundingSource = nil
	var verifyErr error
	if user.Email == strings.TrimSpace(ownerDetails.Email) {
		ret, verifyErr = s.Verify(ctx, &VerifyArgs{
			IdentityID:      identityID,
			FundingSourceID: fundingsourceID,
		})
	} else {
		// TODO: surface to customer service for manual verification.
		verifyErr = fmt.Errorf(
			"%w user does not own bank account. userID=%s, fundingsourceID=%s",
			ErrUnauthorized,
			user.ID,
			fundingsourceID,
		)
	}

	return ret, verifyErr
}

func (s *service) CreateUnitCounterPartyFromMxAccount(
	ctx context.Context,
	fundingsourceID string,
) (*UnitCounterParty, error) {
	mxFs, err := s.Get(ctx, fundingsourceID)
	if err != nil {
		return nil, err
	}
	if mxFs.Type != "mx" {
		return nil, fmt.Errorf("%w Funding source is not an mx account.", ErrInternal)
	}

	unitCustomer, err := s.unit.GetCustomerByAccountID(ctx, mxFs.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w No unit customer found for accountID=%s.", ErrInternal, mxFs.AccountID)
	}

	// perform this just before creating the counter party as we get charged for Mx api calls.
	accountNumbers, err := s.mx.GetMxAccount(ctx, fundingsourceID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	idempotencyKey := sha256.Sum256([]byte(fundingsourceID))
	cp, err := s.unit.CreateCounterParty(ctx, &_unit.CreateCounterPartyArgs{
		Name:           mxFs.Name,
		RoutingNumber:  accountNumbers.RoutingNumber,
		AccountNumber:  accountNumbers.AccountNumber,
		AccountType:    accountNumbers.Type,
		Type:           "person",
		IdempotencyKey: string(idempotencyKey[0:]),
		UnitCustomerID: unitCustomer.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	ret, err := s.CreateUnitCounterParty(ctx, fundingsourceID, cp.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, err
}

func (s *service) CreateUnitCounterParty(
	ctx context.Context,
	fundingsourceID string,
	unitCounterPartyID string,
) (*UnitCounterParty, error) {
	ret := &UnitCounterParty{}
	err := s.db.GetContext(
		ctx,
		ret,
		"INSERT INTO unit_counterparties (id, unit_counterparty_id) VALUES ($1, $2) RETURNING *;",
		fundingsourceID,
		unitCounterPartyID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
	}

	return ret, nil
}

func (s *service) GetUnitCounterParty(
	ctx context.Context,
	fundingsourceID string,
) (*UnitCounterParty, error) {
	ret := &UnitCounterParty{}
	err := s.db.GetContext(
		ctx,
		ret,
		"SELECT * FROM unit_counterparties WHERE id=$1;",
		fundingsourceID)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else {
		if err != nil {
			return nil, fmt.Errorf("%w %s", ErrInternal, err)
		}
	}

	return ret, nil
}

func IsVerified(fs *FundingSource) bool {
	return fs.VerificationState == "verified"
}
