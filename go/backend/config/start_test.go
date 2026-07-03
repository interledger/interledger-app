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

const personaFakeZAIDGuardrailErr = "persona.sandbox_fake_za_id must be false when environment.mode is prod"

// TestValidateStartPersonaFakeZAIDProdGuardrail asserts that the Persona sandbox
// fake ZA ID cannot be enabled in production but is permitted in other modes.
func TestValidateStartPersonaFakeZAIDProdGuardrail(t *testing.T) {
	t.Run("prod with sandbox_fake_za_id enabled is rejected", func(t *testing.T) {
		cfg := twilioConfig("prod", true)
		cfg.Persona.SandboxFakeZAID = true
		err := validateStart(cfg)
		if err == nil {
			t.Fatalf("expected error when persona.sandbox_fake_za_id is enabled in prod, got nil")
		}
		if !strings.Contains(err.Error(), personaFakeZAIDGuardrailErr) {
			t.Fatalf("expected error %q, got %q", personaFakeZAIDGuardrailErr, err.Error())
		}
	})

	t.Run("prod without sandbox_fake_za_id passes the guardrail", func(t *testing.T) {
		cfg := twilioConfig("prod", true)
		cfg.Persona.SandboxFakeZAID = false
		err := validateStart(cfg)
		if err != nil && strings.Contains(err.Error(), personaFakeZAIDGuardrailErr) {
			t.Fatalf("did not expect persona guardrail error when disabled, got %q", err.Error())
		}
	})

	for _, mode := range []string{"sandbox", "dev", "local", "test"} {
		t.Run(mode+" with sandbox_fake_za_id enabled is allowed", func(t *testing.T) {
			cfg := twilioConfig(mode, false)
			cfg.Persona.SandboxFakeZAID = true
			if err := validateStart(cfg); err != nil {
				t.Fatalf("expected no error for mode %q with sandbox_fake_za_id enabled, got %q", mode, err.Error())
			}
		})
	}
}
