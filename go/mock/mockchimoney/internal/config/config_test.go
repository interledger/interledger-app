package config

import "testing"

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("MOCKCHIMONEY_PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MOCKCHIMONEY_API_KEY", "")
	t.Setenv("MOCKCHIMONEY_ENFORCE_AUTHENTICATION", "")

	cfg := Load()
	if cfg.Port != "41800" {
		t.Fatalf("Port mismatch: got %q want %q", cfg.Port, "41800")
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel mismatch: got %q want %q", cfg.LogLevel, "info")
	}
	if cfg.APIKey != "local-test-api-key" {
		t.Fatalf("APIKey mismatch: got %q want %q", cfg.APIKey, "local-test-api-key")
	}
	if !cfg.EnforceAuthentication {
		t.Fatalf("EnforceAuthentication mismatch: got %v want %v", cfg.EnforceAuthentication, true)
	}
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("MOCKCHIMONEY_PORT", "9000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("MOCKCHIMONEY_API_KEY", "abc123")
	t.Setenv("MOCKCHIMONEY_ENFORCE_AUTHENTICATION", "false")

	cfg := Load()
	if cfg.Port != "9000" {
		t.Fatalf("Port mismatch: got %q want %q", cfg.Port, "9000")
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel mismatch: got %q want %q", cfg.LogLevel, "debug")
	}
	if cfg.APIKey != "abc123" {
		t.Fatalf("APIKey mismatch: got %q want %q", cfg.APIKey, "abc123")
	}
	if cfg.EnforceAuthentication {
		t.Fatalf("EnforceAuthentication mismatch: got %v want %v", cfg.EnforceAuthentication, false)
	}
}

func TestGetEnvAsBoolInvalidFallsBackToDefault(t *testing.T) {
	t.Setenv("MOCKCHIMONEY_ENFORCE_AUTHENTICATION", "not-a-bool")

	if !getEnvAsBool("MOCKCHIMONEY_ENFORCE_AUTHENTICATION", true) {
		t.Fatalf("getEnvAsBool() should return default true when parsing fails")
	}
}
