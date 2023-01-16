package workflows

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/providers/machnet"
	"go.temporal.io/sdk/temporal"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
}

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
	// Our friends may have provided the full ID with the payment pointer and the `incoming-payments` prefix.
	idxSlash := strings.LastIndex(outgoingID, "/")
	if idxSlash > 0 {
		outgoingID = outgoingID[idxSlash+1:]
	}

	op, err := ops.GetOutgoingPayment(ctx, a.b, outgoingID)
	if err != nil {
		return nil, err
	}

	ip, err := ops.GetIncomingPayment(ctx, a.b, op.Receiver)
	if err != nil {
		return nil, err
	}

	incomingID := ip.ID
	idxSlash = strings.LastIndex(incomingID, "/")
	if idxSlash > 0 {
		incomingID = incomingID[idxSlash+1:]
	}

	recvPPURL, _, err := ops.ExtractPaymentPointer(op.Receiver)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("failed to parse payment pointer URL from receiver (%s)", op.Receiver), "ErrInvalidURL", err)
	}

	recvAcc, err := getProviderLinkedAccount(ctx, a.b, recvPPURL, machnet.ProviderName, machnet.TypeWallet)
	if errors.Is(err, openpayments.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	sendAccID := op.FromLinkedAccount

	if sendAccID == "" {
		sendAcc, err := getProviderLinkedAccount(ctx, a.b, op.PaymentPointer, machnet.ProviderName, machnet.TypeSendCard)
		if errors.Is(err, openpayments.ErrNotFound) {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
		}
		if err != nil {
			return nil, err
		}

		sendAccID = sendAcc.ID
	}

	return &machnet.CreateTransactionArgs{
		FromForeignID:       outgoingID,
		ToForeignID:         incomingID,
		FromPaymentPointer:  op.PaymentPointer,
		ToPaymentPointer:    op.ToPaymentPointer,
		FromLinkedAccountID: sendAccID,
		ToLinkedAccountID:   recvAcc.ID,
		Amount:              op.SendAmount.Float64(),
		Currency:            op.SendAmount.Currency.String(),
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

func (a *Activity) SendOutgoingPaymentReceipt(ctx context.Context, outgoingID string, extID string) error {
	op, err := ops.GetOutgoingPayment(ctx, a.b, outgoingID)
	if errors.Is(err, openpayments.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	pp, err := ops.GetPaymentPointer(ctx, a.b, op.PaymentPointer)
	if err != nil {
		return err
	}

	// TODO verify that these are formatted correctly.
	err = a.b.Email().SendMailTemplate(ctx, pp.WalletID, email.ReceiptTemplateID, map[string]interface{}{
		"transactionID":      op.ID,
		"paymentDate":        op.UpdatedAt.Format(time.RFC1123),
		"sendAmount":         op.SentAmount.Format(),
		"fees":               "$ 0.00",
		"receiveAmount":      op.ReceiveAmount.Format(),
		"note":               op.Description,
		"toPaymentPointer":   op.ToPaymentPointer,
		"fromPaymentPointer": op.PaymentPointer,
	}, []email.Attachment{})

	return err
}

func (a *Activity) SendIncomingPaymentReceipt(ctx context.Context, outgoingID string) error {
	op, err := ops.GetOutgoingPayment(ctx, a.b, outgoingID)
	if errors.Is(err, openpayments.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	ip, err := ops.GetIncomingPayment(ctx, a.b, op.Receiver)
	if errors.Is(err, openpayments.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	pp, err := ops.GetPaymentPointer(ctx, a.b, ip.PaymentPointer)
	if err != nil {
		return err
	}

	err = a.b.Email().SendMailTemplate(ctx, pp.WalletID, email.ReceivedReceiptTemplateID, map[string]interface{}{
		"fromPaymentPointer": ip.FromPaymentPointer,
		"toPaymentPointer":   ip.PaymentPointer,
		"transactionID":      ip.ID,
		"paymentDate":        ip.UpdatedAt.Format(time.RFC1123),
		"receiveAmount":      ip.ReceivedAmount.Format(),
		"note":               ip.Description,
		"subject":            "Fynbos payment received.",
	}, []email.Attachment{})

	return err
}
