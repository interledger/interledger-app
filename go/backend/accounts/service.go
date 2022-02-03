package accounts

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	_identity "gitlab.com/fynbos/backend/identity"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
)

type Account struct {
	ID              string
	DebitsAccepted  uint64
	DebitsReserved  uint64
	CreditsAccepted uint64
	CreditsReserved uint64
	IdentityID      string `db:"identity_id"`
	LedgerAccountID string `db:"ledger_account_id"` // id returned by Pacioli.
	CreatedAt       string `db:"created_at"`
	UpdatedAt       string `db:"updated_at"`
}

type Service interface {
	Create(ctx context.Context, tx *sqlx.Tx, args *CreateAccountArgs) (*Account, error)
	GetByIdentityID(ctx context.Context, tx *sqlx.Tx, id string) (*Account, error)
}

type service struct {
	identity        _identity.Service
	pacioliLedgerID string
	pacioliClient   pacioliv1.PacioliServiceClient
}

func NewService(
	identity _identity.Service,
	pacioliLederID string,
	pacioliClient pacioliv1.PacioliServiceClient,
) (Service, error) {
	if pacioliLederID == "" {
		return nil, &ErrInvalidArgument{Err: "Pacioli ledger ID is required."}
	}
	if identity == nil {
		return nil, &ErrInvalidArgument{Err: "Identity is required."}
	}
	if pacioliClient == nil {
		return nil, &ErrInvalidArgument{Err: "Pacioli client is required."}
	}

	return &service{
		identity:        identity,
		pacioliLedgerID: pacioliLederID,
		pacioliClient:   pacioliClient,
	}, nil
}

type CreateAccountArgs struct {
	IdentityID string
	Country    string // 3-letter ISO 4217 country code
}

// This will create the ledger account in Pacioli first and then inserts an account entry into
// CRDB. It therefore suffers from the dual-write problem so we could end up with dangling accounts
// in Pacioli.
// We create the Pacioli account first so that the account in CRDB will always have a Pacioli ID.
func (s *service) Create(ctx context.Context, tx *sqlx.Tx, args *CreateAccountArgs) (*Account, error) {
	currencyCode, err := currencyCode(args.Country)
	if err != nil {
		return nil, &ErrInvalidArgument{Err: err.Error()}
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
		Unit:     uint32(currencyCode), // grpc does not have uint16 natively
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

	// fetch ledger account from pacioli and merge.
	ledgerAccount, err := s.pacioliClient.GetAccount(ctx, &pacioliv1.GetAccountRequest{
		LedgerID: s.pacioliLedgerID,
		Id:       ret.LedgerAccountID,
	})
	if err != nil {
		return nil, &ErrInternalError{Err: err.Error()}
	}
	if ledgerAccount.Id != ret.LedgerAccountID {
		// Panic-ing here as something is wrong - possibly encoding of uuids to byte slice.
		panic("Ledger account ID does not match that returned from Pacioli.")
	}

	ret.CreditsAccepted = ledgerAccount.CreditsAccepted
	ret.CreditsReserved = ledgerAccount.CreditsReserved
	ret.DebitsAccepted = ledgerAccount.DebitsAccepted
	ret.DebitsReserved = ledgerAccount.DebitsReserved

	return &ret, nil
}

// This returns a u16 to match the TigerBeetle type.
// This maps a 3 letter ISO 3166 code to ISO 4217 currency code.
// Note countries without currency codes under ISO 4217 https://en.wikipedia.org/wiki/ISO_4217#Currencies_without_ISO_4217_currency_codes.
func currencyCode(country string) (uint16, error) {
	switch country {
	case "USA":
		return 840, nil
	default:
		return 0, errors.New("Unknown or unsupported country: " + country)
	}
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
