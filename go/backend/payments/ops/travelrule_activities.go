package ops

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
)

// AddXagoTravelRuleRecord records a GateHub → Xago payment so it can later be
// included in the Xago travel rule report. It is a no-op for any other
// provider combination.
func (a *Activity) AddXagoTravelRuleRecord(ctx context.Context, paymentID string) error {
	payment, err := Lookup(ctx, a.b, paymentID)
	if err != nil {
		return err
	}

	if payment.SenderAccount == "" || payment.ReceiverAccount == "" {
		return nil
	}

	senderAccount, err := a.b.LinkedAccounts().Get(ctx, payment.SenderAccount)
	if err != nil {
		return err
	}
	if senderAccount.Provider != gatehub.ProviderName {
		return nil
	}

	receiverAccount, err := a.b.LinkedAccounts().Get(ctx, payment.ReceiverAccount)
	if err != nil {
		return err
	}
	if receiverAccount.Provider != xago.ProviderName {
		return nil
	}

	err = a.b.Xago().InsertTravelRuleRecord(ctx, xago.TravelRuleRecordArgs{
		PaymentID:        payment.ID,
		SenderWalletID:   payment.Sender.WalletID,
		ReceiverWalletID: payment.Receiver.WalletID,
	})
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return nil
	}
	return err
}
