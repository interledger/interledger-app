package config

import "testing"

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("MOCKCHIMONEY_PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("MOCKCHIMONEY_API_KEY", "")
	t.Setenv("MOCKCHIMONEY_ENFORCE_AUTHENTICATION", "")
	t.Setenv("MOCKCHIMONEY_REDIS_URL", "")
	t.Setenv("MOCKCHIMONEY_REDIS_DB", "")
	t.Setenv("WEBHOOK_URL", "")
	t.Setenv("CHIMONEY_WEBHOOK_SECRET", "")
	t.Setenv("WEBHOOK_MIN_DELAY_SEC", "")
	t.Setenv("INTERAC_FEE_FLAT", "")
	t.Setenv("CAD_TO_USD_RATE", "")
	t.Setenv("MOCKCHIMONEY_PUBLIC_BASE_URL", "")

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
	if cfg.RedisURL != "" {
		t.Fatalf("RedisURL mismatch: got %q want empty", cfg.RedisURL)
	}
	if cfg.RedisDB != 5 {
		t.Fatalf("RedisDB mismatch: got %d want %d", cfg.RedisDB, 5)
	}
	if cfg.WebhookURL == "" || cfg.WebhookSecret == "" {
		t.Fatalf("webhook defaults should be populated")
	}
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("MOCKCHIMONEY_PORT", "9000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("MOCKCHIMONEY_API_KEY", "abc123")
	t.Setenv("MOCKCHIMONEY_ENFORCE_AUTHENTICATION", "false")
	t.Setenv("MOCKCHIMONEY_REDIS_URL", "redis://localhost:6379")
	t.Setenv("MOCKCHIMONEY_REDIS_DB", "7")
	t.Setenv("WEBHOOK_URL", "http://localhost:9999/webhooks")
	t.Setenv("CHIMONEY_WEBHOOK_SECRET", "prefix_bXlzZWNyZXQ=")
	t.Setenv("WEBHOOK_MIN_DELAY_SEC", "1.25")
	t.Setenv("INTERAC_FEE_FLAT", "2.75")
	t.Setenv("CAD_TO_USD_RATE", "0.72")
	t.Setenv("MOCKCHIMONEY_PUBLIC_BASE_URL", "https://example.test")

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
	if cfg.RedisURL != "redis://localhost:6379" {
		t.Fatalf("RedisURL mismatch: got %q", cfg.RedisURL)
	}
	if cfg.RedisDB != 7 {
		t.Fatalf("RedisDB mismatch: got %d want %d", cfg.RedisDB, 7)
	}
	if cfg.WebhookURL != "http://localhost:9999/webhooks" {
		t.Fatalf("WebhookURL mismatch: got %q", cfg.WebhookURL)
	}
	if cfg.WebhookSecret != "prefix_bXlzZWNyZXQ=" {
		t.Fatalf("WebhookSecret mismatch: got %q", cfg.WebhookSecret)
	}
	if cfg.WebhookMinDelaySec != 1.25 || cfg.InteracFeeFlat != 2.75 || cfg.CADToUSDRate != 0.72 {
		t.Fatalf("float config mismatch: %#v", cfg)
	}
	if cfg.PublicBaseURL != "https://example.test" {
		t.Fatalf("PublicBaseURL mismatch: got %q", cfg.PublicBaseURL)
	}
}

func TestGetEnvAsBoolInvalidFallsBackToDefault(t *testing.T) {
	t.Setenv("MOCKCHIMONEY_ENFORCE_AUTHENTICATION", "not-a-bool")

	if !getEnvAsBool("MOCKCHIMONEY_ENFORCE_AUTHENTICATION", true) {
		t.Fatalf("getEnvAsBool() should return default true when parsing fails")
	}
}

func TestGetEnvAsFloatInvalidFallsBackToDefault(t *testing.T) {
	t.Setenv("INTERAC_FEE_FLAT", "not-a-float")

	if getEnvAsFloat("INTERAC_FEE_FLAT", 3.5) != 3.5 {
		t.Fatalf("getEnvAsFloat() should return default when parsing fails")
	}
}

func TestGetEnvAsIntInvalidFallsBackToDefault(t *testing.T) {
	t.Setenv("MOCKCHIMONEY_REDIS_DB", "not-an-int")

	if getEnvAsInt("MOCKCHIMONEY_REDIS_DB", 11) != 11 {
		t.Fatalf("getEnvAsInt() should return default when parsing fails")
	}
}
