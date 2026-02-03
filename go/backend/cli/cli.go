package cli

import (
	"errors"
	"log"
	"os"

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
	Port                string
	AuthorisationPort   string
	DbConnectionString  string
	PacioliDBConString  string
	KratosUrl           string
	KratosAdminUrl      string
	LogLevel            string
	LogOutputPath       string
	TemporalUrl         string
	TwilioSid           string
	TwilioSecret        string
	TwilioServiceSid    string
	ZendeskUser         string
	ZendeskToken        string
	AdminPolicyAud      string
	AdminTeamDomain     string
	SendgridAPIKey      string
	SmartyAuthID        string
	SmartyAuthToken     string
	PusherAddr          string
	SegmentKey          string
	TwitterClientID     string
	TwitterClientSecret string
	TwitterRedirectURL  string
	TwitterBearerToken  string
	DiscordClientID              string
	DiscordClientSecret          string
	DiscordRedirectURL           string
	GatehubAppID                 string
	GatehubSecret                string
	GatehubCardAppID             string
	GatehubGatewayID             string
	GatehubCardAccountProductCode string
	GatehubPaywiserEuroVaultID   string
	GatehubSendingUserID         string
	GatehubSendingUserAddress    string
	GatehubWebhookSecret         string
	GatehubFallbackWebhookURL    string
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

	authorisationPort := os.Getenv("AUTHORISATION_PORT")
	if authorisationPort == "" {
		authorisationPort = "8082"
	}
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		dbUrl = "cockroach://backend@cockroachdb-public:26257/backend?sslmode=verify-full&sslrootcert=/cockroach-certs/ca.crt&sslcert=/cockroach-certs/client.backend.crt&sslkey=/cockroach-certs/client.backend.key&max_conns=20&max_idle_conns=4"
	}

	pacDB := os.Getenv("PACIOLI_DB_URL")
	if pacDB == "" {
		dbUrl = "cockroach://backend@cockroachdb-public:26257/pacioli?sslmode=verify-full&sslrootcert=/cockroach-certs/ca.crt&sslcert=/cockroach-certs/client.backend.crt&sslkey=/cockroach-certs/client.backend.key&max_conns=20&max_idle_conns=4"
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

	zendeskUser := os.Getenv("ZENDESK_USER")
	if zendeskUser == "" {
		return nil, errors.New("ZENDESK_USER is required, provide an email address")
	}

	zendeskToken := os.Getenv("ZENDESK_TOKEN")
	if zendeskToken == "" {
		return nil, errors.New("ZENDESK_TOKEN is required")
	}

	personaToken := os.Getenv("PERSONA_TOKEN")
	if personaToken == "" && env.IsProd() {
		return nil, errors.New("PERSONA_TOKEN is required in prod")
	}

	personaWebhook := os.Getenv("PERSONA_WEBHOOK_TOKEN")
	if personaWebhook == "" && env.IsProd() {
		return nil, errors.New("PERSONA_WEBHOOK_TOKEN is required in prod")
	}

	twitterClientId := os.Getenv("TWITTER_CLIENT_ID")
	if twitterClientId == "" && env.IsProd() {
		return nil, errors.New("TWITTER_CLIENT_ID is required")
	}

	twitterClientSecret := os.Getenv("TWITTER_CLIENT_SECRET")
	if twitterClientSecret == "" && env.IsProd() {
		return nil, errors.New("TWITTER_CLIENT_SECRET is required")
	}

	twitterBearerToken := os.Getenv("TWITTER_BEARER_TOKEN")
	if twitterBearerToken == "" && env.IsProd() {
		return nil, errors.New("TWITTER_BEARER_TOKEN is required")
	}

	twitterRedirectURL := os.Getenv("TWITTER_REDIRECT_URL")
	if twitterClientSecret == "" && env.IsProd() {
		return nil, errors.New("TWITTER_REDIRECT_URL is required")
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

	return &StartArgs{
		Port:                port,
		AuthorisationPort:   authorisationPort,
		DbConnectionString:  dbUrl,
		PacioliDBConString:  pacDB,
		KratosUrl:           kratosUrl,
		KratosAdminUrl:      kratosAdminUrl,
		LogLevel:            logLevel,
		LogOutputPath:       logOutputPath,
		TemporalUrl:         temporalUrl,
		TwilioSid:           TwilioSid,
		TwilioSecret:        TwilioSecret,
		TwilioServiceSid:    twilioServiceSid,
		ZendeskUser:         zendeskUser,
		ZendeskToken:        zendeskToken,
		TwitterClientID:     twitterClientId,
		TwitterClientSecret: twitterClientSecret,
		TwitterRedirectURL:  twitterRedirectURL,
		TwitterBearerToken:  twitterBearerToken,
		AdminPolicyAud:      os.Getenv("ADMIN_POLICY_AUD"),
		AdminTeamDomain:     os.Getenv("ADMIN_TEAM_DOMAIN"),
		SendgridAPIKey:      os.Getenv("SENDGRID_API_KEY"),
		SmartyAuthID:        os.Getenv("SMARTY_AUTH_ID"),
		SmartyAuthToken:     os.Getenv("SMARTY_AUTH_TOKEN"),
		PusherAddr:          os.Getenv("PUSHER_ADDR"),
		SegmentKey:                   os.Getenv("SEGMENT_KEY"),
		DiscordClientID:              os.Getenv("DISCORD_CLIENT_ID"),
		DiscordClientSecret:          os.Getenv("DISCORD_CLIENT_SECRET"),
		DiscordRedirectURL:           os.Getenv("DISCORD_REDIRECT_URL"),
		GatehubAppID:                 gatehubAppID,
		GatehubSecret:                gatehubSecret,
		GatehubCardAppID:             gatehubCardAppID,
		GatehubGatewayID:             gatehubGatewayID,
		GatehubCardAccountProductCode: gatehubCardAccountProductCode,
		GatehubPaywiserEuroVaultID:   gatehubPaywiserEuroVaultID,
		GatehubSendingUserID:         gatehubSendingUserID,
		GatehubSendingUserAddress:    gatehubSendingUserAddress,
		GatehubWebhookSecret:         gatehubWebhookSecret,
		GatehubFallbackWebhookURL:    gatehubFallbackWebhookURL,
	}, nil
}
