package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	aasa_assetlinks "github.com/interledger/interledger-app/go/backend/aasa_assetlinks"
	"github.com/interledger/interledger-app/go/backend/accountdeletion"
	accountdeletion_client "github.com/interledger/interledger-app/go/backend/accountdeletion/client"
	"github.com/interledger/interledger-app/go/backend/admin"
	"github.com/interledger/interledger-app/go/backend/admin/auth"
	"github.com/interledger/interledger-app/go/backend/agreements"
	agreements_client "github.com/interledger/interledger-app/go/backend/agreements/client"
	agreements_migrations "github.com/interledger/interledger-app/go/backend/agreements/migrations"
	"github.com/interledger/interledger-app/go/backend/analytics"
	analytics_client "github.com/interledger/interledger-app/go/backend/analytics/client"
	analytics_webhook "github.com/interledger/interledger-app/go/backend/analytics/webhook"
	"github.com/interledger/interledger-app/go/backend/api"
	"github.com/interledger/interledger-app/go/backend/cli"
	"github.com/interledger/interledger-app/go/backend/config"
	"github.com/interledger/interledger-app/go/backend/contacts"
	contacts_client "github.com/interledger/interledger-app/go/backend/contacts/client"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/jobs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/riandyrn/otelchi"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"

	"github.com/interledger/interledger-app/go/backend/email"
	email_client "github.com/interledger/interledger-app/go/backend/email/client"
	"github.com/interledger/interledger-app/go/backend/features"
	features_client "github.com/interledger/interledger-app/go/backend/features/client"
	_grpc "github.com/interledger/interledger-app/go/backend/grpc"
	"github.com/interledger/interledger-app/go/backend/healthcheck"
	"github.com/interledger/interledger-app/go/backend/identities"
	identities_client "github.com/interledger/interledger-app/go/backend/identities/client"
	"github.com/interledger/interledger-app/go/backend/images"
	img_client "github.com/interledger/interledger-app/go/backend/images/client"
	"github.com/interledger/interledger-app/go/backend/keys"
	keys_client "github.com/interledger/interledger-app/go/backend/keys/client"
	"github.com/interledger/interledger-app/go/backend/kyc"
	kyc_client "github.com/interledger/interledger-app/go/backend/kyc/client"
	kyc_ops "github.com/interledger/interledger-app/go/backend/kyc/ops"
	"github.com/interledger/interledger-app/go/backend/kyc/persona"
	"github.com/interledger/interledger-app/go/backend/limits"
	limits_client "github.com/interledger/interledger-app/go/backend/limits/client"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	linked_account_client "github.com/interledger/interledger-app/go/backend/linkedaccounts/client"
	"github.com/interledger/interledger-app/go/backend/notify"
	notify_client "github.com/interledger/interledger-app/go/backend/notify/client"
	"github.com/interledger/interledger-app/go/backend/payments"
	payments_client "github.com/interledger/interledger-app/go/backend/payments/client"
	"github.com/interledger/interledger-app/go/backend/providers/chimoney"
	chimoney_client "github.com/interledger/interledger-app/go/backend/providers/chimoney/client"
	chimoney_ops "github.com/interledger/interledger-app/go/backend/providers/chimoney/ops"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	gatehub_client "github.com/interledger/interledger-app/go/backend/providers/gatehub/client"
	gatehub_ops "github.com/interledger/interledger-app/go/backend/providers/gatehub/ops"
	"github.com/interledger/interledger-app/go/backend/providers/plaid"
	plaid_client "github.com/interledger/interledger-app/go/backend/providers/plaid/client"
	plaid_ops "github.com/interledger/interledger-app/go/backend/providers/plaid/ops"
	"github.com/interledger/interledger-app/go/backend/providers/pti"
	pti_client "github.com/interledger/interledger-app/go/backend/providers/pti/client"
	pti_ops "github.com/interledger/interledger-app/go/backend/providers/pti/ops"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	xago_client "github.com/interledger/interledger-app/go/backend/providers/xago/client"
	xago_external "github.com/interledger/interledger-app/go/backend/providers/xago/external"
	"github.com/interledger/interledger-app/go/backend/rafiki"
	rafiki_client "github.com/interledger/interledger-app/go/backend/rafiki/client"
	rafiki_external "github.com/interledger/interledger-app/go/backend/rafiki/external"
	"github.com/interledger/interledger-app/go/backend/signup"
	signup_client "github.com/interledger/interledger-app/go/backend/signup/client"
	"github.com/interledger/interledger-app/go/backend/slack"
	slack_client "github.com/interledger/interledger-app/go/backend/slack/client"
	slack_external "github.com/interledger/interledger-app/go/backend/slack/external"
	"github.com/interledger/interledger-app/go/backend/temporal"
	"github.com/interledger/interledger-app/go/backend/transactions"
	transactions_client "github.com/interledger/interledger-app/go/backend/transactions/client"
	_twilio "github.com/interledger/interledger-app/go/backend/twilio"
	"github.com/interledger/interledger-app/go/backend/twitter"
	twitter_client "github.com/interledger/interledger-app/go/backend/twitter/client"
	"github.com/interledger/interledger-app/go/backend/user"
	user_client "github.com/interledger/interledger-app/go/backend/user/client"
	"github.com/interledger/interledger-app/go/backend/vault"
	"github.com/interledger/interledger-app/go/backend/waitlist"
	waitlist_client "github.com/interledger/interledger-app/go/backend/waitlist/client"
	"github.com/interledger/interledger-app/go/backend/wallets"
	wallets_client "github.com/interledger/interledger-app/go/backend/wallets/client"
	wallet_handler "github.com/interledger/interledger-app/go/backend/wallets/handler"
	"github.com/interledger/interledger-app/go/log"
	"github.com/interledger/interledger-app/go/pacioli"
	pacioli_client "github.com/interledger/interledger-app/go/pacioli/client"
	pacioli_db "github.com/interledger/interledger-app/go/pacioli/db"
	"github.com/interledger/interledger-app/go/tracing"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	// bradu
	fiant "github.com/interledger/interledger-app/go/backend/providers/fiant/v1"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

// Version is set at build time via -ldflags "-X main.Version=<tag>".
var Version = "v0.0.0"

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
		initSentry(args.Sentry.DSN, Version, args.Label)
		defer sentry.Flush(2 * time.Second)
		migrate(args)
	case "start":
		args, err := cli.ParseStartArgs()
		if err != nil {
			log.Fatalln(err)
		}
		initSentry(args.Sentry.DSN, Version, args.Environment.Label)
		defer sentry.Flush(2 * time.Second)
		start(args)
	case "worker":
		args, err := cli.ParseStartArgs()
		if err != nil {
			log.Fatalln(err)
		}
		initSentry(args.Sentry.DSN, Version, args.Environment.Label)
		defer sentry.Flush(2 * time.Second)
		startWorker(args)
	case "dev":
		args, err := cli.ParseStartArgs()
		if err != nil {
			log.Fatalln(err)
		}
		initSentry(args.Sentry.DSN, Version, args.Environment.Label)
		defer sentry.Flush(2 * time.Second)
		go func() {
			startWorker(args)
		}()
		start(args)
	default:
		log.Fatal("Unknown command:", zap.String("command", command))
	}
}

func initSentry(dsn, release, environment string) {
	if dsn == "" {
		return
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Release:          release,
		Environment:      environment,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Fatal("sentry.Init: %s", zap.Error(err))
	}
}

func start(args *cli.StartArgs) {
	traceShutdown, err := tracing.InitTraceProvider("backend", Version, args.OTEL.Enabled, args.OTEL.Endpoint, args.OTEL.Headers)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		ctx := context.Background()
		if err := traceShutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider", zap.Error(err))
		}
	}()

	b := NewBackends(args, false)
	defer CloseBackends(b)

	// Atomically claim unnotified agreements
	var newAgreementIDs []string
	if err := b.DB().SelectContext(context.Background(), &newAgreementIDs, "UPDATE agreements SET notified = true WHERE notified = false RETURNING id"); err != nil {
		log.Warn("failed to claim unnotified agreements", zap.Error(err))
	} else if len(newAgreementIDs) > 0 {
		deadlineDate := time.Now().UTC().Add(jobs.AgreementChangeDeadlineDays * 24 * time.Hour).Format("January 2, 2006")
		sort.Strings(newAgreementIDs)
		h := sha256.Sum256([]byte(strings.Join(newAgreementIDs, ",")))
		wo := client.StartWorkflowOptions{
			ID:                    "agreement_change_notify_" + hex.EncodeToString(h[:8]),
			TaskQueue:             "backend",
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
		}
		if _, err := b.Temporal().ExecuteWorkflow(context.Background(), wo, jobs.NotifyAgreementChangedWorkflow, newAgreementIDs, deadlineDate, 0, nil, nil); err != nil {
			log.Warn("failed to start agreement change notification workflow", zap.Error(err), zap.Strings("agreementIDs", newAgreementIDs))
		}
	}

	router := chi.NewRouter()
	router.Routes()
	router.Use(otelchi.Middleware("backend", otelchi.WithChiRoutes(router)))
	router.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	router.Mount("/api", api.NewRouter(b.Users(), b.Wallets(), b.Gatehub()))
	router.Handle("/kratos/signup", analytics_webhook.NewHandleSignup(b))
	router.Handle("/kratos/login", analytics_webhook.NewHandleLogin(b))
	router.Handle("/kratos/logout", analytics_webhook.NewHandleLogout(b))
	router.Handle("/rafiki", b.rafiki.WebhookHandler())
	router.Handle("/webhooks/xago", b.xago.WebhookHandler())
	personaClient := persona.New(persona.Config{
		BaseURL:       args.Persona.BaseURL,
		BearerToken:   args.Persona.Token,
		WebhookSecret: args.Persona.WebhookToken,
	})
	router.Handle("/webhooks/persona", kyc_ops.NewHandlePersonaWebhook(b, personaClient))
	router.Handle("/webhooks/chimoney", chimoney_ops.NewWebhook(b, args.Chimoney.WebhookSecret, args.Chimoney.Token, args.Environment.IsModeProd()))
	router.Handle("/.well-known/apple-app-site-association", aasa_assetlinks.AppSiteAssociationHandler(b.aasaConfig))
	router.Handle("/.well-known/assetlinks.json", aasa_assetlinks.AssetLinksHandler(b.aasaConfig))

	if args.PTI.Enabled {
		ptiWebhook, err := pti_ops.Webhook(b, args.PTI.ClientID, args.PTI.PublicKeyJWK)
		if err != nil {
			log.Fatalln(err)
		}
		router.Handle("/webhooks/pti", ptiWebhook)
	}
	router.Handle("/webhooks/gatehub", gatehub_ops.NewWebhook(b, b.gatehubConfig))
	router.Handle("/webhooks/gatehub/v1/users/managed/{userId}/2fa", gatehub_ops.NewSCAHandler(b, b.gatehubConfig))
	router.Handle("/{wallet_id}/identities/{identity_sig_hash}", wallet_handler.GetIdentityHandler(b))

	if b.plaidClient != nil {
		var linker plaid_ops.FiantLinker
		if args.PTI.Enabled {
			fl, err := pti_ops.NewFiantLinker(b, args.PTI.BaseURL, args.PTI.ClientID, args.PTI.JWK)
			if err != nil {
				log.Fatalln(err)
			}
			linker = fl
		}
		router.Mount("/api/plaid", plaid_ops.NewRouter(b.plaidClient, b.Users(), linker, args.Plaid.Processor))
	}

	router.NotFound(wallet_handler.WalletRedirectHandler(b))

	// fiant sandbox actions (only when PTI is enabled)
	if args.PTI.Enabled {
		ptiPrivateKey, err := jwk.ParseKey([]byte(args.PTI.JWK))
		if err != nil {
			log.Fatalln(err)
		}

		ctrl, err := fiant.NewController(
			fiant.WithBaseURL(args.PTI.BaseURL),
			fiant.WithClientID(args.PTI.ClientID),
			fiant.WithDerivedKeys(ptiPrivateKey),
		)
		if err != nil {
			log.Fatalln(err)
		}

		router.Handle("/settle/{transaction_id}", ctrl.SettleTransactionHook())
		router.Handle("/return/{transaction_id}", ctrl.ReturnTransactionHook())
	}
	// ~fiant sandbox actions

	var wg sync.WaitGroup

	log.Info("backend running at http://localhost:%s", zap.String("port", args.Port))
	serveHTTP(&http.Server{Addr: ":" + args.Port, Handler: router}, &wg)

	server, err := _grpc.NewServer(b)
	if err != nil {
		log.Fatalln(err)
	}

	serveGrpc("8443", server, &wg)

	adminServer, err := admin.NewServer(b)
	if err != nil {
		log.Fatalln(err)
	}

	serveGrpc("8448", adminServer, &wg)

	log.Info("waiting for shutdown")
	wg.Wait()
	b.analytics.Close()
	log.Info("clean shutdown")
}

func serveGrpc(port string, server *grpc.Server, wg *sync.WaitGroup) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		log.Fatalln(err)
	}

	wg.Add(1)
	go func(sigCh chan os.Signal, wg *sync.WaitGroup) {
		defer wg.Done()
		<-sigCh
		log.Info(fmt.Sprintf("got signal attempting graceful shutdown: 0.0.0.0:%s", port))
		server.GracefulStop()
	}(ch, wg)

	go func() {
		log.Info(fmt.Sprintf("grpc server: 0.0.0.0:%s", port))
		err = server.Serve(listener)
		if err != nil {
			log.Fatalln(err)
		}
	}()
}

func serveHTTP(server *http.Server, wg *sync.WaitGroup) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	wg.Add(1)
	go func(sigCh chan os.Signal, wg *sync.WaitGroup) {
		defer wg.Done()
		<-sigCh

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Info(fmt.Sprintf("got signal attempting graceful HTTP shutdown: %s", server.Addr))
		_ = server.Shutdown(shutdownCtx)
	}(ch, wg)

	go func() {
		err := server.ListenAndServe()
		// http.ErrServerClosed is returned immediately after Shutdown is called.
		// Don't panic and let the HTTP shutdown inside the 30-second timeout.
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("failed to start HTTP server", zap.Error(err))
		}
	}()
}

func migrate(args *cli.MigrationArgs) {
	err := db.Migrate(context.Background(), args.DBUrl, args.OpenPaymentsBaseURL)
	if err != nil {
		log.Fatalln(err)
	}

	dbConn, err := sqlx.Connect("postgres", args.DBUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	err = db.CreateExpIndex(context.Background(), dbConn)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = agreements_migrations.MigrateFromEmbeddedMarkdowns(context.Background(), dbConn, args.MigrationConfig)
	if err != nil {
		log.Fatalln(err)
	}

	err = pacioli_db.Migrate(context.Background(), args.PacioliDBUrl)
	if err != nil {
		log.Fatalln(err)
	}

	pacCon, err := sqlx.Connect("postgres", args.PacioliDBUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := pacCon.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	pc := pacioli_client.NewLocal(pacCon)
	ledgers, err := pc.ConfigureLedgers(context.Background(), []pacioli.ConfigureLedgerArgs{
		{
			ID:    xago.LedgerIDZAR,
			Name:  "Xago ZAR Ledger",
			Asset: currency.ZAR.String(),
			Scale: uint8(currency.ZAR.Scale()),
		},
		{
			ID:    xago.LedgerIDUSD,
			Name:  "Xago USD Ledger",
			Asset: currency.USD.String(),
			Scale: uint8(currency.USD.Scale()),
		},
		{
			ID:    pti.LedgerIDUSD,
			Name:  "PTI USD Ledger",
			Asset: currency.USD.String(),
			Scale: uint8(currency.USD.Scale()),
		},
		{
			ID:    gatehub.LedgerIDEUR,
			Name:  "Gatehub EUR Ledger",
			Asset: currency.EUR.String(),
			Scale: uint8(currency.EUR.Scale()),
		},
		{
			ID:    chimoney.LedgerIDCAD,
			Name:  "Gatehub CAD Ledger",
			Asset: currency.CAD.String(),
			Scale: uint8(currency.CAD.Scale()),
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	for _, l := range ledgers {
		if l.Code != pacioli.LedgerOK {
			log.Fatal("failed to configure pacioli ledgers", zap.String("code", l.Code.String()))
		}
	}

	accs, err := pc.ConfigureAccounts(context.Background(), []pacioli.ConfigureAccountArgs{
		{
			ID:                         xago.ZAROpsAccount,
			LedgerID:                   xago.LedgerIDZAR,
			Code:                       1,
			DebitsMustNotExceedCredits: false,
			CreditsMustNotExceedDebits: false,
		},
		{
			ID:                         xago.USDOpsAccount,
			LedgerID:                   xago.LedgerIDUSD,
			Code:                       1,
			DebitsMustNotExceedCredits: false,
			CreditsMustNotExceedDebits: false,
		},
		{
			ID:                         rafiki.ZARBalanceAccount,
			LedgerID:                   xago.LedgerIDZAR,
			Code:                       1,
			DebitsMustNotExceedCredits: false,
			CreditsMustNotExceedDebits: false,
		},
		{
			ID:                         pti.USDOpsAccount,
			LedgerID:                   pti.LedgerIDUSD,
			Code:                       1,
			DebitsMustNotExceedCredits: false,
			CreditsMustNotExceedDebits: false,
		},
		{
			ID:                         gatehub.EUROpsAccount,
			LedgerID:                   gatehub.LedgerIDEUR,
			Code:                       1,
			DebitsMustNotExceedCredits: false,
			CreditsMustNotExceedDebits: false,
		},
		{
			ID:                         chimoney.CADOpsAccount,
			LedgerID:                   chimoney.LedgerIDCAD,
			Code:                       1,
			DebitsMustNotExceedCredits: false,
			CreditsMustNotExceedDebits: false,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	for _, acc := range accs {
		if acc.Code != pacioli.AccountExists && acc.Code != pacioli.AccountOK {
			log.Fatal("failed to configure pacioli accounts", zap.String("code", acc.Code.String()))
		}
	}
}

func startWorker(args *cli.StartArgs) {
	log.Info("begin worker start")

	traceShutdown, err := tracing.InitTraceProvider("backend-worker", Version, args.OTEL.Enabled, args.OTEL.Endpoint, args.OTEL.Headers)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		ctx := context.Background()
		if err := traceShutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider", zap.Error(err))
		}
	}()

	b := NewBackends(args, true)

	router := chi.NewRouter()
	router.Routes()
	router.Use(otelchi.Middleware("worker", otelchi.WithChiRoutes(router)))
	router.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	var wg sync.WaitGroup

	// TODO: Replace the port
	log.Info("worker http healthcheck running at http://localhost:%s", zap.String("port", "8081"))
	serveHTTP(&http.Server{Addr: ":8081", Handler: router}, &wg)

	log.Info("Worker creating")
	w, err := temporal.NewTemporalWorker(b, b.gatehubConfig, b.xagoConfig, args.PTI.JWK, args.PTI.BaseURL, args.PTI.ClientID, args.Chimoney.Token, args.Rafiki.NodeEnabled, jobs.Config{
		KratosURL:         args.Kratos.URL,
		KratosAdminURL:    args.Kratos.AdminURL,
		PTIJWK:            args.PTI.JWK,
		PTIBaseURL:        args.PTI.BaseURL,
		PTIClientID:       args.PTI.ClientID,
		RafikiDBURL:       args.Rafiki.DBURL,
		RafikiAuthDBURL:   args.Rafiki.AuthDBURL,
		TempGatehubAppID:  "",
		TempGatehubSecret: "",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = w.Run(worker.InterruptCh())
	log.Info("Worker started")
	if err != nil {
		log.Fatal("Unable to start worker", zap.Error(err))
	}
}

type backends struct {
	val            *validator.Validate
	db             *sqlx.DB
	txRunner       *db.TxRunner
	twitter        twitter.Client
	adminAuth      auth.Service
	agreements     agreements.Client
	linkedaccounts linkedaccounts.Client
	healthcheck    healthcheck.Service
	signup         signup.Client
	temporal       client.Client
	twilio         _twilio.Service
	users          user.Client
	waitlist       waitlist.Client
	kyc            kyc.Client
	keys           keys.Client
	email          email.Client
	transactions   transactions.Client
	notify         notify.Client
	analytics      analytics.Client
	contacts       contacts.Client
	limits         limits.Client
	ident          identities.Client
	vault          vault.Client
	feat           features.Client
	img            images.Client
	wallet         wallets.Client
	payment        payments.Client
	slack          slack.Client
	rafiki         rafiki.Client
	xago           xago.Client
	xagoConfig     xago_external.Config
	pac            pacioli.Client
	pti            pti.Client
	gatehub        gatehub.Client
	gatehubConfig  gatehub.Config
	chimoney       chimoney.Client
	plaidConfig    plaid.Config
	plaidClient    plaid.Client
	aasaConfig     aasa_assetlinks.Config

	accountDeletion accountdeletion.Client
	cfg             *config.StartConfig
}

func (b backends) Chimoney() chimoney.Client {
	return b.chimoney
}

func (b backends) Gatehub() gatehub.Client {
	return b.gatehub
}

func (b backends) Pacioli() pacioli.Client {
	return b.pac
}

func (b backends) Xago() xago.Client {
	return b.xago
}

func (b backends) Slack() slack.Client {
	return b.slack
}

func (b backends) Rafiki() rafiki.Client {
	return b.rafiki
}

func (b backends) Payments() payments.Client {
	return b.payment
}

func (b backends) Wallets() wallets.Client {
	return b.wallet
}

func (b backends) Features() features.Client {
	return b.feat
}

func (b backends) Transactions() transactions.Client {
	return b.transactions
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

func (b backends) Waitlist() waitlist.Client {
	return b.waitlist
}

func (b backends) Users() user.Client {
	return b.users
}

func (b backends) Temporal() client.Client {
	return b.temporal
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	return b.txRunner.WithTx(ctx, fn)
}

func (b backends) Validator() *validator.Validate {
	return b.val
}

func (b backends) Twilio() _twilio.Service {
	return b.twilio
}

func (b backends) Twitter() twitter.Client {
	return b.twitter
}

func (b backends) LinkedAccounts() linkedaccounts.Client {
	return b.linkedaccounts
}

func (b backends) Email() email.Client {
	return b.email
}

func (b backends) Notify() notify.Client {
	return b.notify
}

func (b backends) Analytics() analytics.Client {
	return b.analytics
}

func (b backends) Contacts() contacts.Client {
	return b.contacts
}

func (b backends) Limits() limits.Client {
	return b.limits
}

func (b backends) Identities() identities.Client {
	return b.ident
}

func (b backends) Keys() keys.Client {
	return b.keys
}

func (b backends) Vault() vault.Client {
	return b.vault
}

func (b backends) Images() images.Client {
	return b.img
}

func (b backends) PTI() pti.Client {
	return b.pti
}

func (b backends) AccountDeletion() accountdeletion.Client {
	return b.accountDeletion
}

func (b backends) Config() *config.StartConfig {
	return b.cfg
}

func NewBackends(args *cli.StartArgs, isWorker bool) *backends {
	b := &backends{}
	b.cfg = args.StartConfig

	dbConn, err := otelsqlx.Connect("postgres", args.DB.URL, otelsql.WithAttributes(semconv.DBSystemCockroachdb), otelsql.WithDBName("cockroachdb"))
	if err != nil {
		log.Fatalln(err)
	}
	b.db = dbConn
	b.txRunner = db.NewTxRunner(dbConn)

	// Initialises the logger we will use throughout
	err = log.Initialize(args.LogLevel)
	if err != nil {
		log.Fatalln(err)
	}

	tp, err := temporal.NewTemporalClient(args.Temporal.URL)
	if err != nil {
		log.Fatalln(err)
	}
	b.temporal = tp

	b.users = user_client.New(b, args.Kratos.URL, args.Kratos.AdminURL)

	b.wallet = wallets_client.New(b)

	b.payment = payments_client.New(b)

	b.linkedaccounts = linked_account_client.New(b)

	b.signup = signup_client.New(b)

	b.accountDeletion = accountdeletion_client.New(b)

	b.waitlist = waitlist_client.New(b, log.Logger())

	b.twitter = twitter_client.New(b, &twitter_client.NewClientArgs{
		ClientID:      "DEPRECATED",
		ClientSecret:  "DEPRECATED",
		AuthEndpoint:  "https://twitter.com/i/oauth2/authorize",
		TokenEndpoint: "https://api.twitter.com/2/oauth2/token",
		RedirectURL:   "DEPRECATED",
		BearerToken:   "DEPRECATED",
	})

	slack.Init(args.Slack.Token, map[slack.Channel]string{
		slack.ChannelSignupKYC:   args.Slack.ChannelSignupKYC,
		slack.ChannelTransaction: args.Slack.ChannelTransaction,
		slack.ChannelError:       args.Slack.ChannelError,
	})
	_grpc.InitAgreementIDs(args.Agreements.SignupAgreementIDs)

	b.slack, err = slack_client.New(b, slack_external.Config{
		ClientID:       args.Slack.ClientID,
		ClientSecret:   args.Slack.ClientSecret,
		RedirectURL:    "",
		BotRedirectURL: "",
		ApplicationURL: args.ApplicationURL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	b.analytics = analytics_client.New(b, args.Segment.Key)

	b.feat = features_client.New(b, true)

	// When Twilio is disabled we run a no-op service that approves any code.
	// This is only reachable outside environment.mode=prod: config validation
	// (and the Helm chart) reject twilio.enabled=false in production.
	if args.Twilio.Enabled {
		twilioService, err := _twilio.NewService(&_twilio.ServiceArgs{
			AccountSid:   args.Twilio.AccountSID,
			AccountToken: args.Twilio.AccountToken,
			ServiceSid:   args.Twilio.ServiceSID,
		})
		if err != nil {
			log.Fatalln(err)
		}
		b.twilio = twilioService
	} else {
		b.twilio = _twilio.NewNoOp()
	}

	if !isWorker {
		health, err := healthcheck.NewService()
		if err != nil {
			log.Fatalln(err)
		}
		b.healthcheck = health

		adminUsers, err := auth.NewService(args.Admin.PolicyAud, args.Admin.TeamDomain, b.DB(), args.Environment.IsModeLocal())
		if err != nil {
			log.Fatalln(err)
		}
		b.adminAuth = auth.NewLoggingService(adminUsers, log.Logger())
	}

	b.agreements = agreements_client.New(b)

	b.kyc, err = kyc_client.NewWithPersonaConfig(
		b,
		args.Smarty.AuthID,
		args.Smarty.AuthToken,
		persona.Config{
			BaseURL:       args.Persona.BaseURL,
			BearerToken:   args.Persona.Token,
			WebhookSecret: args.Persona.WebhookToken,
			FakeZAID:      args.Persona.SandboxFakeZAID,
		},
		args.Environment.IsModeProd(),
	)
	if err != nil {
		log.Fatalln(err)
	}

	if args.Email.Enabled {
		log.Debug("initialising SendGrid email client")
	} else {
		log.Debug("email disabled; initialising noop email client")
	}
	b.email = email_client.New(
		b,
		args.Email.Enabled,
		args.Email.Sendgrid.APIKey,
		args.Email.Sendgrid.FromName,
		args.Email.Sendgrid.FromEmail,
		args.Email.Sendgrid.OneTemplateID,
		args.Email.Sendgrid.SupportEmail,
	)

	log.Debug("initialising transactions")
	b.transactions = transactions_client.New(b)

	log.Debug("initialising notify")
	b.notify = notify_client.New(b, args.Pusher.Addr)

	log.Debug("initialising limits")
	b.limits = limits_client.New(b)

	log.Debug("initialising identities")
	b.ident = identities_client.New(b)

	log.Debug("initialising contacts")
	b.contacts = contacts_client.New(b)

	log.Debug("initialising images")
	b.img = img_client.New(b)

	log.Debug("initialising vault")
	vc, err := vault.NewClient(vault.Config{
		Addr:              args.Vault.Addr,
		TransitEnginePath: args.Vault.TransitEnginePath,
		Token:             args.Vault.Token,
		IsLocalOrTest:     args.Environment.IsModeLocal() || args.Environment.IsModeTest(),
	})
	if err != nil {
		log.Error("error vault", zap.Error(err))
	}
	b.vault = vc

	log.Debug("initialising keys")
	b.keys = keys_client.New(b)

	log.Debug("initialising validator")
	b.val = validator.New()

	log.Debug("initialising rafiki")
	b.rafiki = rafiki_client.New(b, rafiki_external.AdminSigningConfig{
		OperatorTenantID:  args.Rafiki.OperatorTenantID,
		AdminAPISecret:    args.Rafiki.AdminAPISecret,
		SignatureVersion:  args.Rafiki.SignatureVersion,
		BackendGraphQLURL: args.Rafiki.BackendGraphQLURL,
		AuthGraphQLURL:    args.Rafiki.AuthGraphQLURL,
	})

	log.Debug("initialising pacioli")
	pacDB, err := otelsqlx.Connect("postgres", args.DB.PacioliURL, otelsql.WithAttributes(semconv.DBSystemCockroachdb), otelsql.WithDBName("cockroachdb"))
	if err != nil {
		log.Fatalln(err)
	}
	b.pac = pacioli_client.NewLocal(pacDB)

	log.Debug("initialising xago")
	b.xagoConfig = xago_external.Config{
		APIBaseURL:      args.Xago.APIBaseURL,
		IdentityBaseURL: args.Xago.IdentityBaseURL,
		ExchangeBaseURL: args.Xago.ExchangeBaseURL,
		PublicKey:       args.Xago.APIPublicKey,
		Secret:          args.Xago.APISecret,
		PolicyID:        args.Xago.PolicyID,
	}
	b.xago = xago_client.New(b, b.xagoConfig, args.Environment.IsModeTest())

	log.Debug("initialising FIANT")
	pti_ops.ConfigureWidgetURLs(args.PTI.SDKURL, args.PTI.FormsURL, args.PTI.ClientID)
	pti.ConfigureScenarios(args.PTI.ScenarioTransfer, args.PTI.ScenarioDeposit, args.PTI.ScenarioWithdrawal)
	b.pti = pti_client.New(b, args.PTI.JWK, args.PTI.BaseURL, args.PTI.ClientID)

	log.Debug("initialising Gatehub")
	b.gatehubConfig = gatehub.Config{
		AppID:                   args.Gatehub.AppID,
		Secret:                  args.Gatehub.Secret,
		CardAppID:               args.Gatehub.CardAppID,
		GatewayID:               args.Gatehub.GatewayID,
		CardAccountProductCode:  args.Gatehub.CardAccountProductCode,
		PaywiserEuroVaultID:     args.Gatehub.PaywiserEuroVaultID,
		SendingUserID:           args.Gatehub.SendingUserID,
		SendingUserAddress:      args.Gatehub.SendingUserAddress,
		IntermediaryUserID:      args.Gatehub.IntermediaryUserID,
		IntermediaryUserAddress: args.Gatehub.IntermediaryUserAddress,
		WebhookSecret:           args.Gatehub.WebhookSecret,
		FallbackWebhookURL:      args.Gatehub.FallbackWebhookURL,
		OnOffRampClientID:       args.Gatehub.OnOffRampClientID,
		OnboardingClientID:      args.Gatehub.OnboardingClientID,
		ExchangeClientID:        args.Gatehub.ExchangeClientID,
		APIBaseURL:              args.Gatehub.APIBaseURL,
		OnboardingBaseURL:       args.Gatehub.OnboardingBaseURL,
		OnOffRampBaseURL:        args.Gatehub.OnOffRampBaseURL,
		EUROpsAccount:           args.Gatehub.EUROpsAccount,
		EUROpsLedgerID:          args.Gatehub.EUROpsLedgerID,
		OrganizationID:          args.Gatehub.OrganizationID,
	}
	b.gatehub = gatehub_client.New(b, b.gatehubConfig)
	if b.gatehub == nil {
		log.Fatalln(errors.New("failed to initialize Gatehub client; check Gatehub configuration"))
	}

	log.Debug("initialising Chimoney")
	b.chimoney = chimoney_client.New(b, args.Chimoney.Token, args.Environment.IsModeProd())

	if args.Plaid.Enabled {
		log.Debug("initialising Plaid")
		b.plaidConfig = plaid.Config{
			Enabled:      args.Plaid.Enabled,
			ClientID:     args.Plaid.ClientID,
			Secret:       args.Plaid.Secret,
			Env:          args.Plaid.Env,
			Products:     args.Plaid.Products,
			CountryCodes: args.Plaid.CountryCodes,
			Processor:    args.Plaid.Processor,
			APIURL:       args.Plaid.APIURL,
		}
		plaidC, err := plaid_client.New(b.plaidConfig)
		if err != nil {
			log.Fatalln(err)
		}
		b.plaidClient = plaidC
		log.Info("plaid client initialized",
			zap.String("env", args.Plaid.Env),
			zap.Strings("products", args.Plaid.Products),
			zap.Strings("country_codes", args.Plaid.CountryCodes),
			zap.String("processor", args.Plaid.Processor),
			zap.String("api_url", args.Plaid.APIURL),
		)
	} else {
		log.Debug("Plaid disabled (plaid.enabled=false)")
	}

	b.aasaConfig = aasa_assetlinks.Config{
		AppleAppID:         args.Mobile.AppleAppID,
		AndroidPackageName: args.Mobile.AndroidPackageName,
		AndroidSHA256:      args.Mobile.AndroidSHA256,
	}

	return b
}

func CloseBackends(b *backends) {
	if b == nil {
		return
	}

	if b.db != nil {
		if err := b.db.Close(); err != nil {
			log.Fatalln(err)
		}
	}
}
