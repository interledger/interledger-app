package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/pacioli"
)

// This will create the ledger account in Pacioli first and then inserts an account entry into
// CRDB. It therefore suffers from the dual-write problem so we could end up with dangling accounts
// in Pacioli.
// We create the Pacioli account first so that the account in CRDB will always have a Pacioli ID.
func Create(ctx context.Context, b Backends, pacioliLedgerID uint32, args *accounts.CreateAccountArgs) (*accounts.Account, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", accounts.ErrInvalidArgument, err.Error())
	}

	identity, err := b.Identity().Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", accounts.ErrInternal, err.Error())
	}

	ledgerAccountID := uuid.NewString()
	eventErrors, err := b.Pacioli().ConfigureAccounts(ctx, []pacioli.ConfigureAccountArgs{
		{
			ID:       ledgerAccountID,
			LedgerID: pacioliLedgerID,
			Code:     1, // TODO: Get code from somewhere.
			Flags: pacioli.AccountFlags{
				Linked:                     args.Linked,
				DebitsMustNotExceedCredits: args.DebitMustNotExceedCredits,
				CreditsMustNotExceedDebits: args.CreditsMustNotExceedDebits,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", accounts.ErrInternal, err)
	}

	if len(eventErrors) > 0 {
		return nil, fmt.Errorf("Failed to create account in pacioli. %w %+v", accounts.ErrInternal, eventErrors)
	}

	var ret accounts.Account
	err = b.DB().GetContext(
		ctx,
		&ret,
		`
		INSERT INTO accounts (identity_id, ledger_account_id, provider, provider_id)
		VALUES ($1, $2, $3, $4) RETURNING *;
		`,
		identity.ID,
		ledgerAccountID,
		args.Provider,
		args.ProviderID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", accounts.ErrInternal, err)
	}

	err = fetchFromPacioli(ctx, b, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func GetByIdentityIDWithTrx(ctx context.Context, b Backends, tx *sqlx.Tx, identityID string) (*accounts.Account, error) {
	if identityID == "" {
		return nil, fmt.Errorf("%w IdentityID is required.", accounts.ErrInvalidArgument)
	}

	var ret accounts.Account
	err := tx.Get(&ret, "SELECT * FROM accounts WHERE identity_id=$1 LIMIT 1", identityID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, accounts.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", accounts.ErrInternal, err.Error())
	}

	err = fetchFromPacioli(ctx, b, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func GetByIdentityID(ctx context.Context, b Backends, identityID string) (*accounts.Account, error) {
	var acc *accounts.Account
	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		_acc, err := GetByIdentityIDWithTrx(ctx, b, tx, identityID)
		if err != nil {
			return err
		}
		acc = _acc
		return nil
	})

	return acc, err
}

func Get(ctx context.Context, b Backends, accountID string) (*accounts.Account, error) {
	if accountID == "" {
		return nil, fmt.Errorf("%w AccountID is required.", accounts.ErrInvalidArgument)
	}

	var ret accounts.Account
	err := b.DB().Get(&ret, "SELECT * FROM accounts WHERE id=$1 LIMIT 1", accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, accounts.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", accounts.ErrInternal, err.Error())
	}

	err = fetchFromPacioli(ctx, b, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

// Fetches ledger account from Pacioli and merges with the specified account.
func fetchFromPacioli(ctx context.Context, b Backends, account *accounts.Account) error {
	if account == nil {
		return fmt.Errorf("%w Account is required.", accounts.ErrInvalidArgument)
	}
	accs, err := b.Pacioli().GetAccounts(ctx, []string{account.LedgerAccountID})
	if err != nil {
		return fmt.Errorf("%w %s", accounts.ErrInternal, err.Error())
	}
	if len(accs) != 1 {
		return fmt.Errorf("%w %s", accounts.ErrInternal, "Account not found in pacioli.")
	}

	ledgerAccount := accs[0]
	if ledgerAccount.ID != account.LedgerAccountID {
		// Panic-ing here as something is wrong - possibly encoding of uuids to byte slice.
		panic("Accounts ops: Ledger account ID does not match that returned from Pacioli.")
	}

	account.CreditsAccepted = ledgerAccount.CreditsPosted
	account.CreditsReserved = ledgerAccount.CreditsPending
	account.DebitsAccepted = ledgerAccount.DebitsPosted
	account.DebitsReserved = ledgerAccount.DebitsPending
	account.CreditsMustNotExceedDebits = ledgerAccount.Flags.CreditsMustNotExceedDebits
	account.DebitsMustNotExceedCredits = ledgerAccount.Flags.DebitsMustNotExceedCredits

	// Calculate available balance
	account.AvailableBalance = int64(account.DebitsAccepted - account.CreditsAccepted - account.CreditsReserved)

	return nil
}

func CanMakeOutgoingPayment(acc *accounts.Account, identityID string) bool {
	if acc == nil {
		return false
	}

	return acc.IdentityID == identityID
}

func CanMakeDeposit(acc *accounts.Account, identityID string) bool {
	if acc == nil {
		return false
	}

	return acc.IdentityID == identityID
}

func CanCreateFundingSource(acc *accounts.Account, identityID string) bool {
	if acc == nil {
		return false
	}

	return acc.IdentityID == identityID
}

func CanVerifyFundingSource(acc *accounts.Account, identityID string) bool {
	if acc == nil {
		return false
	}

	return acc.IdentityID == identityID
}
