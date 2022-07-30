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

	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/temporal"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc/credentials/insecure"

	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/withdrawals"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	kratos "github.com/ory/kratos-client-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/cli"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/graph"
	_grpc "gitlab.com/fynbos/backend/grpc"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/migrations"
	_mx "gitlab.com/fynbos/backend/providers/mx"
	_mxexternal "gitlab.com/fynbos/backend/providers/mx/external"
	_noop "gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/providers/rafiki"
	_unit "gitlab.com/fynbos/backend/providers/unit"
	_twilio "gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
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
		err = migrations.MigrateFromEmbeddedFiles(args.ConnectionString, fs)
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
		err = agreements.Migrate(context.Background(), &agreements.MigrateArgs{
			Db:            db,
			DirectoryPath: "./utils/agreements/live",
		})
		if err != nil {
			log.Fatalln(err)
		}
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
	db, err := sqlx.Connect("postgres", args.DbConnectionString)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

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

	users, err := user.NewService(kratosClient)
	if err != nil {
		log.Fatalln(err)
	}
	users = user.NewLoggingService(users, logger)

	cs := country.NewService(db)
	id, err := identity.NewService(identity.ServiceArgs{
		CountryService: cs,
		Db:             db,
	})
	if err != nil {
		log.Fatalln(err)
	}
	id = identity.NewLoggingService(id, logger)

	conn, err := grpc.Dial(args.PacioliUrl, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}

	pClient := pacioliv1.NewPacioliServiceClient(conn)
	as, err := accounts.NewService(&accounts.ServiceArgs{
		Db:              db,
		Is:              id,
		Cs:              cs,
		PacioliClient:   pClient,
		PacioliLedgerID: args.UsdLedgerID,
		PacioliTenant:   "dev",
	})
	if err != nil {
		log.Fatalln(err)
	}
	as = accounts.NewLoggingService(as, logger)

	ts, err := transactions.NewService(&transactions.ServiceArgs{
		AccountService: as,
		PacioliClient:  pClient,
		Db:             db,
	})
	if err != nil {
		log.Fatalln(err)
	}
	ts = transactions.NewLoggingService(ts, logger)

	nos, err := _noop.NewService(_noop.ServiceArgs{
		LedgerID:      args.NoopLedgerID,
		EquityAccID:   args.NoopEquityAccountID,
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		log.Fatalln(err)
	}

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
		AccountService:  as,
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

	fs, err := fundingsources.NewService(&fundingsources.ServiceArgs{
		Is:   id,
		As:   as,
		Db:   db,
		Noop: nos,
		Unit: us,
		Tp:   tp,
	})
	if err != nil {
		log.Fatalln(err)
	}
	fs = fundingsources.NewLoggingService(fs, logger)

	os, err := onboarding.NewService(&onboarding.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   id,
		Noop: nos,
		Tp:   tp,
	})
	if err != nil {
		log.Fatalln(err)
	}

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

	ps, err := payments.NewService(&payments.ServiceArgs{
		Db: db,
		As: as,
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
		Account:                          as,
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
		Db:            db,
		AgreementsDir: "./utils/agreements/live",
	})
	if err != nil {
		log.Fatalln(err)
	}

	server, err := _grpc.NewServer(&_grpc.ServerArgs{
		HealthCheckService:   health,
		IdentityService:      id,
		AccountsService:      as,
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

func startWorker(args *cli.StartArgs) {
	log.Printf("begin worker start")
	db, err := sqlx.Connect("postgres", args.DbConnectionString)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

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
	cs := country.NewService(db)
	id, err := identity.NewService(identity.ServiceArgs{
		CountryService: cs,
		Db:             db,
	})
	if err != nil {
		log.Fatalln(err)
	}
	id = identity.NewLoggingService(id, logger)

	conn, err := grpc.Dial(args.PacioliUrl, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}

	pClient := pacioliv1.NewPacioliServiceClient(conn)
	as, err := accounts.NewService(&accounts.ServiceArgs{
		Db:              db,
		Is:              id,
		Cs:              cs,
		PacioliClient:   pClient,
		PacioliLedgerID: args.UsdLedgerID,
		PacioliTenant:   "dev",
	})
	if err != nil {
		log.Fatalln(err)
	}
	as = accounts.NewLoggingService(as, logger)

	ts, err := transactions.NewService(&transactions.ServiceArgs{
		AccountService: as,
		PacioliClient:  pClient,
		Db:             db,
	})
	if err != nil {
		log.Fatalln(err)
	}
	ts = transactions.NewLoggingService(ts, logger)

	nos, err := _noop.NewService(_noop.ServiceArgs{
		LedgerID:      args.NoopLedgerID,
		EquityAccID:   args.NoopEquityAccountID,
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		log.Fatalln(err)
	}

	tp, err := temporal.NewTemporalClient(args.TemporalUrl)
	if err != nil {
		log.Fatalln(err)
	}

	unit, err := _unit.NewService(_unit.ServiceArgs{
		BaseURL:         args.UnitBaseURL,
		Token:           args.UnitToken,
		WebhookToken:    args.UnitWebhookToken,
		Db:              db,
		IdentityService: id,
		AccountService:  as,
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

	fs, err := fundingsources.NewService(&fundingsources.ServiceArgs{
		Is:   id,
		As:   as,
		Db:   db,
		Noop: nos,
		Unit: unit,
		Tp:   tp,
	})
	if err != nil {
		log.Fatalln(err)
	}
	fs = fundingsources.NewLoggingService(fs, logger)

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

	os, err := onboarding.NewService(&onboarding.ServiceArgs{
		Db:   db,
		As:   as,
		Is:   id,
		Noop: nos,
		Tp:   tp,
	})
	if err != nil {
		log.Fatal(err)
	}

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
