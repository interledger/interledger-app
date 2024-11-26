package pti

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/currency"
)

type Client interface {
	CreateWallet(ctx context.Context, walletID string, currency currency.Currency) (Await, error)
	GetWallet(ctx context.Context, linkedAccountID string) (*Wallet, error)
	GetBalance(ctx context.Context, linkedAccountID string) (*Balance, error)
	DepositToWallet(ctx context.Context, args TransactionArgs) (string, error)
	WithdrawalFromWallet(ctx context.Context, args TransactionArgs) (string, error)
	UpdateTransactionStatus(ctx context.Context, args TransactionStatusArgs) error

	ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*Balance, error)
	FinaliseReserve(ctx context.Context, txID string) error
	RollbackReserve(ctx context.Context, txID string) error
	AssignBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount) (*Balance, error)

	ReserveTransfer(ctx context.Context, fromAccount, toAccount, txID string, amt currency.Amount, timeout time.Duration) error

	CreateJWT(ctx context.Context, args TokenArgs) (*TokenResponse, error)
	GetWidget(ctx context.Context, walletID string) (*WidgetDetails, error)
	CreateCard(ctx context.Context, walletID, tokenID string) (Await, error)
	GetLinkedAccountCardDetails(ctx context.Context, id string) (*EncryptedCreditCardPaymentInformation, error)
}
