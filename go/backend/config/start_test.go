package config

import (
	"strings"
	"testing"
)

const twilioProdGuardrailErr = "twilio.enabled must be true when environment.mode is prod"

// twilioEnabledConfig returns a StartConfig with Twilio configured with the
// given enabled state and mode, and the required Twilio credentials filled in
// so that the conditional field validation passes when enabled.
func twilioConfig(mode string, enabled bool) *StartConfig {
	cfg := &StartConfig{}
	cfg.Environment.Mode = mode
	cfg.Twilio.Enabled = enabled
	cfg.Twilio.AccountSID = "AC_test"
	cfg.Twilio.AccountToken = "token_test"
	cfg.Twilio.ServiceSID = "VA_test"
	return cfg
}

// TestValidateStartTwilioProdGuardrail asserts that disabling Twilio is rejected
// in production but permitted (defaulting to disabled) in every other mode.
func TestValidateStartTwilioProdGuardrail(t *testing.T) {
	t.Run("prod with twilio disabled is rejected", func(t *testing.T) {
		err := validateStart(twilioConfig("prod", false))
		if err == nil {
			t.Fatalf("expected error when twilio is disabled in prod, got nil")
		}
		if !strings.Contains(err.Error(), twilioProdGuardrailErr) {
			t.Fatalf("expected error %q, got %q", twilioProdGuardrailErr, err.Error())
		}
	})

	t.Run("prod with twilio enabled passes the guardrail", func(t *testing.T) {
		// prod may still fail later validation (e.g. gatehub); we only assert the
		// twilio guardrail no longer fires once twilio is enabled.
		err := validateStart(twilioConfig("prod", true))
		if err != nil && strings.Contains(err.Error(), twilioProdGuardrailErr) {
			t.Fatalf("did not expect twilio guardrail error when enabled, got %q", err.Error())
		}
	})

	for _, mode := range []string{"sandbox", "dev", "local", "test"} {
		t.Run(mode+" with twilio disabled is allowed", func(t *testing.T) {
			cfg := twilioConfig(mode, false)
			// Clear credentials to prove none are required when disabled.
			cfg.Twilio.AccountSID = ""
			cfg.Twilio.AccountToken = ""
			cfg.Twilio.ServiceSID = ""
			if err := validateStart(cfg); err != nil {
				t.Fatalf("expected no error for mode %q with twilio disabled, got %q", mode, err.Error())
			}
		})
	}
}
