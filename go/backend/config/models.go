package config

// EnvironmentConfig describes how the running instance behaves and how it labels
// telemetry. Mode is the behavioural switch (prod, sandbox, dev, local, test);
// Label is the human-readable tag attached to monitoring signals (Sentry, OTEL, etc.).
type EnvironmentConfig struct {
	Mode  string `yaml:"mode"  validate:"required,oneof=prod sandbox dev local test"`
	Label string `yaml:"label"`
}

func (e EnvironmentConfig) IsModeLocal() bool   { return e.Mode == "local" }
func (e EnvironmentConfig) IsModeProd() bool    { return e.Mode == "prod" }
func (e EnvironmentConfig) IsModeDev() bool     { return e.Mode == "dev" }
func (e EnvironmentConfig) IsModeTest() bool    { return e.Mode == "test" }
func (e EnvironmentConfig) IsModeSandbox() bool { return e.Mode == "sandbox" }

// StartConfig is the typed configuration for the backend start and worker commands.
// It is loaded from YAML files listed in the CONFIG environment variable.
// Secrets may be embedded as {{ secret "k8s-secret-name" "key" }} template expressions,
// which configa resolves against the Kubernetes Secrets API at startup.
type StartConfig struct {
	Environment          EnvironmentConfig `yaml:"environment"`
	Port                 string            `yaml:"port"`
	ApplicationURL       string            `yaml:"application_url"       validate:"required"`
	OpenPaymentsBaseURL  string            `yaml:"open_payments_base_url" validate:"required"`
	AuthBaseURL          string            `yaml:"auth_base_url"          validate:"required"`
	LogLevel             string            `yaml:"log_level"`
	LogOutputPath        string            `yaml:"log_output_path"`
	AllowedWalletIDs     []string          `yaml:"allowed_wallet_ids"`
	BlockedRegions       []string          `yaml:"blocked_regions"`
	DeleteAccountEnabled bool              `yaml:"delete_account_enabled"`

	DB             DBConfig             `yaml:"db"`
	Kratos         KratosConfig         `yaml:"kratos"`
	Temporal       TemporalConfig       `yaml:"temporal"`
	Rafiki         RafikiConfig         `yaml:"rafiki"`
	Gatehub        GatehubConfig        `yaml:"gatehub"`
	Xago           XagoConfig           `yaml:"xago"`
	PTI            PTIConfig            `yaml:"pti"`
	Persona        PersonaConfig        `yaml:"persona"`
	Twilio         TwilioConfig         `yaml:"twilio"`
	Email          EmailConfig          `yaml:"email"`
	Slack          SlackConfig          `yaml:"slack"`
	Chimoney       ChimoneyConfig       `yaml:"chimoney"`
	Admin          AdminConfig          `yaml:"admin"`
	Mobile         MobileConfig         `yaml:"mobile"`
	Vault          VaultConfig          `yaml:"vault"`
	Sentry         SentryConfig         `yaml:"sentry"`
	Smarty         SmartyConfig         `yaml:"smarty"`
	Pusher         PusherConfig         `yaml:"pusher"`
	Segment        SegmentConfig        `yaml:"segment"`
	Agreements     AgreementsConfig     `yaml:"agreements"`
	OTEL           OTELConfig           `yaml:"otel"`
	Plaid          PlaidConfig          `yaml:"plaid"`
	WalletFeatures WalletFeaturesConfig `yaml:"wallet_features"`
}

// OTELConfig configures the OpenTelemetry trace exporter. Enabled is an explicit
// on/off switch for tracing. Endpoint and headers configure the OTLP/gRPC
// exporter, and let the Honeycomb key flow through configa's {{ secret ... }}
// templating like the rest of the config.
//
// The standard OTEL_EXPORTER_OTLP_* environment variables take priority over
// endpoint/headers when set (a warning is logged); see tracing.InitTraceProvider.
type OTELConfig struct {
	Enabled  bool              `yaml:"enabled"`
	Endpoint string            `yaml:"endpoint"`
	Headers  map[string]string `yaml:"headers"`
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
	Enabled            bool   `yaml:"enabled"`
	BaseURL            string `yaml:"base_url"`
	JWK                string `yaml:"jwk"`
	ClientID           string `yaml:"client_id"`
	SDKURL             string `yaml:"sdk_url"`
	FormsURL           string `yaml:"forms_url"`
	PublicKeyJWK       string `yaml:"public_key_jwk"`
	ScenarioTransfer   string `yaml:"scenario_transfer"`
	ScenarioDeposit    string `yaml:"scenario_deposit"`
	ScenarioWithdrawal string `yaml:"scenario_withdrawal"`
}

// PlaidConfig configures the Plaid bank-linking integration. Gated by Enabled;
// the remaining fields are required only when enabled. APIURL overrides the SDK
// base URL (e.g. to point at mockplaid locally) — empty selects the real Plaid
// environment matching Env.
type PlaidConfig struct {
	Enabled      bool     `yaml:"enabled"`
	ClientID     string   `yaml:"client_id"`
	Secret       string   `yaml:"secret"`
	Env          string   `yaml:"env"`
	Products     []string `yaml:"products"`
	CountryCodes []string `yaml:"country_codes"`
	Processor    string   `yaml:"processor"`
	APIURL       string   `yaml:"api_url"`
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

// MigrationConfig is the typed configuration for the backend migrate command.
// It is loaded from YAML files listed in the CONFIG environment variable.
type MigrationConfig struct {
	Environment         EnvironmentConfig `yaml:"environment"          validate:"required"`
	Agreements          AgreementsConfig  `yaml:"agreements"           validate:"required"`
	DBUrl               string            `yaml:"db_url"                validate:"required"`
	PacioliDBUrl        string            `yaml:"pacioli_db_url"        validate:"required"`
	OpenPaymentsBaseURL string            `yaml:"open_payments_base_url" validate:"required"`
	KratosUrl           string            `yaml:"kratos_url"`
	LogLevel            string            `yaml:"log_level"`
	LogOutputPath       string            `yaml:"log_output_path"`
	// Label is the telemetry tag attached to monitoring signals (Sentry, etc.).
	Label  string       `yaml:"label"`
	Sentry SentryConfig `yaml:"sentry"`
}

// WalletFeaturesConfig holds deployment-level feature-flag defaults. These seed the
// default value of a per-wallet flag when no wallet_features row exists yet; once
// an admin persists a wallet's features, the stored value takes precedence and
// these defaults no longer apply to that wallet.
type WalletFeaturesConfig struct {
	XagoGatehubPaymentsDefaultEnabled bool `yaml:"xago_gatehub_payments_default_enabled"`
}
