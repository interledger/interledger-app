package gatehub

import (
	"context"
	"io"
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
	GetCardToken(ctx context.Context, args GetCardTokenArgs) (*external.TokenResponse, error)
	FreezeCard(ctx context.Context, args FreezeCardArgs) error
	UnfreezeCard(ctx context.Context, args UnfreezeCardArgs) error
	BlockCard(ctx context.Context, args BlockCardArgs) error
	ValidateCardProductCode(ctx context.Context, cardProductCode string) error
	GetPendingThreeDSConfirmations(ctx context.Context, userID string) ([]external.PendingThreeDSConfirmation, error)
	ThreeDSPaymentConfirmation(ctx context.Context, userID, txID string, confirmed bool) error

	ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*Balance, error)
	FinaliseReserve(ctx context.Context, txID string) error
	RollbackReserve(ctx context.Context, txID string) error
	AssignBalance(ctx context.Context, linkedAccountID, trxID string, amount currency.Amount) (*Balance, error)
	TransferUserToOmnibus(ctx context.Context, senderLinkedAccountID string, amount currency.Amount) (*external.Transaction, error)
	TransferOmnibusToUser(ctx context.Context, receiverLinkedAccountID string, amount currency.Amount) (*external.Transaction, error)
	LinkUserToGatewayByWalletID(ctx context.Context, walletID string) error
	LinkUserToGatewayByExternalID(ctx context.Context, ExternalID string) error
	UpdateOrganizationConfiguration(ctx context.Context, apiBaseURL, twoFAType string) (*external.UpdateOrganizationConfigurationResponse, error)
	GetAccountConfirmation(ctx context.Context, walletID string) (io.ReadCloser, error)
	GetAccountStatement(ctx context.Context, walletID string, year, month int) (io.ReadCloser, error)
}

type Await func(ctx context.Context, result interface{}) error
