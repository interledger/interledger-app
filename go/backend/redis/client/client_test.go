package client

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisTargetSanitizesCredentialsAndQuery(t *testing.T) {
	target := redisTarget("redis://user:super-secret@cache.internal:6381/0?ssl=true")

	if target != "cache.internal:6381" {
		t.Fatalf("unexpected redis target: %s", target)
	}
	if strings.Contains(target, "super-secret") || strings.Contains(target, "user") {
		t.Fatalf("redis target leaked credentials: %s", target)
	}
}

func TestRedisTargetDefaultsPort(t *testing.T) {
	target := redisTarget("redis://cache.internal/0")
	if target != "cache.internal:6379" {
		t.Fatalf("expected default port 6379, got: %s", target)
	}
}

func TestRedisTargetInvalidURL(t *testing.T) {
	target := redisTarget("not a url")
	if target != "<invalid redis url>" {
		t.Fatalf("expected invalid redis url marker, got: %s", target)
	}
}

func TestPingHealthyRedis(t *testing.T) {
	mr := miniredis.RunT(t)

	c := New("redis://" + mr.Addr() + "/0")
	t.Cleanup(func() {
		_ = c.Close()
	})

	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("expected healthy ping, got error: %v", err)
	}
}

func TestNewRetriesUntilRedisBecomesAvailable(t *testing.T) {
	addr := reserveAddr(t)
	started := make(chan *miniredis.Miniredis, 1)
	errCh := make(chan error, 1)

	go func() {
		time.Sleep(1 * time.Second)
		mr := miniredis.NewMiniRedis()
		if err := mr.StartAddr(addr); err != nil {
			errCh <- err
			return
		}
		started <- mr
	}()

	start := time.Now()
	c := New("redis://" + addr + "/0?dial_timeout=100ms")
	elapsed := time.Since(start)

	select {
	case err := <-errCh:
		t.Fatalf("failed to start miniredis: %v", err)
	default:
	}

	select {
	case mr := <-started:
		t.Cleanup(func() {
			mr.Close()
		})
	case <-time.After(2 * time.Second):
		t.Fatal("miniredis did not start")
	}

	t.Cleanup(func() {
		_ = c.Close()
	})

	if elapsed < retryDelay {
		t.Fatalf("expected at least one retry delay, elapsed=%s retryDelay=%s", elapsed, retryDelay)
	}

	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("expected ping to succeed after retry, got: %v", err)
	}
}

func reserveAddr(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve tcp address: %v", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	return listener.Addr().String()
}
