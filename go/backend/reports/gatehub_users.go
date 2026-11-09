package reports

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/user"
)

var gatehubUsersReport = providerReport{
	name:                  "gatehub",
	displayName:           "GateHub",
	title:                 "GateHub users report",
	linkedAccountProvider: gatehub.ProviderName,
	linkedAccountType:     gatehub.AccTypeBalance,
	selectUsers:           selectGatehubUsers,
	balance: func(ctx context.Context, b Backends, linkedAccountID string) (currency.Amount, currency.Amount, error) {
		bal, err := b.Gatehub().GetBalance(ctx, linkedAccountID)
		if err != nil {
			return currency.Amount{}, currency.Amount{}, err
		}
		return bal.Total, bal.Available, nil
	},
	externalUserID: gatehubUserID,
}

// GatehubUsers returns a report of every wallet that has a GateHub user or a GateHub
// balance account, with its owner's details. Balances come from the local ledger, so
// building it makes no calls to the GateHub API.
func GatehubUsers(ctx context.Context, b Backends, format Format, now time.Time) (*File, error) {
	return gatehubUsersReport.build(ctx, b, format, now)
}

// selectGatehubUsers picks wallets that were onboarded to GateHub. GateHub users are
// created on demand (when a user opens GateHub onboarding), not by country, so the
// population comes from the data rather than a country rule. Only GateHub wallets are
// listed, and users without one are left out.
func selectGatehubUsers(ctx context.Context, b Backends, _ []user.User) ([]string, []walletRow, error) {
	var rows []walletRow
	err := b.DB().SelectContext(ctx, &rows, `
		SELECT uw.user_id, w.id AS wallet_id, w.country
		FROM user_wallets uw
		JOIN wallets w ON w.id = uw.wallet_id
		WHERE EXISTS (SELECT 1 FROM gatehub_users g WHERE g.wallet_id = w.id)
		   OR EXISTS (SELECT 1 FROM linked_accounts la WHERE la.wallet_id = w.id AND la.provider = $1 AND la.deleted_at IS NULL)
		ORDER BY uw.user_id, w.created_at`,
		gatehub.ProviderName)
	if err != nil {
		return nil, nil, fmt.Errorf("listing GateHub wallets: %w", err)
	}
	return nil, rows, nil
}

// gatehubUserID returns the wallet's most recent GateHub user ID, or "" if it has none.
func gatehubUserID(ctx context.Context, b Backends, walletID string) (string, error) {
	var id string
	err := b.DB().GetContext(ctx, &id, `
		SELECT external_id
		FROM gatehub_users
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT 1`,
		walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return id, err
}
