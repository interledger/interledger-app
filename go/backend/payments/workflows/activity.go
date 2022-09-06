package workflows

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/payments"

	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/providers/noop"
	"go.temporal.io/sdk/activity"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
}

func (s *Activity) CreatePendingOutgoingPayment(ctx context.Context, outgoingPaymentId, outgoingPaymentTransferID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating pending transaction")

	outgoingPayment, err := s.b.Payments().GetUnauthenticated(ctx, outgoingPaymentId)
	if err != nil {
		return "", err
	}

	acc, err := s.b.Accounts().Get(ctx, outgoingPayment.AccountID)
	if err != nil {
		return "", err
	}

	// TODO this needs to be better way to handle getting the data.
	trx, err := s.b.Transactions().CreatePending(ctx, &transactions.CreatePendingTransactionArgs{
		AccountID:   outgoingPayment.AccountID,
		Type:        "outgoingPayment",
		NetAmount:   outgoingPayment.Amount,
		Description: "Sent to " + outgoingPayment.Destination,
		LedgerTransfers: []transactions.CreateLedgerTransferArgs{
			{
				ID:              outgoingPaymentTransferID,
				LedgerID:        s.b.Noop().GetLedgerID(),
				DebitAccountID:  s.b.Noop().GetEquityAccountID(),
				CreditAccountID: acc.LedgerAccountID,
				Amount:          outgoingPayment.Amount,
				// Code: uint16
				Flags: transactions.LedgerTransferFlags{
					TwoPhaseCommit: true,
				},
			},
		},
	})

	// TODO need to work out retryable vs non retryable errors
	if err != nil {
		return "", err
	}

	if err != nil {
		switch {
		case errors.Is(err, transactions.ErrExceedsDebits):
			return "", payments.ErrInsufficientBalance
		default:
			return "", fmt.Errorf("%s %w", err, payments.ErrInternal)
		}
	}

	return trx.ID, nil
}

func (s *Activity) ProcessNoopOutgoingPayment(ctx context.Context, outgoingPaymentId string) error {

	outgoingPayment, err := s.b.Payments().GetUnauthenticated(ctx, outgoingPaymentId)
	if err != nil {
		return err
	}

	_, err = s.b.Accounts().Get(ctx, outgoingPayment.AccountID)
	if err != nil {
		return err
	}

	err = s.b.Noop().InitiateOutgoingPayment(ctx, &noop.OutgoingPaymentArgs{
		Amount: outgoingPayment.Amount,
		To:     outgoingPayment.Destination,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) VoidPendingOutgoingPayment(ctx context.Context, trxId string) error {

	_, err := s.b.Transactions().VoidPending(ctx, trxId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) PostPendingOutgoingPayment(ctx context.Context, trxId string) error {

	_, err := s.b.Transactions().PostPending(ctx, trxId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Activity) SetOutgoingPaymentState(ctx context.Context, outgoingPaymentId string, state payments.State) error {
	err := s.b.Payments().SetState(ctx, outgoingPaymentId, state)
	if err != nil {
		return err
	}
	return nil
}
