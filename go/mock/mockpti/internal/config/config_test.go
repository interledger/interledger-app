package config

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	_ = os.Unsetenv("MOCKPTI_PORT")
	_ = os.Unsetenv("LOG_LEVEL")
	_ = os.Unsetenv("MOCKPTI_REDIS_URL")
	_ = os.Unsetenv("MOCKPTI_REDIS_DB")
	_ = os.Unsetenv("MOCKPTI_CLIENT_ID")
	_ = os.Unsetenv("MOCKPTI_WEBHOOK_SIGNING_KEY_B64")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level info, got %s", cfg.LogLevel)
	}
	if cfg.RedisURL != "" {
		t.Errorf("expected empty RedisURL, got %s", cfg.RedisURL)
	}
	if cfg.RedisDB != "0" {
		t.Errorf("expected default RedisDB 0, got %s", cfg.RedisDB)
	}
	if cfg.ClientID != "test-client-id" {
		t.Errorf("expected default ClientID test-client-id, got %s", cfg.ClientID)
	}
	if cfg.WebhookSigningKey != "" {
		t.Errorf("expected empty WebhookSigningKey, got %s", cfg.WebhookSigningKey)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("MOCKPTI_PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("MOCKPTI_CLIENT_ID", "my-client")
	t.Setenv("MOCKPTI_WEBHOOK_SIGNING_KEY_B64", base64.StdEncoding.EncodeToString([]byte("signing-pem")))

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.LogLevel)
	}
	if cfg.ClientID != "my-client" {
		t.Errorf("expected ClientID my-client, got %s", cfg.ClientID)
	}
	if cfg.WebhookSigningKey != "signing-pem" {
		t.Errorf("expected WebhookSigningKey signing-pem, got %s", cfg.WebhookSigningKey)
	}
}

func TestLoad_InvalidB64_YieldsEmptySigningKey(t *testing.T) {
	t.Setenv("MOCKPTI_WEBHOOK_SIGNING_KEY_B64", "not-valid-base64!!!")

	cfg := Load()

	if cfg.WebhookSigningKey != "" {
		t.Errorf("expected empty WebhookSigningKey for invalid b64, got %s", cfg.WebhookSigningKey)
	}
}
