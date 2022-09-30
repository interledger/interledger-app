package main

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"gitlab.com/fynbos/backend/kyc"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	kratos "github.com/ory/kratos-client-go"
	"github.com/riandyrn/otelchi"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	agreements_client "gitlab.com/fynbos/backend/agreements/client"
	agreements_migrations "gitlab.com/fynbos/backend/agreements/migrations"
	"gitlab.com/fynbos/backend/cli"
	"gitlab.com/fynbos/backend/country"
	country_client "gitlab.com/fynbos/backend/country/client"
	_grpc "gitlab.com/fynbos/backend/grpc"
	"gitlab.com/fynbos/backend/healthcheck"
	kyc_client "gitlab.com/fynbos/backend/kyc/client"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linked_account_client "gitlab.com/fynbos/backend/linkedaccounts/client"
	"gitlab.com/fynbos/backend/migrations"
	open_server "gitlab.com/fynbos/backend/openpayments/server"
	"gitlab.com/fynbos/backend/providers/fakecash"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_webhook "gitlab.com/fynbos/backend/providers/machnet/webhook"
	"gitlab.com/fynbos/backend/providers/rafiki"
	"gitlab.com/fynbos/backend/signup"
	signup_client "gitlab.com/fynbos/backend/signup/client"
	"gitlab.com/fynbos/backend/supporttickets"
	support_client "gitlab.com/fynbos/backend/supporttickets/client"
	"gitlab.com/fynbos/backend/temporal"
	_twilio "gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client"
	"gitlab.com/fynbos/backend/waitlist"
	waitlist_client "gitlab.com/fynbos/backend/waitlist/client"
	"gitlab.com/fynbos/log"
	"gitlab.com/fynbos/pacioli"
	pacioli_client "gitlab.com/fynbos/pacioli/client"
	pacioli_migrations "gitlab.com/fynbos/pacioli/migrations"
	"gitlab.com/fynbos/tracing"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var fs embed.FS

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Expected `start` or `migrate`.")
	}

	// Set the timezone globally
	time.Local = time.UTC

	command := os.Args[1]
	switch command {
	case "migrate":
		args, err := cli.ParseMigrationArgs()
		if err != nil {
			log.Fatalln(err)
		}
		migrate(args)
	case "start":
		args, err := cli.ParseStartArgs()
		if err != nil {
			log.Fatalln(err)
		}
		start(args)
	case "worker":
		args, err := cli.ParseStartArgs()
		if err != nil {
			log.Fatalln(err)
		}
		startWorker(args)
	default:
		log.Fatal("Unknown command:", zap.String("command", command))
	}
}

func start(args *cli.StartArgs) {
	var b = new(backends)
	b.val = validator.New()

	traceShutdown, err := tracing.InitTraceProvider("backend")
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
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalln(err)
		}
	}()
	b.db = db

	cfg := zap.NewProductionConfig()
	err = cfg.Level.UnmarshalText([]byte(args.LogLevel))
	if err != nil {
		log.Fatalln(err)
	}
	cfg.OutputPaths = []string{args.LogOutputPath}
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalln(err)
	}
	log.Setup(logger)

	configuration := kratos.NewConfiguration()
	configuration.Servers = kratos.ServerConfigurations{
		{
			URL:         args.KratosUrl,
			Description: "Dev Kratos",
		},
	}

	tp, err := temporal.NewTemporalClient(args.TemporalUrl)
	if err != nil {
		log.Fatalln(err)
	}
	b.temporal = tp

	b.users = user_client.New(b, args.KratosUrl)

	b.countries = country_client.New(b)

	pacioliDb, err := otelsqlx.Connect("postgres", args.DbConnectionString, otelsql.WithAttributes(semconv.DBSystemCockroachdb), otelsql.WithDBName("pacioli"))
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := pacioliDb.Close(); err != nil {
			log.Fatalln(err)
		}
	}()
	b.pacioli = newLocalPacioliClient(pacioliDb)

	twilioService, err := _twilio.NewService(&_twilio.ServiceArgs{
		AccountSid:   args.TwilioSid,
		AccountToken: args.TwilioSecret,
		ServiceSid:   args.TwilioServiceSid,
	})
	if err != nil {
		log.Fatalln(err)
	}
	b.twilio = twilioService

	//b.mxProvider = mx_client.New(b, args.MxClientID, args.MxApiKey)

	b.linkedaccounts = linked_account_client.New(b, logger)

	b.signup = signup_client.New(b)

	b.waitlist = waitlist_client.New(b, logger)

	rafikiProvider, err := rafiki.NewService(&rafiki.ServiceArgs{
		Db:  db,
		Url: args.RafikiGraphqlUrl,
	})
	if err != nil {
		log.Fatalln(err)
	}
	b.rafiki = rafikiProvider

	router := chi.NewRouter()
	router.Routes()
	router.Use(otelchi.Middleware("backend", otelchi.WithChiRoutes(router)))
	router.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	router.Handle("/webhooks/machnet", machnet_webhook.New(b))

	open_server.StartOpenPaymentsHTTP(b, args.OpenPaymentsPort)

	log.Info("connect to http://localhost:%s/playground for GraphQL playground", zap.String("port", args.Port))
	go func() {
		log.Fatalln(http.ListenAndServe(":"+args.Port, router))
	}()

	health, err := healthcheck.NewService()
	if err != nil {
		log.Fatalln(err)
	}
	b.healthcheck = health

	adminUsers, err := auth.NewService(args.GoogleOauth2ClientID)
	if err != nil {
		log.Fatalln(err)
	}
	b.adminAuth = auth.NewLoggingService(adminUsers, logger)

	b.agreements = agreements_client.New(b)

	b.supportTickets = support_client.NewClient(b, args.ZendeskUser, args.ZendeskToken)

	b.kyc = kyc_client.New(b)

	server, err := _grpc.NewServer(b)
	if err != nil {
		log.Fatalln(err)
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", "8443"))
	if err != nil {
		log.Fatalln(err)
	}
	log.Info(fmt.Sprintf("grpc server: 0.0.0.0:%s", "8443"))
	err = server.Serve(listener)
	if err != nil {
		log.Fatalln(err)
	}
}

func migrate(args *cli.MigrationArgs) {
	err := migrations.MigrateFromEmbeddedFiles(args.ConnectionString, fs)
	if err != nil {
		log.Fatalln(err)
	}

	db, err := sqlx.Connect("postgres", args.ConnectionString)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	err = agreements_migrations.MigrateFromEmbeddedMarkdowns(context.Background(), db)
	if err != nil {
		log.Fatalln(err)
	}

	err = pacioli_migrations.Migrate(args.PacioliDbConnectionString)
	if err != nil {
		log.Fatalln(err)
	}
}

func startWorker(args *cli.StartArgs) {
	log.Info("begin worker start")

	traceShutdown, err := tracing.InitTraceProvider("backend-worker")
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		ctx := context.Background()
		if err := traceShutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider", zap.Error(err))
		}
	}()

	var b = new(backends)
	b.val = validator.New()

	db, err := otelsqlx.Connect("postgres", args.DbConnectionString, otelsql.WithAttributes(semconv.DBSystemCockroachdb), otelsql.WithDBName("cockroachdb"))
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalln(err)
		}
	}()
	b.db = db

	cfg := zap.NewProductionConfig()
	err = cfg.Level.UnmarshalText([]byte(args.LogLevel))
	if err != nil {
		log.Fatalln(err)
	}
	cfg.OutputPaths = []string{args.LogOutputPath}
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalln(err)
	}
	log.Setup(logger)

	b.countries = country_client.New(b)

	//pClient, err := pacioli_client.New(args.PacioliUrl)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//b.pacioli = pClient

	tp, err := temporal.NewTemporalClient(args.TemporalUrl)
	if err != nil {
		log.Fatalln(err)
	}
	b.temporal = tp

	twilioService, err := _twilio.NewService(&_twilio.ServiceArgs{
		AccountSid:   args.TwilioSid,
		AccountToken: args.TwilioSecret,
		ServiceSid:   args.TwilioServiceSid,
	})
	if err != nil {
		log.Fatalln(err)
	}
	b.twilio = twilioService

	//mxImpl := mx_client.New(b, args.MxClientID, args.MxApiKey)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//b.mxProvider = mxImpl

	b.linkedaccounts = linked_account_client.New(b, logger)

	b.signup = signup_client.New(b)

	log.Info("Worker creating")
	w, err := temporal.NewTemporalWorker(b)
	if err != nil {
		log.Fatalln(err)
	}

	err = w.Run(worker.InterruptCh())
	log.Info("Worker started")
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}
}

type backends struct {
	val *validator.Validate
	db  *sqlx.DB

	adminAuth      auth.Service
	agreements     agreements.Client
	countries      country.Client
	fakecash       fakecash.Client
	linkedaccounts linkedaccounts.Client
	machnet        machnet.Client
	healthcheck    healthcheck.Service
	signup         signup.Client
	pacioli        pacioli.Client
	rafiki         rafiki.Service
	supportTickets supporttickets.Client
	temporal       client.Client
	twilio         _twilio.Service
	users          user.Client
	waitlist       waitlist.Client
	kyc            kyc.Client
}

func (b backends) KYC() kyc.Client {
	return b.kyc
}

func (b backends) HealthCheck() healthcheck.Service {
	return b.healthcheck
}

func (b backends) Signup() signup.Client {
	return b.signup
}

func (b backends) AdminAuth() auth.Service {
	return b.adminAuth
}

func (b backends) Agreements() agreements.Client {
	return b.agreements
}

func (b backends) FakeCash() fakecash.Client {
	return b.fakecash
}

func (b backends) Rafiki() rafiki.Service {
	return b.rafiki
}

func (b backends) SupportTickets() supporttickets.Client {
	return b.supportTickets
}

func (b backends) Waitlist() waitlist.Client {
	return b.waitlist
}

func (b backends) Users() user.Client {
	return b.users
}

func (b backends) Temporal() client.Client {
	return b.temporal
}

func (b backends) Countries() country.Client {
	return b.countries
}

func (b backends) Pacioli() pacioli.Client {
	return b.pacioli
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) Validator() *validator.Validate {
	return b.val
}

func (b backends) Twilio() _twilio.Service {
	return b.twilio
}

//func (b backends) MX() _mx.Client {
//	return b.mxProvider
//}

func (b backends) LinkedAccounts() linkedaccounts.Client {
	return b.linkedaccounts
}

func (b backends) Machnet() machnet.Client {
	return b.machnet
}

func newLocalPacioliClient(db *sqlx.DB) pacioli.Client {
	b := backends{
		db:  db,
		val: validator.New(),
	}

	return pacioli_client.NewLocal(b)
}
