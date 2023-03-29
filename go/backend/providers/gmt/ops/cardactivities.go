package ops

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"go.temporal.io/sdk/temporal"
)

type PullFromCardArgs struct {
	TransactionID       string
	CardLinkedAccountID string
	Amount              currency.Amount
}

// Pulls from the card to the GMT account.
func (a *Activity) PullFromCard(ctx context.Context, args PullFromCardArgs) (string, error) {
	// fetch the linked account
	linkedCard, err := a.b.LinkedAccounts().Get(ctx, args.CardLinkedAccountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("Linked card not found.", "ErrNotFound", err)
	}
	if err != nil {
		return "", err
	}

	if linkedCard.Provider != tabapay.ProviderName {
		return "", temporal.NewNonRetryableApplicationError("Linked account is not a card.", "ErrInternal", err)
	}

	externalTransactionID, err := a.b.Tabapay().PullFromCard(ctx, tabapay.PullFromCardArgs{
		WalletID:    linkedCard.WalletID,
		ProviderID:  linkedCard.ProviderID,
		ReferenceID: args.TransactionID,
		Amount:      args.Amount,
	})
	if err != nil {
		return "", err
	}

	return externalTransactionID, nil
}
