package workflows

import (
	"context"
	"errors"
	"fmt"
	"math"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/providers/machnet"
	"go.temporal.io/sdk/temporal"
)

func getProviderLinkedAccount(ctx context.Context, b Backends, pointer, providerName, providerType string) (*linkedaccounts.LinkedAccount, error) {
	pp, err := ops.GetPaymentPointer(ctx, b, pointer)
	if err != nil {
		return nil, err
	}

	accs, err := b.LinkedAccounts().ListByWalletId(ctx, pp.WalletID)
	if err != nil {
		return nil, err
	}

	var found *linkedaccounts.LinkedAccount
	for _, ra := range accs {
		if ra.Provider != providerName ||
			ra.Type != providerType {
			continue
		}
		found = &ra
		break
	}

	if found == nil {
		return nil, fmt.Errorf("%w no account type (%s) for provider (%s) account payment pointer (%s)", openpayments.ErrNotFound, providerType, providerName, pointer)
	}

	return found, nil
}

func (a *Activity) GetProviderArgs(ctx context.Context, outgoingID string) (*machnet.CreateTransactionArgs, error) {
	op, err := ops.GetOutgoingPayment(ctx, a.b, outgoingID)
	if err != nil {
		return nil, err
	}

	recvPPURL, _, err := ops.ExtractPaymentPointer(op.Receiver)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("failed to parse payment pointer URL from receiver (%s)", op.Receiver), "ErrInvalidURL", err)
	}

	recvAcc, err := getProviderLinkedAccount(ctx, a.b, recvPPURL, machnet.ProviderName, machnet.TypeReceiveBankAccount)
	if errors.Is(err, openpayments.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	sendAcc, err := getProviderLinkedAccount(ctx, a.b, recvPPURL, machnet.ProviderName, machnet.TypeSendCard)
	if errors.Is(err, openpayments.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	amnt := float64(op.SendAmount.Value)
	if op.SendAmount.AssetScale > 0 {
		amnt /= math.Pow(10, float64(op.SendAmount.AssetScale))
	}

	return &machnet.CreateTransactionArgs{
		FromLinkedAccountID: sendAcc.ID,
		ToLinkedAccountID:   recvAcc.ID,
		Amount:              amnt,
		Currency:            op.SendAmount.Asset,
	}, nil
}

func (a *Activity) FailOutgoingPayment(ctx context.Context, outgoingID string) error {
	return ops.FailOutgoingPayment(ctx, a.b, outgoingID)
}

func (a *Activity) CompleteOutgoingPayment(ctx context.Context, outgoingID, extID string) error {
	// TODO: lookup the send amount from the provider, for now just assume the full amount was sent
	op, err := ops.GetOutgoingPayment(ctx, a.b, outgoingID)
	if errors.Is(err, openpayments.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	err = ops.CompleteOutgoingPayment(ctx, a.b, openpayments.CompleteOutgoingPaymentArgs{
		ID:         outgoingID,
		SentAmount: op.SendAmount,
	})
	if errors.Is(err, openpayments.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}

	return err
}

func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if !temporal.IsApplicationError(err) {
		return false
	}

	var applicationError *temporal.ApplicationError
	errors.As(err, &applicationError)

	return applicationError.NonRetryable()
}
