package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
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
func New(valkeyURL string) *Client {
	if valkeyURL == "" {
		log.Fatal("VALKEY_URL is empty; cannot initialise Valkey client")
	}

	redisURL := normalizeValkeyURLScheme(valkeyURL)
	target := valkeyTarget(redisURL)
	log.Info("initialising Valkey client", zap.String("valkeyTarget", target))

	opts, err := goredis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("failed to parse Valkey URL", zap.String("valkeyTarget", target), zap.Error(err))
	}

	var lastErr error
	for attempt := 1; attempt <= startupAttempts; attempt++ {
		rc := goredis.NewClient(opts)

		ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
		_, pingErr := rc.Ping(ctx).Result()
		cancel()

		if pingErr == nil {
			log.Info("Valkey connected",
				zap.Int("attempt", attempt),
				zap.String("valkeyTarget", target),
			)
			return &Client{redis: rc}
		}

		lastErr = pingErr
		log.Error("failed to connect to Valkey during startup",
			zap.Int("attempt", attempt),
			zap.Int("maxAttempts", startupAttempts),
			zap.String("valkeyTarget", target),
			zap.Error(pingErr),
		)

		_ = rc.Close()

		if attempt < startupAttempts {
			time.Sleep(retryDelay)
		}
	}

	log.Fatal("unable to establish Valkey connection at startup; exiting",
		zap.String("valkeyTarget", target),
		zap.Int("maxAttempts", startupAttempts),
		zap.Error(lastErr),
	)
	return nil // unreachable — log.Fatal exits the process
}

// normalizeValkeyURLScheme maps Valkey URL schemes to redis-compatible schemes
// expected by go-redis ParseURL.
func normalizeValkeyURLScheme(valkeyURL string) string {
	const (
		valkeyPrefix  = "valkey://"
		valkeysPrefix = "valkeys://"
	)

	lower := strings.ToLower(valkeyURL)
	if strings.HasPrefix(lower, valkeysPrefix) {
		return "rediss://" + valkeyURL[len(valkeysPrefix):]
	}
	if strings.HasPrefix(lower, valkeyPrefix) {
		return "redis://" + valkeyURL[len(valkeyPrefix):]
	}

	return valkeyURL
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
		return fmt.Errorf("valkey ping failed: %w", err)
	}
	if strings.ToUpper(result) != "PONG" {
		return fmt.Errorf("unexpected Valkey ping response: %s", result)
	}
	return nil
}

// Close closes the underlying Redis connection.
func (c *Client) Close() error {
	return c.redis.Close()
}

// valkeyTarget extracts host:port from a Valkey URL for use in log messages,
// avoiding leaking credentials or query parameters.
func valkeyTarget(valkeyURL string) string {
	parsed, err := url.Parse(valkeyURL)
	if err != nil {
		return "<invalid valkey url>"
	}
	hostname := parsed.Hostname()
	if hostname == "" {
		return "<invalid valkey url>"
	}
	port := parsed.Port()
	if port == "" {
		port = "6379"
	}
	return fmt.Sprintf("%s:%s", hostname, port)
}
