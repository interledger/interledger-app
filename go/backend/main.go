package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	kratos "github.com/ory/kratos-client-go"
	"go.uber.org/zap"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gitlab.com/fynbos/backend/authorization"
	"gitlab.com/fynbos/backend/db/utils"
	"gitlab.com/fynbos/backend/graph"
	_org "gitlab.com/fynbos/backend/organisation"
	"gitlab.com/fynbos/backend/user"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	baseDbUrl := os.Getenv("DB_URL")
	if baseDbUrl == "" {
		baseDbUrl = "cockroach://backend@cockroachdb-public:26257/backend?sslmode=verify-full&max_conns=20&max_idle_conns=4"
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
	connString, err := utils.InlineSslCreds(
		strings.Replace(baseDbUrl, "cockroach", "postgres", 1), // replace cockroach protocol with postgres so that we can use pq driver.
		"/cockroach-certs/client.backend.key",
		"/cockroach-certs/client.backend.crt",
		"/cockroach-certs/ca.crt",
	)
	if err != nil {
		log.Fatalln(err)
	}
	db, err := sqlx.Connect("postgres", connString)
	defer db.Close()

	if err != nil {
		log.Fatalln(err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level.UnmarshalText([]byte(logLevel))
	cfg.OutputPaths = []string{logOutputPath}
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalln(err)
	}

	authz, err := authorization.NewService()
	if err != nil {
		log.Fatalln(err)
	}

	configuration := kratos.NewConfiguration()
	configuration.Servers = kratos.ServerConfigurations{
		{
			URL:         kratosUrl,
			Description: "Dev Kratos",
		},
	}
	kratosClient := kratos.NewAPIClient(configuration)
	users, err := user.NewService(kratosClient)
	if err != nil {
		log.Fatalln(err)
	}

	org, err := _org.NewService(db, authz)
	if err != nil {
		log.Fatalln(err)
	}
	org = _org.NewLoggingService(org, logger)

	graphql, err := graph.NewService(graph.GraphqlOpts{
		Organisation:                     org,
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

	log.Printf("connect to http://localhost:%s/playground for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
