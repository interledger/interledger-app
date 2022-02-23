package accounts

import (
	"context"
	"database/sql"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_country "gitlab.com/fynbos/backend/country"
	_identity "gitlab.com/fynbos/backend/identity"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
)

type Account struct {
	ID               string
	DebitsAccepted   uint64
	DebitsReserved   uint64
	CreditsAccepted  uint64
	CreditsReserved  uint64
	AvailableBalance int64
	IdentityID       string `db:"identity_id"`
	LedgerAccountID  string `db:"ledger_account_id"` // id returned by Pacioli.
	CreatedAt        string `db:"created_at"`
	UpdatedAt        string `db:"updated_at"`
}

type Service interface {
	Create(ctx context.Context, tx *sqlx.Tx, args *CreateAccountArgs) (*Account, error)
	GetByIdentityID(ctx context.Context, tx *sqlx.Tx, id string) (*Account, error)
	Get(ctx context.Context, tx *sqlx.Tx, id string) (*Account, error)
}

type service struct {
	identity        _identity.Service
	country         _country.Service
	validator       *validator.Validate
	pacioliLedgerID string
	pacioliClient   pacioliv1.PacioliServiceClient
}

func NewService(
	identity _identity.Service,
	country _country.Service,
	pacioliLedgerCode uint16,
	pacioliClient pacioliv1.PacioliServiceClient,
) (Service, error) {
	if identity == nil {
		return nil, &ErrInvalidArgument{Err: "Identity is required."}
	}
	if pacioliClient == nil {
		return nil, &ErrInvalidArgument{Err: "Pacioli client is required."}
	}
	// we do not use 0 so it can't be confused with default value.
	if pacioliLedgerCode == 0 {
		return nil, &ErrInvalidArgument{Err: "Pacioli ledger code is required."}
	}

	// TODO: re-work configuration when grpc auth is introduced.
	ctx := context.Background() // TODO: timeout
	ledger, err := pacioliClient.GetLedgerByCode(ctx, &pacioliv1.GetLedgerByCodeRequest{
		Code: uint32(pacioliLedgerCode),
	})
	if err != nil {
		return nil, err
	}

	return &service{
		identity:        identity,
		country:         country,
		pacioliLedgerID: ledger.Id,
		pacioliClient:   pacioliClient,
		validator:       validator.New(),
	}, nil
}

type CreateAccountArgs struct {
	IdentityID string `validate:"required,uuid"`
	Country    string `validate:"iso3166_1_alpha2"`
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
	ctry, err := s.country.GetByAlpha2(ctx, tx, args.Country)
	if err != nil {
		return nil, &ErrInvalidArgument{Err: "Unknown or unsupported country: " + args.Country}
	}
	identity, err := s.identity.Get(ctx, tx, args.IdentityID)
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
	ledgerAccount, err := s.pacioliClient.CreateAccount(ctx, &pacioliv1.CreateAccountRequest{
		LedgerID: s.pacioliLedgerID,
		Unit:     uint32(ctry.NumericCode), // grpc does not have uint16 natively
	})
	if err != nil {
		return nil, &ErrInternalError{Err: err.Error()}
	}

	stmt, err := tx.PrepareNamed("INSERT INTO accounts (identity_id, ledger_account_id) VALUES (:identityid, :ledgeraccountid) RETURNING *")
	if err != nil {
		return nil, err
	}

	var ret Account
	err = stmt.Stmt.Get(&ret, identity.ID, ledgerAccount.Id)
	if err != nil {
		return nil, &ErrInternalError{Err: err.Error()}
	}

	return &ret, nil
}

func (s *service) GetByIdentityID(ctx context.Context, tx *sqlx.Tx, identityID string) (*Account, error) {
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
	ledgerAccount, err := s.pacioliClient.GetAccount(ctx, &pacioliv1.GetAccountRequest{
		LedgerID: s.pacioliLedgerID,
		Id:       account.LedgerAccountID,
	})
	if err != nil {
		return &ErrInternalError{Err: err.Error()}
	}
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
