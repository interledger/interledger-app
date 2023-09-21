package ops

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/payments"
)

type Activity struct {
	b ActivityBackends
}

func NewActivity(b ActivityBackends) *Activity {
	return &Activity{b}
}

type Payout struct {
	ID             string `db:"id"`
	WalletID       string `db:"wallet_id"`
	ReceivedAmount uint64 `db:"received_amount"`
	Asset          string `db:"received_amount_asset"`
}

func (a *Activity) ListPayouts(ctx context.Context) ([]Payout, error) {
	var payouts []Payout
	err := a.b.DB().SelectContext(ctx, &payouts, `SELECT ip.id, ip.received_amount, ip.received_amount_asset, pp.wallet_id FROM rafiki_incoming_payments ip 
		INNER JOIN rafiki_payment_pointers pp ON pp.payment_pointer_id=ip.payment_pointer_id 
		WHERE ip.completed=true AND payment_id is null`)
	if err != nil {
		return nil, err
	}

	return payouts, nil
}

func (a *Activity) CreatePayoutPayment(ctx context.Context, payout Payout) (string, error) {
	p, err := a.b.Payments().Create(ctx, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletURL,
			Identifier: env.OpenPaymentsURL() + "/webmonitization", // TODO: reserve these in all environments
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: payout.WalletID,
		},
		SenderAmount:   currency.FromUInt64(payout.ReceivedAmount, currency.ParseCurrency(payout.Asset)),
		ReceiverAmount: currency.FromUInt64(payout.ReceivedAmount, currency.ParseCurrency(payout.Asset)),
		Note:           "Web Monitization payout",
		IPAddress:      "198.0.0.2", // TODO: Add a our static IP Address
	})

	if err != nil {
		return "", err
	}

	return p.ID, nil
}

func (a *Activity) ConfirmPayment(ctx context.Context, id string) error {
	_, ra, err := a.b.Payments().Confirm(ctx, id)
	if errors.Is(err, payments.ErrRequiredActions) {
		log.Error("required actions outstanding for web monitization payout", zap.String("actions", fmt.Sprintf("%+v", ra)))
	}
	return err
}

func (a *Activity) AddPaymentRef(ctx context.Context, id, paymentID string) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE rafiki_incoming_payments SET payment_id=$1 WHERE id=$2", paymentID, id)
	return err
}
