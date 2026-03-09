package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.com/fynbos/mock/mockxago/internal/models"
)

// RedisStorage is a Redis-backed implementation of the Storage interface
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a new Redis storage client
func NewRedisStorage(addr string, db int) (Storage, error) {
	// Parse redis:// URL if provided, otherwise use as direct host:port
	var redisAddr string
	if u, err := url.Parse(addr); err == nil && u.Scheme == "redis" {
		// Extract host and port from redis://host:port URL
		redisAddr = u.Host
		if redisAddr == "" {
			redisAddr = addr // Fallback to original if parsing fails
		}
	} else {
		// Use directly as host:port
		redisAddr = addr
	}

	client := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		DB:           db,
		Password:     "",
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{client: client}, nil
}

// Redis key patterns
func tokenKey(tokenValue string) string {
	return fmt.Sprintf("token:%s", tokenValue)
}

func tokenAccountKey(tokenValue string) string {
	return fmt.Sprintf("token:account:%s", tokenValue)
}

func subAccountKey(accountID string) string {
	return fmt.Sprintf("subaccount:%s", accountID)
}

func subAccountWalletKey(walletID string) string {
	return fmt.Sprintf("subaccount:wallet:%s", walletID)
}

func beneficiaryKey(beneficiaryID string) string {
	return fmt.Sprintf("beneficiary:%s", beneficiaryID)
}

func beneficiariesWalletKey(walletID string) string {
	return fmt.Sprintf("beneficiaries:wallet:%s", walletID)
}

func beneficiariesAccountKey(accountID string) string {
	return fmt.Sprintf("beneficiaries:account:%s", accountID)
}

func transactionKey(transactionID string) string {
	return fmt.Sprintf("transaction:%s", transactionID)
}

func idempotencyKey(key string) string {
	return fmt.Sprintf("idempotency:%s", key)
}

func transactionsAccountKey(accountID string) string {
	return fmt.Sprintf("transactions:account:%s", accountID)
}

func balanceKey(walletID, currency string) string {
	return fmt.Sprintf("balance:%s:%s", walletID, currency)
}

func depositKey(depositID string) string {
	return fmt.Sprintf("deposit:%s", depositID)
}

func depositReferenceKey(reference string) string {
	return fmt.Sprintf("deposit:ref:%s", reference)
}

func depositsKey() string {
	return "deposits:all"
}

func jobKey(jobID string) string {
	return fmt.Sprintf("job:%s", jobID)
}

func jobsReadyKey() string {
	return "jobs:ready"
}

// Token operations

func (r *RedisStorage) SaveAccessToken(ctx context.Context, token *models.AccessToken) error {
	token.CreatedAt = time.Now()
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	ttl := time.Until(token.ExpiresAt)
	if ttl < 0 {
		ttl = 0
	}

	return r.client.Set(ctx, tokenKey(token.Token), data, ttl).Err()
}

func (r *RedisStorage) GetAccessToken(ctx context.Context, tokenValue string) (*models.AccessToken, error) {
	data, err := r.client.Get(ctx, tokenKey(tokenValue)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	var token models.AccessToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	if token.IsExpired() {
		return nil, ErrTokenExpired
	}

	return &token, nil
}

func (r *RedisStorage) InvalidateAccessToken(ctx context.Context, tokenValue string) error {
	return r.client.Del(ctx, tokenKey(tokenValue)).Err()
}

func (r *RedisStorage) SaveTokenAccount(ctx context.Context, tokenValue string, accountID string) error {
	if tokenValue == "" || accountID == "" {
		return nil
	}
	return r.client.Set(ctx, tokenAccountKey(tokenValue), accountID, 1*time.Hour).Err()
}

func (r *RedisStorage) GetAccountIDByToken(ctx context.Context, tokenValue string) (string, error) {
	accountID, err := r.client.Get(ctx, tokenAccountKey(tokenValue)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrTokenNotFound
		}
		return "", err
	}
	return accountID, nil
}

// Sub-account operations

func (r *RedisStorage) SaveSubAccount(ctx context.Context, account *models.SubAccount) error {
	data, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal sub-account: %w", err)
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, subAccountKey(account.AccountID), data, 0)
	pipe.Set(ctx, subAccountWalletKey(account.WalletID), account.AccountID, 0)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) GetSubAccount(ctx context.Context, accountID string) (*models.SubAccount, error) {
	data, err := r.client.Get(ctx, subAccountKey(accountID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrSubAccountNotFound
		}
		return nil, err
	}

	var account models.SubAccount
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sub-account: %w", err)
	}

	return &account, nil
}

func (r *RedisStorage) GetSubAccountByWalletID(ctx context.Context, walletID string) (*models.SubAccount, error) {
	accountID, err := r.client.Get(ctx, subAccountWalletKey(walletID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrSubAccountNotFound
		}
		return nil, err
	}

	return r.GetSubAccount(ctx, accountID)
}

func (r *RedisStorage) UpdateSubAccount(ctx context.Context, account *models.SubAccount) error {
	return r.SaveSubAccount(ctx, account)
}

// Beneficiary operations

func (r *RedisStorage) SaveBeneficiary(ctx context.Context, beneficiary *models.Beneficiary) error {
	data, err := json.Marshal(beneficiary)
	if err != nil {
		return fmt.Errorf("failed to marshal beneficiary: %w", err)
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, beneficiaryKey(beneficiary.ID), data, 0)
	pipe.RPush(ctx, beneficiariesWalletKey(beneficiary.AccountID), beneficiary.ID)
	pipe.RPush(ctx, beneficiariesAccountKey(beneficiary.AccountID), beneficiary.ID)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) GetBeneficiary(ctx context.Context, beneficiaryID string) (*models.Beneficiary, error) {
	data, err := r.client.Get(ctx, beneficiaryKey(beneficiaryID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrBeneficiaryNotFound
		}
		return nil, err
	}

	var beneficiary models.Beneficiary
	if err := json.Unmarshal(data, &beneficiary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal beneficiary: %w", err)
	}

	return &beneficiary, nil
}

func (r *RedisStorage) ListBeneficiariesByWallet(ctx context.Context, walletID string, limit int, offset int) ([]*models.Beneficiary, int, error) {
	// Get total count
	total, err := r.client.LLen(ctx, beneficiariesWalletKey(walletID)).Result()
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*models.Beneficiary{}, 0, nil
	}

	// Get paginated IDs
	end := int64(offset + limit - 1)
	if end >= total {
		end = total - 1
	}

	ids, err := r.client.LRange(ctx, beneficiariesWalletKey(walletID), int64(offset), end).Result()
	if err != nil {
		return nil, 0, err
	}

	// Fetch beneficiaries
	beneficiaries := make([]*models.Beneficiary, 0, len(ids))
	for _, id := range ids {
		ben, err := r.GetBeneficiary(ctx, id)
		if err != nil {
			// Skip beneficiaries that were deleted
			continue
		}
		beneficiaries = append(beneficiaries, ben)
	}

	return beneficiaries, int(total), nil
}

func (r *RedisStorage) ListBeneficiariesByAccountID(ctx context.Context, accountID string, limit int, offset int) ([]*models.Beneficiary, int, error) {
	// Get total count
	total, err := r.client.LLen(ctx, beneficiariesAccountKey(accountID)).Result()
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*models.Beneficiary{}, 0, nil
	}

	// Get paginated IDs
	end := int64(offset + limit - 1)
	if end >= total {
		end = total - 1
	}

	ids, err := r.client.LRange(ctx, beneficiariesAccountKey(accountID), int64(offset), end).Result()
	if err != nil {
		return nil, 0, err
	}

	// Fetch beneficiaries
	beneficiaries := make([]*models.Beneficiary, 0, len(ids))
	for _, id := range ids {
		ben, err := r.GetBeneficiary(ctx, id)
		if err != nil {
			// Skip beneficiaries that were deleted
			continue
		}
		beneficiaries = append(beneficiaries, ben)
	}

	return beneficiaries, int(total), nil
}

func (r *RedisStorage) UpdateBeneficiaryStatus(ctx context.Context, beneficiaryID string, status string) error {
	ben, err := r.GetBeneficiary(ctx, beneficiaryID)
	if err != nil {
		return err
	}

	ben.Status = status
	return r.SaveBeneficiary(ctx, ben)
}

// Transaction operations

func (r *RedisStorage) SaveTransaction(ctx context.Context, transaction *models.Transaction) error {
	data, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, transactionKey(transaction.ID), data, 0)
	pipe.RPush(ctx, transactionsAccountKey(transaction.AccountID), transaction.ID)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) GetTransaction(ctx context.Context, transactionID string) (*models.Transaction, error) {
	data, err := r.client.Get(ctx, transactionKey(transactionID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	var transaction models.Transaction
	if err := json.Unmarshal(data, &transaction); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &transaction, nil
}

func (r *RedisStorage) GetTransactionByIdempotencyKey(ctx context.Context, key string) (*models.Transaction, error) {
	txID, err := r.client.Get(ctx, idempotencyKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	return r.GetTransaction(ctx, txID)
}

func (r *RedisStorage) SaveIdempotencyKey(ctx context.Context, key string, transactionID string) error {
	return r.client.Set(ctx, idempotencyKey(key), transactionID, 24*time.Hour).Err()
}

func (r *RedisStorage) ListTransactionsByAccount(ctx context.Context, accountID string, limit int, offset int) ([]*models.Transaction, int, error) {
	// Get total count
	total, err := r.client.LLen(ctx, transactionsAccountKey(accountID)).Result()
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*models.Transaction{}, 0, nil
	}

	// Get paginated IDs
	end := int64(offset + limit - 1)
	if end >= total {
		end = total - 1
	}

	ids, err := r.client.LRange(ctx, transactionsAccountKey(accountID), int64(offset), end).Result()
	if err != nil {
		return nil, 0, err
	}

	// Fetch transactions
	transactions := make([]*models.Transaction, 0, len(ids))
	for _, id := range ids {
		tx, err := r.GetTransaction(ctx, id)
		if err != nil {
			continue
		}
		transactions = append(transactions, tx)
	}

	return transactions, int(total), nil
}

func (r *RedisStorage) UpdateTransactionStatus(ctx context.Context, transactionID string, status string) error {
	tx, err := r.GetTransaction(ctx, transactionID)
	if err != nil {
		return err
	}

	tx.Status = status
	if status == "completed" || status == "settled" {
		now := time.Now()
		tx.SettledAt = &now
	}

	return r.SaveTransaction(ctx, tx)
}

// Balance operations

func (r *RedisStorage) GetBalance(ctx context.Context, walletID string, currency string) (available float64, reserved float64, err error) {
	availKey := balanceKey(walletID, currency) + ":available"
	reservedKey := balanceKey(walletID, currency) + ":reserved"

	pipe := r.client.Pipeline()
	availCmd := pipe.Get(ctx, availKey)
	reservedCmd := pipe.Get(ctx, reservedKey)
	_, err = pipe.Exec(ctx)

	// Parse available (default to 0 if not found)
	availStr, err := availCmd.Result()
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}
	if availStr != "" {
		available, _ = strconv.ParseFloat(availStr, 64)
	}

	// Parse reserved (default to 0 if not found)
	reservedStr, err := reservedCmd.Result()
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}
	if reservedStr != "" {
		reserved, _ = strconv.ParseFloat(reservedStr, 64)
	}

	return available, reserved, nil
}

func (r *RedisStorage) SetBalance(ctx context.Context, walletID string, currency string, available float64, reserved float64) error {
	availKey := balanceKey(walletID, currency) + ":available"
	reservedKey := balanceKey(walletID, currency) + ":reserved"

	pipe := r.client.Pipeline()
	pipe.Set(ctx, availKey, fmt.Sprintf("%.2f", available), 0)
	pipe.Set(ctx, reservedKey, fmt.Sprintf("%.2f", reserved), 0)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) AddBalance(ctx context.Context, walletID string, currency string, amount float64) error {
	availKey := balanceKey(walletID, currency) + ":available"
	return r.client.IncrByFloat(ctx, availKey, amount).Err()
}

func (r *RedisStorage) SubtractBalance(ctx context.Context, walletID string, currency string, amount float64) error {
	availKey := balanceKey(walletID, currency) + ":available"

	// Check current balance before subtracting
	current, _, err := r.GetBalance(ctx, walletID, currency)
	if err != nil {
		return err
	}

	if current < amount {
		return ErrInsufficientBalance
	}

	return r.client.IncrByFloat(ctx, availKey, -amount).Err()
}

// Deposit operations

func (r *RedisStorage) SaveDeposit(ctx context.Context, deposit *models.Deposit) error {
	data, err := json.Marshal(deposit)
	if err != nil {
		return fmt.Errorf("failed to marshal deposit: %w", err)
	}

	// Check if this deposit already exists to avoid duplicate list entries
	exists, err := r.client.Exists(ctx, depositKey(deposit.ID)).Result()
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, depositKey(deposit.ID), data, 0)
	if deposit.DepositReference != "" {
		pipe.Set(ctx, depositReferenceKey(deposit.DepositReference), deposit.ID, 0)
	}
	// Only add to deposits list if this is a new deposit (not an update)
	if exists == 0 {
		pipe.RPush(ctx, depositsKey(), deposit.ID)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) GetDeposit(ctx context.Context, depositID string) (*models.Deposit, error) {
	data, err := r.client.Get(ctx, depositKey(depositID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	var deposit models.Deposit
	if err := json.Unmarshal(data, &deposit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deposit: %w", err)
	}

	return &deposit, nil
}

func (r *RedisStorage) GetDepositByReference(ctx context.Context, reference string) (*models.Deposit, error) {
	depositID, err := r.client.Get(ctx, depositReferenceKey(reference)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	return r.GetDeposit(ctx, depositID)
}

func (r *RedisStorage) ListDeposits(ctx context.Context, limit int, offset int) ([]*models.Deposit, int, error) {
	// Get total count
	total, err := r.client.LLen(ctx, depositsKey()).Result()
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*models.Deposit{}, 0, nil
	}

	// Get paginated IDs
	end := int64(offset + limit - 1)
	if end >= total {
		end = total - 1
	}

	ids, err := r.client.LRange(ctx, depositsKey(), int64(offset), end).Result()
	if err != nil {
		return nil, 0, err
	}

	// Fetch deposits
	deposits := make([]*models.Deposit, 0, len(ids))
	for _, id := range ids {
		dep, err := r.GetDeposit(ctx, id)
		if err != nil {
			continue
		}
		deposits = append(deposits, dep)
	}

	return deposits, int(total), nil
}

func (r *RedisStorage) UpdateDepositStatus(ctx context.Context, depositID string, status string) error {
	dep, err := r.GetDeposit(ctx, depositID)
	if err != nil {
		return err
	}

	dep.Status = status
	if status == "settled" || status == "completed" {
		now := time.Now()
		dep.SettledAt = &now
	}

	return r.SaveDeposit(ctx, dep)
}

func (r *RedisStorage) ClearTransactions(ctx context.Context) error {
	// Clear transactions stored in Redis
	// Use pattern matching to find all transaction keys
	keys, err := r.client.Keys(ctx, "transaction:*").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	// Also clear idempotency keys
	idempKeys, err := r.client.Keys(ctx, "idempotency:*").Result()
	if err != nil {
		return err
	}

	if len(idempKeys) > 0 {
		return r.client.Del(ctx, idempKeys...).Err()
	}

	return nil
}

func (r *RedisStorage) ClearDeposits(ctx context.Context) error {
	// Get all deposits to find their references
	ids, err := r.client.LRange(ctx, depositsKey(), 0, -1).Result()
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	// Delete all deposits, references, and the deposits list
	pipe := r.client.Pipeline()
	for _, id := range ids {
		// Get the deposit to find its reference
		depositData, err := r.client.Get(ctx, depositKey(id)).Result()
		if err == nil {
			var deposit models.Deposit
			if err := json.Unmarshal([]byte(depositData), &deposit); err == nil {
				// Delete the reference mapping
				if deposit.DepositReference != "" {
					pipe.Del(ctx, "depositRef:"+deposit.DepositReference)
				}
			}
		}
		pipe.Del(ctx, depositKey(id))
	}
	pipe.Del(ctx, depositsKey())
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) ClearBalances(ctx context.Context) error {
	// This is tricky in Redis without knowing all wallet IDs
	// For now, we'll use a pattern match
	keys, err := r.client.Keys(ctx, "balance:*").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
}

// Job operations

func (r *RedisStorage) SaveJob(ctx context.Context, job *models.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, jobKey(job.ID), data, 0)

	// Add to ready queue with score based on NotBefore
	score := float64(job.NotBefore.Unix())
	pipe.ZAdd(ctx, jobsReadyKey(), redis.Z{Score: score, Member: job.ID})

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	data, err := r.client.Get(ctx, jobKey(jobID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	var job models.Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

func (r *RedisStorage) ListReadyJobs(ctx context.Context, limit int) ([]*models.Job, error) {
	now := time.Now().Unix()

	// Get job IDs that are ready (score <= now)
	ids, err := r.client.ZRangeByScore(ctx, jobsReadyKey(), &redis.ZRangeBy{
		Min:   "-inf",
		Max:   fmt.Sprintf("%d", now),
		Count: int64(limit),
	}).Result()
	if err != nil {
		return nil, err
	}

	// Fetch jobs
	jobs := make([]*models.Job, 0, len(ids))
	for _, id := range ids {
		job, err := r.GetJob(ctx, id)
		if err != nil {
			continue
		}
		if job.Status == "pending" {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

func (r *RedisStorage) UpdateJobStatus(ctx context.Context, jobID string, status string, completedAt *time.Time, lastError string) error {
	job, err := r.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	job.Status = status
	if completedAt != nil {
		job.CompletedAt = completedAt
	}
	if lastError != "" {
		job.LastError = lastError
	}

	// If completed or failed, remove from ready queue
	if status == "completed" || status == "failed" {
		pipe := r.client.Pipeline()
		pipe.ZRem(ctx, jobsReadyKey(), jobID)

		data, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal job: %w", err)
		}
		pipe.Set(ctx, jobKey(jobID), data, 0)
		_, err = pipe.Exec(ctx)
		return err
	}

	return r.SaveJob(ctx, job)
}

func (r *RedisStorage) IncrementJobAttempts(ctx context.Context, jobID string) error {
	job, err := r.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	job.Attempts++
	return r.SaveJob(ctx, job)
}

func (r *RedisStorage) ClearJobs(ctx context.Context) error {
	// Get all job IDs from the ready queue
	ids, err := r.client.ZRange(ctx, jobsReadyKey(), 0, -1).Result()
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	// Delete all jobs
	pipe := r.client.Pipeline()
	for _, id := range ids {
		pipe.Del(ctx, jobKey(id))
	}
	pipe.Del(ctx, jobsReadyKey())
	_, err = pipe.Exec(ctx)
	return err
}

// Reset all data (for testing)
func (r *RedisStorage) Reset(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}
