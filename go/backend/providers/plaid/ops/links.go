package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// PlaidLink is the Plaid-specific data for a linked account, kept 1:1 with a
// linked_accounts row so the provider-agnostic linked_accounts table carries no
// Plaid columns. It exists only to dedupe Plaid links per wallet.
type PlaidLink struct {
	ID              string       `db:"id"`
	LinkedAccountID string       `db:"linked_account_id"`
	WalletID        string       `db:"wallet_id"`
	PlaidAccountID  string       `db:"plaid_account_id"`
	CreatedAt       time.Time    `db:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at"`
	DeletedAt       sql.NullTime `db:"deleted_at"`
}

type CreateLinkArgs struct {
	ID              string `validate:"omitempty,uuid4"`
	LinkedAccountID string `validate:"required,uuid4"`
	WalletID        string `validate:"required,uuid4"`
	PlaidAccountID  string `validate:"required"`
}

var (
	ErrLinkNotFound = errors.New("plaid link: not found.")
	ErrLinkInvalid  = errors.New("plaid link: invalid argument.")
	ErrLinkInternal = errors.New("plaid link: internal error.")
)

// LinkBackends is the minimal, duck-typed dependency set the plaid_links queries
// need. The main *backends struct and the grpc Backends both satisfy it.
type LinkBackends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error
}

const linkFields = "id, linked_account_id, wallet_id, plaid_account_id, created_at, updated_at, deleted_at"

// CreateLinkTx inserts a plaid_links row inside the given transaction, with no
// side effects. Use it to create the link atomically with the linked_accounts row
// it points at, sharing the caller's *sqlx.Tx.
func CreateLinkTx(ctx context.Context, tx *sqlx.Tx, v *validator.Validate, args *CreateLinkArgs) (*PlaidLink, error) {
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", ErrLinkInvalid, err.Error())
	}

	id := args.ID
	if id == "" {
		id = uuid.NewString()
	}

	var link PlaidLink
	err := tx.GetContext(
		ctx,
		&link,
		fmt.Sprintf("INSERT INTO plaid_links (id, linked_account_id, wallet_id, plaid_account_id) VALUES ($1, $2, $3, $4) RETURNING %s;", linkFields),
		id,
		args.LinkedAccountID,
		args.WalletID,
		args.PlaidAccountID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrLinkInternal, err.Error())
	}

	return &link, nil
}

// GetLinkByPlaidAccountID returns the live plaid_links row a wallet provisioned
// from the given Plaid account_id, used to short-circuit duplicate Fiant
// registrations. Returns ErrLinkNotFound when the wallet has not linked it (live).
func GetLinkByPlaidAccountID(ctx context.Context, b LinkBackends, walletID, plaidAccountID string) (*PlaidLink, error) {
	if walletID == "" || plaidAccountID == "" {
		return nil, fmt.Errorf("%w wallet_id and plaid_account_id are required", ErrLinkInvalid)
	}

	var link PlaidLink
	err := b.DB().GetContext(
		ctx,
		&link,
		fmt.Sprintf("SELECT %s FROM plaid_links WHERE wallet_id=$1 AND plaid_account_id=$2 AND deleted_at IS NULL;", linkFields),
		walletID,
		plaidAccountID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, fmt.Errorf("%w %s", ErrLinkInternal, err.Error())
	}

	return &link, nil
}

// ListPlaidAccountIDsByWallet returns the Plaid account_ids a wallet has live
// links for (the "already registered" set).
func ListPlaidAccountIDsByWallet(ctx context.Context, b LinkBackends, walletID string) ([]string, error) {
	if walletID == "" {
		return nil, fmt.Errorf("%w wallet_id is required", ErrLinkInvalid)
	}

	var ids []string
	err := b.DB().SelectContext(
		ctx,
		&ids,
		"SELECT plaid_account_id FROM plaid_links WHERE wallet_id=$1 AND deleted_at IS NULL;",
		walletID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrLinkInternal, err.Error())
	}

	return ids, nil
}

// SoftDeleteLinkByLinkedAccountID marks every live link for a linked account
// deleted, freeing the (wallet_id, plaid_account_id) slot for re-linking.
func SoftDeleteLinkByLinkedAccountID(ctx context.Context, b LinkBackends, linkedAccountID string) error {
	if linkedAccountID == "" {
		return fmt.Errorf("%w linked_account_id is required", ErrLinkInvalid)
	}

	_, err := b.DB().ExecContext(
		ctx,
		"UPDATE plaid_links SET deleted_at=now(), updated_at=now() WHERE linked_account_id=$1 AND deleted_at IS NULL;",
		linkedAccountID,
	)
	if err != nil {
		return fmt.Errorf("%w %s", ErrLinkInternal, err.Error())
	}

	return nil
}
