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

	"gitlab.com/fynbos/mock/mockplaid/internal/config"
	"gitlab.com/fynbos/mock/mockplaid/internal/handler"
	"gitlab.com/fynbos/mock/mockplaid/internal/logger"
	"gitlab.com/fynbos/mock/mockplaid/internal/storage"
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
				logger.Fatal("invalid MOCKPLAID_REDIS_DB value", zap.String("value", cfg.RedisDB), zap.Error(err))
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

	router := chi.NewRouter()
	h := handler.NewHandler(store, cfg)
	setupRoutes(router, h)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		logger.Infof("Starting MockPlaid server on port %s", cfg.Port)
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

// setupRoutes wires the HTTP routes. The Plaid REST surface is registered as
// 501 stubs (filled in one endpoint per task, MP7+).
func setupRoutes(router *chi.Mux, h *handler.Handler) {
	// Health check (public)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Test-only: wipe state for e2e isolation.
	router.Post("/test/reset", h.Reset)

	// Browser-facing mock Link surface (served on the cdn.plaid.com host).
	router.Get("/link/v2/stable/link-initialize.js", h.LinkInitializeJS)
	router.Get("/link", h.LinkPage)

	// Plaid REST surface (the 11 endpoints the backend SDK calls). 501 stubs
	// until their per-task implementations land.
	router.Post("/link/token/create", h.CreateLinkToken)
	// Mock-only control endpoint the Link UI calls to pick a bank + account.
	router.Post("/link/session/select", h.SelectAccount)
	router.Post("/item/public_token/exchange", h.ExchangePublicToken)
	router.Post("/item/get", h.ItemGet)
	router.Post("/institutions/get_by_id", h.InstitutionsGetByID)
	router.Post("/processor/token/create", h.CreateProcessorToken)
	router.Post("/item/remove", h.RemoveItem)
	router.Post("/accounts/get", h.GetAccounts)
	router.Post("/auth/get", h.GetAuth)
	router.Post("/accounts/balance/get", h.GetBalance)
	router.Post("/identity/get", h.GetIdentity)
	router.Post("/transactions/sync", h.TransactionsSync)
}
