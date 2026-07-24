package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lestrrat-go/jwx/v3/jwk"

	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	linkedaccounts_ops "github.com/interledger/interledger-app/go/backend/linkedaccounts/ops"
	"github.com/interledger/interledger-app/go/backend/notify"
	plaid_ops "github.com/interledger/interledger-app/go/backend/providers/plaid/ops"
	"github.com/interledger/interledger-app/go/backend/providers/pti"
	"github.com/interledger/interledger-app/go/backend/providers/pti/external"
)

// FiantLinkerBackends is scoped to exactly what FiantLinker (and the
// linkedaccounts/plaid ops it calls into) need — kept separate from Backends
// so extending it doesn't ripple into every other caller of that shared
// interface (webhooks, temporal activities, etc).
type FiantLinkerBackends interface {
	Backends
	WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error
	Validator() *validator.Validate
	Notify() notify.Client
}

type FiantLinker struct {
	b        FiantLinkerBackends
	external external.Client
}

func NewFiantLinker(b FiantLinkerBackends, ptiBaseURL, ptiClientID, ptiJWK string) (*FiantLinker, error) {
	if ptiJWK == "" {
		return nil, fmt.Errorf("plaid/fiant linker: PTI JWK is required when PTI is enabled")
	}
	pk, err := jwk.ParseKey([]byte(ptiJWK))
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: parse PTI JWK: %w", err)
	}
	ext, err := external.NewWithOptions(
		external.WithBaseURL(ptiBaseURL),
		external.WithOTELLHTTPClient(),
		external.WithClientID(ptiClientID),
		external.WithDerivedKeys(pk),
	)
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: build PTI external client: %w", err)
	}
	return &FiantLinker{b: b, external: ext}, nil
}

// WithAccountLock runs fn while holding a transaction-scoped Postgres advisory
// lock keyed on (userID, plaidAccountID). This serializes the dedupe-check →
// processor-token mint → Fiant register sequence so two concurrent requests for
// the same Plaid account can't both call Fiant and create duplicate
// payment-information records. The partial unique index
// `plaid_links_wallet_plaid_uniq` is the final backstop; this lock prevents
// the wasted/orphaning external calls before that index ever trips.
func (l *FiantLinker) WithAccountLock(ctx context.Context, userID, plaidAccountID string, fn func(context.Context) error) error {
	tx, err := l.b.DB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("plaid/fiant linker: begin lock tx: %w", err)
	}
	// The advisory lock is released automatically when the tx ends (commit or
	// rollback). fn's own DB writes run on the pool and autocommit; the lock
	// only acts as a per-key mutex across the critical section.
	key := userID + ":" + plaidAccountID
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, key); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("plaid/fiant linker: acquire advisory lock: %w", err)
	}

	if fnErr := fn(ctx); fnErr != nil {
		_ = tx.Rollback()
		return fnErr
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("plaid/fiant linker: release advisory lock: %w", err)
	}
	return nil
}

func (l *FiantLinker) ExistingLink(ctx context.Context, userID, plaidAccountID string) (*plaid_ops.LinkedIDs, error) {
	walletID, err := l.getWalletIdByUserId(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	link, err := plaid_ops.GetLinkByPlaidAccountID(ctx, l.b, walletID, plaidAccountID)
	if err != nil {
		if errors.Is(err, plaid_ops.ErrLinkNotFound) {
			return nil, nil
		}
		return nil, err
	}

	la, err := linkedaccounts_ops.Get(ctx, l.b, link.LinkedAccountID)
	if err != nil {
		return nil, err
	}
	return &plaid_ops.LinkedIDs{
		LinkedAccountID:      la.ID,
		PaymentInformationID: la.ProviderID,
	}, nil
}

func (l *FiantLinker) Register(ctx context.Context, args plaid_ops.LinkPlaidArgs) (*plaid_ops.LinkedIDs, error) {
	walletID, err := l.getWalletIdByUserId(ctx, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: resolve wallet: %w", err)
	}

	ptiUser, err := GetUser(ctx, l.b, walletID)
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: resolve pti user: %w", err)
	}
	if ptiUser == nil || ptiUser.ExternalID == "" {
		return nil, fmt.Errorf("plaid/fiant linker: pti user missing external id")
	}

	bank, err := l.external.CreateBankAccountFromPlaid(ctx, ptiUser.ExternalID, args.ProcessorToken)
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: fiant create: %w", err)
	}
	if bank == nil || bank.ID == "" {
		return nil, fmt.Errorf("plaid/fiant linker: fiant returned empty payment-information id")
	}

	name := strings.TrimSpace(args.AccountName)
	if name == "" {
		name = "Plaid bank"
	}
	mask := args.AccountMask

	var la *linkedaccounts.LinkedAccount
	err = l.b.WithTx(ctx, func(tx *sqlx.Tx) error {
		created, e := linkedaccounts_ops.CreateTx(ctx, tx, l.b.Validator(), &linkedaccounts.CreateArgs{
			WalletID:            walletID,
			Name:                name,
			Nickname:            name,
			Mask:                mask,
			Provider:            pti.ProviderName,
			ProviderID:          bank.ID,
			Type:                pti.TypeBank,
			State:               linkedaccounts.Verified,
			CanSend:             true,
			CanReceive:          true,
			SendCountry:         country.US,
			SendCurrency:        currency.USD,
			SendNetwork:         "ACH",
			SendAvailability:    linkedaccounts.Few,
			ReceiveCountry:      country.US,
			ReceiveCurrency:     currency.USD,
			ReceiveNetwork:      "ACH",
			ReceiveAvailability: linkedaccounts.Few,
		})
		if e != nil {
			return fmt.Errorf("persist linked_account: %w", e)
		}

		if _, e = plaid_ops.CreateLinkTx(ctx, tx, l.b.Validator(), &plaid_ops.CreateLinkArgs{
			LinkedAccountID: created.ID,
			WalletID:        walletID,
			PlaidAccountID:  args.PlaidAccountID,
		}); e != nil {
			return fmt.Errorf("persist plaid_link: %w", e)
		}

		la = created
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: %w", err)
	}

	linkedaccounts_ops.EmitCreated(ctx, l.b, la)

	return &plaid_ops.LinkedIDs{
		LinkedAccountID:      la.ID,
		PaymentInformationID: bank.ID,
	}, nil
}

func (l *FiantLinker) ListLinkedPlaidAccountIDs(ctx context.Context, userID string) ([]string, error) {
	walletID, err := l.getWalletIdByUserId(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	ids, err := plaid_ops.ListPlaidAccountIDsByWallet(ctx, l.b, walletID)
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: list plaid account ids: %w", err)
	}
	return ids, nil
}

// IsActivated reports whether the user's wallet holds a US balance — the PTI
// balance linked account provisioned asynchronously by CreateWalletWorkflow
// after KYC. Mirrors the GetBalances predicate (grpc/balances.go) so the backend
// guard and the frontend gate agree on "activated". No wallet / no rows → false.
func (l *FiantLinker) IsActivated(ctx context.Context, userID string) (bool, error) {
	walletID, err := l.getWalletIdByUserId(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	las, err := linkedaccounts_ops.ListByWalletId(ctx, l.b, walletID)
	if err != nil {
		if errors.Is(err, linkedaccounts.ErrNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("plaid/fiant linker: list linked accounts: %w", err)
	}

	for _, la := range las {
		if la.Provider == pti.ProviderName &&
			la.Type == pti.AccTypeBalance &&
			la.ReceiveCountry == country.US &&
			la.DeletedAt.Time.IsZero() {
			return true, nil
		}
	}
	return false, nil
}

func (l *FiantLinker) getWalletIdByUserId(ctx context.Context, userID string) (string, error) {
	var id string
	err := l.b.DB().GetContext(ctx, &id, "SELECT wallet_id FROM user_wallets WHERE user_id = $1 LIMIT 1;", userID)
	return id, err
}
