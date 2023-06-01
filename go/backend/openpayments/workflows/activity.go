package workflows

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/providers"

	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	"gitlab.com/fynbos/backend/paymentpointers"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/env"
	"go.temporal.io/sdk/temporal"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
}

func getDefaultSendAcc(ctx context.Context, b Backends, pointer string) (*linkedaccounts.LinkedAccount, error) {
	pp, err := ops.GetPaymentPointer(ctx, b, pointer)
	if err != nil {
		return nil, err
	}

	accs, err := b.LinkedAccounts().ListByWalletId(ctx, pp.WalletID)
	if err != nil {
		return nil, err
	}

	// Search all linked accounts and check for the first account that can send funds.
	// TODO: Add default flags to linked accounts and add more providers
	for _, ra := range accs {
		if ra.CanSend {
			return &ra, nil
		}
	}

	return nil, fmt.Errorf("%w no account capable of receiving found for payment pointer (%s)", openpayments.ErrNotFound, pointer)
}

func getDefaultRecvAcc(ctx context.Context, b Backends, pointer string) (*linkedaccounts.LinkedAccount, error) {
	pp, err := ops.GetPaymentPointer(ctx, b, pointer)
	if err != nil {
		return nil, err
	}

	accs, err := b.LinkedAccounts().ListByWalletId(ctx, pp.WalletID)
	if err != nil {
		return nil, err
	}

	// Search all linked accounts and check for the first account that can receive funds.
	// TODO: Add default flags to linked accounts and add more providers
	for _, ra := range accs {
		if ra.CanReceive {
			return &ra, nil
		}
	}

	return nil, fmt.Errorf("%w no account capable of receiving found for payment pointer (%s)", openpayments.ErrNotFound, pointer)
}

type ProviderWorkflowArgs struct {
	Args providers.TransfersArgs
	Key  providers.WorkflowKey
}

func (a *Activity) GetProviderWorkflowArgs(ctx context.Context, outgoingID string) (*ProviderWorkflowArgs, error) {
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

	recvPPURL := ip.PaymentPointer
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("failed to parse payment pointer URL from receiver (%s)", op.Receiver), "ErrInvalidURL", err)
	}

	recvAcc, err := getDefaultRecvAcc(ctx, a.b, recvPPURL)
	if errors.Is(err, openpayments.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	var sendAcc *linkedaccounts.LinkedAccount
	if op.FromLinkedAccount == "" {
		sendAcc, err = getDefaultSendAcc(ctx, a.b, op.PaymentPointer)
	} else {
		sendAcc, err = a.b.LinkedAccounts().Get(ctx, op.FromLinkedAccount)
	}
	if errors.Is(err, openpayments.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	var key providers.WorkflowKey
	if sendAcc.Provider == tabapay.ProviderName && recvAcc.Provider == mx.ProviderName {
		key = providers.GMTCARD2ACH
	} else if sendAcc.Provider == mx.ProviderName && recvAcc.Provider == tabapay.ProviderName {
		key = providers.GMTACH2CARD
	} else if sendAcc.Provider == mx.ProviderName && recvAcc.Provider == mx.ProviderName {
		key = providers.GMTACH2ACH
	} else {
		key = providers.GMTUNSUPPORTED
	}

	return &ProviderWorkflowArgs{
		Args: providers.TransfersArgs{
			FromForeignID:       outgoingID,
			ToForeignID:         incomingID,
			FromPaymentPointer:  op.PaymentPointer,
			ToPaymentPointer:    op.ToPaymentPointer,
			FromLinkedAccountID: sendAcc.ID,
			ToLinkedAccountID:   recvAcc.ID,
			FromWalletID:        sendAcc.WalletID,
			ToWalletID:          recvAcc.WalletID,
			Amount:              op.SendAmount,
		},
		Key: key,
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

func (a *Activity) SendFailedOutgoingPaymentMail(ctx context.Context, outgoingID string) error {
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

	user, err := a.b.KYC().GetIndividualDetails(ctx, pp.WalletID)
	if errors.Is(err, openpayments.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	actionUrl, err := url.JoinPath(env.GetUrl(), "pay")
	if err != nil {
		return err
	}

	err = a.b.Email().SendMailTemplate(ctx, pp.WalletID, email.FailedTransactionTemplateID, map[string]interface{}{
		"subject":         fmt.Sprintf(email.FailedTransactionTemplateID.Subject(), "payment"),
		"transactionType": "payment",
		"name":            user.FirstName,
		"actionUrl":       actionUrl,
	}, []email.Attachment{})

	return err
}

func (a *Activity) AddContact(ctx context.Context, fromPaymentPointer, toPaymentPointer string) error {
	issuedFromPaymentPointer, err := ops.GetPaymentPointer(ctx, a.b, fromPaymentPointer)
	if err != nil {
		return err
	}
	issuedToPaymentPointer, err := ops.GetPaymentPointer(ctx, a.b, toPaymentPointer)
	if err != nil {
		return err
	}

	tpp, err := paymentpointers.Parse(toPaymentPointer)
	if err != nil {
		return err
	}

	// Check if it exists and don't error on not found
	c, err := a.b.Contacts().Get(ctx, issuedFromPaymentPointer.WalletID, tpp)
	if err != nil && !errors.Is(err, contacts.ErrNotFound) {
		return err
	}
	// Exists so we can move on
	if c != nil {
		return nil
	}

	toWallet, err := a.b.Users().GetWallet(ctx, issuedToPaymentPointer.WalletID)
	if err != nil {
		return err
	}

	// Create new contact
	_, err = a.b.Contacts().Create(ctx, contacts.CreateContactArgs{
		Name:           toWallet.Name,
		PaymentPointer: tpp,
		WalletID:       issuedFromPaymentPointer.WalletID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) MarkContactLastPaid(ctx context.Context, fromPaymentPointer, toPaymentPointer string) error {
	issuedFromPaymentPointer, err := ops.GetPaymentPointer(ctx, a.b, fromPaymentPointer)
	if err != nil {
		return err
	}

	tpp, err := paymentpointers.Parse(toPaymentPointer)
	if err != nil {
		return err
	}

	//Also mark as last paid for now
	err = a.b.Contacts().SetLastPaidAtNow(ctx, issuedFromPaymentPointer.WalletID, tpp)
	if err != nil {
		return err
	}

	return nil
}
