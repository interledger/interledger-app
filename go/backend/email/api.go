package email

import (
	"context"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/payments"
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

	SendPaymentSentEmailV2(ctx context.Context, walletID string, payment *payments.Payment)
	SendPaymentReceivedEmailV2(ctx context.Context, walletID string, payment *payments.Payment)
}
