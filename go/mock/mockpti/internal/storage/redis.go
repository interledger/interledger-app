package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
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
