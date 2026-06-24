package client

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestValkeyTargetSanitizesCredentialsAndQuery(t *testing.T) {
	target := valkeyTarget("valkey://user:super-secret@cache.internal:6381/0?ssl=true")

	if target != "cache.internal:6381" {
		t.Fatalf("unexpected valkey target: %s", target)
	}
	if strings.Contains(target, "super-secret") || strings.Contains(target, "user") {
		t.Fatalf("valkey target leaked credentials: %s", target)
	}
}

func TestValkeyTargetDefaultsPort(t *testing.T) {
	target := valkeyTarget("valkey://cache.internal/0")
	if target != "cache.internal:6379" {
		t.Fatalf("expected default port 6379, got: %s", target)
	}
}

func TestValkeyTargetInvalidURL(t *testing.T) {
	target := valkeyTarget("not a url")
	if target != "<invalid valkey url>" {
		t.Fatalf("expected invalid valkey url marker, got: %s", target)
	}
}

func TestNormalizeValkeyURLScheme(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want string
	}{
		{name: "valkey unchanged", in: "valkey://cache.internal:6379/0", want: "redis://cache.internal:6379/0"},
		{name: "valkey mapped to redis", in: "valkey://cache.internal:6379/0", want: "redis://cache.internal:6379/0"},
		{name: "valkeys mapped to rediss", in: "valkeys://cache.internal:6380/1", want: "rediss://cache.internal:6380/1"},
		{name: "mixed-case valkey", in: "VaLkEy://cache.internal:6379/0", want: "redis://cache.internal:6379/0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeValkeyURLScheme(tc.in)
			if got != tc.want {
				t.Fatalf("unexpected normalized URL: got=%s want=%s", got, tc.want)
			}
		})
	}
}

func TestPingHealthyValkey(t *testing.T) {
	mr := miniredis.RunT(t)

	c := New("valkey://" + mr.Addr() + "/0")
	t.Cleanup(func() {
		_ = c.Close()
	})

	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("expected healthy ping, got error: %v", err)
	}
}

func TestNewWithValkeyURL(t *testing.T) {
	mr := miniredis.RunT(t)

	c := New("valkey://" + mr.Addr() + "/0")
	t.Cleanup(func() {
		_ = c.Close()
	})

	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("expected healthy ping for valkey URL, got error: %v", err)
	}
}

func TestNewRetriesUntilValkeyBecomesAvailable(t *testing.T) {
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
	c := New("valkey://" + addr + "/0?dial_timeout=100ms")
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
