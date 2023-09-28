package ops

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"gitlab.com/fynbos/backend/wallets"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

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
	IDs            []string
	WalletID       string
	ReceivedAmount uint64
	Asset          string
}

type dbPayout struct {
	ID             string `db:"id"`
	WalletID       string `db:"wallet_id"`
	ReceivedAmount uint64 `db:"received_amount"`
	Asset          string `db:"received_amount_asset"`
}

func (a *Activity) ListPayouts(ctx context.Context) ([]Payout, error) {
	var payouts []dbPayout
	err := a.b.DB().SelectContext(ctx, &payouts, `SELECT ip.id, ip.received_amount, ip.received_amount_asset, pp.wallet_id FROM rafiki_incoming_payments ip 
		INNER JOIN rafiki_payment_pointers pp ON pp.payment_pointer_id=ip.payment_pointer_id 
		WHERE ip.completed=true AND payment_id is null`)
	if err != nil {
		return nil, err
	}

	// Group by walletID
	walletPayouts := make(map[string]Payout)
	for _, p := range payouts {
		key := fmt.Sprintf("%s|%s", p.WalletID, p.Asset)
		payout, ok := walletPayouts[key]
		if ok {
			payout.ReceivedAmount += p.ReceivedAmount
			payout.IDs = append(payout.IDs, p.ID)
			walletPayouts[key] = payout
			continue
		}
		walletPayouts[key] = Payout{
			IDs:            []string{p.ID},
			WalletID:       p.WalletID,
			ReceivedAmount: p.ReceivedAmount,
			Asset:          p.Asset,
		}
	}

	// Flatten into array
	var resp []Payout
	for _, v := range walletPayouts {
		resp = append(resp, v)
	}

	return resp, nil
}

func (a *Activity) CreatePayoutPayment(ctx context.Context, payout Payout) (string, error) {

	senderAcc, err := a.b.LinkedAccounts().GetDefaultReceive(ctx, wallets.WebMonetizationWalletID)
	if err != nil {
		return "", err
	}

	p, err := a.b.Payments().Create(ctx, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: wallets.WebMonetizationWalletID,
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: payout.WalletID,
		},
		SenderAccount:  senderAcc.ID,
		Type:           payments.TypeWebMonetization,
		SenderAmount:   currency.FromUInt64(payout.ReceivedAmount, currency.ParseCurrency(payout.Asset)),
		ReceiverAmount: currency.FromUInt64(payout.ReceivedAmount, currency.ParseCurrency(payout.Asset)),
		Note:           "Web Monetization payout",
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

func (a *Activity) AddPaymentRef(ctx context.Context, payout Payout, paymentID string) error {
	query, args, err := sqlx.In("UPDATE rafiki_incoming_payments SET payment_id=? WHERE id IN (?)", paymentID, payout.IDs)
	if err != nil {
		return err
	}
	_, err = a.b.DB().ExecContext(ctx, a.b.DB().Rebind(query), args...)
	return err
}
