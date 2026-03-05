//go:build e2e
// +build e2e

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Global shared webhook server state across all scenarios
var (
	globalWebhookServer   *http.Server
	globalWebhookMu       sync.Mutex
	globalWebhookEvents   []webhookEvent
	globalWebhookEventsMu sync.Mutex
)

// resetWebhookEvents clears all webhook events (called between scenarios)
func resetWebhookEvents() {
	globalWebhookEventsMu.Lock()
	globalWebhookEvents = nil
	globalWebhookEventsMu.Unlock()
}

func (tc *TestContext) startWebhookServer(webhookURL string) error {
	globalWebhookMu.Lock()
	defer globalWebhookMu.Unlock()

	// If server already running, just update the URL and return
	if globalWebhookServer != nil {
		tc.webhookURL = webhookURL
		return nil
	}

	parsed, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	path := parsed.Path
	if path == "" {
		path = "/"
	}

	// Extract host and port from URL
	host, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		// No port specified, use the whole host
		host = parsed.Host
		port = "3000"
	}

	// Replace localhost with 0.0.0.0 to make server accessible from Docker
	if host == "localhost" || host == "127.0.0.1" {
		host = "0.0.0.0"
	}

	addr := host + ":" + port

	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Webhook received: %s %s from %s\n", r.Method, r.URL.Path, r.RemoteAddr)

		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()

		var payload webhookPayload
		_ = json.Unmarshal(body, &payload)

		// Store in global webhook events
		globalWebhookEventsMu.Lock()
		globalWebhookEvents = append(globalWebhookEvents, webhookEvent{
			Body:       payload,
			Headers:    r.Header.Clone(),
			RawBody:    body,
			ReceivedAt: time.Now(),
		})
		globalWebhookEventsMu.Unlock()

		fmt.Printf("Webhook stored, total count: %d\n", len(globalWebhookEvents))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	if parsed.Scheme == "https" {
		return fmt.Errorf("https webhook server not supported in test harness")
	}

	globalWebhookServer = server
	tc.webhookURL = webhookURL

	// Start server in background
	started := make(chan error, 1)
	go func() {
		// Signal that we're attempting to listen
		started <- nil
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("Webhook server error: %v\n", err)
		}
	}()

	// Wait for server to start (or fail immediately)
	select {
	case err := <-started:
		if err != nil {
			return err
		}
		// Give the server a moment to actually bind to the port
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("Webhook server started on %s%s\n", addr, path)
	case <-time.After(2 * time.Second):
		return fmt.Errorf("timeout waiting for webhook server to start")
	}

	return nil
}

func (tc *TestContext) waitForWebhookCount(count int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// Copy global webhooks to local context
		globalWebhookEventsMu.Lock()
		tc.webhookEvents = make([]webhookEvent, len(globalWebhookEvents))
		copy(tc.webhookEvents, globalWebhookEvents)
		currentCount := len(tc.webhookEvents)
		globalWebhookEventsMu.Unlock()

		if currentCount >= count {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Final check
	globalWebhookEventsMu.Lock()
	tc.webhookEvents = make([]webhookEvent, len(globalWebhookEvents))
	copy(tc.webhookEvents, globalWebhookEvents)
	currentCount := len(tc.webhookEvents)
	globalWebhookEventsMu.Unlock()

	return fmt.Errorf("expected %d webhooks, got %d", count, currentCount)
}

// waitForNextWebhook waits for 1 new webhook to arrive beyond the current count.
func (tc *TestContext) waitForNextWebhook(timeout time.Duration) error {
	globalWebhookEventsMu.Lock()
	baseline := len(globalWebhookEvents)
	globalWebhookEventsMu.Unlock()

	target := baseline + 1
	return tc.waitForWebhookCount(target, timeout)
}

func (tc *TestContext) lastWebhookEvent() (webhookEvent, error) {
	// Sync from global state
	globalWebhookEventsMu.Lock()
	defer globalWebhookEventsMu.Unlock()

	if len(globalWebhookEvents) == 0 {
		return webhookEvent{}, fmt.Errorf("no webhook events received")
	}
	return globalWebhookEvents[len(globalWebhookEvents)-1], nil
}
