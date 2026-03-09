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

	"gitlab.com/fynbos/mock/mockxago/internal/handler"
	"gitlab.com/fynbos/mock/mockxago/internal/logger"
	"gitlab.com/fynbos/mock/mockxago/internal/storage"
)

func main() {
	// Initialize logger with configured log level
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // Default to info level
	}
	if err := logger.Initialize(logLevel); err != nil {
		logger.Fatalln(err)
	}

	// Load configuration
	port := os.Getenv("XAGO_MOCK_PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("XAGO_API_PUBLIC_KEY") == "" {
		defaultPublicKey := "test-public-key"
		if err := os.Setenv("XAGO_API_PUBLIC_KEY", defaultPublicKey); err != nil {
			logger.Errorf("Failed to set default XAGO_API_PUBLIC_KEY: %v", err)
		} else {
			logger.Infof("Using default XAGO_API_PUBLIC_KEY: %s", defaultPublicKey)
		}
	}

	if os.Getenv("XAGO_API_SECRET") == "" {
		defaultSecret := "test-secret"
		if err := os.Setenv("XAGO_API_SECRET", defaultSecret); err != nil {
			logger.Errorf("Failed to set default XAGO_API_SECRET: %v", err)
		} else {
			logger.Infof("Using default XAGO_API_SECRET: %s", defaultSecret)
		}
	}

	// Initialize in-memory storage
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
			logger.Fatalln(err)
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
	// Canonical /v1 prefix — matches XAGO_API_BASE_URL=http://mockxago:8080/v1 in wallet.yaml
	router.Route("/v1", func(r chi.Router) {
		r.Post("/login", h.Login)

		r.Group(func(pr chi.Router) {
			pr.Use(h.AuthMiddleware)
			pr.Get("/example-route", h.ExampleProtectedRoute)
			pr.Post("/company/accounts", h.CreateSubAccount)
			pr.Put("/company/accounts/{accountId}", h.UpdateSubAccount)
			pr.Get("/company/accounts", h.GetSubAccountByWallet)

			// Beneficiary management
			pr.Post("/accounts/{accountId}/beneficiaries", h.AddBeneficiary)
			pr.Get("/accounts/{accountId}/beneficiaries", h.ListBeneficiaries)
		})
	})

	// Test-only endpoint (outside auth middleware)
	router.Post("/v1/test/reset", h.TestReset)

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})
}
