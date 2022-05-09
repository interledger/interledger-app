package cli

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"gitlab.com/fynbos/backend/migrations"
)

type MigrationArgs struct {
	ConnectionString    string
	NoopLedgerID        uint16
	NoopEquityAccountID string
	UnitWebhookToken		string
	PacioliUrl          string
	UsdLedgerID         uint16
	KratosUrl           string
	LogLevel            string
	LogOutputPath       string
}

func ParseMigrationArgs() (*MigrationArgs, error) {
	baseDbUrl := os.Getenv("DB_URL")
	if baseDbUrl == "" {
		baseDbUrl = "cockroach://backend@cockroachdb-public:26257/backend?sslmode=verify-full&max_conns=20&max_idle_conns=4"
	}

	connString, err := migrations.InlineSslCreds(
		strings.Replace(baseDbUrl, "cockroach", "postgres", 1), // replace cockroach protocol with postgres so that we can use pq driver.
		"/cockroach-certs/client.backend.key",
		"/cockroach-certs/client.backend.crt",
		"/cockroach-certs/ca.crt",
	)
	if err != nil {
		log.Fatalln(err)
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

	pacioliUrl := os.Getenv("PACIOLI_URL")
	if pacioliUrl == "" {
		pacioliUrl = "pacioli:443"
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
		ConnectionString:    connString,
		PacioliUrl:          pacioliUrl,
		NoopEquityAccountID: noopEquityAccount,
		NoopLedgerID:        uint16(usdLedgerID),
		UsdLedgerID:         uint16(usdLedgerID),
		KratosUrl:           kratosUrl,
		LogLevel:            logLevel,
		LogOutputPath:       logOutputPath,
	}, nil
}

type StartArgs struct {
	Port                 string
	DbConnectionString   string
	KratosUrl            string
	PacioliUrl           string
	UsdLedgerID          uint16
	LogLevel             string
	LogOutputPath        string
	NoopLedgerID         uint16
	NoopEquityAccountID  string
	UnitWebhookToken 	   string
	GoogleOauth2ClientID string
}

func ParseStartArgs() (*StartArgs, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	baseDbUrl := os.Getenv("DB_URL")
	if baseDbUrl == "" {
		baseDbUrl = "cockroach://pacioli@cockroachdb-public:26257/pacioli?sslmode=verify-full&max_conns=20&max_idle_conns=4"
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
	pacioliUrl := os.Getenv("PACIOLI_URL")
	if pacioliUrl == "" {
		pacioliUrl = "pacioli:443"
	}
	usdLedgerIDString := os.Getenv("USD_LEDGER_ID")
	if usdLedgerIDString == "" {
		return nil, errors.New("USD ledger code not specified.")
	}
	usdLedgerID, err := strconv.ParseUint(usdLedgerIDString, 10, 16)
	if err != nil {
		return nil, errors.New("USD_LEDGER_ID must be a uint16.")
	}

	connString, err := migrations.InlineSslCreds(
		strings.Replace(baseDbUrl, "cockroach", "postgres", 1), // replace cockroach protocol with postgres so that we can use pq driver.
		"/cockroach-certs/client.backend.key",
		"/cockroach-certs/client.backend.crt",
		"/cockroach-certs/ca.crt",
	)
	if err != nil {
		return nil, err
	}

	noopEquityAccount := os.Getenv("NOOP_EQUITY_ACCOUNT_ID")
	if noopEquityAccount == "" {
		return nil, errors.New("NOOP_EQUITY_ACCOUNT_ID is required.")
	}

	unitWebhookToken := os.Getenv("UNIT_WEBHOOK_TOKEN")
	if unitWebhookToken == "" {
		return nil, errors.New("UNIT_WEBHOOK_TOKEN is required")
	}

	googleOauth2ClientID := os.Getenv("GOOGLE_OUATH2_CLIENT_ID")
	if googleOauth2ClientID == "" {
		return nil, errors.New("GOOGLE_OUATH2_CLIENT_ID is required.")
	}

	return &StartArgs{
		Port:                 port,
		DbConnectionString:   connString,
		KratosUrl:            kratosUrl,
		PacioliUrl:           pacioliUrl,
		UsdLedgerID:          uint16(usdLedgerID),
		LogLevel:             logLevel,
		LogOutputPath:        logOutputPath,
		NoopLedgerID:         uint16(usdLedgerID), // all on the same ledger at the moment.
		NoopEquityAccountID:  noopEquityAccount,
		UnitWebhookToken:     unitWebhookToken,
		GoogleOauth2ClientID: googleOauth2ClientID,
	}, nil
}
