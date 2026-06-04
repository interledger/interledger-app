package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"

	"gitlab.com/fynbos/mock/mockplaid/internal/models"
)

// RedisStorage is a Redis-backed implementation of Storage.
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

// Key patterns.

func linkKey(linkToken string) string      { return "plaid:link:" + linkToken }
func itemPubKey(publicToken string) string { return "plaid:item:pub:" + publicToken }
func itemAccKey(accessToken string) string { return "plaid:item:acc:" + accessToken }

const accountSeqKey = "plaid:accseq"

func (r *RedisStorage) SaveLinkSession(ctx context.Context, s models.LinkSession) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, linkKey(s.LinkToken), data, 0).Err()
}

func (r *RedisStorage) GetLinkSession(ctx context.Context, linkToken string) (models.LinkSession, error) {
	val, err := r.client.Get(ctx, linkKey(linkToken)).Result()
	if err == redis.Nil {
		return models.LinkSession{}, ErrLinkSessionNotFound
	}
	if err != nil {
		return models.LinkSession{}, err
	}
	var s models.LinkSession
	if err := json.Unmarshal([]byte(val), &s); err != nil {
		return models.LinkSession{}, err
	}
	return s, nil
}

func (r *RedisStorage) SaveItem(ctx context.Context, item models.Item) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	pipe := r.client.TxPipeline()
	if item.PublicToken != "" {
		pipe.Set(ctx, itemPubKey(item.PublicToken), data, 0)
	}
	if item.AccessToken != "" {
		pipe.Set(ctx, itemAccKey(item.AccessToken), data, 0)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) getItem(ctx context.Context, key string) (models.Item, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return models.Item{}, ErrItemNotFound
	}
	if err != nil {
		return models.Item{}, err
	}
	var item models.Item
	if err := json.Unmarshal([]byte(val), &item); err != nil {
		return models.Item{}, err
	}
	return item, nil
}

func (r *RedisStorage) GetItemByPublicToken(ctx context.Context, publicToken string) (models.Item, error) {
	return r.getItem(ctx, itemPubKey(publicToken))
}

func (r *RedisStorage) GetItemByAccessToken(ctx context.Context, accessToken string) (models.Item, error) {
	return r.getItem(ctx, itemAccKey(accessToken))
}

func (r *RedisStorage) DeleteItemByAccessToken(ctx context.Context, accessToken string) error {
	item, err := r.getItem(ctx, itemAccKey(accessToken))
	if err == ErrItemNotFound {
		// idempotent: no-op success
		return nil
	}
	if err != nil {
		return err
	}
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, itemAccKey(accessToken))
	if item.PublicToken != "" {
		pipe.Del(ctx, itemPubKey(item.PublicToken))
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStorage) NextAccountSeq(ctx context.Context) (uint64, error) {
	v, err := r.client.Incr(ctx, accountSeqKey).Result()
	if err != nil {
		return 0, err
	}
	return uint64(v), nil
}

func (r *RedisStorage) Reset(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}
