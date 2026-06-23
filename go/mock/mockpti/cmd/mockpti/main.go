package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
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
	"go.uber.org/zap"

	"github.com/interledger/interledger-app/go/mock/mockpti/internal/config"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/handler"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/jobs"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/logger"
	"github.com/interledger/interledger-app/go/mock/mockpti/internal/storage"
)

func main() {
	if handled, exitCode := handleCLI(os.Args[1:], os.Stdout, os.Stderr); handled {
		os.Exit(exitCode)
	}

	runServer()
}

func runServer() {
	b64Key := os.Getenv("MOCKPTI_WEBHOOK_SIGNING_KEY_B64")
	if b64Key == "" {
		fmt.Fprintln(os.Stderr, "fatal: MOCKPTI_WEBHOOK_SIGNING_KEY_B64 is required")
		os.Exit(1)
	}
	if _, err := base64.StdEncoding.DecodeString(b64Key); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: MOCKPTI_WEBHOOK_SIGNING_KEY_B64 is not valid base64: %v\n", err)
		os.Exit(1)
	}

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
	case "derive-public-pem":
		return true, runDerivePublicPEM(args[1:], stdout, stderr)
	default:
		return false, 0
	}
}

func printCLIHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "MockPTI CLI")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Commands:")
	_, _ = fmt.Fprintln(w, "  gen-webhook-settings   Generate Ed25519 webhook key pair for wallet/backend + mockpti")
	_, _ = fmt.Fprintln(w, "  derive-public-pem      Read Ed25519 private key PEM from stdin and print matching public key PEM")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Examples:")
	_, _ = fmt.Fprintln(w, "  mockpti gen-webhook-settings")
	_, _ = fmt.Fprintln(w, "  cat private.pem | mockpti derive-public-pem")
}

func runGenerateWebhookSettings(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gen-webhook-settings", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error generating Ed25519 key: %v\n", err)
		return 1
	}

	privPEM, err := marshalPrivateKeyPEM(priv)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error marshaling private key: %v\n", err)
		return 1
	}

	pubPEM, err := marshalPublicKeyPEM(pub)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error marshaling public key: %v\n", err)
		return 1
	}

	pubOneLine := strings.ReplaceAll(strings.TrimSpace(pubPEM), "\n", `\n`)
	privB64 := base64.StdEncoding.EncodeToString([]byte(privPEM))

	_, _ = fmt.Fprintln(stdout, "# Private key (for mockpti) — keep secret")
	_, _ = fmt.Fprintln(stdout, privPEM)
	_, _ = fmt.Fprintln(stdout, "# Public key (for backend wallet)")
	_, _ = fmt.Fprintln(stdout, pubPEM)

	_, _ = fmt.Fprintln(stdout, "# local/wallet.yaml (backend service env) — \\n will be expanded by the backend")
	_, _ = fmt.Fprintf(stdout, "- PTI_PUBLIC_KEY_JWK=${BACKEND_PTI_PUBLIC_KEY_JWK:-%s}\n", pubOneLine)

	_, _ = fmt.Fprintln(stdout, "")
	_, _ = fmt.Fprintln(stdout, "# local/mockpti.yaml (mockpti service env) — base64-encoded PEM")
	_, _ = fmt.Fprintf(stdout, "MOCKPTI_WEBHOOK_SIGNING_KEY_B64: ${MOCKPTI_WEBHOOK_SIGNING_KEY_B64:-%s}\n", privB64)

	return 0
}

func runDerivePublicPEM(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("derive-public-pem", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	privPEMRaw, err := io.ReadAll(os.Stdin)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error reading stdin: %v\n", err)
		return 1
	}

	privPEMText := strings.TrimSpace(string(privPEMRaw))
	if privPEMText == "" {
		_, _ = fmt.Fprintln(stderr, "error: provide Ed25519 private key PEM on stdin")
		return 2
	}

	block, _ := pem.Decode([]byte(privPEMText))
	if block == nil {
		_, _ = fmt.Fprintln(stderr, "error: failed to decode PEM block")
		return 1
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error parsing private key: %v\n", err)
		return 1
	}

	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		_, _ = fmt.Fprintln(stderr, "error: key is not Ed25519")
		return 1
	}

	pubPEM, err := marshalPublicKeyPEM(edKey.Public().(ed25519.PublicKey))
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error marshaling public key: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprint(stdout, pubPEM)
	return 0
}

func marshalPrivateKeyPEM(key ed25519.PrivateKey) (string, error) {
	b, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b})), nil
}

func marshalPublicKeyPEM(key ed25519.PublicKey) (string, error) {
	b, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: b})), nil
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
