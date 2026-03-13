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

	"gitlab.com/fynbos/mock/mockxago/internal/handler"
	"gitlab.com/fynbos/mock/mockxago/internal/jobs"
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

	// Initialize storage (memory or Redis)
	var store storage.Storage
	redisURL := os.Getenv("MOCKXAGO_REDIS_URL")
	if redisURL != "" {
		redisDB := 0
		if dbStr := os.Getenv("MOCKXAGO_REDIS_DB"); dbStr != "" {
			var err error
			redisDB, err = strconv.Atoi(dbStr)
			if err != nil {
				logger.Fatal("invalid MOCKXAGO_REDIS_DB value", zap.String("value", dbStr), zap.Error(err))
			}
		}
		var err error
		store, err = storage.NewRedisStorage(redisURL, redisDB)
		if err != nil {
			logger.Fatal("failed to initialize Redis storage", zap.Error(err))
		}
		logger.Infof("Initialized Redis storage at %s (DB %d)", redisURL, redisDB)
	} else {
		store = storage.NewMemoryStorage()
		logger.Infof("Initialized in-memory storage")
	}

	// Initialize job queue and worker
	queue := jobs.NewQueue(store)
	worker := jobs.NewWorker(queue)

	// Initialize router
	router := chi.NewRouter()

	// Create handler
	h := handler.NewHandler(store, queue)

	// Register job handlers BEFORE starting worker
	worker.RegisterHandler(handler.JobTypeProcessDeposit, h.NewProcessDepositHandler())
	logger.Infof("Registered deposit job handler")

	// Start job worker in background
	worker.StartAsync()
	logger.Infof("Job worker started")

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

	// Stop worker first
	worker.Stop()
	logger.Infof("Job worker stopped")

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

		// Public endpoint (no auth required)
		r.Get("/currencies", h.ListCurrencies)

		// Persona SDK compatible endpoints (no auth required)
		// These are accessed by the backend Persona client and the browser iframe.
		r.Post("/inquiries", h.PersonaCreateInquiry)
		r.Get("/inquiries/{inquiryId}", h.PersonaGetInquiry)
		r.Get("/inquiries/{inquiryId}/iframe", h.PersonaGetInquiryIframe)
		r.Post("/inquiries/{inquiryId}/submit", h.PersonaInquirySubmit)
		r.Get("/accounts/{accountId}", h.PersonaGetAccount)
		r.Post("/accounts/{accountId}/remove-tag", h.PersonaRemoveTag)

		r.Group(func(pr chi.Router) {
			pr.Use(h.AuthMiddleware)
			pr.Post("/company/accounts", h.CreateSubAccount)
			pr.Put("/company/accounts/{accountId}", h.UpdateSubAccount)
			pr.Get("/company/accounts", h.GetSubAccountByWallet)
			pr.Post("/company/accounts/testdeposit", h.SimulateTestDeposit)
			pr.Get("/company/deposits", h.ListCompanyDeposits)
			pr.Get("/accounts/{accountId}/balance", h.GetBalance)
			pr.Post("/accounts/{accountId}/beneficiaries", h.AddBeneficiary)
			pr.Get("/accounts/{accountId}/beneficiaries", h.ListBeneficiaries)

			// Global beneficiary endpoints (API compliance aliases)
			pr.Post("/beneficiaries", h.AddBeneficiaryGlobal)
			pr.Get("/beneficiaries", h.ListBeneficiariesGlobal)

			pr.Post("/transfers", h.CreateTransfer)
			pr.Get("/transfers", h.ListTransfers)
			pr.Get("/transfers/{id}", h.GetTransaction)
			pr.Get("/company/transactions", h.ListTransactions)
			pr.Get("/company/transactions/{id}", h.GetTransaction)

			// Transaction query endpoint (API compliance)
			pr.Get("/transactions", h.GetTransactionByQuery)

			pr.Post("/test/balances/set", h.TestSetBalance)
			pr.Post("/test/balances/deposit", h.TestDeposit)
			pr.Post("/test/balances/transfer", h.TestTransfer)
			pr.Post("/test/transactions", h.TestCreateTransaction)
			pr.Post("/test/deposits/clear", h.TestClearDeposits)
		})

		// Reset endpoint (outside auth middleware, but protected by ensureTestMode)
		r.Post("/test/reset", h.TestReset)
	})

	// KYC endpoints (public, not under /v1)
	router.Get("/kyc/iframe", h.KYCIframe)
	router.Post("/kyc/submit", h.KYCIframeSubmit)

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})
}
