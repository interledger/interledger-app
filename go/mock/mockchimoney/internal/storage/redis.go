package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
)

const chimoneyKeyPrefix = "chimoney"

// RedisStore is a Redis-backed Store implementation.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a Redis-backed store and verifies connectivity.
func NewRedisStore(redisURL string, db int) (*RedisStore, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}
	opt.DB = db

	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{client: client}, nil
}

// Close closes the Redis connection.
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// FlushAll removes all keys in the selected Redis DB.
func (s *RedisStore) FlushAll(ctx context.Context) error {
	if err := s.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("flush redis db: %w", err)
	}
	return nil
}

func (s *RedisStore) CreateSubAccount(ctx context.Context, account models.SubAccount) (models.SubAccount, error) {
	created, err := s.setIfAbsent(ctx, subAccountKey(account.ID), account)
	if err != nil {
		return models.SubAccount{}, err
	}
	if !created {
		return models.SubAccount{}, ErrAlreadyExists
	}

	if err := s.client.SAdd(ctx, subAccountsIndexKey(), account.ID).Err(); err != nil {
		return models.SubAccount{}, fmt.Errorf("add sub-account index: %w", err)
	}

	return account, nil
}

func (s *RedisStore) GetSubAccount(ctx context.Context, id string) (models.SubAccount, error) {
	var account models.SubAccount
	if err := s.getJSON(ctx, subAccountKey(id), &account); err != nil {
		return models.SubAccount{}, err
	}
	return account, nil
}

func (s *RedisStore) ListSubAccounts(ctx context.Context) ([]models.SubAccount, error) {
	ids, err := s.client.SMembers(ctx, subAccountsIndexKey()).Result()
	if err != nil {
		return nil, fmt.Errorf("list sub-account ids: %w", err)
	}

	accounts := make([]models.SubAccount, 0, len(ids))
	for _, id := range ids {
		account, getErr := s.GetSubAccount(ctx, id)
		if getErr == ErrNotFound {
			continue
		}
		if getErr != nil {
			return nil, getErr
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (s *RedisStore) UpdateSubAccountKYCStatus(ctx context.Context, id string, status string) (models.SubAccount, error) {
	account, err := s.GetSubAccount(ctx, id)
	if err != nil {
		return models.SubAccount{}, err
	}
	account.KYCStatus = status

	if err := s.setJSON(ctx, subAccountKey(id), account); err != nil {
		return models.SubAccount{}, err
	}
	return account, nil
}

func (s *RedisStore) CreatePayment(ctx context.Context, payment models.Payment) (models.Payment, error) {
	created, err := s.setIfAbsent(ctx, paymentKey(payment.IssueID), payment)
	if err != nil {
		return models.Payment{}, err
	}
	if !created {
		return models.Payment{}, ErrAlreadyExists
	}
	return payment, nil
}

func (s *RedisStore) GetPaymentByIssueID(ctx context.Context, issueID string) (models.Payment, error) {
	var payment models.Payment
	if err := s.getJSON(ctx, paymentKey(issueID), &payment); err != nil {
		return models.Payment{}, err
	}
	return payment, nil
}

func (s *RedisStore) UpdatePaymentStatus(ctx context.Context, issueID string, status string) (models.Payment, error) {
	payment, err := s.GetPaymentByIssueID(ctx, issueID)
	if err != nil {
		return models.Payment{}, err
	}
	payment.Status = status

	if err := s.setJSON(ctx, paymentKey(issueID), payment); err != nil {
		return models.Payment{}, err
	}
	return payment, nil
}

func (s *RedisStore) CreatePayout(ctx context.Context, payout models.Payout) (models.Payout, error) {
	created, err := s.setIfAbsent(ctx, payoutKey(payout.IssueID), payout)
	if err != nil {
		return models.Payout{}, err
	}
	if !created {
		return models.Payout{}, ErrAlreadyExists
	}

	mapped, err := s.client.SetNX(ctx, payoutChiRefKey(payout.ChiRef), payout.IssueID, 0).Result()
	if err != nil {
		return models.Payout{}, fmt.Errorf("map payout chiRef: %w", err)
	}
	if !mapped {
		_ = s.client.Del(ctx, payoutKey(payout.IssueID)).Err()
		return models.Payout{}, ErrAlreadyExists
	}

	return payout, nil
}

func (s *RedisStore) GetPayoutByChiRef(ctx context.Context, chiRef string) (models.Payout, error) {
	issueID, err := s.client.Get(ctx, payoutChiRefKey(chiRef)).Result()
	if err == redis.Nil {
		return models.Payout{}, ErrNotFound
	}
	if err != nil {
		return models.Payout{}, fmt.Errorf("get payout issue id by chiRef: %w", err)
	}

	return s.GetPayoutByIssueID(ctx, issueID)
}

func (s *RedisStore) GetPayoutByIssueID(ctx context.Context, issueID string) (models.Payout, error) {
	var payout models.Payout
	if err := s.getJSON(ctx, payoutKey(issueID), &payout); err != nil {
		return models.Payout{}, err
	}
	return payout, nil
}

func (s *RedisStore) UpdatePayoutStatus(ctx context.Context, issueID string, status string) (models.Payout, error) {
	payout, err := s.GetPayoutByIssueID(ctx, issueID)
	if err != nil {
		return models.Payout{}, err
	}
	payout.Status = status

	if err := s.setJSON(ctx, payoutKey(issueID), payout); err != nil {
		return models.Payout{}, err
	}
	return payout, nil
}

func (s *RedisStore) getJSON(ctx context.Context, key string, dest any) error {
	encoded, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("get %s: %w", key, err)
	}
	if err := json.Unmarshal([]byte(encoded), dest); err != nil {
		return fmt.Errorf("decode %s: %w", key, err)
	}
	return nil
}

func (s *RedisStore) setJSON(ctx context.Context, key string, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s: %w", key, err)
	}
	if err := s.client.Set(ctx, key, encoded, 0).Err(); err != nil {
		return fmt.Errorf("set %s: %w", key, err)
	}
	return nil
}

func (s *RedisStore) setIfAbsent(ctx context.Context, key string, value any) (bool, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("encode %s: %w", key, err)
	}
	created, err := s.client.SetNX(ctx, key, encoded, 0).Result()
	if err != nil {
		return false, fmt.Errorf("setnx %s: %w", key, err)
	}
	return created, nil
}

func subAccountKey(id string) string {
	return fmt.Sprintf("%s:subaccount:%s", chimoneyKeyPrefix, id)
}

func subAccountsIndexKey() string {
	return fmt.Sprintf("%s:subaccounts", chimoneyKeyPrefix)
}

func paymentKey(issueID string) string {
	return fmt.Sprintf("%s:payment:%s", chimoneyKeyPrefix, issueID)
}

func payoutKey(issueID string) string {
	return fmt.Sprintf("%s:payout:%s", chimoneyKeyPrefix, issueID)
}

func payoutChiRefKey(chiRef string) string {
	return fmt.Sprintf("%s:payout:chiref:%s", chimoneyKeyPrefix, chiRef)
}
