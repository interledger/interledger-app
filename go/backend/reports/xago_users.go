package reports

import (
	"context"
	"fmt"

	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/lib/pq"
)

var xagoUsersReport = providerReport{
	name:                  "xago",
	linkedAccountProvider: xago.ProviderName,
	linkedAccountType:     xago.AccTypeBalance,
	selectUsers:           selectXagoUsers,
	balance: func(ctx context.Context, b Backends, linkedAccountID string) (currency.Amount, currency.Amount, error) {
		bal, err := b.Xago().GetBalance(ctx, linkedAccountID)
		if err != nil {
			return currency.Amount{}, currency.Amount{}, err
		}
		return bal.Total, bal.Available, nil
	},
}

// XagoUsersCSV returns a CSV of every user that is ZA in Kratos (the only country that
// gets a Xago balance account) or owns a ZA wallet, with all of their wallets.
func XagoUsersCSV(ctx context.Context, b Backends) (string, error) {
	return xagoUsersReport.build(ctx, b)
}

// selectXagoUsers picks users that are ZA in Kratos or own a ZA wallet, with all their
// wallets. ZA users in Kratos are reported even without a wallet.
func selectXagoUsers(ctx context.Context, b Backends, identities []user.User) ([]string, []walletRow, error) {
	// A ZA wallet owner may have another country in Kratos, so both are checked.
	var kratosZAUserIDs []string
	for _, id := range identities {
		if id.Country == country.ZA {
			kratosZAUserIDs = append(kratosZAUserIDs, id.ID)
		}
	}

	var rows []walletRow
	err := b.DB().SelectContext(ctx, &rows, `
		SELECT uw.user_id, w.id AS wallet_id, w.country
		FROM user_wallets uw
		JOIN wallets w ON w.id = uw.wallet_id
		WHERE uw.user_id IN (SELECT user_id FROM user_wallets uw2 JOIN wallets w2 ON w2.id = uw2.wallet_id WHERE w2.country = $1)
		   OR uw.user_id = ANY($2)
		ORDER BY uw.user_id, w.created_at`,
		country.ZA, pq.Array(kratosZAUserIDs))
	if err != nil {
		return nil, nil, fmt.Errorf("listing ZA wallets: %w", err)
	}
	return kratosZAUserIDs, rows, nil
}
