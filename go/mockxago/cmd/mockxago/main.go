package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"gitlab.com/fynbos/mockxago/internal/handler"
	"gitlab.com/fynbos/mockxago/internal/logger"
	"gitlab.com/fynbos/mockxago/internal/storage"
)

func main() {
	// Load configuration
	port := os.Getenv("XAGO_MOCK_PORT")
	if port == "" {
		port = "8080"
	}

	publicKey := os.Getenv("XAGO_API_PUBLIC_KEY")
	if publicKey == "" {
		publicKey = "test-public-key"
		logger.Infof("Using default XAGO_API_PUBLIC_KEY: %s", publicKey)
	}

	secret := os.Getenv("XAGO_API_SECRET")
	if secret == "" {
		secret = "test-secret"
		logger.Infof("Using default XAGO_API_SECRET: %s", secret)
	}

	// Initialize storage
	store := storage.NewMemoryStorage()
	logger.Infof("Initialized in-memory storage")

	// Initialize router
	router := chi.NewRouter()

	// Create handler
	h := handler.NewHandler(store)

	// Setup routes
	setupRoutes(router, h)

	// Create server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Infof("Starting MockXago server on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	logger.Infof("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Shutdown error: %v", err)
	}

	logger.Infof("Server stopped")
}

func setupRoutes(router *chi.Mux, h *handler.Handler) {
	router.Route("/xago/v1", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Get("/currencies", h.ListCurrencies)

		r.Group(func(pr chi.Router) {
			pr.Use(h.AuthMiddleware)
			pr.Post("/company/accounts", h.CreateSubAccount)
			pr.Put("/company/accounts/{accountId}", h.UpdateSubAccount)
			pr.Get("/company/accounts", h.GetSubAccountByWallet)
			pr.Get("/accounts/{accountId}/balance", h.GetBalance)
			pr.Post("/test/balances/set", h.TestSetBalance)
			pr.Post("/test/balances/deposit", h.TestDeposit)
			pr.Post("/test/balances/transfer", h.TestTransfer)
		})
	})

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})
}
