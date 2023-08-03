package client

import (
	"context"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/email/ops"
	"gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/openpayments"
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

func (c *client) SendPaymentSentEmail(ctx context.Context, walletID, trxID string, op openpayments.OutgoingPayment) {
	ops.SendPaymentSentEmail(ctx, c.b, walletID, trxID, op)
}

func (c *client) SendPaymentReceivedEmail(ctx context.Context, walletID, trxID string, ip openpayments.IncomingPayment) {
	ops.SendPaymentReceivedEmail(ctx, c.b, walletID, trxID, ip)
}

func (c *client) SendPaymentFailedEmail(ctx context.Context, walletID string) {
	ops.SendPaymentFailedEmail(ctx, c.b, walletID)
}
