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
	"go.uber.org/zap"

	"github.com/interledger/interledger-app/go/mock/mockxago/internal/config"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/handler"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/jobs"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/logger"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/storage"
)

func main() {
	cfg := config.Load()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		logger.Fatalln(err)
	}

	var store storage.Storage
	if cfg.RedisURL != "" {
		var err error
		store, err = storage.NewRedisStorage(cfg.RedisURL, cfg.RedisDB)
		if err != nil {
			logger.Fatal("failed to initialize Redis storage", zap.Error(err))
		}
		logger.Infof("Initialized Redis storage at %s (DB %d)", cfg.RedisURL, cfg.RedisDB)
	} else {
		store = storage.NewMemoryStorage()
		logger.Infof("Initialized in-memory storage")
	}

	queue := jobs.NewQueue(store)
	worker := jobs.NewWorker(queue)
	router := chi.NewRouter()
	h := handler.NewHandler(store, queue, cfg)

	worker.RegisterHandler(handler.JobTypeProcessDeposit, h.NewProcessDepositHandler())
	logger.Infof("Registered deposit job handler")

	worker.StartAsync()
	logger.Infof("Job worker started")

	setupRoutes(router, h)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		logger.Infof("Starting MockXago server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalln(err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Infof("Shutting down server...")
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
		r.Get("/persona-sdk.js", h.PersonaSDK)

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
