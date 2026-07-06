package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/interledger/interledger-app/go/tracing"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/log"
	"github.com/interledger/interledger-app/go/pacioli/cli"
	"github.com/interledger/interledger-app/go/pacioli/db"
	"github.com/interledger/interledger-app/go/pacioli/healthcheck"
	"github.com/interledger/interledger-app/go/pacioli/ledger"
	"github.com/interledger/interledger-app/go/pacioli/rpcserver"
	"github.com/interledger/interledger-app/go/pacioli/seed"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("Expected 'start' or 'migrate'.")
	}

	// Set the timezone globally
	time.Local = time.UTC

	command := args[1]
	switch command {
	case "start":
		args, err := cli.ParseStartArgs()
		if err != nil {
			log.Fatalln(err)
		}

		start(args)
	case "init":
		args, err := cli.ParseInitArgs()
		if err != nil {
			log.Fatalln(err)
		}

		runInit(args)
	default:
		log.Fatal("Unknown command", zap.String("command", command))
	}
}

func runInit(args *cli.InitArgs) {
	// run migrations
	err := db.Migrate(context.Background(), args.DbConnectionString)
	if err != nil {
		log.Fatalln(err)
	}

	db, err := sqlx.Connect("postgres", args.DbConnectionString)
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(db)

	if err != nil {
		log.Fatalln(err)
	}

	b := NewBackends(db)

	log.Info("tigerbeetle seeding starting")
	err = seed.Seed(b, args.SeedFile)
	if err != nil {
		log.Fatalln(err)
	}
	log.Info("tigerbeetle seeding complete")
}

func start(args *cli.StartArgs) {
	err := log.Initialize(args.LogLevel)
	if err != nil {
		log.Fatalln(err)
	}

	traceShutdown, err := tracing.InitTraceProvider("pacioli")
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		ctx := context.Background()
		if err := traceShutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider", zap.Error(err))
		}
	}()

	db, err := otelsqlx.Connect("postgres", args.DbConnectionString, otelsql.WithAttributes(semconv.DBSystemCockroachdb), otelsql.WithDBName("cockroachdb"))
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(db)

	if err != nil {
		log.Fatalln(err)
	}

	b := NewBackends(db)

	// Start time-ing out transactions
	go ledger.TimeoutTransfersForever(b)

	hs, err := healthcheck.NewService()
	if err != nil {
		log.Fatalln(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", args.Port))
	if err != nil {
		log.Fatalln(err)
	}
	log.Info(fmt.Sprintf("grpc server: 0.0.0.0:%s", args.Port))
	server := rpcserver.NewServer(b, hs)
	err = server.Serve(listener)
	if err != nil {
		log.Fatalln(err)
	}
}

type Backends interface {
	DB() *sqlx.DB
	Validator() *validator.Validate
}

var _ Backends = backends{}

type backends struct {
	db  *sqlx.DB
	val *validator.Validate
}

func NewBackends(db *sqlx.DB) Backends {
	return &backends{
		db:  db,
		val: validator.New(),
	}
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) Validator() *validator.Validate {
	return b.val
}
