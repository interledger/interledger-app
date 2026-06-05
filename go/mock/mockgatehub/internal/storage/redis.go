package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisStorage implements Storage using Redis
type RedisStorage struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient creates a standalone Redis client (for webhook queue in memory mode)
func NewRedisClient(redisURL string, db int) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	opt.DB = db

	client := redis.NewClient(opt)
	ctx := context.Background()

	// Ping to verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("created standalone redis client", zap.String("url", redisURL), zap.Int("db", db))
	return client, nil
}

// NewRedisStorage creates a new Redis storage instance
func NewRedisStorage(redisURL string, db int) (*RedisStorage, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	opt.DB = db

	client := redis.NewClient(opt)
	ctx := context.Background()

	// Ping to verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("connected to redis", zap.String("url", redisURL), zap.Int("db", db))

	return &RedisStorage{
		client: client,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis connection
func (s *RedisStorage) Close() error {
	return s.client.Close()
}

// GetClient returns the underlying Redis client (for webhook queue)
func (s *RedisStorage) GetClient() *redis.Client {
	return s.client
}

// User operations

func (s *RedisStorage) CreateUser(user *models.User) error {
	if user.Email == "" {
		return errors.New("email is required")
	}

	// Generate ID and timestamps similar to memory storage
	if user.ID == "" {
		user.ID = utils.GenerateUUID()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	// Check if user exists
	exists, err := s.client.Exists(s.ctx, s.userKey(user.ID)).Result()
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists > 0 {
		return errors.New("user already exists")
	}

	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Store user by ID
	if err := s.client.Set(s.ctx, s.userKey(user.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store user: %w", err)
	}

	// Store email → ID mapping
	if err := s.client.Set(s.ctx, s.emailKey(user.Email), user.ID, 0).Err(); err != nil {
		return fmt.Errorf("failed to store email mapping: %w", err)
	}

	return nil
}

func (s *RedisStorage) GetUser(id string) (*models.User, error) {
	data, err := s.client.Get(s.ctx, s.userKey(id)).Result()
	if err == redis.Nil {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var user models.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

func (s *RedisStorage) GetUserByEmail(email string) (*models.User, error) {
	// Get user ID from email mapping
	userID, err := s.client.Get(s.ctx, s.emailKey(email)).Result()
	if err == redis.Nil {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return s.GetUser(userID)
}

func (s *RedisStorage) UpdateUser(user *models.User) error {
	if user.ID == "" {
		return errors.New("user ID is required")
	}

	// Check if user exists
	exists, err := s.client.Exists(s.ctx, s.userKey(user.ID)).Result()
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists == 0 {
		return errors.New("user not found")
	}

	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	if err := s.client.Set(s.ctx, s.userKey(user.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Card customer operations

func (s *RedisStorage) CreateCustomer(customer *models.Customer) error {
	if customer.ID == nil || *customer.ID == "" {
		id := utils.GenerateUUID()
		customer.ID = &id
	}
	if customer.SourceID == "" {
		return errors.New("sourceId is required")
	}
	if customer.CreatedAt.IsZero() {
		customer.CreatedAt = time.Now()
	}

	data, err := json.Marshal(customer)
	if err != nil {
		return fmt.Errorf("failed to marshal customer: %w", err)
	}

	if err := s.client.Set(s.ctx, s.customerKey(*customer.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store customer: %w", err)
	}

	if err := s.client.Set(s.ctx, s.customerSourceKey(customer.SourceID), *customer.ID, 0).Err(); err != nil {
		return fmt.Errorf("failed to store customer source mapping: %w", err)
	}

	return nil
}

func (s *RedisStorage) GetCustomer(id string) (*models.Customer, error) {
	data, err := s.client.Get(s.ctx, s.customerKey(id)).Result()
	if err == redis.Nil {
		return nil, errors.New("customer not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	var customer models.Customer
	if err := json.Unmarshal([]byte(data), &customer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal customer: %w", err)
	}

	return &customer, nil
}

func (s *RedisStorage) GetCustomerBySourceID(sourceID string) (*models.Customer, error) {
	customerID, err := s.client.Get(s.ctx, s.customerSourceKey(sourceID)).Result()
	if err == redis.Nil {
		return nil, errors.New("customer not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer by sourceId: %w", err)
	}

	return s.GetCustomer(customerID)
}

func (s *RedisStorage) UpdateCustomer(customer *models.Customer) error {
	if customer.ID == nil || *customer.ID == "" {
		return errors.New("customer ID is required")
	}

	data, err := json.Marshal(customer)
	if err != nil {
		return fmt.Errorf("failed to marshal customer: %w", err)
	}

	if err := s.client.Set(s.ctx, s.customerKey(*customer.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	if customer.SourceID != "" {
		if err := s.client.Set(s.ctx, s.customerSourceKey(customer.SourceID), *customer.ID, 0).Err(); err != nil {
			return fmt.Errorf("failed to update customer source mapping: %w", err)
		}
	}

	return nil
}

// Card account operations

func (s *RedisStorage) CreateAccount(account *models.Account) error {
	if account.ID == nil || *account.ID == "" {
		id := utils.GenerateUUID()
		account.ID = &id
	}
	if account.CreatedAt.IsZero() {
		account.CreatedAt = time.Now()
	}

	data, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	if err := s.client.Set(s.ctx, s.accountKey(*account.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store account: %w", err)
	}

	if account.CustomerID != nil {
		if err := s.client.SAdd(s.ctx, s.customerAccountsKey(*account.CustomerID), *account.ID).Err(); err != nil {
			return fmt.Errorf("failed to add account to customer set: %w", err)
		}
	}

	return nil
}

func (s *RedisStorage) GetAccount(id string) (*models.Account, error) {
	data, err := s.client.Get(s.ctx, s.accountKey(id)).Result()
	if err == redis.Nil {
		return nil, errors.New("account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	var account models.Account
	if err := json.Unmarshal([]byte(data), &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	return &account, nil
}

func (s *RedisStorage) UpdateAccount(account *models.Account) error {
	if account.ID == nil || *account.ID == "" {
		return errors.New("account ID is required")
	}

	data, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	if err := s.client.Set(s.ctx, s.accountKey(*account.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// Card delivery address operations

func (s *RedisStorage) CreateCustomerAddress(customerID string, address *models.CustomerDeliveryAddress) error {
	if customerID == "" {
		return errors.New("customer ID is required")
	}
	if address == nil {
		return errors.New("address is required")
	}
	if address.ID == "" {
		address.ID = utils.GenerateUUID()
	}

	data, err := json.Marshal(address)
	if err != nil {
		return fmt.Errorf("failed to marshal address: %w", err)
	}

	if err := s.client.Set(s.ctx, s.customerAddressKey(address.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store address: %w", err)
	}

	if err := s.client.SAdd(s.ctx, s.customerAddressesKey(customerID), address.ID).Err(); err != nil {
		return fmt.Errorf("failed to add address to customer set: %w", err)
	}

	return nil
}

func (s *RedisStorage) GetCustomerAddresses(customerID string) ([]*models.CustomerDeliveryAddress, error) {
	ids, err := s.client.SMembers(s.ctx, s.customerAddressesKey(customerID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get customer addresses: %w", err)
	}

	addresses := make([]*models.CustomerDeliveryAddress, 0, len(ids))
	for _, id := range ids {
		data, err := s.client.Get(s.ctx, s.customerAddressKey(id)).Result()
		if err != nil {
			continue
		}
		var addr models.CustomerDeliveryAddress
		if err := json.Unmarshal([]byte(data), &addr); err != nil {
			continue
		}
		addresses = append(addresses, &addr)
	}

	return addresses, nil
}

// Card operations

func (s *RedisStorage) CreateCard(card *models.Card) error {
	if card.ID == "" {
		card.ID = utils.GenerateUUID()
	}
	if card.CreatedAt.IsZero() {
		card.CreatedAt = time.Now()
	}

	data, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed to marshal card: %w", err)
	}

	if err := s.client.Set(s.ctx, s.cardKey(card.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store card: %w", err)
	}

	if card.CustomerID != "" {
		if err := s.client.SAdd(s.ctx, s.customerCardsKey(card.CustomerID), card.ID).Err(); err != nil {
			return fmt.Errorf("failed to add card to customer set: %w", err)
		}
	}
	if card.AccountID != "" {
		if err := s.client.SAdd(s.ctx, s.accountCardsKey(card.AccountID), card.ID).Err(); err != nil {
			return fmt.Errorf("failed to add card to account set: %w", err)
		}
	}

	return nil
}

func (s *RedisStorage) GetCard(id string) (*models.Card, error) {
	data, err := s.client.Get(s.ctx, s.cardKey(id)).Result()
	if err == redis.Nil {
		return nil, errors.New("card not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get card: %w", err)
	}

	var card models.Card
	if err := json.Unmarshal([]byte(data), &card); err != nil {
		return nil, fmt.Errorf("failed to unmarshal card: %w", err)
	}

	return &card, nil
}

func (s *RedisStorage) UpdateCard(card *models.Card) error {
	if card.ID == "" {
		return errors.New("card ID is required")
	}

	data, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed to marshal card: %w", err)
	}

	if err := s.client.Set(s.ctx, s.cardKey(card.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	return nil
}

func (s *RedisStorage) GetCardsByCustomer(customerID string) ([]*models.Card, error) {
	cardIDs, err := s.client.SMembers(s.ctx, s.customerCardsKey(customerID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get customer cards: %w", err)
	}

	var cards []*models.Card
	for _, cardID := range cardIDs {
		card, err := s.GetCard(cardID)
		if err == nil {
			cards = append(cards, card)
		}
	}

	return cards, nil
}

func (s *RedisStorage) GetCardsByAccount(accountID string) ([]*models.Card, error) {
	cardIDs, err := s.client.SMembers(s.ctx, s.accountCardsKey(accountID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get account cards: %w", err)
	}

	var cards []*models.Card
	for _, cardID := range cardIDs {
		card, err := s.GetCard(cardID)
		if err == nil {
			cards = append(cards, card)
		}
	}

	return cards, nil
}

func (s *RedisStorage) GetCardLimits(cardID string) ([]models.CardLimit, error) {
	data, err := s.client.Get(s.ctx, s.cardLimitsKey(cardID)).Result()
	if err == redis.Nil {
		return []models.CardLimit{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get card limits: %w", err)
	}

	var limits []models.CardLimit
	if err := json.Unmarshal([]byte(data), &limits); err != nil {
		return nil, fmt.Errorf("failed to unmarshal card limits: %w", err)
	}

	return limits, nil
}

func (s *RedisStorage) SetCardLimits(cardID string, limits []models.CardLimit) error {
	data, err := json.Marshal(limits)
	if err != nil {
		return fmt.Errorf("failed to marshal card limits: %w", err)
	}

	if err := s.client.Set(s.ctx, s.cardLimitsKey(cardID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store card limits: %w", err)
	}

	return nil
}

// Card transaction operations

func (s *RedisStorage) CreateCardTransaction(tx *models.CardTransaction) error {
	if tx.TransactionID == "" {
		return errors.New("transactionId is required")
	}

	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal card transaction: %w", err)
	}

	if err := s.client.Set(s.ctx, s.cardTransactionKey(tx.TransactionID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store card transaction: %w", err)
	}

	return nil
}

func (s *RedisStorage) GetCardTransaction(id string) (*models.CardTransaction, error) {
	data, err := s.client.Get(s.ctx, s.cardTransactionKey(id)).Result()
	if err == redis.Nil {
		return nil, errors.New("card transaction not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get card transaction: %w", err)
	}

	var tx models.CardTransaction
	if err := json.Unmarshal([]byte(data), &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal card transaction: %w", err)
	}

	return &tx, nil
}

func (s *RedisStorage) UpdateCardTransactionStatus(txID string, status string) error {
	tx, err := s.GetCardTransaction(txID)
	if err != nil {
		return err
	}
	tx.TxStatus = &status
	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal card transaction: %w", err)
	}
	if err := s.client.Set(s.ctx, s.cardTransactionKey(txID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to update card transaction: %w", err)
	}

	if rawData, err := s.GetRawCardTransaction(txID); err == nil {
		var m map[string]interface{}
		if err := json.Unmarshal(rawData, &m); err == nil {
			m["txStatus"] = status
			if updated, err := json.Marshal(m); err == nil {
				_ = s.StoreRawCardTransaction(txID, updated)
			}
		}
	}

	return nil
}

func (s *RedisStorage) AddCardTransactionIndex(cardID string, transactionID string) error {
	if cardID == "" || transactionID == "" {
		return errors.New("cardID and transactionID are required")
	}

	if err := s.client.LPush(s.ctx, s.cardTransactionsKey(cardID), transactionID).Err(); err != nil {
		return fmt.Errorf("failed to index card transaction: %w", err)
	}

	return nil
}

func (s *RedisStorage) GetCardTransactionIDs(cardID string) ([]string, error) {
	ids, err := s.client.LRange(s.ctx, s.cardTransactionsKey(cardID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get card transaction IDs: %w", err)
	}

	return ids, nil
}

func (s *RedisStorage) StoreRawCardTransaction(txID string, data json.RawMessage) error {
	if txID == "" {
		return errors.New("txID is required")
	}
	if err := s.client.Set(s.ctx, s.cardTransactionRawKey(txID), []byte(data), 0).Err(); err != nil {
		return fmt.Errorf("failed to store raw card transaction: %w", err)
	}
	return nil
}

func (s *RedisStorage) GetRawCardTransaction(txID string) (json.RawMessage, error) {
	data, err := s.client.Get(s.ctx, s.cardTransactionRawKey(txID)).Bytes()
	if err == redis.Nil {
		return nil, errors.New("raw card transaction not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get raw card transaction: %w", err)
	}
	return json.RawMessage(data), nil
}

func (s *RedisStorage) NextCardTransactionSeqID() (int, error) {
	n, err := s.client.Incr(s.ctx, "card_tx_seq").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment card transaction seq: %w", err)
	}
	return int(n), nil
}

func (s *RedisStorage) PeekCardTransactionSeqID() (int, error) {
	val, err := s.client.Get(s.ctx, "card_tx_seq").Int64()
	if err == redis.Nil {
		return 1, nil // next would be 1
	}
	if err != nil {
		return 0, fmt.Errorf("failed to peek card transaction seq: %w", err)
	}
	return int(val) + 1, nil // return what the next call to Next would return
}

// Wallet operations

func (s *RedisStorage) CreateWallet(wallet *models.Wallet) error {
	if wallet.Address == "" {
		return errors.New("wallet address is required")
	}

	wallet.CreatedAt = time.Now()

	data, err := json.Marshal(wallet)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet: %w", err)
	}

	if err := s.client.Set(s.ctx, s.walletKey(wallet.Address), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store wallet: %w", err)
	}

	// Add to user's wallet list
	if err := s.client.SAdd(s.ctx, s.userWalletsKey(wallet.UserID), wallet.Address).Err(); err != nil {
		return fmt.Errorf("failed to add wallet to user list: %w", err)
	}

	return nil
}

func (s *RedisStorage) GetWallet(address string) (*models.Wallet, error) {
	data, err := s.client.Get(s.ctx, s.walletKey(address)).Result()
	if err == redis.Nil {
		return nil, errors.New("wallet not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	var wallet models.Wallet
	if err := json.Unmarshal([]byte(data), &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet: %w", err)
	}

	return &wallet, nil
}

func (s *RedisStorage) GetWalletsByUser(userID string) ([]*models.Wallet, error) {
	addresses, err := s.client.SMembers(s.ctx, s.userWalletsKey(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user wallets: %w", err)
	}

	var wallets []*models.Wallet
	for _, addr := range addresses {
		wallet, err := s.GetWallet(addr)
		if err == nil {
			wallets = append(wallets, wallet)
		}
	}

	return wallets, nil
}

// Transaction operations

func (s *RedisStorage) CreateTransaction(tx *models.Transaction) error {
	if tx.ID == "" {
		tx.ID = utils.GenerateUUID()
	}

	tx.CreatedAt = time.Now()

	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	if err := s.client.Set(s.ctx, s.txKey(tx.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store transaction: %w", err)
	}

	// Maintain per-user transaction index for admin UI enumeration
	if tx.UserID != "" {
		if err := s.client.LPush(s.ctx, s.userTransactionsKey(tx.UserID), tx.ID).Err(); err != nil {
			return fmt.Errorf("failed to index transaction for user: %w", err)
		}
	}

	return nil
}

func (s *RedisStorage) ListTransactionsByUser(userID string) ([]*models.Transaction, error) {
	ids, err := s.client.LRange(s.ctx, s.userTransactionsKey(userID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list transaction IDs for user: %w", err)
	}

	transactions := make([]*models.Transaction, 0, len(ids))
	for _, id := range ids {
		tx, err := s.GetTransaction(id)
		if err != nil {
			continue
		}
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

func (s *RedisStorage) GetTransaction(id string) (*models.Transaction, error) {
	data, err := s.client.Get(s.ctx, s.txKey(id)).Result()
	if err == redis.Nil {
		return nil, errors.New("transaction not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	var tx models.Transaction
	if err := json.Unmarshal([]byte(data), &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &tx, nil
}

func (s *RedisStorage) UpdateTransactionStatus(id string, status int) error {
	tx, err := s.GetTransaction(id)
	if err != nil {
		return err
	}

	tx.Status = status

	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	if err := s.client.Set(s.ctx, s.txKey(id), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// Balance operations

func (s *RedisStorage) GetBalance(userID, currency string) (float64, error) {
	val, err := s.client.Get(s.ctx, s.balanceKey(userID, currency)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	balance, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse balance: %w", err)
	}

	return balance, nil
}

func (s *RedisStorage) AddBalance(userID, currency string, amount float64) error {
	if _, err := s.client.IncrByFloat(s.ctx, s.balanceKey(userID, currency), amount).Result(); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}

func (s *RedisStorage) DeductBalance(userID, currency string, amount float64) error {
	return s.AddBalance(userID, currency, -amount)
}

// 3DS Challenges

func (s *RedisStorage) CreateThreeDSChallenge(challenge *models.ThreeDSChallenge) error {
	if challenge.TransactionID == "" {
		return errors.New("transaction ID is required")
	}

	if challenge.CreatedAt.IsZero() {
		challenge.CreatedAt = time.Now()
	}

	data, err := json.Marshal(challenge)
	if err != nil {
		return fmt.Errorf("failed to marshal 3DS challenge: %w", err)
	}

	// Calculate TTL based on timeout
	ttl := time.Until(challenge.Timeout)
	if ttl < 0 {
		return errors.New("3DS challenge timeout is in the past")
	}
	// Add a buffer to TTL to ensure we can mark it as expired
	ttl = ttl + 10*time.Minute

	// Store the challenge
	if err := s.client.Set(s.ctx, s.threeDSChallengeKey(challenge.TransactionID), data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store 3DS challenge: %w", err)
	}

	// Add to user's pending challenges set (sorted by timeout)
	score := float64(challenge.Timeout.Unix())
	if err := s.client.ZAdd(s.ctx, s.userThreeDSChallengesKey(challenge.UserID), redis.Z{
		Score:  score,
		Member: challenge.TransactionID,
	}).Err(); err != nil {
		return fmt.Errorf("failed to index 3DS challenge: %w", err)
	}

	return nil
}

func (s *RedisStorage) GetThreeDSChallenge(txID string) (*models.ThreeDSChallenge, error) {
	data, err := s.client.Get(s.ctx, s.threeDSChallengeKey(txID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("3DS challenge not found")
		}
		return nil, fmt.Errorf("failed to get 3DS challenge: %w", err)
	}

	var challenge models.ThreeDSChallenge
	if err := json.Unmarshal([]byte(data), &challenge); err != nil {
		return nil, fmt.Errorf("failed to unmarshal 3DS challenge: %w", err)
	}

	return &challenge, nil
}

func (s *RedisStorage) GetPendingThreeDSChallenges(userID string) ([]*models.ThreeDSChallenge, error) {
	now := time.Now().Unix()

	// Get all transaction IDs from the sorted set that haven't timed out yet
	txIDs, err := s.client.ZRangeByScore(s.ctx, s.userThreeDSChallengesKey(userID), &redis.ZRangeBy{
		Min: strconv.FormatInt(now, 10),
		Max: "+inf",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending 3DS challenges: %w", err)
	}

	var pending []*models.ThreeDSChallenge
	for _, txID := range txIDs {
		challenge, err := s.GetThreeDSChallenge(txID)
		if err != nil {
			// Challenge may have expired or been deleted
			continue
		}

		// Only return if still pending
		if challenge.Status == "pending" {
			pending = append(pending, challenge)
		}
	}

	return pending, nil
}

func (s *RedisStorage) UpdateThreeDSChallenge(challenge *models.ThreeDSChallenge) error {
	// Check if exists
	exists, err := s.client.Exists(s.ctx, s.threeDSChallengeKey(challenge.TransactionID)).Result()
	if err != nil {
		return fmt.Errorf("failed to check 3DS challenge existence: %w", err)
	}
	if exists == 0 {
		return errors.New("3DS challenge not found")
	}

	data, err := json.Marshal(challenge)
	if err != nil {
		return fmt.Errorf("failed to marshal 3DS challenge: %w", err)
	}

	// Maintain existing TTL
	ttl, err := s.client.TTL(s.ctx, s.threeDSChallengeKey(challenge.TransactionID)).Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL: %w", err)
	}

	if err := s.client.Set(s.ctx, s.threeDSChallengeKey(challenge.TransactionID), data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update 3DS challenge: %w", err)
	}

	// If status changed from pending, remove from user's pending set
	if challenge.Status != "pending" {
		if err := s.client.ZRem(s.ctx, s.userThreeDSChallengesKey(challenge.UserID), challenge.TransactionID).Err(); err != nil {
			return fmt.Errorf("failed to remove from pending set: %w", err)
		}
	}

	return nil
}

// Organization operations

func (s *RedisStorage) GetOrganization(orgID string) (*models.Organization, error) {
	key := s.organizationKey(orgID)
	data, err := s.client.Get(s.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("organization not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	var org models.Organization
	if err := json.Unmarshal(data, &org); err != nil {
		return nil, fmt.Errorf("failed to unmarshal organization: %w", err)
	}
	return &org, nil
}

func (s *RedisStorage) CreateOrganization(org *models.Organization) error {
	key := s.organizationKey(org.ID)

	exists, err := s.client.Exists(s.ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check organization existence: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("organization already exists")
	}

	data, err := json.Marshal(org)
	if err != nil {
		return fmt.Errorf("failed to marshal organization: %w", err)
	}

	return s.client.Set(s.ctx, key, data, 0).Err()
}

func (s *RedisStorage) UpdateOrganization(org *models.Organization) error {
	key := s.organizationKey(org.ID)

	exists, err := s.client.Exists(s.ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check organization existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("organization not found")
	}

	org.UpdatedAt = time.Now()
	data, err := json.Marshal(org)
	if err != nil {
		return fmt.Errorf("failed to marshal organization: %w", err)
	}

	return s.client.Set(s.ctx, key, data, 0).Err()
}

// Key helpers

func (s *RedisStorage) userKey(id string) string {
	return fmt.Sprintf("user:%s", id)
}

func (s *RedisStorage) emailKey(email string) string {
	return fmt.Sprintf("email:%s", email)
}

func (s *RedisStorage) customerKey(id string) string {
	return fmt.Sprintf("customer:%s", id)
}

func (s *RedisStorage) customerSourceKey(sourceID string) string {
	return fmt.Sprintf("customer:source:%s", sourceID)
}

func (s *RedisStorage) customerAccountsKey(customerID string) string {
	return fmt.Sprintf("customer:%s:accounts", customerID)
}

func (s *RedisStorage) accountKey(id string) string {
	return fmt.Sprintf("account:%s", id)
}

func (s *RedisStorage) customerCardsKey(customerID string) string {
	return fmt.Sprintf("customer:%s:cards", customerID)
}

func (s *RedisStorage) accountCardsKey(accountID string) string {
	return fmt.Sprintf("account:%s:cards", accountID)
}

func (s *RedisStorage) cardKey(id string) string {
	return fmt.Sprintf("card:%s", id)
}

func (s *RedisStorage) cardLimitsKey(cardID string) string {
	return fmt.Sprintf("card:%s:limits", cardID)
}

func (s *RedisStorage) customerAddressesKey(customerID string) string {
	return fmt.Sprintf("customer:%s:addresses", customerID)
}

func (s *RedisStorage) customerAddressKey(addressID string) string {
	return fmt.Sprintf("customer:address:%s", addressID)
}

func (s *RedisStorage) cardTransactionKey(id string) string {
	return fmt.Sprintf("cardtx:%s", id)
}

func (s *RedisStorage) cardTransactionsKey(cardID string) string {
	return fmt.Sprintf("card:%s:transactions", cardID)
}

func (s *RedisStorage) cardTransactionRawKey(id string) string {
	return fmt.Sprintf("cardtx_raw:%s", id)
}

func (s *RedisStorage) walletKey(address string) string {
	return fmt.Sprintf("wallet:%s", address)
}

func (s *RedisStorage) userWalletsKey(userID string) string {
	return fmt.Sprintf("user:%s:wallets", userID)
}

func (s *RedisStorage) txKey(id string) string {
	return fmt.Sprintf("tx:%s", id)
}

func (s *RedisStorage) userTransactionsKey(userID string) string {
	return fmt.Sprintf("user_txs:%s", userID)
}

func (s *RedisStorage) balanceKey(userID, currency string) string {
	return fmt.Sprintf("balance:%s:%s", userID, currency)
}

func (s *RedisStorage) threeDSChallengeKey(txID string) string {
	return fmt.Sprintf("3ds:challenge:%s", txID)
}

func (s *RedisStorage) userThreeDSChallengesKey(userID string) string {
	return fmt.Sprintf("user:%s:3ds:challenges", userID)
}

func (s *RedisStorage) organizationKey(orgID string) string {
	return fmt.Sprintf("organization:%s", orgID)
}

func (s *RedisStorage) userTransactionsKey(userID string) string {
	return fmt.Sprintf("user:%s:transactions", userID)
}

// ListUsers returns all users stored in Redis by scanning for keys matching
// the pattern "user:*" and filtering out composite keys (e.g. "user:{id}:wallets").
func (s *RedisStorage) ListUsers() ([]*models.User, error) {
	var users []*models.User
	var cursor uint64

	for {
		keys, nextCursor, err := s.client.Scan(s.ctx, cursor, "user:*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan user keys: %w", err)
		}

		for _, key := range keys {
			// Skip composite keys like "user:{id}:wallets", "user:{id}:3ds:challenges", etc.
			// A plain user key has exactly one colon: "user:{id}"
			suffix := key[len("user:"):]
			if strings.Contains(suffix, ":") {
				continue
			}

			data, err := s.client.Get(s.ctx, key).Result()
			if err != nil {
				continue
			}

			var user models.User
			if err := json.Unmarshal([]byte(data), &user); err != nil {
				continue
			}
			users = append(users, &user)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return users, nil
}

// ListTransactionsByUser returns all transactions for the given user by reading
// from the per-user transaction index list "user:{id}:transactions".
func (s *RedisStorage) ListTransactionsByUser(userID string) ([]*models.Transaction, error) {
	ids, err := s.client.LRange(s.ctx, s.userTransactionsKey(userID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user transaction index: %w", err)
	}

	txns := make([]*models.Transaction, 0, len(ids))
	for _, id := range ids {
		tx, err := s.GetTransaction(id)
		if err != nil {
			continue
		}
		txns = append(txns, tx)
	}
	return txns, nil
}

// GetAllBalances returns a map of currency → amount for the given user by
// scanning keys matching "balance:{userID}:*".
func (s *RedisStorage) GetAllBalances(userID string) (map[string]float64, error) {
	result := make(map[string]float64)
	pattern := fmt.Sprintf("balance:%s:*", userID)
	prefix := fmt.Sprintf("balance:%s:", userID)
	var cursor uint64

	for {
		keys, nextCursor, err := s.client.Scan(s.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan balance keys: %w", err)
		}

		for _, key := range keys {
			currency := strings.TrimPrefix(key, prefix)
			if currency == "" {
				continue
			}

			val, err := s.client.Get(s.ctx, key).Result()
			if err != nil {
				continue
			}

			amount, err := strconv.ParseFloat(val, 64)
			if err != nil {
				continue
			}
			result[currency] = amount
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return result, nil
}
