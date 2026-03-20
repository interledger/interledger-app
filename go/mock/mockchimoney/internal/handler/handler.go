package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockchimoney/internal/config"
	"gitlab.com/fynbos/mock/mockchimoney/internal/logger"

	"go.uber.org/zap"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	config *config.Config
}

// NewHandler creates a new handler using configuration loaded from the environment.
func NewHandler(cfg *config.Config) *Handler {
	if cfg == nil {
		cfg = config.Load()
	}
	logger.Info("initializing http handlers")
	return &Handler{
		config: cfg,
	}
}

// RequestLogger middleware logs all incoming requests with full details
func (h *Handler) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Read and re-buffer the body so downstream handlers can still read it
		var bodyStr string
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				bodyStr = string(bodyBytes)
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		// Collect all headers into a map for structured logging
		headers := make(map[string]string, len(r.Header))
		for k, v := range r.Header {
			headers[k] = strings.Join(v, ", ")
		}

		logger.Debug("request incoming",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Any("headers", headers),
			zap.String("body", bodyStr),
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
