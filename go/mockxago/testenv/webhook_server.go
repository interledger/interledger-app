//go:build e2e

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (tc *TestContext) startWebhookServer(webhookURL string) error {
	if tc.webhookServer != nil {
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

	addr := parsed.Host
	if addr == "" {
		addr = ":3000"
	}
	if !strings.Contains(addr, ":") {
		addr = addr + ":80"
	}

	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()

		var payload webhookPayload
		_ = json.Unmarshal(body, &payload)

		tc.webhookMu.Lock()
		tc.webhookEvents = append(tc.webhookEvents, webhookEvent{
			Body:       payload,
			Headers:    r.Header.Clone(),
			RawBody:    body,
			ReceivedAt: time.Now(),
		})
		tc.webhookMu.Unlock()

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

	tc.webhookServer = server
	tc.webhookURL = webhookURL

	go func() {
		_ = server.ListenAndServe()
	}()

	return nil
}

func (tc *TestContext) waitForWebhookCount(count int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		tc.webhookMu.Lock()
		got := len(tc.webhookEvents)
		tc.webhookMu.Unlock()
		if got >= count {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	tc.webhookMu.Lock()
	got := len(tc.webhookEvents)
	tc.webhookMu.Unlock()
	return fmt.Errorf("expected %d webhooks, got %d", count, got)
}

func (tc *TestContext) lastWebhookEvent() (webhookEvent, error) {
	tc.webhookMu.Lock()
	defer tc.webhookMu.Unlock()
	if len(tc.webhookEvents) == 0 {
		return webhookEvent{}, fmt.Errorf("no webhook events received")
	}
	return tc.webhookEvents[len(tc.webhookEvents)-1], nil
}
