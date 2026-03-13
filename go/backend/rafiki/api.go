package rafiki

import (
	"context"
	"net/http"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/wallets"
)

type Client interface {
	WebhookHandler() http.HandlerFunc
	CreatePaymentPointer(ctx context.Context, address wallets.Wallet) (string, error)
	GetWalletAddress(ctx context.Context, walletID string) (*WalletAddress, error)
	CreatePaymentPointerKey(ctx context.Context, keyID string, walletID string) error
	RevokePaymentPointerKey(ctx context.Context, keyID string) error
	FundOutgoingPayment(ctx context.Context, paymentID string) error
	FinalizeWebMonetization(ctx context.Context, paymentID string) error
	RollbackWebMonetization(ctx context.Context, paymentID string) error
	ListGrants(ctx context.Context, walletID string) ([]Grant, error)
	GetGrant(ctx context.Context, grantID string) (*Grant, error)
	RevokeGrant(ctx context.Context, grantID string) error
	ListPendingTransactions(ctx context.Context, walletID string) ([]transactions.Transaction, error)
	UpdateWalletAddressStatus(ctx context.Context, walletId UpdateAddressStatus, isActive bool) error
	GetIncomingPayment(ctx context.Context, id string) (*IncomingPayment, error)
	CancelOutgoingPayment(ctx context.Context, paymentPointerID, reason string) error
	WithdrawIncomingPaymentLiquidity(ctx context.Context, incomingPaymentID string) error
	WithdrawOutgoingPaymentLiquidity(ctx context.Context, outgoingPaymentID string) error
}
