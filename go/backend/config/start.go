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

// EnvironmentConfig describes how the running instance behaves and how it labels
// telemetry. Mode is the behavioural switch consumed by env.SetFynbosEnv;
// Label is the human-readable tag attached to monitoring signals (Sentry, OTEL, etc.).
type EnvironmentConfig struct {
	Mode  string `yaml:"mode"  validate:"required,oneof=prod sandbox dev local test"`
	Label string `yaml:"label"`
}

func (e EnvironmentConfig) IsModeLocal() bool { return e.Mode == "local" }
func (e EnvironmentConfig) IsModeProd() bool  { return e.Mode == "prod" }
func (e EnvironmentConfig) IsModeDev() bool   { return e.Mode == "dev" }
func (e EnvironmentConfig) IsModeTest() bool  { return e.Mode == "test" }

// IsTestExecution returns true when running under go test or when mode is "test".
func IsTestExecution(mode string) bool {
	return mode == "test" || flag.Lookup("test.v") != nil
}

// StartConfig is the typed configuration for the backend start and worker commands.
// It is loaded from YAML files listed in the CONFIG environment variable.
// Secrets may be embedded as {{ secret "k8s-secret-name" "key" }} template expressions,
// which configa resolves against the Kubernetes Secrets API at startup.
type StartConfig struct {
	Environment         EnvironmentConfig `yaml:"environment"`
	Port                string            `yaml:"port"`
	ApplicationURL      string            `yaml:"application_url"       validate:"required"`
	OpenPaymentsBaseURL string            `yaml:"open_payments_base_url" validate:"required"`
	AuthBaseURL         string            `yaml:"auth_base_url"          validate:"required"`
	LogLevel            string            `yaml:"log_level"`
	LogOutputPath       string            `yaml:"log_output_path"`
	AllowedWalletIDs    []string          `yaml:"allowed_wallet_ids"`
	BlockedRegions      []string          `yaml:"blocked_regions"`

	DB         DBConfig         `yaml:"db"`
	Kratos     KratosConfig     `yaml:"kratos"`
	Temporal   TemporalConfig   `yaml:"temporal"`
	Rafiki     RafikiConfig     `yaml:"rafiki"`
	Gatehub    GatehubConfig    `yaml:"gatehub"`
	Xago       XagoConfig       `yaml:"xago"`
	PTI        PTIConfig        `yaml:"pti"`
	Persona    PersonaConfig    `yaml:"persona"`
	Twilio     TwilioConfig     `yaml:"twilio"`
	Email      EmailConfig      `yaml:"email"`
	Slack      SlackConfig      `yaml:"slack"`
	Chimoney   ChimoneyConfig   `yaml:"chimoney"`
	Admin      AdminConfig      `yaml:"admin"`
	Mobile     MobileConfig     `yaml:"mobile"`
	Vault      VaultConfig      `yaml:"vault"`
	Sentry     SentryConfig     `yaml:"sentry"`
	Smarty     SmartyConfig     `yaml:"smarty"`
	Pusher     PusherConfig     `yaml:"pusher"`
	Segment    SegmentConfig    `yaml:"segment"`
	Agreements AgreementsConfig `yaml:"agreements"`
}

type AgreementsConfig struct {
	SignupAgreementIDs   []string `yaml:"signup_agreement_ids"`
	AgreementsFolderName string   `yaml:"folder_name" validate:"required"`
}

type DBConfig struct {
	URL        string `yaml:"url"        validate:"required"`
	PacioliURL string `yaml:"pacioli_url" validate:"required"`
}

type KratosConfig struct {
	URL      string `yaml:"url"       validate:"required"`
	AdminURL string `yaml:"admin_url" validate:"required"`
}

type TemporalConfig struct {
	URL string `yaml:"url"`
}

// RafikiConfig covers both the ILP node integration and the Rafiki-specific database
// connections used by background worker jobs.
type RafikiConfig struct {
	NodeEnabled       bool   `yaml:"node_enabled"`
	BackendGraphQLURL string `yaml:"backend_graphql_url" validate:"required"`
	AuthGraphQLURL    string `yaml:"auth_graphql_url"    validate:"required"`
	OperatorTenantID  string `yaml:"operator_tenant_id"  validate:"required"`
	AdminAPISecret    string `yaml:"admin_api_secret"    validate:"required"`
	SignatureVersion  string `yaml:"signature_version"   validate:"required"`
	DBURL             string `yaml:"db_url"`
	AuthDBURL         string `yaml:"auth_db_url"`
}

type GatehubConfig struct {
	AppID                   string `yaml:"app_id"`
	Secret                  string `yaml:"secret"`
	CardAppID               string `yaml:"card_app_id"`
	GatewayID               string `yaml:"gateway_id"`
	CardAccountProductCode  string `yaml:"card_account_product_code"`
	PaywiserEuroVaultID     string `yaml:"paywiser_euro_vault_id"`
	SendingUserID           string `yaml:"sending_user_id"`
	SendingUserAddress      string `yaml:"sending_user_address"`
	IntermediaryUserID      string `yaml:"intermediary_user_id"`
	IntermediaryUserAddress string `yaml:"intermediary_user_address"`
	WebhookSecret           string `yaml:"webhook_secret"`
	FallbackWebhookURL      string `yaml:"fallback_webhook_url"`
	OnOffRampClientID       string `yaml:"on_off_ramp_client_id"`
	OnboardingClientID      string `yaml:"onboarding_client_id"`
	ExchangeClientID        string `yaml:"exchange_client_id"`
	APIBaseURL              string `yaml:"api_base_url"`
	OnboardingBaseURL       string `yaml:"onboarding_base_url"`
	OnOffRampBaseURL        string `yaml:"on_off_ramp_base_url"`
	EUROpsAccount           string `yaml:"eur_ops_account"`
	EUROpsLedgerID          uint32 `yaml:"eur_ops_ledger_id"`
	OrganizationID          string `yaml:"organization_id"`
}

type XagoConfig struct {
	APIBaseURL      string `yaml:"api_base_url"      validate:"required"`
	IdentityBaseURL string `yaml:"identity_base_url" validate:"required"`
	APIPublicKey    string `yaml:"api_public_key"    validate:"required"`
	APISecret       string `yaml:"api_secret"        validate:"required"`
	PolicyID        string `yaml:"policy_id"         validate:"required"`
}

type PTIConfig struct {
	Enabled      bool   `yaml:"enabled"`
	BaseURL      string `yaml:"base_url"`
	JWK          string `yaml:"jwk"`
	ClientID     string `yaml:"client_id"`
	SDKURL       string `yaml:"sdk_url"`
	FormsURL     string `yaml:"forms_url"`
	PublicKeyJWK string `yaml:"public_key_jwk"`
}

type PersonaConfig struct {
	BaseURL         string `yaml:"base_url"         validate:"required"`
	Token           string `yaml:"token"            validate:"required"`
	WebhookToken    string `yaml:"webhook_token"    validate:"required"`
	SandboxFakeZAID bool   `yaml:"sandbox_fake_za_id"`
}

type TwilioConfig struct {
	Enabled      bool   `yaml:"enabled"`
	AccountSID   string `yaml:"account_sid"`
	AccountToken string `yaml:"account_token"`
	ServiceSID   string `yaml:"service_sid"`
}

type SendgridConfig struct {
	APIKey        string `yaml:"api_key"`
	FromName      string `yaml:"from_name"`
	FromEmail     string `yaml:"from_email"`
	OneTemplateID string `yaml:"one_template_id"`
	SupportEmail  string `yaml:"support_email"`
}

type EmailConfig struct {
	Enabled  bool           `yaml:"enabled"`
	Sendgrid SendgridConfig `yaml:"sendgrid"`
}

type SlackConfig struct {
	Token              string `yaml:"token"`
	ChannelSignupKYC   string `yaml:"channel_signup_kyc"`
	ChannelTransaction string `yaml:"channel_transaction"`
	ChannelError       string `yaml:"channel_error"`
	ClientID           string `yaml:"client_id"`
	ClientSecret       string `yaml:"client_secret"`
}

type ChimoneyConfig struct {
	Token         string `yaml:"token"`
	WebhookSecret string `yaml:"webhook_secret"`
}

type AdminConfig struct {
	PolicyAud  string `yaml:"policy_aud"  validate:"required"`
	TeamDomain string `yaml:"team_domain" validate:"required"`
	BaseURL    string `yaml:"base_url"    validate:"required"`
}

type MobileConfig struct {
	AppleAppID         string `yaml:"apple_app_id"          validate:"required"`
	AndroidPackageName string `yaml:"android_package_name"  validate:"required"`
	AndroidSHA256      string `yaml:"android_sha256"        validate:"required"`
}

type VaultConfig struct {
	Addr              string `yaml:"addr"`
	TransitEnginePath string `yaml:"transit_engine_path"`
	Token             string `yaml:"token"`
}

type SentryConfig struct {
	DSN string `yaml:"dsn"`
}

type SmartyConfig struct {
	AuthID    string `yaml:"auth_id"    validate:"required"`
	AuthToken string `yaml:"auth_token" validate:"required"`
}

type PusherConfig struct {
	Addr string `yaml:"addr" validate:"required"`
}

type SegmentConfig struct {
	Key string `yaml:"key" validate:"required"`
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

	if cfg.Environment.Mode == "prod" {
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
