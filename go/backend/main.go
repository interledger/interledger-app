package main

import (
	"embed"
	"log"
	"net/http"
	"os"
	"time"

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
	"gitlab.com/fynbos/backend/graph"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/migrations"
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
		migrations.MigrateFromEmbeddedFiles(args.ConnectionString, fs)
	case "start":
		args, err := cli.ParseStartArgs()
		if err != nil {
			log.Fatalln(err)
		}
		start(args)
	default:
		log.Fatalln("Unknown command:", command)
	}
}

func start(args *cli.StartArgs) {
	db, err := sqlx.Connect("postgres", args.DbConnectionString)
	defer db.Close()
	if err != nil {
		log.Fatalln(err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level.UnmarshalText([]byte(args.LogLevel))
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

	users, err := user.NewService(kratosClient)
	if err != nil {
		log.Fatalln(err)
	}
	users = user.NewLoggingService(users, logger)

	cs := country.NewService()
	id, err := identity.NewService(cs)
	if err != nil {
		log.Fatalln(err)
	}
	id = identity.NewLoggingService(id, logger)

	conn, err := grpc.Dial(args.PacioliUrl, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}

	pClient := pacioliv1.NewPacioliServiceClient(conn)
	as, err := accounts.NewService(id, cs, args.UsdLedgerCode, pClient)
	if err != nil {
		log.Fatalln(err)
	}
	as = accounts.NewLoggingService(as, logger)

	graphql, err := graph.NewService(graph.GraphqlOpts{
		Db:                               db,
		Identity:                         id,
		Account:                          as,
		User:                             users,
		QueryCacheSize:                   1000,
		AutomaticPersistedQueryCacheSize: 100,
	})
	if err != nil {
		log.Fatalln(err)
	}
	graphql = graph.NewLoggingService(graphql, logger)

	router := chi.NewRouter()
	router.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	router.Handle("/graphql", user.MakeMiddleware(users)(graph.MakeHandler(graphql, graph.GraphqlHttpHandlerOpts{
		WebSocketKeepAlivePingInterval: 10 * time.Second,
	})))
	router.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	log.Printf("connect to http://localhost:%s/playground for GraphQL playground", args.Port)
	log.Fatal(http.ListenAndServe(":"+args.Port, router))
}
