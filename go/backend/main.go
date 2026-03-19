package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/riandyrn/otelchi"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
	aassassetlinks "gitlab.com/fynbos/backend/aass_assetlinks"
	"gitlab.com/fynbos/backend/admin"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	agreements_client "gitlab.com/fynbos/backend/agreements/client"
	agreements_migrations "gitlab.com/fynbos/backend/agreements/migrations"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	analytics_webhook "gitlab.com/fynbos/backend/analytics/webhook"
	"gitlab.com/fynbos/backend/cli"
	"gitlab.com/fynbos/backend/contacts"
	contacts_client "gitlab.com/fynbos/backend/contacts/client"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/discord"
	discord_client "gitlab.com/fynbos/backend/discord/client"
	"gitlab.com/fynbos/backend/email"
	email_client "gitlab.com/fynbos/backend/email/client"
	"gitlab.com/fynbos/backend/features"
	features_client "gitlab.com/fynbos/backend/features/client"
	_grpc "gitlab.com/fynbos/backend/grpc"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identities"
	identities_client "gitlab.com/fynbos/backend/identities/client"
	"gitlab.com/fynbos/backend/images"
	img_client "gitlab.com/fynbos/backend/images/client"
	"gitlab.com/fynbos/backend/keys"
	keys_client "gitlab.com/fynbos/backend/keys/client"
	"gitlab.com/fynbos/backend/kyc"
	kyc_client "gitlab.com/fynbos/backend/kyc/client"
	kyc_ops "gitlab.com/fynbos/backend/kyc/ops"
	"gitlab.com/fynbos/backend/limits"
	limits_client "gitlab.com/fynbos/backend/limits/client"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linked_account_client "gitlab.com/fynbos/backend/linkedaccounts/client"
	"gitlab.com/fynbos/backend/notify"
	notify_client "gitlab.com/fynbos/backend/notify/client"
	"gitlab.com/fynbos/backend/payments"
	payments_client "gitlab.com/fynbos/backend/payments/client"
	"gitlab.com/fynbos/backend/paymentsv2"
	"gitlab.com/fynbos/backend/providers/chimoney"
	chimoney_client "gitlab.com/fynbos/backend/providers/chimoney/client"
	chimoney_ops "gitlab.com/fynbos/backend/providers/chimoney/ops"
	"gitlab.com/fynbos/backend/providers/gatehub"
	gatehub_client "gitlab.com/fynbos/backend/providers/gatehub/client"
	gatehub_ops "gitlab.com/fynbos/backend/providers/gatehub/ops"
	"gitlab.com/fynbos/backend/providers/pti"
	pti_client "gitlab.com/fynbos/backend/providers/pti/client"
	pti_ops "gitlab.com/fynbos/backend/providers/pti/ops"
	"gitlab.com/fynbos/backend/providers/xago"
	xago_client "gitlab.com/fynbos/backend/providers/xago/client"
	xago_external "gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/backend/rafiki"
	rafiki_client "gitlab.com/fynbos/backend/rafiki/client"
	"gitlab.com/fynbos/backend/signup"
	signup_client "gitlab.com/fynbos/backend/signup/client"
	"gitlab.com/fynbos/backend/slack"
	slack_client "gitlab.com/fynbos/backend/slack/client"
	"gitlab.com/fynbos/backend/temporal"
	"gitlab.com/fynbos/backend/transactions"
	transactions_client "gitlab.com/fynbos/backend/transactions/client"
	_twilio "gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/twitter"
	twitter_client "gitlab.com/fynbos/backend/twitter/client"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client"
	"gitlab.com/fynbos/backend/vault"
	"gitlab.com/fynbos/backend/waitlist"
	waitlist_client "gitlab.com/fynbos/backend/waitlist/client"
	"gitlab.com/fynbos/backend/wallets"
	wallets_client "gitlab.com/fynbos/backend/wallets/client"
	wallet_handler "gitlab.com/fynbos/backend/wallets/handler"
	"gitlab.com/fynbos/log"
	"gitlab.com/fynbos/pacioli"
	pacioli_client "gitlab.com/fynbos/pacioli/client"
	pacioli_db "gitlab.com/fynbos/pacioli/db"
	"gitlab.com/fynbos/tracing"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	// bradu
	"github.com/lestrrat-go/jwx/v3/jwk"
	fiant "gitlab.com/fynbos/backend/providers/fiant/v1"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Expected `start` or `migrate`.")
	}

	// Set the timezone globally
	time.Local = time.UTC

	if os.Getenv("SENTRY_DSN") != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              os.Getenv("SENTRY_DSN"),
			Release:          os.Getenv("SENTRY_RELEASE"),
			TracesSampleRate: 1.0,
		})
		if err != nil {
			log.Fatal("sentry.Init: %s", zap.Error(err))
		}
		// Flush buffered events before the program terminates.
		defer sentry.Flush(2 * time.Second)
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
	case "dev":
		args, err := cli.ParseStartArgs()
		if err != nil {
			log.Fatalln(err)
		}

		go func() {
			startWorker(args)
		}()

		start(args)
	default:
		log.Fatal("Unknown command:", zap.String("command", command))
	}
}

func start(args *cli.StartArgs) {
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

	b := NewBackends(args, false)
	defer CloseBackends(b)

	router := chi.NewRouter()
	router.Routes()
	router.Use(otelchi.Middleware("backend", otelchi.WithChiRoutes(router)))
	router.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	router.Handle("/kratos/signup", analytics_webhook.NewHandleSignup(b))
	router.Handle("/kratos/login", analytics_webhook.NewHandleLogin(b))
	router.Handle("/kratos/logout", analytics_webhook.NewHandleLogout(b))
	router.Handle("/rafiki", b.rafiki.WebhookHandler())
	router.Handle("/webhooks/xago", b.xago.WebhookHandler())
	router.Handle("/webhooks/persona", kyc_ops.NewHandlePersonaWebhook(b))
	router.Handle("/webhooks/chimoney", chimoney_ops.NewWebhook(b))
	router.Handle("/.well-known/apple-app-site-association", aassassetlinks.AppSiteAssociationHandler(b.aasaConfig))
	router.Handle("/.well-known/assetlinks.json", aassassetlinks.AssetLinksHandler(b.aasaConfig))

	if args.PTIEnabled {
		ptiWebhook, err := pti_ops.Webhook(b)
		if err != nil {
			log.Fatalln(err)
		}
		router.Handle("/webhooks/pti", ptiWebhook)
	}
	router.Handle("/webhooks/gatehub", gatehub_ops.NewWebhook(b, b.gatehubConfig))
	router.Handle("/webhooks/gatehub/v1/users/managed/{userId}/2fa", gatehub_ops.NewSCAHandler(b, b.gatehubConfig))
	router.Handle("/{wallet_id}/identities/{identity_sig_hash}", wallet_handler.GetIdentityHandler(b))
	router.NotFound(wallet_handler.WalletRedirectHandler(b))

	// fiant sandbox actions (only when PTI is enabled)
	if args.PTIEnabled {
		ptiPrivateKey, err := jwk.ParseKey([]byte(args.PTIJWK))
		if err != nil {
			log.Fatalln(err)
		}

		ctrl, err := fiant.NewController(
			fiant.WithBaseURL(args.PTIBaseURL),
			fiant.WithClientID(args.PTIClientID),
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
	err := db.Migrate(context.Background(), args.ConnectionString)
	if err != nil {
		log.Fatalln(err)
	}

	dbConn, err := sqlx.Connect("postgres", args.ConnectionString)
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

	err = agreements_migrations.MigrateFromEmbeddedMarkdowns(context.Background(), dbConn)
	if err != nil {
		log.Fatalln(err)
	}

	err = pacioli_db.Migrate(context.Background(), args.PacioliConnectionString)
	if err != nil {
		log.Fatalln(err)
	}

	pacCon, err := sqlx.Connect("postgres", args.PacioliConnectionString)
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
	w, err := temporal.NewTemporalWorker(b, b.gatehubConfig, b.xagoConfig)
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
	discord        discord.Client
	slack          slack.Client
	rafiki         rafiki.Client
	xago           xago.Client
	xagoConfig     xago_external.Config
	pac            pacioli.Client
	pti            pti.Client
	gatehub        gatehub.Client
	gatehubConfig  gatehub.Config
	chimoney       chimoney.Client
	aasaConfig     aassassetlinks.Config
	paymentsv2     *paymentsv2.Service
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

func (b backends) Discord() discord.Client {
	return b.discord
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

func (b backends) PaymentsV2() *paymentsv2.Service {
	return b.paymentsv2
}

func (b backends) PTI() pti.Client {
	return b.pti
}

func NewBackends(args *cli.StartArgs, isWorker bool) *backends {
	b := &backends{}

	db, err := otelsqlx.Connect("postgres", args.DbConnectionString, otelsql.WithAttributes(semconv.DBSystemCockroachdb), otelsql.WithDBName("cockroachdb"))
	if err != nil {
		log.Fatalln(err)
	}
	b.db = db

	// Initialises the logger we will use throughout
	err = log.Initialize(args.LogLevel)
	if err != nil {
		log.Fatalln(err)
	}

	tp, err := temporal.NewTemporalClient(args.TemporalUrl)
	if err != nil {
		log.Fatalln(err)
	}
	b.temporal = tp

	b.users = user_client.New(b, args.KratosUrl, args.KratosAdminUrl)

	b.wallet = wallets_client.New(b)

	b.payment = payments_client.New(b)

	b.linkedaccounts = linked_account_client.New(b)

	b.signup = signup_client.New(b)

	b.waitlist = waitlist_client.New(b, log.Logger())

	b.twitter = twitter_client.New(b, &twitter_client.NewClientArgs{
		ClientID:      args.TwitterClientID,
		ClientSecret:  args.TwitterClientSecret,
		AuthEndpoint:  "https://twitter.com/i/oauth2/authorize",
		TokenEndpoint: "https://api.twitter.com/2/oauth2/token",
		RedirectURL:   args.TwitterRedirectURL,
		BearerToken:   args.TwitterBearerToken,
	})

	b.discord = discord_client.New(b, &discord_client.NewClientArgs{
		ClientID:      args.DiscordClientID,
		ClientSecret:  args.DiscordClientSecret,
		AuthEndpoint:  "https://discord.com/oauth2/authorize",
		TokenEndpoint: "https://discord.com/api/oauth2/token",
		RedirectURL:   args.DiscordRedirectURL,
	})

	b.slack, err = slack_client.New(b)
	if err != nil {
		log.Fatalln(err)
	}

	b.analytics = analytics_client.New(b, args.SegmentKey)

	b.feat = features_client.New(b)

	twilioService, err := _twilio.NewService(&_twilio.ServiceArgs{
		AccountSid:   args.TwilioSid,
		AccountToken: args.TwilioSecret,
		ServiceSid:   args.TwilioServiceSid,
	})
	if err != nil {
		log.Fatalln(err)
	}
	b.twilio = twilioService

	if !isWorker {
		health, err := healthcheck.NewService()
		if err != nil {
			log.Fatalln(err)
		}
		b.healthcheck = health

		adminUsers, err := auth.NewService(args.AdminPolicyAud, args.AdminTeamDomain, b.DB())
		if err != nil {
			log.Fatalln(err)
		}
		b.adminAuth = auth.NewLoggingService(adminUsers, log.Logger())
	}

	b.agreements = agreements_client.New(b)

	b.kyc, err = kyc_client.New(b, args.SmartyAuthID, args.SmartyAuthToken)
	if err != nil {
		log.Fatalln(err)
	}

	log.Debug("initialising SendGrid")
	b.email = email_client.New(b, args.SendgridAPIKey)

	log.Debug("initialising transactions")
	b.transactions = transactions_client.New(b)

	log.Debug("initialising notify")
	b.notify = notify_client.New(b, args.PusherAddr)

	log.Debug("initialising limits")
	b.limits = limits_client.New(b)

	log.Debug("initialising identities")
	b.ident = identities_client.New(b)

	log.Debug("initialising contacts")
	b.contacts = contacts_client.New(b)

	log.Debug("initialising images")
	b.img = img_client.New(b)

	log.Debug("initialising vault")
	vc, err := vault.NewClient()
	if err != nil {
		log.Error("error vault", zap.Error(err))
	}
	b.vault = vc

	log.Debug("initialising keys")
	b.keys = keys_client.New(b)

	log.Debug("initialising validator")
	b.val = validator.New()

	log.Debug("initialising rafiki")
	b.rafiki = rafiki_client.New(b)

	log.Debug("initialising pacioli")
	pacDB, err := otelsqlx.Connect("postgres", args.PacioliDBConString, otelsql.WithAttributes(semconv.DBSystemCockroachdb), otelsql.WithDBName("cockroachdb"))
	if err != nil {
		log.Fatalln(err)
	}
	b.pac = pacioli_client.NewLocal(pacDB)

	log.Debug("initialising xago")
	b.xagoConfig = xago_external.Config{
		APIBaseURL:      args.XagoAPIBaseURL,
		IdentityBaseURL: args.XagoIdentityBaseURL,
		PublicKey:       args.XagoPublicKey,
		Secret:          args.XagoSecret,
		PolicyID:        args.XagoPolicyID,
	}
	b.xago = xago_client.New(b, b.xagoConfig)

	log.Debug("initialising FIANT")
	b.pti = pti_client.New(b)

	log.Debug("initialising Gatehub")
	b.gatehubConfig = gatehub.Config{
		AppID:                  args.GatehubAppID,
		Secret:                 args.GatehubSecret,
		CardAppID:              args.GatehubCardAppID,
		GatewayID:              args.GatehubGatewayID,
		CardAccountProductCode: args.GatehubCardAccountProductCode,
		PaywiserEuroVaultID:    args.GatehubPaywiserEuroVaultID,
		SendingUserID:          args.GatehubSendingUserID,
		SendingUserAddress:     args.GatehubSendingUserAddress,
		WebhookSecret:          args.GatehubWebhookSecret,
		FallbackWebhookURL:     args.GatehubFallbackWebhookURL,
		OnOffRampClientID:      args.GatehubOnOffRampClientID,
		OnboardingClientID:     args.GatehubOnboardingClientID,
		ExchangeClientID:       args.GatehubExchangeClientID,
		APIBaseURL:             args.GatehubAPIBaseURL,
		OnboardingBaseURL:      args.GatehubOnboardingBaseURL,
		OnOffRampBaseURL:       args.GatehubOnOffRampBaseURL,
		EUROpsAccount:          args.GatehubEUROpsAccount,
		EUROpsLedgerID:         args.GatehubEUROpsLedgerID,
		OrganizationID:         args.GatehubOrganizationID,
	}
	b.gatehub = gatehub_client.New(b, b.gatehubConfig)
	if b.gatehub == nil {
		log.Fatalln(errors.New("failed to initialize Gatehub client; check Gatehub configuration"))
	}

	log.Debug("initialising Chimoney")
	b.chimoney = chimoney_client.New(b)

	b.aasaConfig = aassassetlinks.Config{
		AppleAppID:         args.AppleAppID,
		AndroidPackageName: args.AndroidPackageName,
		AndroidSHA256:      args.AndroidSHA256,
	}

	paymentsv2Repo := paymentsv2.NewRepository(db)
	paymentsv2Service := paymentsv2.NewService(paymentsv2Repo, b.pac, tp)

	b.paymentsv2 = paymentsv2Service

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
