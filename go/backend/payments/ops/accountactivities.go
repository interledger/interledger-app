package ops

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
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
	if linkedCard.SendCurrency != currency.ParseCurrency(dbp.SenderCurrency) {
		return nil, temporal.NewNonRetryableApplicationError("Sender currency does not match that of sending linked account.", "ErrInternal", payments.ErrInternal)
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
		WalletID:       linkedCard.WalletID,
		ProviderID:     linkedCard.ProviderID,
		ReferenceID:    externalID,
		Amount:         currency.FromUInt64(dbp.SenderAmount, currency.ParseCurrency(dbp.SenderCurrency)),
		ThreeDSID:      dbp.ThreeDSID.String,
		SoftDescriptor: fmt.Sprintf("fynbos.me*%s", dbp.PublicID),
	})
	if err != nil {
		return nil, err
	}
	if linkedCard.DeletedAt.Valid {
		sendAmount := currency.FromUInt64(dbp.SenderAmount, currency.ParseCurrency(dbp.SenderCurrency))
		slack.SendToChannel(ctx, slack.ChannelNotifyReview, "Fynbot", fmt.Sprintf("Pulled from a deleted card. paymentID: %s\n tabapay transactionID: %s\n amount: %s\n linkedaccountID: %s\n Transaction link: %s", paymentID, externalTransaction.ID, sendAmount.Format(), linkedCard.ID, env.AdminURL()+"/wallet/"+linkedCard.WalletID+"/transactions/"+dbp.SendTransactionID.String))
	}

	return externalTransaction, nil
}

func (a *Activity) PushToAccount(ctx context.Context, paymentID, externalRef string) (*tabapay.Transaction, error) {
	p, err := getPayment(ctx, a.b, paymentID)
	if err != nil {
		return nil, err
	}

	// Check if the receiving account is configured else lookup default
	accountID := p.ReceiverAccount.String
	if accountID == "" {
		w, err := lookupWallet(ctx, a.b, payments.Identity{Identifier: p.ReceiverID, Type: p.ReceiverIDType})
		if err != nil {
			if errors.Is(err, identities.ErrNotFound) {
				return nil, temporal.NewNonRetryableApplicationError("Linked identity not found.", "ErrNotFound", err)
			}
			return nil, err
		}

		account, err := defaultReceiveAccount(ctx, a.b, w, currency.ParseCurrency(p.ReceiverCurrency))
		if err != nil {
			if errors.Is(err, linkedaccounts.ErrNotFound) {
				return nil, temporal.NewNonRetryableApplicationError("Default linked card not found.", "ErrNotFound", err)
			}
			return nil, err
		}
		accountID = account.ID

		// Set the linked account ID on the payment
		_, err = update(ctx, a.b, payments.UpdateArgs{ID: paymentID, ReceiverAccount: accountID}, p)
		if errors.Is(err, payments.ErrIncompatibleAccounts) {
			return nil, temporal.NewNonRetryableApplicationError("default receive account incompatible", "ErrIncompatible", err)
		}
		if err != nil {
			return nil, err
		}
	}
	// fetch the linked account
	linkedCard, err := a.b.LinkedAccounts().Get(ctx, accountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("Linked card not found.", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}
	if linkedCard.ReceiveCurrency != currency.ParseCurrency(p.ReceiverCurrency) {
		return nil, temporal.NewNonRetryableApplicationError("Receiver currency does not match that of receiving linked account.", "ErrInternal", payments.ErrInternal)
	}

	if linkedCard.Provider != tabapay.ProviderName {
		return nil, temporal.NewNonRetryableApplicationError("Linked account is not a card.", "ErrInternal", err)
	}

	externalTransaction, err := a.b.Tabapay().PushToCard(ctx, tabapay.PushToCardArgs{
		WalletID:       linkedCard.WalletID,
		ProviderID:     linkedCard.ProviderID,
		ReferenceID:    externalRef,
		Amount:         currency.FromUInt64(p.ReceiverAmount, currency.ParseCurrency(p.ReceiverCurrency)),
		SoftDescriptor: fmt.Sprintf("fynbos.me*%s", p.PublicID),
	})
	if err != nil {
		return nil, err
	}

	if linkedCard.DeletedAt.Valid {
		receiveAmount := currency.FromUInt64(p.SenderAmount, currency.ParseCurrency(p.SenderCurrency))
		slack.SendToChannel(ctx, slack.ChannelNotifyReview, "Fynbot", fmt.Sprintf("Pushed to a deleted card. paymentID: %s\n tabapay transactionID: %s\n amount: %s\n linkedAccountID: %s\n  Transaction link:%s", paymentID, externalTransaction.ID, receiveAmount.Format(), linkedCard.ID, env.AdminURL()+"/wallet/"+linkedCard.WalletID+"/transactions/"+p.ReceiveTransactionID.String))
	}

	return externalTransaction, nil
}

func (a *Activity) GetSenderCardTransaction(ctx context.Context, txID, paymentID string) (*tabapay.Transaction, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return nil, err
	}

	externalTransaction, err := a.b.Tabapay().GetTransaction(ctx, txID, p.SenderAmount.Currency)
	if err != nil {
		return nil, err
	}

	return externalTransaction, nil
}

func (a *Activity) GetReceiverCardTransaction(ctx context.Context, txID, paymentID string) (*tabapay.Transaction, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return nil, err
	}

	externalTransaction, err := a.b.Tabapay().GetTransaction(ctx, txID, p.ReceiverAmount.Currency)
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

	xfers, err := a.b.Transactions().ListTransfers(ctx, p.SendTransactionID)
	if err != nil {
		return err
	}

	var externalTX string
	for _, xfer := range xfers {
		if xfer.Type == transactions.TransferTypeDebitCard {
			externalTX = xfer.ForeignID
		}
	}

	if externalTX == "" {
		return nil
	}
	// TODO: Get if the transaction was actually settled from the reports.
	return a.b.Tabapay().ReverseTransaction(ctx, externalTX, false, p.SenderAmount.Currency)
}

func (a *Activity) WithdrawFromXagoBalance(ctx context.Context, paymentID string) (string, error) {
	p, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return "", err
	}

	if p.Type != payments.TypeWithdrawal {
		return uuid.NewString(), nil
	}

	tx, err := a.b.Xago().CreateTransaction(ctx, xago.CreateTransactionArgs{
		WalletID:        p.Sender.WalletID,
		LinkedAccountID: p.ReceiverAccount,
		TransactionID:   p.SendTransactionID,
		Amount:          p.ReceiverAmount,
	})
	if err != nil {
		return "", err
	}

	return tx.ID, nil

}
