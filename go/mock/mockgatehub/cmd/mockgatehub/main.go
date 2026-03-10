package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/fynbos/mock/mockgatehub/internal/auth"
	"gitlab.com/fynbos/mock/mockgatehub/internal/config"
	"gitlab.com/fynbos/mock/mockgatehub/internal/handler"
	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"gitlab.com/fynbos/mock/mockgatehub/internal/storage"
	"gitlab.com/fynbos/mock/mockgatehub/internal/webhook"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

var buildTime = "unknown"

func main() {
	cfg := config.Load()

	// Initialize logger with configured log level
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		logger.Fatal("failed to initialize logger", zap.Error(err))
	}

	logger.Info("gitlab.com/fynbos/mock/mockgatehub build info", zap.String("build_time", buildTime))

	logger.Info("starting MockGatehub")

	// Log all startup configuration at INFO level
	logger.Info("configuration loaded",
		zap.String("log_level", cfg.LogLevel),
		zap.String("port", cfg.Port),
		zap.String("redis_url", cfg.RedisURL),
		zap.Int("redis_db", cfg.RedisDB),
		zap.Bool("use_redis", cfg.UseRedis),
		zap.Bool("enforce_authentication", cfg.EnforceAuthentication),
		zap.String("webhook_url", cfg.WebhookURL),
		zap.String("webhook_secret", cfg.WebhookSecret),
		zap.Strings("valid_app_ids", getAppIDList(cfg.ValidCredentials)),
	)

	var store storage.Storage
	if cfg.UseRedis {
		logger.Info("setting up redis storage", zap.String("url", cfg.RedisURL), zap.Int("db", cfg.RedisDB))
		redisStore, err := storage.NewRedisStorage(cfg.RedisURL, cfg.RedisDB)
		if err != nil {
			logger.Fatal("failed to connect to redis", zap.Error(err))
		}
		defer redisStore.Close()
		store = redisStore
	} else {
		logger.Info("setting up in-memory storage")
		store = storage.NewMemoryStorage()
	}

	if err := storage.SeedTestUsersWithOrgID(store, cfg.DefaultOrganizationID); err != nil {
		logger.Fatal("failed to seed test users", zap.Error(err))
	}

	// Initialize webhook queue and worker
	var webhookQueue *webhook.Queue
	var webhookWorker *webhook.Worker

	if cfg.UseRedis {
		// Use Redis-backed queue for webhook delivery
		redisStore, ok := store.(*storage.RedisStorage)
		if !ok {
			logger.Fatal("redis storage type assertion failed")
		}
		webhookQueue = webhook.NewQueue(redisStore.GetClient(), cfg.WebhookMinDelaySec)
		logger.Info("using redis-backed webhook queue")
	} else {
		// In in-memory mode, use Redis-backed webhook queue only if a Redis URL is configured
		if cfg.RedisURL == "" {
			logger.Warn("no redis_url configured in in-memory mode; webhook queue will be disabled")
		} else {
			// Create a dedicated Redis connection just for webhooks
			logger.Info("connecting to redis for webhook queue", zap.String("url", cfg.RedisURL), zap.Int("db", cfg.RedisDB))
			redisClient, err := storage.NewRedisClient(cfg.RedisURL, cfg.RedisDB)
			if err != nil {
				logger.Fatal("failed to connect to redis for webhook queue", zap.Error(err))
			}
			webhookQueue = webhook.NewQueue(redisClient, cfg.WebhookMinDelaySec)
			logger.Info("using redis-backed webhook queue in in-memory storage mode")
		}
	}

	webhookManager := webhook.NewManager(cfg.WebhookURL, cfg.WebhookSecret, webhookQueue, store, cfg.DefaultOrganizationID)
	if webhookQueue != nil {
		webhookWorker = webhook.NewWorker(webhookQueue, webhookManager)

		// Start webhook worker in background
		webhookWorker.StartAsync()
		logger.Info("webhook worker started")
	} else {
		logger.Info("webhook queue is disabled; webhook worker will not be started")
	}
	h := handler.NewHandlerWithConfig(cfg, store, webhookManager)
	r := chi.NewRouter()

	// Built-in middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Custom request logger with detailed output
	r.Use(func(next http.Handler) http.Handler {
		return h.RequestLogger(next)
	})

	// Authentication middleware (applied to protected routes)
	if cfg.EnforceAuthentication {
		logger.Info("authentication enforcement enabled", zap.Strings("valid_app_ids", getAppIDList(cfg.ValidCredentials)))
		authMiddleware := auth.Middleware(cfg.ValidCredentials)
		r.Use(authMiddleware)
	} else {
		logger.Warn("authentication enforcement disabled")
	}

	setupRoutes(r, h)

	// Log unmatched routes to surface any misrouted traffic
	r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("route not found - no handler registered",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
		)
		http.NotFound(w, r)
	}))

	// Log method-not-allowed for visibility
	r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("method not allowed", zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.String("remote_addr", r.RemoteAddr))
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("gitlab.com/fynbos/mock/mockgatehub listening", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// Stop webhook worker
	if webhookWorker != nil {
		logger.Info("stopping webhook worker")
		webhookWorker.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped")
}

func setupRoutes(r chi.Router, h *handler.Handler) {
	logger.Info("========== SETTING UP ROUTES ==========")
	r.Get("/", h.RootHandler)
	r.Post("/transaction/complete", h.TransactionCompleteHandler)
	r.Get("/health", h.HealthCheck)
	r.Get("/api/user-currencies", h.GetUserCurrencies)
	r.Route("/auth/v1", func(r chi.Router) {
		logger.Info("REGISTERING /auth/v1 ROUTES")
		r.Post("/tokens", h.CreateToken)
		r.Post("/users/managed", h.CreateManagedUser)
		r.Get("/users/managed", h.GetManagedUser)
		r.Put("/users/managed/email", h.UpdateManagedUserEmail)
		r.Patch("/users/organization/{organizationID}", h.UpdateOrganizationConfiguration)
	})
	r.Route("/id/v1", func(r chi.Router) {
		logger.Info("REGISTERING /id/v1 ROUTES")
		r.Get("/users/{userID}", h.GetUser)
		r.Post("/users/{userID}/hubs/{gatewayID}", h.StartKYC)
		r.Put("/hubs/{gatewayID}/users/{userID}", h.UpdateKYCState)
		r.Post("/hubs/{gatewayID}/users/{userID}/overrideRiskLevel", h.OverrideRiskLevel)
	})
	r.Get("/iframe/onboarding", h.KYCIframe)
	r.Post("/iframe/submit", h.KYCIframeSubmit)
	r.Get("/admin/fees", h.GetFees)
	r.Put("/admin/fees", h.SetFees)
	r.Get("/admin/users/{userID}/fees", h.GetUserFees)
	r.Put("/admin/users/{userID}/fees", h.SetUserFees)
	r.Delete("/admin/users/{userID}/fees", h.ClearUserFees)
	r.Route("/core/v1", func(r chi.Router) {
		logger.Info("REGISTERING /core/v1 ROUTES")
		r.Get("/users/{userID}", h.GetUserWallets)
		r.Post("/users/{userID}/wallets", h.CreateWallet)
		r.Get("/users/{userID}/wallets/{walletID}", h.GetWallet)
		r.Get("/wallets/{walletID}/balances", h.GetWalletBalance)
		r.Post("/transactions", h.CreateTransaction)
		r.Get("/transactions/{txID}", h.GetTransaction)
	})
	r.Route("/rates/v1", func(r chi.Router) {
		logger.Info("REGISTERING /rates/v1 ROUTES")
		r.Get("/rates/current", h.GetCurrentRates)
		r.Get("/liquidity_provider/vaults", h.GetVaults)
	})
	r.Route("/cards/v1", func(r chi.Router) {
		logger.Info("========== REGISTERING /cards/v1 ROUTES ==========")
		// Generic customer handler
		r.Post("/customers", h.CreateCustomer)

		// Handlers for managed customers
		r.Post("/customers/managed", h.CreateManagedCustomer)

		// Handlers for customer addresses
		r.Post("/customers/{customerID}/addresses", h.CreateCustomerAddress)
		r.Get("/customers/{customerID}/addresses", h.GetCustomerAddresses)

		// Handlers for additional cards
		r.Post("/accounts/{accountID}/cards", h.OrderAdditionalCard)
		r.Post("/cards/{accountID}/card", h.OrderAdditionalCard)

		// Card handlers - note: order matters for chi routing
		r.Get("/cards/{customerID}", h.ListCards)
		r.Post("/cards", h.CreateCard)
		r.Get("/cards/{cardID}/card", h.GetCard)
		r.Delete("/cards/{cardID}/card", h.DeleteCard)
		r.Put("/cards/{cardID}/lock", h.LockCard)
		r.Put("/cards/{cardID}/unlock", h.UnlockCard)
		r.Put("/cards/{cardID}/block", h.BlockCard)

		// Card limits
		r.Get("/cards/{cardID}/limits", h.GetCardLimits)
		r.Put("/cards/{cardID}/limits", h.UpdateCardLimits)

		// Card tokenization and security
		r.Post("/token/card-data", h.GetCardToken)

		// Card transactions
		r.Post("/transactions", h.CreateCardTransaction)
		r.Get("/transactions/{txID}", h.GetCardTransaction)
		r.Get("/cards/{cardID}/transactions", h.ListCardTransactions)

		// 3DS and confirmations
		r.Get("/transaction/pending-confirmations", h.GetPendingConfirmations)
		r.Post("/test/3ds/challenge", h.CreateThreeDSChallenge)
		r.Post("/transaction/{txID}", h.ConfirmThreeDS)

		// Card products and plastic ordering
		r.Get("/card-applications/{appID}/card-products", h.GetCardApplicationProducts)
		r.Post("/cards/{cardID}/plastic", h.OrderPlasticCard)

		logger.Info("========== /cards/v1 ROUTES REGISTERED ==========")
	})
	logger.Info("========== ALL ROUTES REGISTERED SUCCESSFULLY ==========")
}

// getAppIDList returns a list of registered app IDs for logging
func getAppIDList(validCredentials map[string]string) []string {
	appIDs := make([]string, 0, len(validCredentials))
	for appID := range validCredentials {
		appIDs = append(appIDs, appID)
	}
	return appIDs
}
