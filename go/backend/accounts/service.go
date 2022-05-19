package accounts

//go:generate mockgen -destination=./mock.go -package=accounts -source=./service.go

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_country "gitlab.com/fynbos/backend/country"
	_identity "gitlab.com/fynbos/backend/identity"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
)

const (
	Verified string = "verified"
)

var (
	ErrInternal        = errors.New("accounts service: internal error.")
	ErrDuplicate       = errors.New("accounts service: duplicate.")
	ErrInvalidArgument = errors.New("accounts service: invalid argument.")
	ErrNotFound        = errors.New("accounts service: not found.")
)

type Account struct {
	ID                         string
	DebitsAccepted             uint64
	DebitsReserved             uint64
	CreditsAccepted            uint64
	CreditsReserved            uint64
	AvailableBalance           int64
	IdentityID                 string `db:"identity_id"`
	LedgerAccountID            string `db:"ledger_account_id"` // id returned by Pacioli.
	Provider                   string
	ProviderID                 string `db:"provider_id"`
	VerificationState          string `db:"verification_state"`
	DebitsMustNotExceedCredits bool
	CreditsMustNotExceedDebits bool
	CreatedAt                  string `db:"created_at"`
	UpdatedAt                  string `db:"updated_at"`
}

func (s Account) IsVerified() bool {
	return s.VerificationState == Verified
}

type Service interface {
	Init(ctx context.Context) error
	Create(ctx context.Context, args *CreateAccountArgs) (*Account, error)
	GetByIdentityIDWithTrx(ctx context.Context, tx *sqlx.Tx, id string) (*Account, error)
	GetByIdentityID(ctx context.Context, id string) (*Account, error)
	Get(ctx context.Context, id string) (*Account, error)
	VerifyWithTx(ctx context.Context, tx *sqlx.Tx, args *VerifyArgs) (*Account, error)
	CanMakeOutgoingPayment(acc *Account, identityID string) bool
	CanMakeDeposit(acc *Account, identityID string) bool
	CanCreateFundingSource(acc *Account, identityID string) bool
	CanVerifyFundingSource(acc *Account, identityID string) bool
}

type service struct {
	db              *sqlx.DB
	is              _identity.Service
	cs              _country.Service
	validator       *validator.Validate
	pacioliLedgerID uint16
	pacioliClient   pacioliv1.PacioliServiceClient
}

type ServiceArgs struct {
	Is              _identity.Service `validate:"required"`
	Cs              _country.Service  `validate:"required"`
	PacioliTenant   string
	PacioliLedgerID uint16
	PacioliClient   pacioliv1.PacioliServiceClient `validate:"required"`
	Db              *sqlx.DB                       `validate:"required"`
}

func NewService(args *ServiceArgs) (Service, error) {
	validator := validator.New()
	if err := validator.Struct(args); err != nil {
		return nil, err
	}

	return &service{
		db:              args.Db,
		is:              args.Is,
		cs:              args.Cs,
		pacioliLedgerID: args.PacioliLedgerID,
		pacioliClient:   args.PacioliClient,
		validator:       validator,
	}, nil
}

func (s *service) Init(ctx context.Context) error {
	// TODO: create tenant when auth is working
	response, err := s.pacioliClient.ConfigureLedgers(ctx, &pacioliv1.ConfigureLedgersRequest{
		Args: []*pacioliv1.Ledger{
			{
				Id:    uint32(s.pacioliLedgerID),
				Name:  "Fynbos ledger",
				Asset: "840", // US dollars
				Scale: 2,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("Failed to configure ledgers. %w %s", ErrInternal, err.Error())
	}
	eventErrors := response.GetErrors()
	if len(eventErrors) > 0 {
		return fmt.Errorf("Failed to configure ledgers: %w %+v", ErrInternal, eventErrors)
	}

	return nil
}

type CreateAccountArgs struct {
	IdentityID string `validate:"required,uuid"`
	Country    string `validate:"iso3166_1_alpha2"`

	// Points to the next account in array. Last one in array cannot have linked flag set.
	Linked                     bool
	DebitMustNotExceedCredits  bool
	CreditsMustNotExceedDebits bool
}

// This will create the ledger account in Pacioli first and then inserts an account entry into
// CRDB. It therefore suffers from the dual-write problem so we could end up with dangling accounts
// in Pacioli.
// We create the Pacioli account first so that the account in CRDB will always have a Pacioli ID.
func (s *service) Create(ctx context.Context, args *CreateAccountArgs) (*Account, error) {
	err := s.validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	identity, err := s.is.Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	var ret Account
	err = crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		// pacioli client must be configured with retries and exponential backoff at higher level.
		ledgerAccountID := uuid.NewString()
		response, err := s.pacioliClient.ConfigureAccounts(ctx, &pacioliv1.ConfigureAccountsRequest{
			Args: []*pacioliv1.ConfigureAccountsArgs{
				{
					Id:       ledgerAccountID,
					LedgerId: uint32(s.pacioliLedgerID),
					Flags: &pacioliv1.AccountFlags{
						Linked:                     args.Linked,
						DebitsMustNotExceedCredits: args.DebitMustNotExceedCredits,
						CreditsMustNotExceedDebits: args.CreditsMustNotExceedDebits,
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}
		eventErrors := response.GetErrors()
		if len(eventErrors) > 0 {
			return fmt.Errorf("Failed to create account in pacioli. %w %+v", ErrInternal, eventErrors)
		}

		stmt, err := tx.PrepareNamed("INSERT INTO accounts (identity_id, ledger_account_id) VALUES (:identityid, :ledgeraccountid) RETURNING *")
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}

		err = stmt.Stmt.Get(&ret, identity.ID, ledgerAccountID)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = s.fetchFromPacioli(ctx, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *service) GetByIdentityIDWithTrx(ctx context.Context, tx *sqlx.Tx, identityID string) (*Account, error) {
	if identityID == "" {
		return nil, fmt.Errorf("%w IdentityID is required.", ErrInvalidArgument)
	}

	var ret Account
	err := tx.Get(&ret, "SELECT * FROM accounts WHERE identity_id=$1 LIMIT 1", identityID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	err = s.fetchFromPacioli(ctx, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *service) GetByIdentityID(ctx context.Context, identityID string) (*Account, error) {
	var acc *Account
	err := crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		_acc, err := s.GetByIdentityIDWithTrx(ctx, tx, identityID)
		if err != nil {
			return err
		}
		acc = _acc
		return nil
	})

	return acc, err
}

func (s *service) Get(ctx context.Context, accountID string) (*Account, error) {
	if accountID == "" {
		return nil, fmt.Errorf("%w AccountID is required.", ErrInvalidArgument)
	}

	var ret Account
	err := s.db.Get(&ret, "SELECT * FROM accounts WHERE id=$1 LIMIT 1", accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	err = s.fetchFromPacioli(ctx, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

// Fetches ledger account from Pacioli and merges with the specified account.
func (s *service) fetchFromPacioli(ctx context.Context, account *Account) error {
	if account == nil {
		return fmt.Errorf("%w Account is required.", ErrInvalidArgument)
	}
	response, err := s.pacioliClient.GetAccounts(ctx, &pacioliv1.GetAccountsRequest{
		Ids: []string{account.LedgerAccountID},
	})
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err.Error())
	}
	if len(response.GetAccounts()) != 1 {
		return fmt.Errorf("%w %s", ErrInternal, "Account not found in pacioli.")
	}
	ledgerAccount := response.GetAccounts()[0]
	if ledgerAccount.Id != account.LedgerAccountID {
		// Panic-ing here as something is wrong - possibly encoding of uuids to byte slice.
		panic("Accounts service: Ledger account ID does not match that returned from Pacioli.")
	}

	account.CreditsAccepted = ledgerAccount.CreditsAccepted
	account.CreditsReserved = ledgerAccount.CreditsReserved
	account.DebitsAccepted = ledgerAccount.DebitsAccepted
	account.DebitsReserved = ledgerAccount.DebitsReserved
	account.CreditsMustNotExceedDebits = ledgerAccount.Flags.CreditsMustNotExceedDebits
	account.DebitsMustNotExceedCredits = ledgerAccount.Flags.DebitsMustNotExceedCredits

	// Calculate available balance
	account.AvailableBalance = int64(account.DebitsAccepted - account.CreditsAccepted - account.CreditsReserved)

	return nil
}

type VerifyArgs struct {
	AccountID  string `validate:"required,uuid"`
	Provider   string `validate:"oneof=noop"`
	ProviderID string `validate:"required"`
}

func (s *service) VerifyWithTx(ctx context.Context, tx *sqlx.Tx, args *VerifyArgs) (*Account, error) {
	//TODO: refactor errors
	err := s.validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInvalidArgument, err.Error())
	}

	acc, err := s.Get(ctx, args.AccountID)
	if err != nil {
		return nil, err
	}

	stmt, err := tx.PrepareNamed(`
			UPDATE accounts
			SET provider=$1, provider_id=$2, verification_state=$3
			WHERE id=$4
			RETURNING *;
		`)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	var verifiedAccount Account
	err = stmt.Stmt.Get(&verifiedAccount, args.Provider, args.ProviderID, Verified, acc.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err.Error())
	}

	return &verifiedAccount, nil
}

func (s service) CanMakeOutgoingPayment(acc *Account, identityID string) bool {
	if acc == nil {
		return false
	}

	return acc.IdentityID == identityID
}

func (s service) CanMakeDeposit(acc *Account, identityID string) bool {
	if acc == nil {
		return false
	}

	return acc.IdentityID == identityID
}

func (s service) CanCreateFundingSource(acc *Account, identityID string) bool {
	if acc == nil {
		return false
	}

	return acc.IdentityID == identityID
}

func (s service) CanVerifyFundingSource(acc *Account, identityID string) bool {
	if acc == nil {
		return false
	}

	return acc.IdentityID == identityID
}
