package cli

import (
	"errors"
	"os"
)

type InitArgs struct {
	DbConnectionString string
	SeedFile           string
}

// We choose to specify the ledger's for the consuming services here so that
// it can be configured from deploy config.
func ParseInitArgs() (*InitArgs, error) {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		dbUrl = "cockroach://backend@cockroachdb-public:26257/pacioli?sslmode=verify-full&sslrootcert=/cockroach-certs/ca.crt&sslkey=/cockroach-certs/client.backend.key&sslcert=/cockroach-certs/client.backend.crt&max_conns=20&max_idle_conns=4"
	}

	seedFile := os.Getenv("TB_SEED_FILE")
	if seedFile == "" {
		return nil, errors.New("seed file not specified")
	}

	return &InitArgs{
		DbConnectionString: dbUrl,
		SeedFile:           seedFile,
	}, nil
}

type StartArgs struct {
	Port               string
	DbConnectionString string
	TbUrls             []string
	TbClusterID        uint32
	LogLevel           string
	LogOutputPath      string
	SentryRelease      string
	SentryEnvironment  string
	OtelEnabled        bool
}

func ParseStartArgs() (*StartArgs, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "443" // default port for grpc
	}
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		dbUrl = "cockroach://backend@cockroachdb-public:26257/pacioli?sslmode=verify-full&sslrootcert=/cockroach-certs/ca.crt&sslkey=/cockroach-certs/client.backend.key&sslcert=/cockroach-certs/client.backend.crt&max_conns=20&max_idle_conns=4"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logOutputPath := os.Getenv("LOG_OUTPUT_PATH")
	if logOutputPath == "" {
		logOutputPath = "stderr"
	}

	return &StartArgs{
		Port:               port,
		DbConnectionString: dbUrl,
		LogLevel:           logLevel,
		LogOutputPath:      logOutputPath,
		SentryRelease:      os.Getenv("SENTRY_RELEASE"),
		SentryEnvironment:  os.Getenv("SENTRY_ENVIRONMENT"),
		OtelEnabled:        os.Getenv("OTEL_ENABLED") == "true",
	}, nil
}
