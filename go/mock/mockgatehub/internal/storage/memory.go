package storage

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/utils"
)

// MemoryStorage implements Storage using in-memory maps
type MemoryStorage struct {
	mu                     sync.RWMutex
	users                  map[string]*models.User                      // userID -> User
	usersByEmail           map[string]*models.User                      // email -> User
	customers              map[string]*models.Customer                  // customerID -> Customer
	customersBySource      map[string]*models.Customer                  // sourceID -> Customer
	accounts               map[string]*models.Account                   // accountID -> Account
	cards                  map[string]*models.Card                      // cardID -> Card
	cardTransactions       map[string]*models.CardTransaction           // transactionID -> CardTransaction
	cardTransactionsByCard map[string][]string                          // cardID -> transactionIDs
	rawCardTransactions    map[string]json.RawMessage                   // transactionID -> raw JSON payload
	cardTxSeqID            atomic.Int64                                 // globally incrementing card tx sequence id
	cardLimits             map[string][]models.CardLimit                // cardID -> limits
	customerAddresses      map[string][]*models.CustomerDeliveryAddress // customerID -> addresses
	threeDSChallenges      map[string]*models.ThreeDSChallenge          // transactionID -> ThreeDSChallenge
	wallets                map[string]*models.Wallet                    // address -> Wallet
	transactions           map[string]*models.Transaction               // txID -> Transaction
	balances               map[string]map[string]float64                // userID -> currency -> amount
	organizations          map[string]*models.Organization              // orgID -> Organization
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:                  make(map[string]*models.User),
		usersByEmail:           make(map[string]*models.User),
		customers:              make(map[string]*models.Customer),
		customersBySource:      make(map[string]*models.Customer),
		accounts:               make(map[string]*models.Account),
		cards:                  make(map[string]*models.Card),
		cardTransactions:       make(map[string]*models.CardTransaction),
		cardTransactionsByCard: make(map[string][]string),
		rawCardTransactions:    make(map[string]json.RawMessage),
		cardLimits:             make(map[string][]models.CardLimit),
		customerAddresses:      make(map[string][]*models.CustomerDeliveryAddress),
		threeDSChallenges:      make(map[string]*models.ThreeDSChallenge),
		wallets:                make(map[string]*models.Wallet),
		transactions:           make(map[string]*models.Transaction),
		balances:               make(map[string]map[string]float64),
		organizations:          make(map[string]*models.Organization),
	}
}

// CreateUser creates a new user
func (s *MemoryStorage) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user.Email == "" {
		return fmt.Errorf("email is required")
	}

	// Check if email already exists
	if _, exists := s.usersByEmail[user.Email]; exists {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	// Generate ID if not provided
	if user.ID == "" {
		user.ID = utils.GenerateUUID()
	}

	// Set defaults
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	s.users[user.ID] = user
	s.usersByEmail[user.Email] = user

	return nil
}

// GetUser retrieves a user by ID
func (s *MemoryStorage) GetUser(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *MemoryStorage) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByEmail[email]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *MemoryStorage) UpdateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.users[user.ID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Update email index if changed
	if existing.Email != user.Email {
		delete(s.usersByEmail, existing.Email)
		s.usersByEmail[user.Email] = user
	}

	s.users[user.ID] = user
	return nil
}

// Card customer operations

func (s *MemoryStorage) CreateCustomer(customer *models.Customer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if customer.ID == nil || *customer.ID == "" {
		id := utils.GenerateUUID()
		customer.ID = &id
	}

	if customer.SourceID == "" {
		return fmt.Errorf("sourceId is required")
	}

	if customer.CreatedAt.IsZero() {
		customer.CreatedAt = time.Now()
	}

	if _, exists := s.customers[*customer.ID]; exists {
		return fmt.Errorf("customer already exists")
	}

	s.customers[*customer.ID] = customer
	s.customersBySource[customer.SourceID] = customer
	return nil
}

func (s *MemoryStorage) GetCustomer(id string) (*models.Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	customer, exists := s.customers[id]
	if !exists {
		return nil, fmt.Errorf("customer not found")
	}

	return customer, nil
}

func (s *MemoryStorage) GetCustomerBySourceID(sourceID string) (*models.Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	customer, exists := s.customersBySource[sourceID]
	if !exists {
		return nil, fmt.Errorf("customer not found")
	}

	return customer, nil
}

func (s *MemoryStorage) UpdateCustomer(customer *models.Customer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if customer.ID == nil || *customer.ID == "" {
		return fmt.Errorf("customer ID is required")
	}

	if _, exists := s.customers[*customer.ID]; !exists {
		return fmt.Errorf("customer not found")
	}

	s.customers[*customer.ID] = customer
	s.customersBySource[customer.SourceID] = customer
	return nil
}

// Card account operations

func (s *MemoryStorage) CreateAccount(account *models.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if account.ID == nil || *account.ID == "" {
		id := utils.GenerateUUID()
		account.ID = &id
	}

	if _, exists := s.accounts[*account.ID]; exists {
		return fmt.Errorf("account already exists")
	}

	if account.CreatedAt.IsZero() {
		account.CreatedAt = time.Now()
	}

	s.accounts[*account.ID] = account
	return nil
}

func (s *MemoryStorage) GetAccount(id string) (*models.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, exists := s.accounts[id]
	if !exists {
		return nil, fmt.Errorf("account not found")
	}

	return account, nil
}

func (s *MemoryStorage) UpdateAccount(account *models.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if account.ID == nil || *account.ID == "" {
		return fmt.Errorf("account ID is required")
	}

	if _, exists := s.accounts[*account.ID]; !exists {
		return fmt.Errorf("account not found")
	}

	s.accounts[*account.ID] = account
	return nil
}

// Card delivery address operations

func (s *MemoryStorage) CreateCustomerAddress(customerID string, address *models.CustomerDeliveryAddress) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if customerID == "" {
		return fmt.Errorf("customer ID is required")
	}
	if address == nil {
		return fmt.Errorf("address is required")
	}
	if address.ID == "" {
		address.ID = utils.GenerateUUID()
	}

	s.customerAddresses[customerID] = append(s.customerAddresses[customerID], address)
	return nil
}

func (s *MemoryStorage) GetCustomerAddresses(customerID string) ([]*models.CustomerDeliveryAddress, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	addresses := s.customerAddresses[customerID]
	if addresses == nil {
		return []*models.CustomerDeliveryAddress{}, nil
	}

	return append([]*models.CustomerDeliveryAddress{}, addresses...), nil
}

// Card operations

func (s *MemoryStorage) CreateCard(card *models.Card) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if card.ID == "" {
		card.ID = utils.GenerateUUID()
	}

	if card.CreatedAt.IsZero() {
		card.CreatedAt = time.Now()
	}

	if _, exists := s.cards[card.ID]; exists {
		return fmt.Errorf("card already exists")
	}

	s.cards[card.ID] = card
	return nil
}

func (s *MemoryStorage) GetCard(id string) (*models.Card, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	card, exists := s.cards[id]
	if !exists {
		return nil, fmt.Errorf("card not found")
	}

	return card, nil
}

func (s *MemoryStorage) UpdateCard(card *models.Card) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if card.ID == "" {
		return fmt.Errorf("card ID is required")
	}

	if _, exists := s.cards[card.ID]; !exists {
		return fmt.Errorf("card not found")
	}

	s.cards[card.ID] = card
	return nil
}

func (s *MemoryStorage) GetCardsByCustomer(customerID string) ([]*models.Card, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cards []*models.Card
	for _, card := range s.cards {
		if card.CustomerID == customerID {
			cards = append(cards, card)
		}
	}

	return cards, nil
}

func (s *MemoryStorage) GetCardsByAccount(accountID string) ([]*models.Card, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cards []*models.Card
	for _, card := range s.cards {
		if card.AccountID == accountID {
			cards = append(cards, card)
		}
	}

	return cards, nil
}

func (s *MemoryStorage) GetCardLimits(cardID string) ([]models.CardLimit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limits, ok := s.cardLimits[cardID]
	if !ok {
		return []models.CardLimit{}, nil
	}

	result := make([]models.CardLimit, len(limits))
	copy(result, limits)
	return result, nil
}

func (s *MemoryStorage) SetCardLimits(cardID string, limits []models.CardLimit) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]models.CardLimit, len(limits))
	copy(result, limits)
	s.cardLimits[cardID] = result
	return nil
}

// Card transaction operations

func (s *MemoryStorage) CreateCardTransaction(tx *models.CardTransaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tx.TransactionID == "" {
		return fmt.Errorf("transactionId is required")
	}

	if _, exists := s.cardTransactions[tx.TransactionID]; exists {
		return fmt.Errorf("card transaction already exists")
	}

	s.cardTransactions[tx.TransactionID] = tx
	return nil
}

func (s *MemoryStorage) GetCardTransaction(id string) (*models.CardTransaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tx, exists := s.cardTransactions[id]
	if !exists {
		return nil, fmt.Errorf("card transaction not found")
	}

	return tx, nil
}

func (s *MemoryStorage) AddCardTransactionIndex(cardID string, transactionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cardID == "" || transactionID == "" {
		return fmt.Errorf("cardID and transactionID are required")
	}

	s.cardTransactionsByCard[cardID] = append(s.cardTransactionsByCard[cardID], transactionID)
	return nil
}

func (s *MemoryStorage) GetCardTransactionIDs(cardID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := s.cardTransactionsByCard[cardID]
	if ids == nil {
		return []string{}, nil
	}

	result := make([]string, len(ids))
	copy(result, ids)
	return result, nil
}

func (s *MemoryStorage) StoreRawCardTransaction(txID string, data json.RawMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if txID == "" {
		return fmt.Errorf("txID is required")
	}
	cloned := make([]byte, len(data))
	copy(cloned, data)
	s.rawCardTransactions[txID] = cloned
	return nil
}

func (s *MemoryStorage) GetRawCardTransaction(txID string) (json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.rawCardTransactions[txID]
	if !ok {
		return nil, fmt.Errorf("raw card transaction not found")
	}
	cloned := make([]byte, len(data))
	copy(cloned, data)
	return cloned, nil
}

func (s *MemoryStorage) NextCardTransactionSeqID() (int, error) {
	return int(s.cardTxSeqID.Add(1)), nil
}

func (s *MemoryStorage) PeekCardTransactionSeqID() (int, error) {
	return int(s.cardTxSeqID.Load()) + 1, nil
}

// CreateWallet creates a new wallet
func (s *MemoryStorage) CreateWallet(wallet *models.Wallet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if wallet.Address == "" {
		return fmt.Errorf("address is required")
	}

	if _, exists := s.wallets[wallet.Address]; exists {
		return fmt.Errorf("wallet with address %s already exists", wallet.Address)
	}

	if wallet.CreatedAt.IsZero() {
		wallet.CreatedAt = time.Now()
	}

	s.wallets[wallet.Address] = wallet
	return nil
}

// GetWallet retrieves a wallet by address
func (s *MemoryStorage) GetWallet(address string) (*models.Wallet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wallet, exists := s.wallets[address]
	if !exists {
		return nil, fmt.Errorf("wallet not found")
	}

	return wallet, nil
}

// GetWalletsByUser retrieves all wallets for a user
func (s *MemoryStorage) GetWalletsByUser(userID string) ([]*models.Wallet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var wallets []*models.Wallet
	for _, wallet := range s.wallets {
		if wallet.UserID == userID {
			wallets = append(wallets, wallet)
		}
	}

	return wallets, nil
}

// CreateTransaction creates a new transaction
func (s *MemoryStorage) CreateTransaction(tx *models.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tx.ID == "" {
		tx.ID = utils.GenerateUUID()
	}

	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}

	s.transactions[tx.ID] = tx
	return nil
}

// GetTransaction retrieves a transaction by ID
func (s *MemoryStorage) GetTransaction(id string) (*models.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tx, exists := s.transactions[id]
	if !exists {
		return nil, fmt.Errorf("transaction not found")
	}

	return tx, nil
}

// UpdateTransactionStatus updates the status of an existing transaction
func (s *MemoryStorage) UpdateTransactionStatus(id string, status int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, exists := s.transactions[id]
	if !exists {
		return fmt.Errorf("transaction not found")
	}

	tx.Status = status
	s.transactions[id] = tx

	return nil
}

// GetBalance retrieves balance for a user and currency
func (s *MemoryStorage) GetBalance(userID, currency string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userBalances, exists := s.balances[userID]
	if !exists {
		return 0, nil
	}

	return userBalances[currency], nil
}

// AddBalance adds to a user's balance
func (s *MemoryStorage) AddBalance(userID, currency string, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.balances[userID] == nil {
		s.balances[userID] = make(map[string]float64)
	}

	s.balances[userID][currency] += amount
	return nil
}

// DeductBalance deducts from a user's balance
func (s *MemoryStorage) DeductBalance(userID, currency string, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.balances[userID] == nil {
		return fmt.Errorf("insufficient balance")
	}

	currentBalance := s.balances[userID][currency]
	if currentBalance < amount {
		return fmt.Errorf("insufficient balance: have %.2f, need %.2f", currentBalance, amount)
	}

	s.balances[userID][currency] -= amount
	return nil
}

// CreateThreeDSChallenge creates a new 3DS challenge
func (s *MemoryStorage) CreateThreeDSChallenge(challenge *models.ThreeDSChallenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if challenge.TransactionID == "" {
		return fmt.Errorf("transaction ID is required")
	}

	if challenge.CreatedAt.IsZero() {
		challenge.CreatedAt = time.Now()
	}

	s.threeDSChallenges[challenge.TransactionID] = challenge
	return nil
}

// GetThreeDSChallenge retrieves a 3DS challenge by transaction ID
func (s *MemoryStorage) GetThreeDSChallenge(txID string) (*models.ThreeDSChallenge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	challenge, exists := s.threeDSChallenges[txID]
	if !exists {
		return nil, fmt.Errorf("3DS challenge not found")
	}

	return challenge, nil
}

// GetPendingThreeDSChallenges retrieves all pending 3DS challenges for a user
func (s *MemoryStorage) GetPendingThreeDSChallenges(userID string) ([]*models.ThreeDSChallenge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pending []*models.ThreeDSChallenge
	now := time.Now()

	for _, challenge := range s.threeDSChallenges {
		if challenge.UserID == userID && challenge.Status == "pending" {
			// Check if not expired
			if challenge.Timeout.After(now) {
				pending = append(pending, challenge)
			}
		}
	}

	return pending, nil
}

// UpdateThreeDSChallenge updates a 3DS challenge
func (s *MemoryStorage) UpdateThreeDSChallenge(challenge *models.ThreeDSChallenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.threeDSChallenges[challenge.TransactionID]; !exists {
		return fmt.Errorf("3DS challenge not found")
	}

	s.threeDSChallenges[challenge.TransactionID] = challenge
	return nil
}

// Organization operations

// GetOrganization retrieves an organization by ID
func (s *MemoryStorage) GetOrganization(orgID string) (*models.Organization, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	org, exists := s.organizations[orgID]
	if !exists {
		return nil, fmt.Errorf("organization not found")
	}

	// Return copy to prevent external mutation
	orgCopy := *org
	return &orgCopy, nil
}

// CreateOrganization creates a new organization
func (s *MemoryStorage) CreateOrganization(org *models.Organization) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.organizations[org.ID]; exists {
		return fmt.Errorf("organization already exists")
	}

	// Store a defensive copy to prevent external mutation
	orgCopy := *org
	s.organizations[org.ID] = &orgCopy
	return nil
}

// UpdateOrganization updates an existing organization
func (s *MemoryStorage) UpdateOrganization(org *models.Organization) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.organizations[org.ID]; !exists {
		return fmt.Errorf("organization not found")
	}

	org.UpdatedAt = time.Now()
	// Store a defensive copy to prevent external mutation
	orgCopy := *org
	s.organizations[org.ID] = &orgCopy
	return nil
}

// ListUsers returns all users in storage.
func (s *MemoryStorage) ListUsers() ([]*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*models.User, 0, len(s.users))
	for _, u := range s.users {
		cp := *u
		users = append(users, &cp)
	}
	return users, nil
}

// ListTransactionsByUser returns all transactions belonging to the given user.
func (s *MemoryStorage) ListTransactionsByUser(userID string) ([]*models.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var txns []*models.Transaction
	for _, tx := range s.transactions {
		if tx.UserID == userID {
			cp := *tx
			txns = append(txns, &cp)
		}
	}
	return txns, nil
}

// GetAllBalances returns a map of currency → amount for all currencies that have
// a stored balance entry for the given user. Zero-balance entries are included.
func (s *MemoryStorage) GetAllBalances(userID string) (map[string]float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]float64)
	if balances, ok := s.balances[userID]; ok {
		for currency, amount := range balances {
			result[currency] = amount
		}
	}
	return result, nil
}
