package gatehub

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
)

type Client interface {
	CreateUser(ctx context.Context, walletID string) (Await, error)
	GetUser(ctx context.Context, walletID string) (*User, error)
	GetOnboardingWidget(ctx context.Context, walletID string) (string, error)
	GetOnOffRampWidget(ctx context.Context, walletID string, isDeposit bool) (string, error)
	GetBalance(ctx context.Context, linkedAccountID string) (*Balance, error)
	CreateWithdrawal(ctx context.Context, walletID, externalTransactionID string) (string, error)
	CreateTransfer(ctx context.Context, args CreateTransferArgs) (*external.Transaction, error)
	GetTransaction(ctx context.Context, walletID, id string) (*external.Transaction, error)
	ListDeliveryAddresses(ctx context.Context, walletID string) ([]external.CustomerDeliveryAddress, error)
	ListCards(ctx context.Context, externalIDs ExternalIDs) ([]external.Card, error)
	GetCardApplicationProducts(ctx context.Context) ([]external.CardApplicationProduct, error)
	OrderCard(ctx context.Context, args OrderCardArgs) error
	GetExternalIDs(ctx context.Context, walletID string) (*ExternalIDs, error)

	ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*Balance, error)
	FinaliseReserve(ctx context.Context, txID string) error
	RollbackReserve(ctx context.Context, txID string) error
	AssignBalance(ctx context.Context, linkedAccountID, trxID string, amount currency.Amount) (*Balance, error)
	LinkUserToGatewayByWalletID(ctx context.Context, walletID string) error
	LinkUserToGatewayByExternalID(ctx context.Context, ExternalID string) error
}

type Await func(ctx context.Context, result interface{}) error
