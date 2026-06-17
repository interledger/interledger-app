package ops

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.temporal.io/sdk/temporal"

	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/rafiki"
)

type Activity struct {
	b ActivityBackends
}

func NewActivity(b ActivityBackends) *Activity {
	return &Activity{b}
}

type dbPayment struct {
	ID           string    `db:"id"`
	FromWalletID string    `db:"from_wallet"`
	ToWalletID   string    `db:"to_wallet"`
	Amount       int64     `db:"amount"`
	Asset        string    `db:"amount_asset"`
	Timestamp    time.Time `db:"created_at"`
}

type RafikiPayment struct {
	IDs          pq.StringArray `db:"ids"`
	FromWalletID string         `db:"from_wallet"`
	ToWalletID   string         `db:"to_wallet"`
	Amount       int64          `db:"amount"`
	Asset        string         `db:"amount_asset"`
}

func (a *Activity) ListPaymentsToMake(ctx context.Context) ([]RafikiPayment, error) {
	var dbPayments []RafikiPayment
	err := a.b.DB().SelectContext(ctx, &dbPayments, `SELECT ARRAY_AGG(id::text) AS ids, from_wallet, to_wallet, amount_asset, SUM(amount) AS amount
																FROM 
																	rafiki_outgoing_payments
																WHERE 
																	payment_id IS NULL
																GROUP BY 
																	from_wallet, 
																	to_wallet, 
																	amount_asset;`)
	if err != nil {
		return nil, err
	}
	return dbPayments, nil
}

func (a *Activity) CreateWebMonetizationPayment(ctx context.Context, payment RafikiPayment) (string, error) {
	senderBalances, err := a.b.LinkedAccounts().ListBalances(ctx, payment.FromWalletID)
	if err != nil {
		return "", err
	}

	var senderAcc *linkedaccounts.LinkedAccount
	for _, bal := range senderBalances {
		if bal.SendCurrency == currency.Currency(payment.Asset) {
			senderAcc = &bal
			break
		}
	}
	if senderAcc == nil {
		err = fmt.Errorf("%w failed to find sender account for currency=%s", rafiki.ErrNotFound, payment.Asset)
		return "", temporal.NewNonRetryableApplicationError("web monetization payment cron: no sending account found", "ErrInternal", err)
	}

	receiverAccs, err := a.b.LinkedAccounts().ListBalances(ctx, payment.ToWalletID)
	if err != nil {
		log.Error("failed to lookup balance accounts for receiver", zap.Error(err))
		return "", err
	}

	var receiverAcc *linkedaccounts.LinkedAccount
	for _, la := range receiverAccs {
		if currency.Currency(payment.Asset) == la.ReceiveCurrency {
			receiverAcc = &la
			break
		}
	}
	if receiverAcc == nil {
		err = fmt.Errorf("%w failed to find receiver account for currency=%s", rafiki.ErrNotFound, payment.Asset)
		return "", temporal.NewNonRetryableApplicationError("web monetization payment cron: no receiving account found", "ErrInternal", err)
	}

	p, err := a.b.Payments().Create(ctx, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: payment.FromWalletID,
		},
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: payment.ToWalletID,
		},
		SenderAccount:   senderAcc.ID,
		ReceiverAccount: receiverAcc.ID,
		Type:            payments.TypeWebMonetization,
		SenderAmount:    currency.FromUInt64(payment.Amount, currency.ParseCurrency(payment.Asset)),
		ReceiverAmount:  currency.FromUInt64(payment.Amount, currency.ParseCurrency(payment.Asset)),
		IPAddress:       "198.0.0.2", // TODO: Add a our static IP Address
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

func (a *Activity) AddWebMonetizationPayment(ctx context.Context, payout RafikiPayment, paymentID string) error {
	query, args, err := sqlx.In("UPDATE rafiki_outgoing_payments SET payment_id=?,completed=? WHERE id in (?)", paymentID, true, []string(payout.IDs))
	if err != nil {
		return err
	}
	_, err = a.b.DB().ExecContext(ctx, a.b.DB().Rebind(query), args...)
	return err
}
