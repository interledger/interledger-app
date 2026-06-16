package client

import (
	"context"
	"fmt"
	"net/url"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

const (
	startupAttempts = 3
	retryDelay      = 3 * time.Second
	pingTimeout     = 5 * time.Second
	healthTimeout   = 2 * time.Second
)

// Client is a thin facade over the go-redis client with startup retry logic
// and structured logging, mirroring the behaviour of the TypeScript redis.server.ts.
type Client struct {
	redis *goredis.Client
}

// New creates a Redis client, attempts to connect with up to [startupAttempts]
// retries, and fatals if a connection cannot be established.
func New(redisURL string) *Client {
	if redisURL == "" {
		log.Fatal("REDIS_URL is empty; cannot initialise Redis client")
	}

	target := redisTarget(redisURL)
	log.Info("initialising Redis client", zap.String("redisTarget", target))

	opts, err := goredis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("failed to parse Redis URL", zap.String("redisTarget", target), zap.Error(err))
	}

	var lastErr error
	for attempt := 1; attempt <= startupAttempts; attempt++ {
		rc := goredis.NewClient(opts)

		ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
		_, pingErr := rc.Ping(ctx).Result()
		cancel()

		if pingErr == nil {
			log.Info("Redis connected",
				zap.Int("attempt", attempt),
				zap.String("redisTarget", target),
			)
			return &Client{redis: rc}
		}

		lastErr = pingErr
		log.Error("failed to connect to Redis during startup",
			zap.Int("attempt", attempt),
			zap.Int("maxAttempts", startupAttempts),
			zap.String("redisTarget", target),
			zap.Error(pingErr),
		)

		_ = rc.Close()

		if attempt < startupAttempts {
			time.Sleep(retryDelay)
		}
	}

	log.Fatal("unable to establish Redis connection at startup; exiting",
		zap.String("redisTarget", target),
		zap.Int("maxAttempts", startupAttempts),
		zap.Error(lastErr),
	)
	return nil // unreachable — log.Fatal exits the process
}

// Redis returns the underlying go-redis client for callers that need direct access.
func (c *Client) Redis() *goredis.Client {
	return c.redis
}

// Ping verifies the Redis connection is healthy by issuing a PING command.
func (c *Client) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, healthTimeout)
	defer cancel()

	result, err := c.redis.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	if result != "PONG" {
		return fmt.Errorf("unexpected Redis ping response: %s", result)
	}
	return nil
}

// Close closes the underlying Redis connection.
func (c *Client) Close() error {
	return c.redis.Close()
}

// redisTarget extracts host:port from a Redis URL for use in log messages,
// avoiding leaking credentials or query parameters.
func redisTarget(redisURL string) string {
	parsed, err := url.Parse(redisURL)
	if err != nil {
		return "<invalid redis url>"
	}
	port := parsed.Port()
	if port == "" {
		port = "6379"
	}
	return fmt.Sprintf("%s:%s", parsed.Hostname(), port)
}
