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
	ListCards(ctx context.Context, userID, customerID string) (*ListCardsResponse, error)
	GetDeliveryAddresses(ctx context.Context, userID, customerID string) ([]CustomerDeliveryAddress, error)
	CreateCustomerDeliveryAddress(ctx context.Context, userID, customerID string, args CreateCustomerDeliveryAddressArgs) (string, error)
	GetCardApplicationProducts(ctx context.Context) ([]CardApplicationProduct, error)
	OrderCard(ctx context.Context, userID, accountID string, args OrderCardArgs) (*Card, error)
	CreateCustomerAndCard(ctx context.Context, userID string, args CreateCustomerAndCardArgs) (*Customer, error)
	CreatePlasticForCard(ctx context.Context, userID, cardID string) error
	GetVaultID() string
	LinkUserToGateway(ctx context.Context, gatehubUserId string) error
}
