package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	_ = os.Unsetenv("MOCKPTI_PORT")
	_ = os.Unsetenv("LOG_LEVEL")
	_ = os.Unsetenv("MOCKPTI_REDIS_URL")
	_ = os.Unsetenv("MOCKPTI_REDIS_DB")
	_ = os.Unsetenv("MOCKPTI_CLIENT_ID")

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
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("MOCKPTI_PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("MOCKPTI_CLIENT_ID", "my-client")

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
}
