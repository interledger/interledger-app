package email

import (
	"context"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/openpayments"
)

type Client interface {
	SendApplicationApprovedEmail(ctx context.Context, walletID string)
	SendApplicationPendingEmail(ctx context.Context, walletID string)
	SendApplicationDeniedEmail(ctx context.Context, walletID string)
	SendConnectedAccountEmail(ctx context.Context, la linkedaccounts.LinkedAccount)
	SendConnectedAccountDocumentsNeededEmail(ctx context.Context, walletID string)
	SendPaymentSentEmail(ctx context.Context, walletID, trxID string, op openpayments.OutgoingPayment)
	SendPaymentReceivedEmail(ctx context.Context, walletID, trxID string, ip openpayments.IncomingPayment)
	SendPaymentFailedEmail(ctx context.Context, walletID string)
}
