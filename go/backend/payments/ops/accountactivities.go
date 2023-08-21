package ops

import (
	"context"
	"errors"
	"strings"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
)

// PullFromAccount pulls from the account to the GMT account.
// TODO: For now we expect this to be a Tabapay card account. In the future we need to infer it from the linked account.
func (a *Activity) PullFromAccount(ctx context.Context, paymentID, externalID string) (*tabapay.Transaction, error) {

	dbp, err := getPayment(ctx, a.b, paymentID)
	if err != nil {
		return nil, err
	}

	// fetch the linked account
	linkedCard, err := a.b.LinkedAccounts().Get(ctx, dbp.SenderAccount.String)
	if err != nil {
		return nil, err
	}

	if linkedCard.Provider != tabapay.ProviderName {
		return nil, temporal.NewNonRetryableApplicationError("Linked account is not a card.", "ErrInternal", err)
	}

	if dbp.ThreeDSID.Valid {
		session3DS, err := a.b.Tabapay().Get3DSSession(ctx, dbp.ThreeDSID.String)
		if errors.Is(err, tabapay.ErrNotFound) {
			return nil, temporal.NewNonRetryableApplicationError("3DS session not found.", "ErrNotFound", err)
		}
		if err != nil {
			return nil, err
		}

		// Recommendations from Tabapay https://developers.tabapay.com/reference/3ds-eci-values
		if !strings.Contains(tabapay.ThreeDSFullyAuthenticated, session3DS.ECI) {
			log.Info("3DS session not fully authenticated", zap.String("eci", session3DS.ECI), zap.String("threeDSID", dbp.ThreeDSID.String), zap.String("linkedAccountID", linkedCard.ID))
		}
	}

	externalTransaction, err := a.b.Tabapay().PullFromCard(ctx, tabapay.PullFromCardArgs{
		WalletID:    linkedCard.WalletID,
		ProviderID:  linkedCard.ProviderID,
		ReferenceID: externalID,
		Amount:      currency.FromUInt64(dbp.SenderAmount, currency.ParseCurrency(dbp.SenderCurrency)),
		ThreeDSID:   dbp.ThreeDSID.String,
	})
	if err != nil {
		return nil, err
	}

	return externalTransaction, nil
}

func (a *Activity) PushToAccount(ctx context.Context, paymentID, externalRef string) (*tabapay.Transaction, error) {

	dbp, err := getPayment(ctx, a.b, paymentID)
	if err != nil {
		return nil, err
	}

	// fetch the linked account
	linkedCard, err := a.b.LinkedAccounts().Get(ctx, dbp.ReceiverAccount.String)
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
		ReferenceID: externalRef,
		Amount:      currency.FromUInt64(dbp.ReceiverAmount, currency.ParseCurrency(dbp.SenderCurrency)),
	})
	if err != nil {
		return nil, err
	}

	return externalTransaction, nil
}

func (a *Activity) GetCardTransaction(ctx context.Context, id string) (*tabapay.Transaction, error) {
	externalTransaction, err := a.b.Tabapay().GetTransaction(ctx, id)
	if err != nil {
		return nil, err
	}

	return externalTransaction, nil
}

func (a *Activity) RollbackPullFromAccount(ctx context.Context, paymentID string) error {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	senderWallet, err := lookupWallet(ctx, a.b, p.Sender)
	if err != nil {
		return err
	}

	tx, err := a.b.Transactions().GetTransaction(ctx, senderWallet.ID, p.SendTransactionID)
	if err != nil {
		return err
	}

	var externalTX string
	for _, tr := range tx.Transfers {
		if tr.Type == transactions.TransferTypeDebitCard {
			externalTX = tr.ForeignID
		}
	}

	if externalTX == "" {
		return nil
	}
	// TODO: Get if the transaction was actually settled from the reports.
	return a.b.Tabapay().ReverseTransaction(ctx, externalTX, false)
}
