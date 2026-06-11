package email

import (
	"context"

	"gitlab.com/fynbos/backend/currency"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
)

type Client interface {
	SendApplicationApprovedEmail(ctx context.Context, walletID string)
	SendApplicationPendingEmail(ctx context.Context, walletID string)
	SendApplicationDeniedEmail(ctx context.Context, walletID string)
	SendConnectedAccountEmail(ctx context.Context, la linkedaccounts.LinkedAccount)
	SendConnectedAccountDocumentsNeededEmail(ctx context.Context, walletID string)
	SendPaymentFailedEmail(ctx context.Context, walletID string)

	SendPaymentSentEmailV2(ctx context.Context, walletID string, payment *payments.Payment)
	SendPaymentReceivedEmailV2(ctx context.Context, walletID string, payment *payments.Payment)
	SendDepositReceivedEmail(ctx context.Context, walletID string, amt currency.Amount, sourceAccountName, date string)
	SendDepositFailedEmail(ctx context.Context, walletID string)
	SendWithdrawalEmail(ctx context.Context, walletID string, amt currency.Amount, destinationAccount, date string)
	SendWithdrawalFailedEmail(ctx context.Context, walletID string)
	SendGatehubWithdrawalRejectedEmail(ctx context.Context, walletID, amount, currency, iban, name string)
	SendLimitsExceededEmail(ctx context.Context, walletID string)
	SendCardCreatedEmail(ctx context.Context, walletID, cardID string)
	SendPending3DSConfirmation(ctx context.Context, walletID, confirmationID string)
	SendKYCDocumentsRequiredEmail(ctx context.Context, walletID string)
	SendAuthenticatorResetEmail(ctx context.Context, walletID string)
	SendCardTransactionFXEmail(ctx context.Context, walletID, maskedPAN, merchantName, date, surcharge, transactionAmount, billingAmount string)
	SendSCTITimeoutEmail(ctx context.Context, txID, walletID, amount, name, iban, submittedAt string)
}
