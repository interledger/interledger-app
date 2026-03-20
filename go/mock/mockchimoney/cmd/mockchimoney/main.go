package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/fynbos/mock/mockchimoney/internal/config"
	"gitlab.com/fynbos/mock/mockchimoney/internal/handler"
	"gitlab.com/fynbos/mock/mockchimoney/internal/logger"

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
		zap.Bool("enforce_authentication", cfg.EnforceAuthentication),
	)

	// Initialize handler
	h := handler.NewHandler(cfg)

	// Create chi router
	r := chi.NewRouter()

	// Add middleware
	r.Use(h.RequestLogger)
	r.Use(middleware.Recoverer)

	// Register routes
	r.Get("/health", h.Health)
	r.Group(func(r chi.Router) {
		r.Use(handler.APIKeyMiddleware(cfg.APIKey, cfg.EnforceAuthentication))
		r.Post("/v0.2.4/multicurrency-wallets/create", h.CreateWallet)
		r.Get("/v0.2.4/multicurrency-wallets/get", h.GetWallet)
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

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
	}

	logger.Info("MockChimoney shut down successfully")
}
