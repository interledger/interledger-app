package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lestrrat-go/jwx/v3/jwk"

	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	linkedaccounts_ops "github.com/interledger/interledger-app/go/backend/linkedaccounts/ops"
	plaid_ops "github.com/interledger/interledger-app/go/backend/providers/plaid/ops"
	"github.com/interledger/interledger-app/go/backend/providers/pti"
	pti_external "github.com/interledger/interledger-app/go/backend/providers/pti/external"
	pti_ops "github.com/interledger/interledger-app/go/backend/providers/pti/ops"
)

type plaidFiantLinker struct {
	b        *backends
	external pti_external.Client
}

func newPlaidFiantLinker(b *backends, ptiBaseURL, ptiClientID, ptiJWK string) (*plaidFiantLinker, error) {
	if ptiJWK == "" {
		return nil, nil
	}
	pk, err := jwk.ParseKey([]byte(ptiJWK))
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: parse PTI JWK: %w", err)
	}
	ext, err := pti_external.NewWithOptions(
		pti_external.WithBaseURL(ptiBaseURL),
		pti_external.WithOTELLHTTPClient(),
		pti_external.WithClientID(ptiClientID),
		pti_external.WithDerivedKeys(pk),
	)
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: build PTI external client: %w", err)
	}
	return &plaidFiantLinker{b: b, external: ext}, nil
}

// WithAccountLock runs fn while holding a transaction-scoped Postgres advisory
// lock keyed on (userID, plaidAccountID). This serializes the dedupe-check →
// processor-token mint → Fiant register sequence so two concurrent requests for
// the same Plaid account can't both call Fiant and create duplicate
// payment-information records. The partial unique index
// `plaid_links_wallet_plaid_uniq` is the final backstop; this lock prevents
// the wasted/orphaning external calls before that index ever trips.
func (l *plaidFiantLinker) WithAccountLock(ctx context.Context, userID, plaidAccountID string, fn func(context.Context) error) error {
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

func (l *plaidFiantLinker) ExistingLink(ctx context.Context, userID, plaidAccountID string) (*plaid_ops.LinkedIDs, error) {
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

func (l *plaidFiantLinker) Register(ctx context.Context, args plaid_ops.LinkPlaidArgs) (*plaid_ops.LinkedIDs, error) {
	walletID, err := l.getWalletIdByUserId(ctx, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: resolve wallet: %w", err)
	}

	ptiUser, err := pti_ops.GetUser(ctx, l.b, walletID)
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

func (l *plaidFiantLinker) ListLinkedPlaidAccountIDs(ctx context.Context, userID string) ([]string, error) {
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

func (l *plaidFiantLinker) getWalletIdByUserId(ctx context.Context, userID string) (string, error) {
	var id string
	err := l.b.DB().GetContext(ctx, &id, "SELECT wallet_id FROM user_wallets WHERE user_id = $1 LIMIT 1;", userID)
	return id, err
}
