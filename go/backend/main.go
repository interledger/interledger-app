package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	kratos "github.com/ory/kratos-client-go"
	"gitlab.com/fynbos/backend/accounts"
	accounts_client "gitlab.com/fynbos/backend/accounts/client"
	transactions_client "gitlab.com/fynbos/backend/accounttransactions/client"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/cli"
	"gitlab.com/fynbos/backend/country"
	country_client "gitlab.com/fynbos/backend/country/client"
	"gitlab.com/fynbos/backend/deposits"
	funding_client "gitlab.com/fynbos/backend/fundingsources/client"
	"gitlab.com/fynbos/backend/graph"
	_grpc "gitlab.com/fynbos/backend/grpc"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	identity_client "gitlab.com/fynbos/backend/identity/client"
	"gitlab.com/fynbos/backend/migrations"
	onboarding_client "gitlab.com/fynbos/backend/onboarding/client"
	"gitlab.com/fynbos/backend/payments"
	_mx "gitlab.com/fynbos/backend/providers/mx"
	_mxexternal "gitlab.com/fynbos/backend/providers/mx/external"
	_noop "gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/rafiki"
	_unit "gitlab.com/fynbos/backend/providers/unit"
	support_client "gitlab.com/fynbos/backend/supporttickets/client"
	"gitlab.com/fynbos/backend/temporal"
	_twilio "gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	waitlist_client "gitlab.com/fynbos/backend/waitlist/client"
	"gitlab.com/fynbos/backend/withdrawals"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/pacioli"
	pacioli_client "gitlab.com/fynbos/pacioli/client"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var fs embed.FS

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Expected `start` or `migrate`.")
	}

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
		log.Fatalln("Unknown command:", command)
	}
}

func start(args *cli.StartArgs) {
	var b = new(backends)
	b.val = validator.New()

	db, err := sqlx.Connect("postgres", args.DbConnectionString)
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

	configuration := kratos.NewConfiguration()
	configuration.Servers = kratos.ServerConfigurations{
		{
			URL:         args.KratosUrl,
			Description: "Dev Kratos",
		},
	}
	kratosClient := kratos.NewAPIClient(configuration)

	tp, err := temporal.NewTemporalClient(args.TemporalUrl)
	if err != nil {
		log.Fatalln(err)
	}
	b.temporal = tp

	users, err := user.NewService(kratosClient)
	if err != nil {
		log.Fatalln(err)
	}
	users = user.NewLoggingService(users, logger)
	b.users = users

	cs := country_client.New(b)
	b.countries = cs

	id := identity_client.New(b, logger)
	b.ids = id

	pClient, err := pacioli_client.New(args.PacioliUrl)
	if err != nil {
		log.Fatalln(err)
	}
	b.pacioli = pClient

	accountsClient := accounts_client.New(b, args.UsdLedgerID, logger)
	b.accounts = accountsClient

	ts := transactions_client.New(b, logger)

	nos, err := _noop.NewService(_noop.ServiceArgs{
		LedgerID:      args.NoopLedgerID,
		EquityAccID:   args.NoopEquityAccountID,
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		log.Fatalln(err)
	}
	b.noop = nos

	twilioService, err := _twilio.NewService(&_twilio.ServiceArgs{
		AccountSid:   args.TwilioSid,
		AccountToken: args.TwilioSecret,
		ServiceSid:   args.TwilioServiceSid,
	})
	if err != nil {
		log.Fatalln(err)
	}

	us, err := _unit.NewService(_unit.ServiceArgs{
		BaseURL:         args.UnitBaseURL,
		Token:           args.UnitToken,
		WebhookToken:    args.UnitWebhookToken,
		Db:              db,
		IdentityService: id,
		AccountClient:   accountsClient,
		Logger:          logger,
	})
	if err != nil {
		log.Fatalln(err)
	}

	mx, err := _mx.NewService(&_mx.ServiceArgs{
		ExternalClient:  _mxexternal.NewClient(args.MxBaseURL, args.MxClientID, args.MxApiKey),
		Db:              db,
		AccountsService: accountsClient,
		IdentityService: id,
		Temporal:        tp,
	})
	if err != nil {
		log.Fatalln(err)
	}

	fs := funding_client.New(b, logger)

	os := onboarding_client.New(b)

	ds, err := deposits.NewService(&deposits.ServiceArgs{
		Db: db,
		As: accountsClient,
		Is: id,
		Fs: fs,
		Tp: tp,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ws, err := withdrawals.NewService(&withdrawals.ServiceArgs{
		Db: db,
		As: accountsClient,
		Is: id,
		Fs: fs,
		Tp: tp,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ps, err := payments.NewService(&payments.ServiceArgs{
		Db: db,
		As: accountsClient,
		Is: id,
		Tp: tp,
	})
	if err != nil {
		log.Fatal(err)
	}
	ps = payments.NewLoggingService(ps, logger)

	rafikiProvider, err := rafiki.NewService(&rafiki.ServiceArgs{
		Db:  db,
		Url: args.RafikiGraphqlUrl,
	})
	if err != nil {
		log.Fatal(err)
	}

	graphql, err := graph.NewService(graph.GraphqlOpts{
		Db:                               db,
		Identity:                         id,
		Account:                          accountsClient,
		Country:                          cs,
		User:                             users,
		Noop:                             nos,
		UnitService:                      us,
		Fs:                               fs,
		Ps:                               ps,
		Os:                               os,
		Ws:                               ws,
		AccountTransactions:              ts,
		Ds:                               ds,
		QueryCacheSize:                   1000,
		AutomaticPersistedQueryCacheSize: 100,
	})
	if err != nil {
		log.Fatalln(err)
	}
	graphql = graph.NewLoggingService(graphql, logger)

	unitWebhook, err := _unit.NewWebhook(&_unit.WebhookArgs{
		Db: db,
		Up: us,
		Tp: tp,
	})
	if err != nil {
		log.Fatalln(err)
	}

	router := chi.NewRouter()
	router.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	router.Handle("/graphql", user.MakeMiddleware(users)(graph.MakeHandler(graphql, graph.GraphqlHttpHandlerOpts{
		WebSocketKeepAlivePingInterval: 10 * time.Second,
	})))
	router.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	router.Post("/webhooks/unit", unitWebhook.MakeHttpHandler())

	log.Printf("connect to http://localhost:%s/playground for GraphQL playground", args.Port)
	go func() {
		log.Fatal(http.ListenAndServe(":"+args.Port, router))
	}()

	health, err := healthcheck.NewService()
	if err != nil {
		log.Fatalln(err)
	}
	adminUsers, err := auth.NewService(args.GoogleOauth2ClientID)
	if err != nil {
		log.Fatal(err)
	}
	adminUsers = auth.NewLoggingService(adminUsers, logger)

	ags, err := agreements.NewService(&agreements.ServiceArgs{
		Db: db,
	})
	if err != nil {
		log.Fatalln(err)
	}

	supportTickets := support_client.NewClient(b, args.ZendeskSecret)

	server, err := _grpc.NewServer(&_grpc.ServerArgs{
		HealthCheckService:   health,
		IdentityService:      id,
		AccountsService:      accountsClient,
		AgreementsService:    ags,
		AdminAuthService:     adminUsers,
		UnitProvider:         us,
		UserService:          users,
		FundingSourceService: fs,
		OnboardingService:    os,
		MxProvider:           mx,
		RafikiProvider:       rafikiProvider,
		DepositService:       ds,
		TwilioService:        twilioService,
		WaitlistClient:       waitlist_client.New(b, logger),
		Temporal:             tp,
		TicketClient:         supportTickets,
	})
	if err != nil {
		log.Fatalln(err)
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", "8443"))
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("grpc server: 0.0.0.0:%s", "8443")
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

	err = agreements.MigrateFromEmbeddedMarkdowns(context.Background(), &agreements.MigrateFromEmbeddedMarkdownArgs{
		Db:        db,
		FynbosEnv: env.GetEnv(),
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func startWorker(args *cli.StartArgs) {
	log.Printf("begin worker start")

	var b = new(backends)
	b.val = validator.New()

	db, err := sqlx.Connect("postgres", args.DbConnectionString)
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

	cs := country_client.New(b)
	b.countries = cs

	id := identity_client.New(b, logger)
	b.ids = id

	pClient, err := pacioli_client.New(args.PacioliUrl)
	if err != nil {
		log.Fatal(err)
	}
	b.pacioli = pClient

	as := accounts_client.New(b, args.UsdLedgerID, logger)
	b.accounts = as

	ts := transactions_client.New(b, logger)

	nos, err := _noop.NewService(_noop.ServiceArgs{
		LedgerID:      args.NoopLedgerID,
		EquityAccID:   args.NoopEquityAccountID,
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		log.Fatalln(err)
	}
	b.noop = nos

	tp, err := temporal.NewTemporalClient(args.TemporalUrl)
	if err != nil {
		log.Fatalln(err)
	}
	b.temporal = tp

	unit, err := _unit.NewService(_unit.ServiceArgs{
		BaseURL:         args.UnitBaseURL,
		Token:           args.UnitToken,
		WebhookToken:    args.UnitWebhookToken,
		Db:              db,
		IdentityService: id,
		AccountClient:   as,
		Logger:          logger,
	})
	if err != nil {
		log.Fatalln(err)
	}

	mx, err := _mx.NewService(&_mx.ServiceArgs{
		ExternalClient:  _mxexternal.NewClient(args.MxBaseURL, args.MxClientID, args.MxApiKey),
		Db:              db,
		AccountsService: as,
		IdentityService: id,
		Temporal:        tp,
	})
	if err != nil {
		log.Fatalln(err)
	}

	fs := funding_client.New(b, logger)

	ds, err := deposits.NewService(&deposits.ServiceArgs{
		Db: db,
		As: as,
		Is: id,
		Fs: fs,
		Tp: tp,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ps, err := payments.NewService(&payments.ServiceArgs{
		Db: db,
		As: as,
		Is: id,
		Tp: tp,
	})
	if err != nil {
		log.Fatal(err)
	}

	os := onboarding_client.New(b)

	ws, err := withdrawals.NewService(&withdrawals.ServiceArgs{
		Db: db,
		As: as,
		Is: id,
		Fs: fs,
		Tp: tp,
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Worker creating")
	w, err := temporal.NewTemporalWorker(temporal.WorkerArgs{
		Client: tp,
		Ps:     ps,
		Ds:     ds,
		As:     as,
		Np:     nos,
		Ts:     ts,
		Is:     id,
		Os:     os,
		Up:     unit,
		Ws:     ws,
		Fs:     fs,
		Mx:     mx,
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = w.Run(worker.InterruptCh())
	log.Printf("Worker started")
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}
}

type AllBackends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Identity() identity.Client
	Countries() country.Client
	Pacioli() pacioli.Client
}

var _ AllBackends = backends{}

type backends struct {
	val       *validator.Validate
	db        *sqlx.DB
	ids       identity.Client
	countries country.Client
	pacioli   pacioli.Client
	accounts  accounts.Client
	noop      _noop.Service
	temporal  client.Client
	unit      _unit.Service
	users     user.Service
}

func (b backends) Users() user.Service {
	return b.users
}

func (b backends) Noop() _noop.Service {
	return b.noop
}

func (b backends) Temporal() client.Client {
	return b.temporal
}

func (b backends) Unit() _unit.Service {
	return b.unit
}

func (b backends) Accounts() accounts.Client {
	return b.accounts
}

func (b backends) Identity() identity.Client {
	return b.ids
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
