package email

import (
	"context"

	"gitlab.com/fynbos/backend/email/sendgrid"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
)

type Client interface {
	SendApplicationApprovedEmail(ctx context.Context, walletID string)
	SendApplicationPendingEmail(ctx context.Context, walletID string)
	SendApplicationDeniedEmail(ctx context.Context, walletID string)
	SendConnectedAccountEmail(ctx context.Context, la linkedaccounts.LinkedAccount)
	SendConnectedAccountDocumentsNeededEmail(ctx context.Context, walletID string)

	PaymentSent(ctx context.Context, to []sendgrid.Email, greeting string, payment *payments.Payment)
	PaymentReceived(ctx context.Context, to []sendgrid.Email, greeting string, payment *payments.Payment)
	PaymentFailed(ctx context.Context, to []sendgrid.Email, greeting string)
	GetEmailsAndGreetingForWallet(ctx context.Context, walletID string) ([]sendgrid.Email, string, error)
}
