package ops

import (
	"context"
	"errors"
	"strings"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"go.temporal.io/sdk/temporal"
)

type PullFromCardArgs struct {
	TransactionID       string
	CardLinkedAccountID string
	Amount              currency.Amount
	ThreeDSID           string
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

	session3DS, err := a.b.Tabapay().Get3DSSession(ctx, args.ThreeDSID)
	if errors.Is(err, tabapay.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("3DS session not found.", "ErrNotFound", err)
	}
	if err != nil {
		return "", err
	}

	// Recommendations from Tabapay https://developers.tabapay.com/reference/3ds-eci-values
	if !strings.Contains(tabapay.ThreeDSFullyAuthenticated, session3DS.ECI) {
		return "", temporal.NewNonRetryableApplicationError("3DS not fully authenticated.", "ErrInternal", err)
	}

	externalTransactionID, err := a.b.Tabapay().PullFromCard(ctx, tabapay.PullFromCardArgs{
		WalletID:    linkedCard.WalletID,
		ProviderID:  linkedCard.ProviderID,
		ReferenceID: args.TransactionID,
		Amount:      args.Amount,
		ThreeDSID:   args.ThreeDSID,
	})
	if err != nil {
		return "", err
	}

	return externalTransactionID, nil
}

type PushToCard = PullFromCardArgs

func (a *Activity) PushToCard(ctx context.Context, args PushToCard) (string, error) {
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

	externalTransactionID, err := a.b.Tabapay().PushToCard(ctx, tabapay.PushToCardArgs{
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
