package storage

import (
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
)

// Storage defines the interface for data persistence
type Storage interface {
	// Users
	CreateUser(user *models.User) error
	GetUser(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error

	// Card Customers
	CreateCustomer(customer *models.Customer) error
	GetCustomer(id string) (*models.Customer, error)
	GetCustomerBySourceID(sourceID string) (*models.Customer, error)
	UpdateCustomer(customer *models.Customer) error

	// Card Accounts
	CreateAccount(account *models.Account) error
	GetAccount(id string) (*models.Account, error)
	UpdateAccount(account *models.Account) error

	// Card Delivery Addresses
	CreateCustomerAddress(customerID string, address *models.CustomerDeliveryAddress) error
	GetCustomerAddresses(customerID string) ([]*models.CustomerDeliveryAddress, error)

	// Cards
	CreateCard(card *models.Card) error
	GetCard(id string) (*models.Card, error)
	UpdateCard(card *models.Card) error
	GetCardsByCustomer(customerID string) ([]*models.Card, error)
	GetCardsByAccount(accountID string) ([]*models.Card, error)
	GetCardLimits(cardID string) ([]models.CardLimit, error)
	SetCardLimits(cardID string, limits []models.CardLimit) error

	// Card Transactions
	CreateCardTransaction(tx *models.CardTransaction) error
	GetCardTransaction(id string) (*models.CardTransaction, error)
	AddCardTransactionIndex(cardID string, transactionID string) error
	GetCardTransactionIDs(cardID string) ([]string, error)

	// Wallets
	CreateWallet(wallet *models.Wallet) error
	GetWallet(address string) (*models.Wallet, error)
	GetWalletsByUser(userID string) ([]*models.Wallet, error)

	// Transactions
	CreateTransaction(tx *models.Transaction) error
	GetTransaction(id string) (*models.Transaction, error)
	UpdateTransactionStatus(id string, status int) error

	// Balances (per user, per currency)
	GetBalance(userID, currency string) (float64, error)
	AddBalance(userID, currency string, amount float64) error
	DeductBalance(userID, currency string, amount float64) error

	// 3DS Challenges
	CreateThreeDSChallenge(challenge *models.ThreeDSChallenge) error
	GetThreeDSChallenge(txID string) (*models.ThreeDSChallenge, error)
	GetPendingThreeDSChallenges(userID string) ([]*models.ThreeDSChallenge, error)
	UpdateThreeDSChallenge(challenge *models.ThreeDSChallenge) error

	// Organizations
	GetOrganization(orgID string) (*models.Organization, error)
	CreateOrganization(org *models.Organization) error
	UpdateOrganization(org *models.Organization) error

	// Admin UI enumeration
	ListUsers() ([]*models.User, error)
	ListTransactionsByUser(userID string) ([]*models.Transaction, error)
	GetAllBalances(userID string) (map[string]float64, error)
}
