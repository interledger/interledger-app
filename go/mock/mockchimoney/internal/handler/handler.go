package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/config"
	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/jobs"
	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/logger"
	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/storage"
	"github.com/interledger/interledger-app/go/mock/mockchimoney/internal/webhook"

	"go.uber.org/zap"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	config *config.Config
	store  storage.Store
	queue  jobs.Queue
	sender *webhook.Sender
}

// NewHandler creates a new handler using configuration loaded from the environment.
func NewHandler(cfg *config.Config) *Handler {
	return NewHandlerWithDeps(cfg, storage.NewMemoryStore(), jobs.NewNoopQueue(), webhook.NewSender(&http.Client{Timeout: 10 * time.Second}))
}

// NewHandlerWithStore creates a new handler with an explicit store implementation.
func NewHandlerWithStore(cfg *config.Config, store storage.Store) *Handler {
	return NewHandlerWithDeps(cfg, store, jobs.NewNoopQueue(), webhook.NewSender(&http.Client{Timeout: 10 * time.Second}))
}

// NewHandlerWithDeps creates a new handler with explicit dependencies.
func NewHandlerWithDeps(cfg *config.Config, store storage.Store, queue jobs.Queue, sender *webhook.Sender) *Handler {
	if cfg == nil {
		cfg = config.Load()
	}
	if store == nil {
		store = storage.NewMemoryStore()
	}
	if queue == nil {
		queue = jobs.NewNoopQueue()
	}
	if sender == nil {
		sender = webhook.NewSender(&http.Client{Timeout: 10 * time.Second})
	}

	logger.Info("initializing http handlers")
	return &Handler{
		config: cfg,
		store:  store,
		queue:  queue,
		sender: sender,
	}
}

// sensitiveHeaders lists header names whose values must be redacted in logs.
var sensitiveHeaders = map[string]struct{}{
	"X-Api-Key":     {},
	"Authorization": {},
	"Cookie":        {},
}

// RequestLogger middleware logs incoming requests with safe details.
func (h *Handler) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Re-buffer the body so downstream handlers can still read it
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		// Collect headers, redacting sensitive values
		headers := make(map[string]string, len(r.Header))
		for k, v := range r.Header {
			if _, redact := sensitiveHeaders[http.CanonicalHeaderKey(k)]; redact {
				headers[k] = "[REDACTED]"
			} else {
				headers[k] = strings.Join(v, ", ")
			}
		}

		logger.Debug("request incoming",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Any("headers", headers),
		)

		// Wrap ResponseWriter to capture status code
		wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		logger.Debug("request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", duration),
		)
	})
}

// statusRecorder wraps http.ResponseWriter to capture the status code
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// Health returns a simple health check response
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"status": "ok"}
	json.NewEncoder(w).Encode(response)
}
