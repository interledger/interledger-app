package accounts

import (
	"context"
	"database/sql"
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

type Account struct {
	ID                string
	DebitsAccepted    uint64
	DebitsReserved    uint64
	CreditsAccepted   uint64
	CreditsReserved   uint64
	AvailableBalance  int64
	IdentityID        string `db:"identity_id"`
	LedgerAccountID   string `db:"ledger_account_id"` // id returned by Pacioli.
	Provider          string
	ProviderID        string `db:"provider_id"`
	VerificationState string `db:"verification_state"`
	CreatedAt         string `db:"created_at"`
	UpdatedAt         string `db:"updated_at"`
	// TODO: add flags
}

func (s Account) IsVerified() bool {
	return s.VerificationState == Verified
}

type Service interface {
	Init(ctx context.Context) error
	Create(ctx context.Context, tx *sqlx.Tx, args *CreateAccountArgs) (*Account, error)
	GetByIdentityIDWithTrx(ctx context.Context, tx *sqlx.Tx, id string) (*Account, error)
	GetByIdentityID(ctx context.Context, id string) (*Account, error)
	Get(ctx context.Context, tx *sqlx.Tx, id string) (*Account, error)
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
		return &ErrInternalError{Err: err.Error()}
	}
	eventErrors := response.GetErrors()
	if len(eventErrors) > 0 {
		return &ErrInternalError{Err: fmt.Sprintf("Failed to configure ledgers. %+v", eventErrors)}
	}

	return nil
}

type CreateAccountArgs struct {
	IdentityID string `validate:"required,uuid"`
	Country    string `validate:"iso3166_1_alpha2"`
	// TODO: add flags
}

// This will create the ledger account in Pacioli first and then inserts an account entry into
// CRDB. It therefore suffers from the dual-write problem so we could end up with dangling accounts
// in Pacioli.
// We create the Pacioli account first so that the account in CRDB will always have a Pacioli ID.
func (s *service) Create(ctx context.Context, tx *sqlx.Tx, args *CreateAccountArgs) (*Account, error) {
	err := s.validator.Struct(args)
	if err != nil {
		return nil, &ErrInvalidArgument{Err: err.Error()}
	}
	_, err = s.cs.GetByAlpha2(ctx, tx, args.Country)
	if err != nil {
		return nil, &ErrInvalidArgument{Err: "Unknown or unsupported country: " + args.Country}
	}
	identity, err := s.is.Get(ctx, tx, args.IdentityID)
	// TODO: perhaps a global error set?
	if err != nil {
		switch err.(type) {
		case *_identity.ErrInvalidArgument:
		case *_identity.ErrNotFound:
			return nil, &ErrInvalidArgument{Err: "Identity must exist."}
		default:
			return nil, &ErrInternalError{Err: err.Error()}
		}
	}

	// pacioli client must be configured with retries and exponential backoff at higher level.
	ledgerAccountID := uuid.NewString()
	response, err := s.pacioliClient.ConfigureAccounts(ctx, &pacioliv1.ConfigureAccountsRequest{
		Args: []*pacioliv1.ConfigureAccountsArgs{
			{
				Id:       ledgerAccountID,
				LedgerId: uint32(s.pacioliLedgerID),
				// TODO: add flags
			},
		},
	})
	if err != nil {
		return nil, &ErrInternalError{Err: err.Error()}
	}
	eventErrors := response.GetErrors()
	if len(eventErrors) > 0 {
		return nil, &ErrInternalError{Err: "Failed to create account in pacioli."}
	}

	stmt, err := tx.PrepareNamed("INSERT INTO accounts (identity_id, ledger_account_id) VALUES (:identityid, :ledgeraccountid) RETURNING *")
	if err != nil {
		return nil, err
	}

	var ret Account
	err = stmt.Stmt.Get(&ret, identity.ID, ledgerAccountID)
	if err != nil {
		return nil, &ErrInternalError{Err: err.Error()}
	}

	return &ret, nil
}

func (s *service) GetByIdentityIDWithTrx(ctx context.Context, tx *sqlx.Tx, identityID string) (*Account, error) {
	if identityID == "" {
		return nil, &ErrInvalidArgument{Err: "identityID is required."}
	}

	var ret Account
	err := tx.Get(&ret, "SELECT * FROM accounts WHERE identity_id=$1 LIMIT 1", identityID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{Err: "Not found."}
		}

		return nil, &ErrInternalError{Err: err.Error()}
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

func (s *service) Get(ctx context.Context, tx *sqlx.Tx, accountID string) (*Account, error) {
	if accountID == "" {
		return nil, &ErrInvalidArgument{Err: "Accounts service: accountID is required."}
	}

	var ret Account
	err := tx.Get(&ret, "SELECT * FROM accounts WHERE id=$1 LIMIT 1", accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{Err: "Accounts service: Not found."}
		}

		return nil, &ErrInternalError{Err: err.Error()}
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
		return &ErrInternalError{Err: "Accounts service: account needs to specified to fetch from Pacioli."}
	}
	response, err := s.pacioliClient.GetAccounts(ctx, &pacioliv1.GetAccountsRequest{
		Ids: []string{account.LedgerAccountID},
	})
	if err != nil {
		return &ErrInternalError{Err: err.Error()}
	}
	if len(response.GetAccounts()) != 1 {
		return &ErrNotFound{Err: "Account not in pacioli."}
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
		return nil, &ErrInvalidArgument{Err: "Accounts service: " + err.Error()}
	}

	acc, err := s.Get(ctx, tx, args.AccountID)
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
		return nil, &ErrInternalError{Err: "Accounts service: " + err.Error()}
	}

	var verifiedAccount Account
	err = stmt.Stmt.Get(&verifiedAccount, args.Provider, args.ProviderID, Verified, acc.ID)
	if err != nil {
		return nil, &ErrInternalError{Err: "Accounts service: " + err.Error()}
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

// Error set
// TODO: wrapping errors instead to preserve stack.
type ErrInvalidArgument struct {
	Err string
}

func (r *ErrInvalidArgument) Error() string {
	return r.Err
}

type ErrInternalError struct {
	Err string
}

func (r *ErrInternalError) Error() string {
	return r.Err
}

type ErrNotFound struct {
	Err string
}

func (r *ErrNotFound) Error() string {
	return r.Err
}
