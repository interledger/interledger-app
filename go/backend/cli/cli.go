package cli

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type MigrationArgs struct {
	ConnectionString string
	UsdLedgerID      uint32
	KratosUrl        string
	LogLevel         string
	LogOutputPath    string
}

func ParseMigrationArgs() (*MigrationArgs, error) {
	dbUrl := os.Getenv("DB_URL_WITH_CERTS")
	if dbUrl == "" {
		dbUrl = "cockroach://backend@cockroachdb-public:26257/backend?sslmode=verify-full&sslrootcert=/cockroach-certs/ca.crt&sslcert=/cockroach-certs/client.backend.crt&sslkey=/cockroach-certs/client.backend.key&max_conns=20&max_idle_conns=4"
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

	usdLedgerIDString := os.Getenv("USD_LEDGER_ID")
	if usdLedgerIDString == "" {
		return nil, errors.New("USD ledger code not specified.")
	}
	usdLedgerID, err := strconv.ParseUint(usdLedgerIDString, 10, 16)
	if err != nil {
		return nil, errors.New("USD_LEDGER_ID must be a uint16.")
	}

	noopEquityAccount := os.Getenv("NOOP_EQUITY_ACCOUNT_ID")
	if noopEquityAccount == "" {
		return nil, errors.New("NOOP_EQUITY_ACCOUNT_ID is required.")
	}

	return &MigrationArgs{
		ConnectionString: dbUrl,
		UsdLedgerID:      uint32(usdLedgerID),
		KratosUrl:        kratosUrl,
		LogLevel:         logLevel,
		LogOutputPath:    logOutputPath,
	}, nil
}

type StartArgs struct {
	Port                 string
	OpenPaymentsPort     string
	DbConnectionString   string
	KratosUrl            string
	KratosAdminUrl       string
	UsdLedgerID          uint32
	LogLevel             string
	LogOutputPath        string
	MachnetClientID      string
	MachnetClientSecret  string
	MachnetWebhookSecret string
	RafikiGraphqlUrl     string
	TemporalUrl          string
	TwilioSid            string
	TwilioSecret         string
	TwilioServiceSid     string
	ZendeskUser          string
	ZendeskToken         string
	AdminPolicyAud       string
	AdminTeamDomain      string
	SendgridAPIKey       string
	SmartyAuthID         string
	SmartyAuthToken      string
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
	openPaymentsPort := os.Getenv("OPEN_PAYMENTS_PORT")
	if openPaymentsPort == "" {
		openPaymentsPort = "8081"
	}
	dbUrl := os.Getenv("DB_URL_WITH_CERTS")
	if dbUrl == "" {
		dbUrl = "cockroach://backend@cockroachdb-public:26257/backend?sslmode=verify-full&sslrootcert=/cockroach-certs/ca.crt&sslcert=/cockroach-certs/client.backend.crt&sslkey=/cockroach-certs/client.backend.key&max_conns=20&max_idle_conns=4"
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
		logOutputPath = "stderr"
	}
	usdLedgerIDString := os.Getenv("USD_LEDGER_ID")
	if usdLedgerIDString == "" {
		return nil, errors.New("USD ledger code not specified.")
	}
	usdLedgerID, err := strconv.ParseUint(usdLedgerIDString, 10, 32)
	if err != nil {
		return nil, errors.New("USD_LEDGER_ID must be a uint16.")
	}
	if err != nil {
		log.Fatalln(err)
	}

	rafikiGraphqlUrl := os.Getenv("RAFIKI_GRAPHQL_URL")
	if rafikiGraphqlUrl == "" {
		return nil, errors.New("text RAFIKI_GRAPHQL_URL is required.")
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

	return &StartArgs{
		Port:                 port,
		OpenPaymentsPort:     openPaymentsPort,
		DbConnectionString:   dbUrl,
		KratosUrl:            kratosUrl,
		KratosAdminUrl:       kratosAdminUrl,
		UsdLedgerID:          uint32(usdLedgerID),
		LogLevel:             logLevel,
		LogOutputPath:        logOutputPath,
		MachnetClientID:      os.Getenv("MACHNET_CLIENT_ID"),
		MachnetClientSecret:  os.Getenv("MACHNET_CLIENT_SECRET"),
		MachnetWebhookSecret: os.Getenv("MACHNET_WEBHOOK_SECRET"),
		RafikiGraphqlUrl:     rafikiGraphqlUrl,
		TemporalUrl:          temporalUrl,
		TwilioSid:            TwilioSid,
		TwilioSecret:         TwilioSecret,
		TwilioServiceSid:     twilioServiceSid,
		ZendeskUser:          zendeskUser,
		ZendeskToken:         zendeskToken,
		AdminPolicyAud:       os.Getenv("ADMIN_POLICY_AUD"),
		AdminTeamDomain:      os.Getenv("ADMIN_TEAM_DOMAIN"),
		SendgridAPIKey:       os.Getenv("SENDGRID_API_KEY"),
		SmartyAuthID:         os.Getenv("SMARTY_AUTH_ID"),
		SmartyAuthToken:      os.Getenv("SMARTY_AUTH_TOKEN"),
	}, nil
}
