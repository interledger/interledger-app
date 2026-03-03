package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"

	"go.uber.org/zap"
)

// Helper methods for JSON responses

func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Marshal to log the response
	body, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logger.Error("failed to marshal response", zap.Error(err))
		w.Write([]byte(`{"error":"internal server error"}`))
		return
	}

	logger.Debug("sending json response", zap.Int("status", status))
	w.Write(body)
}

func (h *Handler) sendError(w http.ResponseWriter, status int, message string) {
	logger.Error("sending error response", zap.Int("status", status), zap.String("message", message))
	h.sendJSON(w, status, models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// setCORSHeaders adds CORS headers to the response writer
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

// sendJSONWithCORS sends a JSON response with CORS headers
func (h *Handler) sendJSONWithCORS(w http.ResponseWriter, status int, data interface{}) {
	setCORSHeaders(w)
	h.sendJSON(w, status, data)
}

// sendErrorWithCORS sends an error response with CORS headers
func (h *Handler) sendErrorWithCORS(w http.ResponseWriter, status int, message string) {
	setCORSHeaders(w)
	h.sendError(w, status, message)
}

func (h *Handler) decodeJSON(r *http.Request, v interface{}) error {
	// Read body for logging
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	// Log the raw request body
	logger.Debug("received request body", zap.String("body", string(body)))

	// Restore body for decoding
	r.Body = io.NopCloser(bytes.NewReader(body))

	// Decode
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		logger.Error("failed to decode json", zap.Error(err))
		return err
	}

	// Log the decoded structure
	pretty, _ := json.MarshalIndent(v, "", "  ")
	logger.Debug("decoded request", zap.String("data", string(pretty)))

	return nil
}
