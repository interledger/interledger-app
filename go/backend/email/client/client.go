package client

import (
	"context"

	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/email/ops"
	"gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
)

var _ email.Client = &client{}

type client struct {
	b ops.Backends
}

func New(
	b Backends,
	emailEnabled bool,
	sendgridAPIKey, sendgridFromName, sendgridFromEmail, sendgridOneTemplateID string,
) email.Client {
	if !emailEnabled {
		return &noopClient{}
	}

	externalClient := sendgrid.NewClient(sendgridAPIKey, sendgridFromName, sendgridFromEmail)

	ob := &opsBackends{
		Backends:   b,
		external:   externalClient,
		templateID: sendgridOneTemplateID,
	}

	return &client{
		b: ob,
	}
}

func (c *client) SendApplicationApprovedEmail(ctx context.Context, walletID string) {
	ops.SendApplicationApprovedEmail(ctx, c.b, walletID)
}

func (c *client) SendApplicationDeniedEmail(ctx context.Context, walletID string) {
	ops.SendApplicationDeniedEmail(ctx, c.b, walletID)
}

func (c *client) SendApplicationPendingEmail(ctx context.Context, walletID string) {
	ops.SendApplicationPendingEmail(ctx, c.b, walletID)
}

func (c *client) SendConnectedAccountEmail(ctx context.Context, la linkedaccounts.LinkedAccount) {
	ops.SendConnectedAccountEmail(ctx, c.b, la)
}

func (c *client) SendConnectedAccountDocumentsNeededEmail(ctx context.Context, walletID string) {
	ops.SendConnectedAccountDocumentsNeededEmail(ctx, c.b, walletID)
}

func (c *client) SendPaymentSentEmailV2(ctx context.Context, walletID string, payment *payments.Payment) {
	ops.SendPaymentSentEmailV2(ctx, c.b, walletID, payment)
}

func (c *client) SendPaymentReceivedEmailV2(ctx context.Context, walletID string, payment *payments.Payment) {
	ops.SendPaymentReceivedEmailV2(ctx, c.b, walletID, payment)
}

func (c *client) SendPaymentFailedEmail(ctx context.Context, walletID string) {
	ops.SendPaymentFailedEmail(ctx, c.b, walletID)
}

func (c *client) SendDepositReceivedEmail(ctx context.Context, walletID string, amt currency.Amount, sourceAccount, date string) {
	ops.SendDepositReceivedEmail(ctx, c.b, walletID, amt, sourceAccount, date)
}

func (c *client) SendWithdrawalEmail(ctx context.Context, walletID string, amt currency.Amount, destinationAccount, date string) {
	ops.SendWithdrawalEmail(ctx, c.b, walletID, amt, destinationAccount, date)
}

func (c *client) SendLimitsExceededEmail(ctx context.Context, walletID string) {
	ops.SendLimitsExceededEmail(ctx, c.b, walletID)
}

func (c *client) SendDepositFailedEmail(ctx context.Context, walletID string) {
	ops.SendDepositFailedEmail(ctx, c.b, walletID)
}

func (c *client) SendWithdrawalFailedEmail(ctx context.Context, walletID string) {
	ops.SendWithdrawalFailedEmail(ctx, c.b, walletID)
}

func (c *client) SendCardCreatedEmail(ctx context.Context, walletID, cardID string) {
	ops.SendCardCreatedEmail(ctx, c.b, walletID, cardID)
}

func (c *client) SendPending3DSConfirmation(ctx context.Context, walletID, confirmationID string) {
	ops.SendPending3DSConfirmation(ctx, c.b, walletID, confirmationID)
}

func (c *client) SendKYCDocumentsRequiredEmail(ctx context.Context, walletID string) {
	ops.SendKYCDocumentsRequiredEmail(ctx, c.b, walletID)
}

func (c *client) SendAgreementChangedEmail(ctx context.Context, userID string, agreements []email.AgreementLink, deadlineDate string) error {
	return ops.SendAgreementChangedEmail(ctx, c.b, userID, agreements, deadlineDate)
}

func (c *client) SendAuthenticatorResetEmail(ctx context.Context, walletID string) {
	ops.SendAuthenticatorResetEmail(ctx, c.b, walletID)
}

func (c *client) SendCardTransactionFXEmail(ctx context.Context, walletID, maskedPAN, merchantName, date, surcharge, transactionAmount, billingAmount string) {
	ops.SendCardTransactionFXEmail(ctx, c.b, walletID, maskedPAN, merchantName, date, surcharge, transactionAmount, billingAmount)
}

func (c *client) SendSCTITimeoutEmail(ctx context.Context, txID, walletID, amount, name, iban, submittedAt string) {
	ops.SendSCTITimeoutEmail(ctx, c.b, txID, walletID, amount, name, iban, submittedAt)
}
