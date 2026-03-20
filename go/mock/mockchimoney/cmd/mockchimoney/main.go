package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/fynbos/mock/mockchimoney/internal/config"
	"gitlab.com/fynbos/mock/mockchimoney/internal/handler"
	"gitlab.com/fynbos/mock/mockchimoney/internal/jobs"
	"gitlab.com/fynbos/mock/mockchimoney/internal/logger"
	"gitlab.com/fynbos/mock/mockchimoney/internal/storage"
	"gitlab.com/fynbos/mock/mockchimoney/internal/webhook"

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

	logger.Info("gitlab.com/fynbos/mock/mockchimoney build info", zap.String("build_time", buildTime))
	logger.Info("starting MockChimoney")

	// Log all startup configuration at INFO level
	logger.Info("configuration loaded",
		zap.String("log_level", cfg.LogLevel),
		zap.String("port", cfg.Port),
		zap.String("redis_url", cfg.RedisURL),
		zap.Int("redis_db", cfg.RedisDB),
		zap.Bool("enforce_authentication", cfg.EnforceAuthentication),
	)

	var (
		store       storage.Store
		storeCloser io.Closer
	)

	if cfg.RedisURL != "" {
		redisStore, err := storage.NewRedisStore(cfg.RedisURL, cfg.RedisDB)
		if err != nil {
			logger.Fatal("failed to initialize redis store", zap.Error(err))
		}
		store = redisStore
		storeCloser = redisStore
		logger.Info("using redis-backed store")
	} else {
		store = storage.NewMemoryStore()
		logger.Info("using in-memory store")
	}
	queue := jobs.NewInMemoryQueue(64)
	sender := webhook.NewSender(&http.Client{Timeout: 10 * time.Second})
	h := handler.NewHandlerWithDeps(cfg, store, queue, sender)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go jobs.StartWorker(workerCtx, queue)

	// Create chi router
	r := chi.NewRouter()

	// Add middleware
	r.Use(h.RequestLogger)
	r.Use(middleware.Recoverer)

	// Register routes
	r.Get("/health", h.Health)
	r.Get("/pay/{issueID}", h.PayPage)
	r.Post("/pay/{issueID}/confirm", h.ConfirmPayPage)
	r.Get("/verify/kyc/{externalID}", h.KYCPage)
	r.Post("/verify/kyc/{externalID}/approve", h.KYCApprove)
	r.Post("/verify/kyc/{externalID}/decline", h.KYCDecline)
	r.Group(func(r chi.Router) {
		r.Use(handler.APIKeyMiddleware(cfg.APIKey, cfg.EnforceAuthentication))
		r.Post("/v0.2.4/multicurrency-wallets/create", h.CreateWallet)
		r.Get("/v0.2.4/multicurrency-wallets/get", h.GetWallet)
		r.Post("/v0.2.4/multicurrency-wallets/transfer", h.Transfer)
		r.Post("/v0.2.4/payment/initiate", h.InitiatePayment)
		r.Post("/v0.2.4/payment/verify", h.VerifyPayment)
		r.Post("/v0.2.4/payouts/interac", h.PayoutInterac)
		r.Post("/v0.2.4/payouts/status", h.PayoutStatus)
		r.Post("/v0.2.4/info/fee-estimate", h.FeeEstimate)
		r.Get("/v0.2.4/info/convert/local-amount-to-usd", h.ConvertLocalAmountToUSD)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("starting HTTP server", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down MockChimoney")
	workerCancel()
	queue.Close()
	if storeCloser != nil {
		if err := storeCloser.Close(); err != nil {
			logger.Error("failed to close store", zap.Error(err))
		}
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
	}

	logger.Info("MockChimoney shut down successfully")
}
