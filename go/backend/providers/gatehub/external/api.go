package external

import "context"

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
	ListCards(ctx context.Context, userID, customerID string) ([]Card, error)
	GetDeliveryAddresses(ctx context.Context, userID, customerID string) ([]CustomerDeliveryAddress, error)
	GetCardApplicationProducts(ctx context.Context) ([]CardApplicationProduct, error)
	OrderCard(ctx context.Context, userID string, gatehubWalletAddress string) error
	CreateCustomerAndCard(ctx context.Context, gatehubUserId string, gatehubWalletAddress string) (*CreateCardDTO, error)
	GetVaultID() string
	LinkUserToGateway(ctx context.Context, gatehubUserId string) error
}
