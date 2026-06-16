package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/models"
	"github.com/redis/go-redis/v9"
)

// RedisStorage is a Redis-backed implementation of the Storage interface.
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a new Redis storage client.
func NewRedisStorage(addr string, db int) (Storage, error) {
	var redisAddr string
	if u, err := url.Parse(addr); err == nil && u.Scheme == "redis" {
		redisAddr = u.Host
		if redisAddr == "" {
			redisAddr = addr
		}
	} else {
		redisAddr = addr
	}

	client := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		DB:           db,
		Password:     "",
		PoolSize:     10,
		MinIdleConns: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{client: client}, nil
}

// Redis key patterns

func userKey(userID string) string {
	return fmt.Sprintf("pti:user:%s", userID)
}

func assessmentsKey(userID string) string {
	return fmt.Sprintf("pti:assessments:%s", userID)
}

func walletKey(userID, walletID string) string {
	return fmt.Sprintf("pti:wallet:%s:%s", userID, walletID)
}

func walletsIndexKey(userID string) string {
	return fmt.Sprintf("pti:wallets:%s", userID)
}

func paymentInfoKey(userID, piID string) string {
	return fmt.Sprintf("pti:paymentinfo:%s:%s", userID, piID)
}

func transactionKey(requestID string) string {
	return fmt.Sprintf("pti:transaction:%s", requestID)
}

func transactionUpdatesKey(requestID string) string {
	return fmt.Sprintf("pti:txupdates:%s", requestID)
}

// User operations

func (r *RedisStorage) SaveUser(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}
	return r.client.Set(ctx, userKey(user.ID), data, 0).Err()
}

func (r *RedisStorage) GetUser(ctx context.Context, userID string) (*models.User, error) {
	data, err := r.client.Get(ctx, userKey(userID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}
	return &user, nil
}

func (r *RedisStorage) UpdateUser(ctx context.Context, user *models.User) error {
	exists, err := r.client.Exists(ctx, userKey(user.ID)).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return ErrUserNotFound
	}
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}
	return r.client.Set(ctx, userKey(user.ID), data, 0).Err()
}

// Assessment operations

func (r *RedisStorage) SaveAssessment(ctx context.Context, assessment *models.Assessment) error {
	data, err := json.Marshal(assessment)
	if err != nil {
		return fmt.Errorf("failed to marshal assessment: %w", err)
	}
	return r.client.RPush(ctx, assessmentsKey(assessment.UserID), data).Err()
}

func (r *RedisStorage) GetLatestAssessment(ctx context.Context, userID string) (*models.Assessment, error) {
	data, err := r.client.LRange(ctx, assessmentsKey(userID), -1, -1).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrAssessmentNotFound
	}

	var assessment models.Assessment
	if err := json.Unmarshal([]byte(data[0]), &assessment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal assessment: %w", err)
	}
	return &assessment, nil
}

// Reset clears all data.

func (r *RedisStorage) Reset(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Wallet operations

func (r *RedisStorage) SaveWallet(ctx context.Context, wallet *models.Wallet) error {
	data, err := json.Marshal(wallet)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet: %w", err)
	}
	pipe := r.client.Pipeline()
	pipe.Set(ctx, walletKey(wallet.UserID, wallet.WalletID), data, 0)
	pipe.SAdd(ctx, walletsIndexKey(wallet.UserID), wallet.WalletID)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) GetWallet(ctx context.Context, userID, walletID string) (*models.Wallet, error) {
	data, err := r.client.Get(ctx, walletKey(userID, walletID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}

	var wallet models.Wallet
	if err := json.Unmarshal(data, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet: %w", err)
	}
	return &wallet, nil
}

func (r *RedisStorage) ListWallets(ctx context.Context, userID string) ([]*models.Wallet, error) {
	ids, err := r.client.SMembers(ctx, walletsIndexKey(userID)).Result()
	if err != nil {
		return nil, err
	}

	wallets := make([]*models.Wallet, 0, len(ids))
	for _, id := range ids {
		data, err := r.client.Get(ctx, walletKey(userID, id)).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}
		var wallet models.Wallet
		if err := json.Unmarshal(data, &wallet); err != nil {
			return nil, fmt.Errorf("failed to unmarshal wallet: %w", err)
		}
		wallets = append(wallets, &wallet)
	}
	return wallets, nil
}

// Payment information operations

func (r *RedisStorage) SavePaymentInformation(ctx context.Context, pi *models.PaymentInformation) error {
	data, err := json.Marshal(pi)
	if err != nil {
		return fmt.Errorf("failed to marshal payment information: %w", err)
	}
	return r.client.Set(ctx, paymentInfoKey(pi.UserID, pi.ID), data, 0).Err()
}

func (r *RedisStorage) GetPaymentInformation(ctx context.Context, userID, piID string) (*models.PaymentInformation, error) {
	data, err := r.client.Get(ctx, paymentInfoKey(userID, piID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrPaymentInformationNotFound
		}
		return nil, err
	}

	var pi models.PaymentInformation
	if err := json.Unmarshal(data, &pi); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment information: %w", err)
	}
	return &pi, nil
}

// Transaction operations

func (r *RedisStorage) SaveTransaction(ctx context.Context, tx *models.Transaction) error {
	tx.CreatedAt = time.Now()
	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}
	return r.client.Set(ctx, transactionKey(tx.RequestID), data, 0).Err()
}

func (r *RedisStorage) GetTransaction(ctx context.Context, requestID string) (*models.Transaction, error) {
	data, err := r.client.Get(ctx, transactionKey(requestID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}

	var tx models.Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}
	return &tx, nil
}

func (r *RedisStorage) SaveTransactionUpdate(ctx context.Context, update *models.TransactionUpdate) error {
	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction update: %w", err)
	}
	return r.client.RPush(ctx, transactionUpdatesKey(update.RequestID), data).Err()
}

// Job key helpers

func jobKey(jobID string) string {
	return fmt.Sprintf("pti:job:%s", jobID)
}

const jobsReadySetKey = "pti:jobs:ready"

// Job operations

func (r *RedisStorage) SaveJob(ctx context.Context, job *models.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	pipe := r.client.Pipeline()
	pipe.Set(ctx, jobKey(job.ID), data, 0)
	if job.Status == "queued" {
		// Use sorted set: score = NotBefore Unix timestamp.
		pipe.ZAdd(ctx, jobsReadySetKey, redis.Z{
			Score:  float64(job.NotBefore.Unix()),
			Member: job.ID,
		})
	} else {
		pipe.ZRem(ctx, jobsReadySetKey, job.ID)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	data, err := r.client.Get(ctx, jobKey(jobID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrJobNotFound
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
	now := float64(time.Now().Unix())
	jobs := make([]*models.Job, 0, limit)
	offset := int64(0)
	for len(jobs) < limit {
		remaining := limit - len(jobs)
		batch := int64(remaining * 5)
		if batch < 10 {
			batch = 10
		}

		ids, err := r.client.ZRangeArgs(ctx, redis.ZRangeArgs{
			Key:     jobsReadySetKey,
			Start:   "-inf",
			Stop:    fmt.Sprintf("%f", now),
			ByScore: true,
			Offset:  offset,
			Count:   batch,
		}).Result()
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			break
		}

		offset += int64(len(ids))
		for _, id := range ids {
			job, err := r.GetJob(ctx, id)
			if err != nil {
				continue
			}
			if job.Status == "queued" {
				jobs = append(jobs, job)
				if len(jobs) == limit {
					break
				}
			}
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
	job.LastError = lastError
	if completedAt != nil {
		t := *completedAt
		job.CompletedAt = &t
	}
	// Keep ready set consistent with status transitions.
	if status == "processing" || status == "delivered" || status == "failed" {
		_ = r.client.ZRem(ctx, jobsReadySetKey, jobID).Err()
	}
	if status == "queued" {
		_ = r.client.ZAdd(ctx, jobsReadySetKey, redis.Z{
			Score:  float64(job.NotBefore.Unix()),
			Member: jobID,
		}).Err()
	}
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	return r.client.Set(ctx, jobKey(jobID), data, 0).Err()
}

func (r *RedisStorage) IncrementJobAttempts(ctx context.Context, jobID string) error {
	job, err := r.GetJob(ctx, jobID)
	if err != nil {
		return err
	}
	job.Attempts++
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	return r.client.Set(ctx, jobKey(jobID), data, 0).Err()
}

func (r *RedisStorage) ClearJobs(ctx context.Context) error {
	// Delete all job keys
	keys, err := r.client.Keys(ctx, "pti:job:*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}
	return r.client.Del(ctx, jobsReadySetKey).Err()
}
