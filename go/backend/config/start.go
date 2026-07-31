package config

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/interledger/interledger-app/go/configa"
)

// IsTestExecution returns true when running under go test or when mode is "test".
func IsTestExecution(mode string) bool {
	return mode == "test" || flag.Lookup("test.v") != nil
}

// LoadStart parses the YAML files listed in the CONFIG environment variable
// and returns a validated StartConfig. Panics on fatal configuration errors so
// that callers can rely on a fully initialised struct or not return at all.
func LoadStart() (*StartConfig, error) {
	files, err := configFiles()
	if err != nil {
		return nil, err
	}

	secretClient := configa.NewInClusterSecretClient()
	parsed, err := configa.Parse[StartConfig](files, configa.WithSecretClient(secretClient))
	if err != nil {
		return nil, fmt.Errorf("parse backend config: %w", err)
	}

	cfg, err := parsed.Resolve(context.Background())
	if err != nil {
		return nil, fmt.Errorf("resolve backend config: %w", err)
	}

	applyStartDefaults(&cfg)

	if err := validateStart(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validateStart enforces conditional validation rules that depend on feature
// flags or environment mode and cannot be expressed as struct-level validate tags.
func validateStart(cfg *StartConfig) error {
	// Guardrail: Twilio must not be disabled in production. When disabled the
	// backend runs a no-op verification service that approves any OTP, which is
	// only ever acceptable outside prod.
	if cfg.Environment.IsModeProd() && !cfg.Twilio.Enabled {
		return errors.New("twilio.enabled must be true when environment.mode is prod")
	}

	// Guardrail: the Persona sandbox fake ZA ID generates synthetic identity
	// documents and must never be enabled in production.
	if cfg.Environment.IsModeProd() && cfg.Persona.SandboxFakeZAID {
		return errors.New("persona.sandbox_fake_za_id must be false when environment.mode is prod")
	}

	if cfg.Twilio.Enabled {
		if cfg.Twilio.AccountSID == "" {
			return errors.New("twilio.account_sid is required when twilio.enabled is true")
		}
		if cfg.Twilio.AccountToken == "" {
			return errors.New("twilio.account_token is required when twilio.enabled is true")
		}
		if cfg.Twilio.ServiceSID == "" {
			return errors.New("twilio.service_sid is required when twilio.enabled is true")
		}
	}

	if cfg.Email.Enabled {
		if cfg.Email.Sendgrid.APIKey == "" {
			return errors.New("email.sendgrid.api_key is required when email.enabled is true")
		}
		if cfg.Email.Sendgrid.FromName == "" {
			return errors.New("email.sendgrid.from_name is required when email.enabled is true")
		}
		if cfg.Email.Sendgrid.FromEmail == "" {
			return errors.New("email.sendgrid.from_email is required when email.enabled is true")
		}
		if cfg.Email.Sendgrid.OneTemplateID == "" {
			return errors.New("email.sendgrid.one_template_id is required when email.enabled is true")
		}
		if cfg.Email.Sendgrid.SupportEmail == "" {
			return errors.New("email.sendgrid.support_email is required when email.enabled is true")
		}
	}

	if cfg.PTI.Enabled {
		if cfg.PTI.BaseURL == "" {
			return errors.New("pti.base_url is required when pti.enabled is true")
		}
		if cfg.PTI.JWK == "" {
			return errors.New("pti.jwk is required when pti.enabled is true")
		}
		if cfg.PTI.ClientID == "" {
			return errors.New("pti.client_id is required when pti.enabled is true")
		}
		if cfg.PTI.SDKURL == "" {
			return errors.New("pti.sdk_url is required when pti.enabled is true")
		}
		if cfg.PTI.FormsURL == "" {
			return errors.New("pti.forms_url is required when pti.enabled is true")
		}
		if cfg.PTI.PublicKeyJWK == "" {
			return errors.New("pti.public_key_jwk is required when pti.enabled is true")
		}
	}

	if cfg.Plaid.Enabled {
		if cfg.Plaid.ClientID == "" {
			return errors.New("plaid.client_id is required when plaid.enabled is true")
		}
		if cfg.Plaid.Secret == "" {
			return errors.New("plaid.secret is required when plaid.enabled is true")
		}
		if cfg.Plaid.Env != "sandbox" && cfg.Plaid.Env != "production" {
			return errors.New("plaid.env must be one of: sandbox, production")
		}
		if len(cfg.Plaid.Products) == 0 {
			return errors.New("plaid.products is required when plaid.enabled is true")
		}
		if len(cfg.Plaid.CountryCodes) == 0 {
			return errors.New("plaid.country_codes is required when plaid.enabled is true")
		}
		if cfg.Plaid.Processor != "fiant" {
			return errors.New("plaid.processor must be: fiant")
		}
	}

	if cfg.Environment.IsModeProd() {
		if cfg.Xago.TravelRulePGPPublicKey == "" {
			return errors.New("xago.travel_rule_pgp_public_key is required in production")
		}
		if cfg.Xago.TravelRuleEmail == "" {
			return errors.New("xago.travel_rule_email is required in production")
		}
	}

	if cfg.Environment.IsModeProd() {
		if err := validateGatehubProd(cfg); err != nil {
			return err
		}
	}

	return nil
}

// validateGatehubProd checks GateHub fields that are required in production.
func validateGatehubProd(cfg *StartConfig) error {
	g := &cfg.Gatehub
	required := []struct {
		name  string
		value string
	}{
		{"gatehub.app_id", g.AppID},
		{"gatehub.secret", g.Secret},
		{"gatehub.card_app_id", g.CardAppID},
		{"gatehub.gateway_id", g.GatewayID},
		{"gatehub.card_account_product_code", g.CardAccountProductCode},
		{"gatehub.paywiser_euro_vault_id", g.PaywiserEuroVaultID},
		{"gatehub.sending_user_id", g.SendingUserID},
		{"gatehub.sending_user_address", g.SendingUserAddress},
		{"gatehub.on_off_ramp_client_id", g.OnOffRampClientID},
		{"gatehub.onboarding_client_id", g.OnboardingClientID},
		{"gatehub.exchange_client_id", g.ExchangeClientID},
		{"gatehub.api_base_url", g.APIBaseURL},
		{"gatehub.onboarding_base_url", g.OnboardingBaseURL},
		{"gatehub.on_off_ramp_base_url", g.OnOffRampBaseURL},
		{"gatehub.eur_ops_account", g.EUROpsAccount},
		{"gatehub.organization_id", g.OrganizationID},
		{"gatehub.omnibus_user_id_for_cp_xago", g.OmnibusUserIDForCPXago},
		{"gatehub.omnibus_user_address_for_cp_xago", g.OmnibusUserAddressForCPXago},
	}
	for _, r := range required {
		if r.value == "" {
			return fmt.Errorf("%s is required in production", r.name)
		}
	}
	if g.EUROpsLedgerID == 0 {
		return errors.New("gatehub.eur_ops_ledger_id is required in production")
	}
	if cfg.Rafiki.NodeEnabled {
		if g.IntermediaryUserID == "" {
			return errors.New("gatehub.intermediary_user_id is required in production when rafiki.node_enabled is true")
		}
		if g.IntermediaryUserAddress == "" {
			return errors.New("gatehub.intermediary_user_address is required in production when rafiki.node_enabled is true")
		}
	}
	return nil
}

func applyStartDefaults(cfg *StartConfig) {
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.LogOutputPath == "" {
		cfg.LogOutputPath = "stdout"
	}
	if cfg.Temporal.URL == "" {
		cfg.Temporal.URL = "temporal:7233"
	}
	if cfg.Kratos.URL == "" {
		cfg.Kratos.URL = "http://localhost:4433"
	}
	if cfg.Plaid.Enabled && cfg.Plaid.Processor == "" {
		cfg.Plaid.Processor = "fiant"
	}
}

// configFiles splits the CONFIG environment variable into a list of file paths.
// CONFIG must be a non-empty comma-separated list.
func configFiles() ([]string, error) {
	raw := os.Getenv("CONFIG")
	if raw == "" {
		return nil, fmt.Errorf("CONFIG environment variable is required (comma-separated list of YAML config file paths)")
	}
	var files []string
	for _, f := range strings.Split(raw, ",") {
		if f = strings.TrimSpace(f); f != "" {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("CONFIG environment variable contains no valid file paths")
	}
	return files, nil
}
