package external

import (
	"context"
	"io"
)

type Client interface {
	IssueToken(ctx context.Context, userID string, product Product) (*IssueTokenResponse, error)
	CreateUser(ctx context.Context, email string) (*CreateUserResponse, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	GetOnboardingWidget(ctx context.Context, userID string) (string, error)
	GetOnOffRampWidget(ctx context.Context, userID string, isDeposit bool) (string, error)
	GetUserWallets(ctx context.Context, userID string) (*GetUserWalletsResponse, error)
	GetWalletBalances(ctx context.Context, userID, addressID string) ([]WalletBalance, error)
	CreateTransaction(ctx context.Context, args CreateTransactionRequest) (*Transaction, error)
	GetUserTransactions(ctx context.Context, userID string) ([]Transaction, error)
	GetTransaction(ctx context.Context, userID, id string) (*Transaction, error)
	ListCards(ctx context.Context, userID, customerID string) (*ListCardsResponse, error)
	GetDeliveryAddresses(ctx context.Context, userID, customerID string) ([]CustomerDeliveryAddress, error)
	CreateCustomerDeliveryAddress(ctx context.Context, userID, customerID string, args CreateCustomerDeliveryAddressArgs) (string, error)
	GetCardApplicationProducts(ctx context.Context) ([]CardApplicationProduct, error)
	OrderCard(ctx context.Context, userID, accountID string, args OrderCardArgs) (*Card, error)
	CreateCustomerAndCard(ctx context.Context, userID string, args CreateCustomerAndCardArgs) (*Customer, error)
	CreatePlasticForCard(ctx context.Context, userID, cardID string) error
	GetCardToken(ctx context.Context, userID, tokenPath string, args GetCardTokenArgs) (*TokenResponse, error)
	FreezeCard(ctx context.Context, userID, cardID string, args FreezeCardArgs) error
	UnfreezeCard(ctx context.Context, userID, cardID string, args UnfreezeCardArgs) error
	BlockCard(ctx context.Context, userID, cardID string, args BlockCardArgs) error
	GetVaultID() string
	GetPendingThreeDSConfirmations(ctx context.Context, userID string) ([]PendingThreeDSConfirmation, error)
	ThreeDSPaymentConfirmation(ctx context.Context, userID string, args ThreeDSPaymentConfirmationArgs) error
	LinkUserToGateway(ctx context.Context, gatehubUserID string) error
	GetCardDetails(ctx context.Context, userID, cardID string) (*Card, error)
	GetCardTransaction(ctx context.Context, userID, txID string) (*CardTransaction, error)
	UpdateOrganizationConfiguration(ctx context.Context, args UpdateOrganizationConfigurationArgs) (*UpdateOrganizationConfigurationResponse, error)
	GetAccountConfirmation(ctx context.Context, userID, walletAddress string) (io.ReadCloser, error)
	GetAccountStatement(ctx context.Context, userID, walletAddress string, year, month int) (io.ReadCloser, error)
	GetTransferConfirmation(ctx context.Context, userID, txID string) (io.ReadCloser, error)
}
