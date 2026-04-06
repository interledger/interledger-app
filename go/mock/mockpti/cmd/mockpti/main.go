package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"go.uber.org/zap"

	"gitlab.com/fynbos/mock/mockpti/internal/config"
	"gitlab.com/fynbos/mock/mockpti/internal/handler"
	"gitlab.com/fynbos/mock/mockpti/internal/jobs"
	"gitlab.com/fynbos/mock/mockpti/internal/logger"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
)

func main() {
	if handled, exitCode := handleCLI(os.Args[1:], os.Stdout, os.Stderr); handled {
		os.Exit(exitCode)
	}

	runServer()
}

func runServer() {
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

	// Initialize queue/worker and handler
	queue := jobs.NewQueue(store)
	worker := jobs.NewWorker(queue)

	router := chi.NewRouter()
	h := handler.NewHandler(store, cfg)
	h.SetQueue(queue)

	worker.RegisterHandler(jobs.JobTypeUserAssessmentWebhook, h.NewUserAssessmentWebhookJobHandler())
	worker.RegisterHandler(jobs.JobTypeTransactionStatusWebhook, h.NewTransactionStatusWebhookJobHandler())
	worker.StartAsync()

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
	worker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Shutdown error: %v", err)
	}

	logger.Infof("Server stopped")
}

func handleCLI(args []string, stdout, stderr io.Writer) (bool, int) {
	if len(args) == 0 {
		return false, 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		printCLIHelp(stdout)
		return true, 0
	case "gen-webhook-settings":
		return true, runGenerateWebhookSettings(args[1:], stdout, stderr)
	case "derive-public-jwk":
		return true, runDerivePublicJWK(args[1:], stdout, stderr)
	default:
		return false, 0
	}
}

func printCLIHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "MockPTI CLI")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Commands:")
	_, _ = fmt.Fprintln(w, "  gen-webhook-settings   Generate local webhook JWK settings for wallet/backend + mockpti")
	_, _ = fmt.Fprintln(w, "  derive-public-jwk      Read private JWK from stdin and print matching public JWK")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Examples:")
	_, _ = fmt.Fprintln(w, "  mockpti gen-webhook-settings")
	_, _ = fmt.Fprintln(w, "  cat private.jwk.json | mockpti derive-public-jwk")
}

func runGenerateWebhookSettings(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gen-webhook-settings", flag.ContinueOnError)
	fs.SetOutput(stderr)
	bits := fs.Int("bits", 2048, "RSA key size")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *bits < 2048 {
		_, _ = fmt.Fprintln(stderr, "error: --bits must be >= 2048")
		return 2
	}

	backendDecryptKey, err := rsa.GenerateKey(rand.Reader, *bits)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error generating backend decrypt key: %v\n", err)
		return 1
	}

	webhookSigningKey, err := rsa.GenerateKey(rand.Reader, *bits)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error generating webhook signing key: %v\n", err)
		return 1
	}

	backendDecryptJWK, err := marshalPrivateJWK(backendDecryptKey)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error marshaling backend decrypt JWK: %v\n", err)
		return 1
	}

	webhookSigningPrivateJWK, err := marshalPrivateJWK(webhookSigningKey)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error marshaling webhook signing JWK: %v\n", err)
		return 1
	}

	webhookSigningPublicJWK, err := marshalPublicJWK(&webhookSigningKey.PublicKey)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error marshaling webhook signing public JWK: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintln(stdout, "# .env values")
	_, _ = fmt.Fprintf(stdout, "BACKEND_PTI_JWK=%s\n", backendDecryptJWK)
	_, _ = fmt.Fprintf(stdout, "BACKEND_PTI_PUBLIC_KEY_JWK=%s\n", webhookSigningPublicJWK)
	_, _ = fmt.Fprintf(stdout, "MOCKPTI_WEBHOOK_SIGNING_JWK=%s\n", webhookSigningPrivateJWK)
	_, _ = fmt.Fprintf(stdout, "MOCKPTI_WEBHOOK_ENCRYPTION_JWK=%s\n", backendDecryptJWK)

	_, _ = fmt.Fprintln(stdout, "")
	_, _ = fmt.Fprintln(stdout, "# wallet.yaml (backend service env)")
	_, _ = fmt.Fprintf(stdout, "- PTI_JWK=${BACKEND_PTI_JWK:-%s}\n", backendDecryptJWK)
	_, _ = fmt.Fprintf(stdout, "- PTI_PUBLIC_KEY_JWK=${BACKEND_PTI_PUBLIC_KEY_JWK:-%s}\n", webhookSigningPublicJWK)

	_, _ = fmt.Fprintln(stdout, "")
	_, _ = fmt.Fprintln(stdout, "# mockpti.yaml (mockpti service env)")
	_, _ = fmt.Fprintf(stdout, "MOCKPTI_WEBHOOK_SIGNING_JWK: ${MOCKPTI_WEBHOOK_SIGNING_JWK:-%s}\n", webhookSigningPrivateJWK)
	_, _ = fmt.Fprintf(stdout, "MOCKPTI_WEBHOOK_ENCRYPTION_JWK: ${MOCKPTI_WEBHOOK_ENCRYPTION_JWK:-%s}\n", backendDecryptJWK)

	return 0
}

func runDerivePublicJWK(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("derive-public-jwk", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	privateJWKRaw, err := io.ReadAll(os.Stdin)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error reading stdin: %v\n", err)
		return 1
	}

	privateJWKText := strings.TrimSpace(string(privateJWKRaw))
	if privateJWKText == "" {
		_, _ = fmt.Fprintln(stderr, "error: provide private JWK JSON on stdin")
		return 2
	}

	privateKey, err := jwk.ParseKey([]byte(privateJWKText))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error parsing private JWK: %v\n", err)
		return 1
	}

	publicKey, err := jwk.PublicKeyOf(privateKey)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error deriving public JWK: %v\n", err)
		return 1
	}

	b, err := json.Marshal(publicKey)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error marshaling public JWK: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintln(stdout, string(b))
	return 0
}

func marshalPrivateJWK(key *rsa.PrivateKey) (string, error) {
	jwkKey, err := jwk.FromRaw(key)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(jwkKey)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func marshalPublicJWK(key *rsa.PublicKey) (string, error) {
	jwkKey, err := jwk.FromRaw(key)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(jwkKey)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func setupRoutes(router *chi.Mux, h *handler.Handler) {
	// Health check (public)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})
	router.Get("/sdk/index.js", h.SDKScript)
	router.Get("/forms", h.FormsLanding)
	router.Post("/forms/complete", h.CompleteAssessmentFromForm)

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
