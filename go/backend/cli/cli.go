package cli

import (
	"errors"
	"log"
	"os"
	"strconv"

	"gitlab.com/fynbos/env"

	"github.com/joho/godotenv"
)

type MigrationArgs struct {
	ConnectionString        string
	PacioliConnectionString string
	KratosUrl               string
	LogLevel                string
	LogOutputPath           string
}

func ParseMigrationArgs() (*MigrationArgs, error) {
	dbUrl := os.Getenv("DB_URL_WITH_CERTS")
	if dbUrl == "" {
		dbUrl = "cockroach://backend@cockroachdb-public:26257/backend?sslmode=verify-full&sslrootcert=/cockroach-certs/ca.crt&sslcert=/cockroach-certs/client.backend.crt&sslkey=/cockroach-certs/client.backend.key&max_conns=20&max_idle_conns=4"
	}

	pacDB := os.Getenv("PACIOLI_DB_URL_WITH_CERTS")
	if pacDB == "" {
		dbUrl = "cockroach://backend@cockroachdb-public:26257/pacioli?sslmode=verify-full&sslrootcert=/cockroach-certs/ca.crt&sslcert=/cockroach-certs/client.backend.crt&sslkey=/cockroach-certs/client.backend.key&max_conns=20&max_idle_conns=4"
	}

	kratosUrl := os.Getenv("KRATOS_URL")
	if kratosUrl == "" {
		kratosUrl = "http://localhost:4433"
	}
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logOutputPath := os.Getenv("LOG_OUTPUT_PATH")
	if logOutputPath == "" {
		logOutputPath = "stderr"
	}

	return &MigrationArgs{
		ConnectionString:        dbUrl,
		KratosUrl:               kratosUrl,
		LogLevel:                logLevel,
		LogOutputPath:           logOutputPath,
		PacioliConnectionString: pacDB,
	}, nil
}

type StartArgs struct {
	Port                          string
	DbConnectionString            string
	PacioliDBConString            string
	KratosUrl                     string
	KratosAdminUrl                string
	LogLevel                      string
	LogOutputPath                 string
	TemporalUrl                   string
	TwilioSid                     string
	TwilioSecret                  string
	TwilioServiceSid              string
	AdminPolicyAud                string
	AdminTeamDomain               string
	EmailEnabled                  bool
	SendgridAPIKey                string
	SendgridFromName              string
	SendgridFromEmail             string
	SendgridOneTemplateID         string
	SmartyAuthID                  string
	SmartyAuthToken               string
	PusherAddr                    string
	SegmentKey                    string
	TwitterClientID               string
	TwitterClientSecret           string
	TwitterRedirectURL            string
	TwitterBearerToken            string
	GatehubAppID                  string
	GatehubSecret                 string
	GatehubCardAppID              string
	GatehubGatewayID              string
	GatehubCardAccountProductCode string
	GatehubPaywiserEuroVaultID    string
	GatehubSendingUserID          string
	GatehubSendingUserAddress     string
	GatehubWebhookSecret          string
	GatehubFallbackWebhookURL     string
	GatehubOnOffRampClientID      string
	GatehubOnboardingClientID     string
	GatehubExchangeClientID       string
	GatehubAPIBaseURL             string
	GatehubOnboardingBaseURL      string
	GatehubOnOffRampBaseURL       string
	GatehubEUROpsAccount          string
	GatehubEUROpsLedgerID         uint32
	GatehubOrganizationID         string
	XagoAPIBaseURL                string
	XagoIdentityBaseURL           string
	XagoPublicKey                 string
	XagoSecret                    string
	XagoPolicyID                  string
	PTIEnabled                    bool
	PTIBaseURL                    string
	PTIJWK                        string
	PTIClientID                   string
	PTISDKURL                     string
	PTIFormsURL                   string
	PTIPublicKeyJWK               string
	PersonaBaseURL                string
	PersonaToken                  string
	PersonaWebhookToken           string
	AppleAppID                    string
	AndroidPackageName            string
	AndroidSHA256                 string
	OperatorTenantID              string
	AdminAPISecret                string
	SignatureVersion              string
}

func ParseStartArgs() (*StartArgs, error) {
	envFile := os.Getenv("ENV_FILE")
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			log.Fatal("Error loading .env file")
			return nil, err
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		return nil, errors.New("DB_URL is required.")
	}

	pacDB := os.Getenv("PACIOLI_DB_URL")
	if pacDB == "" {
		return nil, errors.New("PACIOLI_DB_URL is required.")
	}

	kratosUrl := os.Getenv("KRATOS_URL")
	if kratosUrl == "" {
		kratosUrl = "http://localhost:4433"
	}
	kratosAdminUrl := os.Getenv("KRATOS_ADMIN_URL")
	if kratosAdminUrl == "" {
		kratosAdminUrl = "http://localhost:4433"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logOutputPath := os.Getenv("LOG_OUTPUT_PATH")
	if logOutputPath == "" {
		logOutputPath = "stdout"
	}

	temporalUrl := os.Getenv("TEMPORAL_URL")
	if temporalUrl == "" {
		temporalUrl = "temporal:7233"
	}

	TwilioSid := os.Getenv("TWILIO_ACCOUNT_SID")
	if TwilioSid == "" {
		return nil, errors.New("TWILIO_ACCOUNT_SID is required.")
	}

	TwilioSecret := os.Getenv("TWILIO_ACCOUNT_TOKEN")
	if TwilioSecret == "" {
		return nil, errors.New("TWILIO_ACCOUNT_TOKEN is required.")
	}

	twilioServiceSid := os.Getenv("TWILIO_SERVICE_SID")
	if twilioServiceSid == "" {
		return nil, errors.New("TWILIO_SERVICE_SID is required.")
	}

	personaBaseURL := os.Getenv("PERSONA_BASE_URL")
	if personaBaseURL == "" {
		personaBaseURL = "https://api.withpersona.com/api/v1/"
	}

	personaToken := os.Getenv("PERSONA_TOKEN")
	if personaToken == "" {
		return nil, errors.New("PERSONA_TOKEN is required")
	}

	personaWebhook := os.Getenv("PERSONA_WEBHOOK_TOKEN")
	if personaWebhook == "" {
		return nil, errors.New("PERSONA_WEBHOOK_TOKEN is required")
	}

	twitterClientID := "DEPRECATED"
	twitterClientSecret := "DEPRECATED"
	twitterBearerToken := "DEPRECATED"
	twitterRedirectURL := "DEPRECATED"

	adminPolicyAud := os.Getenv("ADMIN_POLICY_AUD")
	if adminPolicyAud == "" {
		return nil, errors.New("ADMIN_POLICY_AUD is required")
	}

	adminTeamDomain := os.Getenv("ADMIN_TEAM_DOMAIN")
	if adminTeamDomain == "" {
		return nil, errors.New("ADMIN_TEAM_DOMAIN is required")
	}

	emailEnabled := true
	if v := os.Getenv("EMAIL_ENABLED"); v != "" {
		var err error
		emailEnabled, err = strconv.ParseBool(v)
		if err != nil {
			return nil, errors.New("EMAIL_ENABLED must be a valid boolean (true/false/1/0)")
		}
	}

	var sendgridAPIKey, sendgridFromName, sendgridFromEmail, sendgridOneTemplateID string
	if emailEnabled {
		sendgridAPIKey = os.Getenv("SENDGRID_API_KEY")
		if sendgridAPIKey == "" {
			return nil, errors.New("SENDGRID_API_KEY is required when EMAIL_ENABLED is true")
		}

		sendgridFromName = os.Getenv("SENDGRID_FROM_NAME")
		if sendgridFromName == "" {
			return nil, errors.New("SENDGRID_FROM_NAME is required when EMAIL_ENABLED is true")
		}

		sendgridFromEmail = os.Getenv("SENDGRID_FROM_EMAIL")
		if sendgridFromEmail == "" {
			return nil, errors.New("SENDGRID_FROM_EMAIL is required when EMAIL_ENABLED is true")
		}

		sendgridOneTemplateID = os.Getenv("SENDGRID_ONE_TEMPLATE_ID")
		if sendgridOneTemplateID == "" {
			return nil, errors.New("SENDGRID_ONE_TEMPLATE_ID is required when EMAIL_ENABLED is true")
		}
	}

	smartyAuthID := os.Getenv("SMARTY_AUTH_ID")
	if smartyAuthID == "" {
		return nil, errors.New("SMARTY_AUTH_ID is required")
	}

	smartyAuthToken := os.Getenv("SMARTY_AUTH_TOKEN")
	if smartyAuthToken == "" {
		return nil, errors.New("SMARTY_AUTH_TOKEN is required")
	}

	pusherAddr := os.Getenv("PUSHER_ADDR")
	if pusherAddr == "" {
		return nil, errors.New("PUSHER_ADDR is required")
	}

	segmentKey := os.Getenv("SEGMENT_KEY")
	if segmentKey == "" {
		return nil, errors.New("SEGMENT_KEY is required")
	}

	gatehubAppID := os.Getenv("GATEHUB_APP_ID")
	if gatehubAppID == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_APP_ID is required in production")
	}

	gatehubSecret := os.Getenv("GATEHUB_SECRET")
	if gatehubSecret == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_SECRET is required in production")
	}

	gatehubCardAppID := os.Getenv("GATEHUB_CARD_APP_ID")
	if gatehubCardAppID == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_CARD_APP_ID is required in production")
	}

	gatehubGatewayID := os.Getenv("GATEHUB_GATEWAY_ID")
	if gatehubGatewayID == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_GATEWAY_ID is required in production")
	}

	gatehubCardAccountProductCode := os.Getenv("GATEHUB_CARD_ACCOUNT_PRODUCT_CODE")
	if gatehubCardAccountProductCode == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_CARD_ACCOUNT_PRODUCT_CODE is required in production")
	}

	gatehubPaywiserEuroVaultID := os.Getenv("GATEHUB_PAYWISER_EURO_VAULT_ID")
	if gatehubPaywiserEuroVaultID == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_PAYWISER_EURO_VAULT_ID is required in production")
	}

	gatehubSendingUserID := os.Getenv("GATEHUB_SENDING_USER_ID")
	if gatehubSendingUserID == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_SENDING_USER_ID is required in production")
	}

	gatehubSendingUserAddress := os.Getenv("GATEHUB_SENDING_USER_ADDRESS")
	if gatehubSendingUserAddress == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_SENDING_USER_ADDRESS is required in production")
	}

	gatehubWebhookSecret := os.Getenv("GATEHUB_WEBHOOK_SECRET")
	// Webhook secret is optional but log if missing

	gatehubFallbackWebhookURL := os.Getenv("GATEHUB_FALLBACK_WEBHOOK_URL")
	// Fallback webhook URL is optional
	gatehubOnOffRampClientID := os.Getenv("GATEHUB_ON_OFF_RAMP_CLIENT_ID")
	if gatehubOnOffRampClientID == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_ON_OFF_RAMP_CLIENT_ID is required in production")
	}

	gatehubOnboardingClientID := os.Getenv("GATEHUB_ONBOARDING_CLIENT_ID")
	if gatehubOnboardingClientID == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_ONBOARDING_CLIENT_ID is required in production")
	}

	gatehubExchangeClientID := os.Getenv("GATEHUB_EXCHANGE_CLIENT_ID")
	if gatehubExchangeClientID == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_EXCHANGE_CLIENT_ID is required in production")
	}

	gatehubAPIBaseURL := os.Getenv("GATEHUB_API_BASE_URL")
	if gatehubAPIBaseURL == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_API_BASE_URL is required in production")
	}

	gatehubOnboardingBaseURL := os.Getenv("GATEHUB_ONBOARDING_BASE_URL")
	if gatehubOnboardingBaseURL == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_ONBOARDING_BASE_URL is required in production")
	}

	gatehubOnOffRampBaseURL := os.Getenv("GATEHUB_ON_OFF_RAMP_BASE_URL")
	if gatehubOnOffRampBaseURL == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_ON_OFF_RAMP_BASE_URL is required in production")
	}

	gatehubEUROpsAccount := os.Getenv("GATEHUB_EUR_OPS_ACCOUNT")
	if gatehubEUROpsAccount == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_EUR_OPS_ACCOUNT is required in production")
	}

	var gatehubEUROpsLedgerID uint32 = 0
	gatehubEUROpsLedgerIDStr := os.Getenv("GATEHUB_EUR_OPS_LEDGER_ID")
	if gatehubEUROpsLedgerIDStr != "" {
		id64, err := strconv.ParseUint(gatehubEUROpsLedgerIDStr, 10, 32)
		if err != nil {
			return nil, errors.New("GATEHUB_EUR_OPS_LEDGER_ID must be a valid uint32")
		}
		gatehubEUROpsLedgerID = uint32(id64)
	}
	if gatehubEUROpsLedgerID == 0 && env.IsProd() {
		return nil, errors.New("GATEHUB_EUR_OPS_LEDGER_ID is required in production")
	}

	gatehubOrganizationID := os.Getenv("GATEHUB_ORGANIZATION_ID")
	if gatehubOrganizationID == "" && env.IsProd() {
		return nil, errors.New("GATEHUB_ORGANIZATION_ID is required in production")
	}

	xagoAPIBaseURL := os.Getenv("XAGO_API_BASE_URL")
	if xagoAPIBaseURL == "" {
		return nil, errors.New("XAGO_API_BASE_URL is required")
	}

	xagoIdentityBaseURL := os.Getenv("XAGO_IDENTITY_BASE_URL")
	if xagoIdentityBaseURL == "" {
		return nil, errors.New("XAGO_IDENTITY_BASE_URL is required")
	}

	xagoPublicKey := os.Getenv("XAGO_API_PUBLIC_KEY")
	if xagoPublicKey == "" {
		return nil, errors.New("XAGO_API_PUBLIC_KEY is required")
	}

	xagoSecret := os.Getenv("XAGO_API_SECRET")
	if xagoSecret == "" {
		return nil, errors.New("XAGO_API_SECRET is required")
	}

	xagoPolicyID := os.Getenv("XAGO_POLICY_ID")
	if xagoPolicyID == "" {
		return nil, errors.New("XAGO_POLICY_ID is required")
	}

	ptiEnabled := os.Getenv("PTI_ENABLED") == "true"
	ptiBaseURL := os.Getenv("PTI_BASE_URL")
	ptiJWK := os.Getenv("PTI_JWK")
	ptiClientID := os.Getenv("PTI_CLIENT_ID")
	ptiSDKURL := os.Getenv("PTI_SDK_URL")
	ptiFormsURL := os.Getenv("PTI_FORMS_URL")
	ptiPublicKeyJWK := os.Getenv("PTI_PUBLIC_KEY_JWK")
	if ptiEnabled {
		if ptiBaseURL == "" {
			return nil, errors.New("PTI_BASE_URL is required when PTI_ENABLED=true")
		}
		if ptiJWK == "" {
			return nil, errors.New("PTI_JWK is required when PTI_ENABLED=true")
		}
		if ptiClientID == "" {
			return nil, errors.New("PTI_CLIENT_ID is required when PTI_ENABLED=true")
		}
		if ptiSDKURL == "" {
			return nil, errors.New("PTI_SDK_URL is required when PTI_ENABLED=true")
		}
		if ptiFormsURL == "" {
			return nil, errors.New("PTI_FORMS_URL is required when PTI_ENABLED=true")
		}
		if ptiPublicKeyJWK == "" {
			return nil, errors.New("PTI_PUBLIC_KEY_JWK is required when PTI_ENABLED=true")
		}
	}

	appleAppID := os.Getenv("APPLE_APP_ID")
	if appleAppID == "" {
		return nil, errors.New("APPLE_APP_ID is required")
	}

	androidPackageName := os.Getenv("ANDROID_PACKAGE_NAME")
	if androidPackageName == "" {
		return nil, errors.New("ANDROID_PACKAGE_NAME is required")
	}

	androidSHA256 := os.Getenv("ANDROID_SHA256")
	if androidSHA256 == "" {
		return nil, errors.New("ANDROID_SHA256 is required")
	}

	operatorTenantID := os.Getenv("OPERATOR_TENANT_ID")
	if operatorTenantID == "" {
		return nil, errors.New("OPERATOR_TENANT_ID is required")
	}

	adminAPISecret := os.Getenv("ADMIN_API_SECRET")
	if adminAPISecret == "" {
		return nil, errors.New("ADMIN_API_SECRET is required")
	}

	signatureVersion := os.Getenv("SIGNATURE_VERSION")
	if signatureVersion == "" {
		return nil, errors.New("SIGNATURE_VERSION is required")
	}

	return &StartArgs{
		Port:                          port,
		DbConnectionString:            dbUrl,
		PacioliDBConString:            pacDB,
		KratosUrl:                     kratosUrl,
		KratosAdminUrl:                kratosAdminUrl,
		LogLevel:                      logLevel,
		LogOutputPath:                 logOutputPath,
		TemporalUrl:                   temporalUrl,
		TwilioSid:                     TwilioSid,
		TwilioSecret:                  TwilioSecret,
		TwilioServiceSid:              twilioServiceSid,
		TwitterClientID:               twitterClientID,
		TwitterClientSecret:           twitterClientSecret,
		TwitterRedirectURL:            twitterRedirectURL,
		TwitterBearerToken:            twitterBearerToken,
		AdminPolicyAud:                adminPolicyAud,
		AdminTeamDomain:               adminTeamDomain,
		EmailEnabled:                  emailEnabled,
		SendgridAPIKey:                sendgridAPIKey,
		SendgridFromName:              sendgridFromName,
		SendgridFromEmail:             sendgridFromEmail,
		SendgridOneTemplateID:         sendgridOneTemplateID,
		SmartyAuthID:                  smartyAuthID,
		SmartyAuthToken:               smartyAuthToken,
		PusherAddr:                    pusherAddr,
		SegmentKey:                    segmentKey,
		GatehubAppID:                  gatehubAppID,
		GatehubSecret:                 gatehubSecret,
		GatehubCardAppID:              gatehubCardAppID,
		GatehubGatewayID:              gatehubGatewayID,
		GatehubCardAccountProductCode: gatehubCardAccountProductCode,
		GatehubPaywiserEuroVaultID:    gatehubPaywiserEuroVaultID,
		GatehubSendingUserID:          gatehubSendingUserID,
		GatehubSendingUserAddress:     gatehubSendingUserAddress,
		GatehubWebhookSecret:          gatehubWebhookSecret,
		GatehubFallbackWebhookURL:     gatehubFallbackWebhookURL,
		GatehubOnOffRampClientID:      gatehubOnOffRampClientID,
		GatehubOnboardingClientID:     gatehubOnboardingClientID,
		GatehubExchangeClientID:       gatehubExchangeClientID,
		GatehubAPIBaseURL:             gatehubAPIBaseURL,
		GatehubOnboardingBaseURL:      gatehubOnboardingBaseURL,
		GatehubOnOffRampBaseURL:       gatehubOnOffRampBaseURL,
		GatehubEUROpsAccount:          gatehubEUROpsAccount,
		GatehubEUROpsLedgerID:         gatehubEUROpsLedgerID,
		GatehubOrganizationID:         gatehubOrganizationID,
		XagoAPIBaseURL:                xagoAPIBaseURL,
		XagoIdentityBaseURL:           xagoIdentityBaseURL,
		XagoPublicKey:                 xagoPublicKey,
		XagoSecret:                    xagoSecret,
		XagoPolicyID:                  xagoPolicyID,
		PTIEnabled:                    ptiEnabled,
		PTIBaseURL:                    ptiBaseURL,
		PTIJWK:                        ptiJWK,
		PTIClientID:                   ptiClientID,
		PTISDKURL:                     ptiSDKURL,
		PTIFormsURL:                   ptiFormsURL,
		PTIPublicKeyJWK:               ptiPublicKeyJWK,
		PersonaBaseURL:                personaBaseURL,
		PersonaToken:                  personaToken,
		PersonaWebhookToken:           personaWebhook,
		AppleAppID:                    appleAppID,
		AndroidPackageName:            androidPackageName,
		AndroidSHA256:                 androidSHA256,
		OperatorTenantID:              operatorTenantID,
		AdminAPISecret:                adminAPISecret,
		SignatureVersion:              signatureVersion,
	}, nil
}
