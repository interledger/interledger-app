package client

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var _ email.Client = &noopClient{}

type noopClient struct{}

func (n *noopClient) SendApplicationApprovedEmail(_ context.Context, walletID string) {
	log.Info("NOT SENDING: application approved email", zap.String("walletID", walletID))
}
func (n *noopClient) SendApplicationPendingEmail(_ context.Context, walletID string) {
	log.Info("NOT SENDING: application pending email", zap.String("walletID", walletID))
}
func (n *noopClient) SendApplicationDeniedEmail(_ context.Context, walletID string) {
	log.Info("NOT SENDING: application denied email", zap.String("walletID", walletID))
}
func (n *noopClient) SendConnectedAccountEmail(_ context.Context, la linkedaccounts.LinkedAccount) {
	log.Info("NOT SENDING: connected account email", zap.String("providerID", la.ProviderID))
}
func (n *noopClient) SendConnectedAccountDocumentsNeededEmail(_ context.Context, walletID string) {
	log.Info("NOT SENDING: connected account documents needed email", zap.String("walletID", walletID))
}
func (n *noopClient) SendPaymentFailedEmail(_ context.Context, walletID string) {
	log.Info("NOT SENDING: payment failed email", zap.String("walletID", walletID))
}
func (n *noopClient) SendPaymentSentEmailV2(_ context.Context, walletID string, payment *payments.Payment) {
	log.Info("NOT SENDING: payment sent email", zap.String("walletID", walletID), zap.String("paymentID", payment.ID))
}
func (n *noopClient) SendPaymentReceivedEmailV2(_ context.Context, walletID string, payment *payments.Payment) {
	log.Info("NOT SENDING: payment received email", zap.String("walletID", walletID), zap.String("paymentID", payment.ID))
}
func (n *noopClient) SendDepositReceivedEmail(_ context.Context, walletID string, amt currency.Amount, sourceAccount, date string) {
	log.Info("NOT SENDING: deposit received email", zap.String("walletID", walletID), zap.String("amount", amt.Format()), zap.String("sourceAccount", sourceAccount), zap.String("date", date))
}
func (n *noopClient) SendDepositFailedEmail(_ context.Context, walletID string) {
	log.Info("NOT SENDING: deposit failed email", zap.String("walletID", walletID))
}
func (n *noopClient) SendWithdrawalEmail(_ context.Context, walletID string, amt currency.Amount, destinationAccount, date string) {
	log.Info("NOT SENDING: withdrawal email", zap.String("walletID", walletID), zap.String("amount", amt.Format()), zap.String("destinationAccount", destinationAccount), zap.String("date", date))
}
func (n *noopClient) SendWithdrawalFailedEmail(_ context.Context, walletID string) {
	log.Info("NOT SENDING: withdrawal failed email", zap.String("walletID", walletID))
}
func (n *noopClient) SendLimitsExceededEmail(_ context.Context, walletID string) {
	log.Info("NOT SENDING: limits exceeded email", zap.String("walletID", walletID))
}
func (n *noopClient) SendCardCreatedEmail(_ context.Context, walletID, cardID string) {
	log.Info("NOT SENDING: card created email", zap.String("walletID", walletID), zap.String("cardID", cardID))
}
func (n *noopClient) SendPending3DSConfirmation(_ context.Context, walletID, confirmationID string) {
	log.Info("NOT SENDING: pending 3DS confirmation email", zap.String("walletID", walletID), zap.String("confirmationID", confirmationID))
}
