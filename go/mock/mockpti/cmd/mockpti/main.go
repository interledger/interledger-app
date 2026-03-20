package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"gitlab.com/fynbos/mock/mockpti/internal/config"
	"gitlab.com/fynbos/mock/mockpti/internal/handler"
	"gitlab.com/fynbos/mock/mockpti/internal/logger"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
)

func main() {
	cfg := config.Load()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		logger.Fatalln(err)
	}

	// Initialize storage (memory or Redis)
	var store storage.Storage
	if cfg.RedisURL != "" {
		redisDB := 0
		if cfg.RedisDB != "" {
			var err error
			redisDB, err = strconv.Atoi(cfg.RedisDB)
			if err != nil {
				logger.Fatal("invalid MOCKPTI_REDIS_DB value", zap.String("value", cfg.RedisDB), zap.Error(err))
			}
		}
		var err error
		store, err = storage.NewRedisStorage(cfg.RedisURL, redisDB)
		if err != nil {
			logger.Fatal("failed to initialize Redis storage", zap.Error(err))
		}
		logger.Infof("Initialized Redis storage at %s (DB %d)", cfg.RedisURL, redisDB)
	} else {
		store = storage.NewMemoryStorage()
		logger.Infof("Initialized in-memory storage")
	}

	// Initialize router and handler
	router := chi.NewRouter()
	h := handler.NewHandler(store, cfg)

	setupRoutes(router, h)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		logger.Infof("Starting MockPTI server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalln(err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Infof("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Shutdown error: %v", err)
	}

	logger.Infof("Server stopped")
}

func setupRoutes(router *chi.Mux, h *handler.Handler) {
	// Health check (public)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Test-only routes (no auth, only present to support e2e scenario isolation)
	router.Post("/test/reset", h.Reset)

	// PTI API routes (require x-pti-client-id header)
	router.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)

		// User endpoints
		r.Post("/users", h.CreateUser)
		r.Get("/users/{id}", h.GetUser)
		r.Patch("/users", h.PatchUser)
		r.Put("/users", h.PutUser)

		// Assessment endpoints
		r.Post("/users/assessments", h.StartUserAssessment)
		r.Get("/users/{id}/assessments", h.GetUserAssessment)

		// Wallet endpoints
		r.Post("/users/{id}/wallets", h.CreateWallet)
		r.Get("/users/{id}/wallets", h.ListWallets)
		r.Get("/users/{id}/wallets/{walletId}", h.GetWallet)

		// Payment information endpoints
		r.Post("/users/{id}/payment-information", h.CreatePaymentInformation)
		r.Get("/users/{id}/payment-information/{piId}", h.GetPaymentInformation)

		// Transaction endpoints
		r.Post("/transactions/deposits", h.CreateDeposit)
		r.Post("/transactions/withdrawals", h.CreateWithdrawal)
		r.Post("/transactions/transfers", h.CreateTransfer)
		r.Get("/transactions/{requestId}", h.GetTransaction)
		r.Post("/transactions/{requestId}/updates", h.UpdateTransaction)

		// Auth endpoints
		r.Post("/auth/jwt", h.CreateJWT)
	})
}
