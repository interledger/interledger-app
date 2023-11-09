package client

import (
	"context"

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

func New(b Backends, sendgridAPIKey string) email.Client {

	externalClient := sendgrid.NewClient(sendgridAPIKey)

	ob := &opsBackends{
		Backends: b,
		external: externalClient,
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

func (c *client) PaymentSent(ctx context.Context, to []sendgrid.Email, greeting string, payment *payments.Payment) {
	ops.PaymentSent(ctx, c.b, to, greeting, payment)
}

func (c *client) PaymentReceived(ctx context.Context, to []sendgrid.Email, greeting string, payment *payments.Payment) {
	ops.PaymentReceived(ctx, c.b, to, greeting, payment)
}

func (c *client) PaymentFailed(ctx context.Context, to []sendgrid.Email, greeting string) {
	ops.PaymentFailed(ctx, c.b, to, greeting)
}

func (c *client) GetEmailsAndGreetingForWallet(ctx context.Context, walletID string) ([]sendgrid.Email, string, error) {
	return ops.GetEmailsAndGreeting(ctx, c.b, walletID)
}
