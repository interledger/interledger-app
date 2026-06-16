package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"gitlab.com/fynbos/backend/providers/plaid"
)

// Redis is the persistent TokenStore used by the Plaid POC. Values are stored
// as JSON-encoded TokenSet under keys of the form `plaid:token:{userID}` with
// no expiry — access tokens are long-lived and a missing key drives the user
// back through Link, which is the desired behaviour after Redis wipes.
type Redis struct {
	client *redis.Client
}

func NewRedis(redisURL string) (*Redis, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("plaid redis store: parse url: %w", err)
	}
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("plaid redis store: ping: %w", err)
	}
	return &Redis{client: client}, nil
}

func (redisStore *Redis) Close() error { return redisStore.client.Close() }

func tokenKey(userID string) string {
	return "plaid:token:" + userID
}

func (redisStore *Redis) Get(ctx context.Context, userID string) (plaid.TokenSet, bool, error) {
	raw, err := redisStore.client.Get(ctx, tokenKey(userID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return plaid.TokenSet{}, false, nil
	}
	if err != nil {
		return plaid.TokenSet{}, false, fmt.Errorf("plaid redis store: get: %w", err)
	}
	var t plaid.TokenSet
	if err := json.Unmarshal(raw, &t); err != nil {
		return plaid.TokenSet{}, false, fmt.Errorf("plaid redis store: decode: %w", err)
	}
	return t, true, nil
}

func (redisStore *Redis) Put(ctx context.Context, userID string, t plaid.TokenSet) error {
	raw, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("plaid redis store: encode: %w", err)
	}
	if err := redisStore.client.Set(ctx, tokenKey(userID), raw, 0).Err(); err != nil {
		return fmt.Errorf("plaid redis store: set: %w", err)
	}
	return nil
}

func (redisStore *Redis) Delete(ctx context.Context, userID string) error {
	if err := redisStore.client.Del(ctx, tokenKey(userID)).Err(); err != nil {
		return fmt.Errorf("plaid redis store: del: %w", err)
	}
	return nil
}
