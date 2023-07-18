package ops

import (
	"context"
	"errors"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
	"strings"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"go.temporal.io/sdk/temporal"
)

type PullFromCardArgs struct {
	ReferenceID         string
	TransactionID       string
	CardLinkedAccountID string
	Amount              currency.Amount
	ThreeDSID           string
}

// Pulls from the card to the GMT account.
func (a *Activity) PullFromCard(ctx context.Context, args PullFromCardArgs) (*tabapay.Transaction, error) {
	// fetch the linked account
	linkedCard, err := a.b.LinkedAccounts().Get(ctx, args.CardLinkedAccountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("Linked card not found.", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	if linkedCard.Provider != tabapay.ProviderName {
		return nil, temporal.NewNonRetryableApplicationError("Linked account is not a card.", "ErrInternal", err)
	}

	if args.ThreeDSID != "" {
		session3DS, err := a.b.Tabapay().Get3DSSession(ctx, args.ThreeDSID)
		if errors.Is(err, tabapay.ErrNotFound) {
			return nil, temporal.NewNonRetryableApplicationError("3DS session not found.", "ErrNotFound", err)
		}
		if err != nil {
			return nil, err
		}

		// Recommendations from Tabapay https://developers.tabapay.com/reference/3ds-eci-values
		if !strings.Contains(tabapay.ThreeDSFullyAuthenticated, session3DS.ECI) {
			log.Info("3DS session not fully authenticated", zap.String("eci", session3DS.ECI), zap.String("threeDSID", args.ThreeDSID), zap.String("linkedAccountID", linkedCard.ID))
		}
	}

	externalTransaction, err := a.b.Tabapay().PullFromCard(ctx, tabapay.PullFromCardArgs{
		WalletID:    linkedCard.WalletID,
		ProviderID:  linkedCard.ProviderID,
		ReferenceID: args.ReferenceID,
		Amount:      args.Amount,
		ThreeDSID:   args.ThreeDSID,
	})
	if err != nil {
		return nil, err
	}

	return externalTransaction, nil
}

type PushToCard = PullFromCardArgs

func (a *Activity) PushToCard(ctx context.Context, args PushToCard) (*tabapay.Transaction, error) {
	// fetch the linked account
	linkedCard, err := a.b.LinkedAccounts().Get(ctx, args.CardLinkedAccountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("Linked card not found.", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	if linkedCard.Provider != tabapay.ProviderName {
		return nil, temporal.NewNonRetryableApplicationError("Linked account is not a card.", "ErrInternal", err)
	}

	externalTransaction, err := a.b.Tabapay().PushToCard(ctx, tabapay.PushToCardArgs{
		WalletID:    linkedCard.WalletID,
		ProviderID:  linkedCard.ProviderID,
		ReferenceID: args.ReferenceID,
		Amount:      args.Amount,
	})
	if err != nil {
		return nil, err
	}

	return externalTransaction, nil
}

func (a *Activity) GetTabapayTransaction(ctx context.Context, id string) (*tabapay.Transaction, error) {
	externalTransaction, err := a.b.Tabapay().GetTransaction(ctx, id)
	if err != nil {
		return nil, err
	}

	return externalTransaction, nil
}

func (a *Activity) ReverseTabapayTransaction(ctx context.Context, id string) error {
	// TODO: Get if the transaction was actually settled from the reports.
	return a.b.Tabapay().ReverseTransaction(ctx, id, false)
}
