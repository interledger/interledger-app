package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lestrrat-go/jwx/v3/jwk"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_ops "gitlab.com/fynbos/backend/linkedaccounts/ops"
	"gitlab.com/fynbos/backend/providers/pti"
	pti_external "gitlab.com/fynbos/backend/providers/pti/external"
	pti_ops "gitlab.com/fynbos/backend/providers/pti/ops"
	plaid_ops "gitlab.com/fynbos/backend/providers/plaid/ops"
)

// plaidFiantLinker is the adapter that the Plaid HTTP layer calls into for
// Phase 2's `/api/plaid/link-to-fiant` route. It hides the cross-package
// orchestration (wallet + PTI user resolution, Fiant external call, linked-
// account persistence) so that `providers/plaid/ops` does not need to import
// pti / linkedaccounts / wallets directly.
//
// Lifetime: one per backend boot. The PTI external client inside is a long-
// lived HTTP client identical in shape to the one wired into the temporal
// Activity — we build a fresh instance here rather than reach into the worker
// to avoid coupling the HTTP path to temporal startup ordering.
type plaidFiantLinker struct {
	b        *backends
	external pti_external.Client
}

// newPlaidFiantLinker constructs the adapter using the same PTI key + URLs as
// the temporal Activity. Returns nil when PTI is disabled (caller treats nil
// as "do not register the /link-to-fiant route").
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

// ExistingLink — see plaid_ops.FiantLinker. Returns nil on absence.
func (l *plaidFiantLinker) ExistingLink(ctx context.Context, userID, plaidAccountID string) (*plaid_ops.LinkedIDs, error) {
	walletID, err := l.walletForUser(ctx, userID)
	if err != nil {
		// User without a wallet can never have a linked Plaid account; treat as "no dupe".
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	la, err := linkedaccounts_ops.GetByPlaidAccountID(ctx, l.b, walletID, plaidAccountID)
	if err != nil {
		if errors.Is(err, linkedaccounts.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plaid_ops.LinkedIDs{
		LinkedAccountID:      la.ID,
		PaymentInformationID: la.ProviderID,
	}, nil
}

// Register — see plaid_ops.FiantLinker. Posts to Fiant then persists.
func (l *plaidFiantLinker) Register(ctx context.Context, args plaid_ops.LinkPlaidArgs) (*plaid_ops.LinkedIDs, error) {
	walletID, err := l.walletForUser(ctx, args.UserID)
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
	la, err := linkedaccounts_ops.Create(ctx, l.b, &linkedaccounts.CreateArgs{
		WalletID:            walletID,
		Name:                name,
		Nickname:            name,
		Mask:                mask,
		Provider:            pti.ProviderName,
		ProviderID:          bank.ID,
		Type:                pti.TypeBank,
		PlaidAccountID:      args.PlaidAccountID,
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
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: persist linked_account: %w", err)
	}

	return &plaid_ops.LinkedIDs{
		LinkedAccountID:      la.ID,
		PaymentInformationID: bank.ID,
	}, nil
}

// ListLinkedPlaidAccountIDs — see plaid_ops.FiantLinker.
func (l *plaidFiantLinker) ListLinkedPlaidAccountIDs(ctx context.Context, userID string) ([]string, error) {
	walletID, err := l.walletForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	err = l.b.DB().SelectContext(
		ctx,
		&ids,
		`SELECT plaid_account_id FROM linked_accounts
		 WHERE wallet_id = $1
		   AND plaid_account_id IS NOT NULL
		   AND deleted_at IS NULL;`,
		walletID,
	)
	if err != nil {
		return nil, fmt.Errorf("plaid/fiant linker: list plaid account ids: %w", err)
	}
	return ids, nil
}

// walletForUser returns the first wallet attached to a Kratos user. Mirrors
// the convention in wallets/middleware which also "picks a default" when the
// user has more than one.
func (l *plaidFiantLinker) walletForUser(ctx context.Context, userID string) (string, error) {
	var id string
	err := l.b.DB().GetContext(ctx, &id, "SELECT wallet_id FROM user_wallets WHERE user_id = $1 LIMIT 1;", userID)
	return id, err
}
