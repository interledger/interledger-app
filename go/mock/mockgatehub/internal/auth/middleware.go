package auth

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"

	"go.uber.org/zap"
)

// PublicEndpoints are endpoints that don't require authentication
var PublicEndpoints = map[string]bool{
	"/health":               true,
	"/":                     true, // Root handler (iframe serving)
	"/iframe/onboarding":    true, // KYC onboarding iframe (loaded in browser)
	"/iframe/submit":        true, // Iframe form submission
	"/transaction/complete": true, // Iframe completion callback
	"/api/user-currencies":  true, // Iframe currency lookup
	"/admin/fees":           true, // Admin fee configuration (test support)
}

// PublicEndpointPatterns are path patterns (with placeholders) that don't require authentication
var PublicEndpointPatterns = []string{
	"/admin/users/*/fees", // User-specific fee configuration (test support)
}

// matchesPublicPattern checks if a path matches any of the public endpoint patterns
func matchesPublicPattern(path string) bool {
	for _, pattern := range PublicEndpointPatterns {
		if matchPattern(pattern, path) {
			return true
		}
	}
	return false
}

// matchPattern checks if a path matches a pattern where * is a wildcard for a single path segment
func matchPattern(pattern, path string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i := range patternParts {
		if patternParts[i] != "*" && patternParts[i] != pathParts[i] {
			return false
		}
	}
	return true
}

// Middleware returns an HTTP middleware that validates HMAC signatures
func Middleware(validCredentials map[string]string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for public endpoints (exact match)
			if PublicEndpoints[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Skip authentication for public endpoint patterns
			if matchesPublicPattern(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract headers
			appID := r.Header.Get("x-gatehub-app-id")
			timestamp := r.Header.Get("x-gatehub-timestamp")
			signature := r.Header.Get("x-gatehub-signature")

			// Missing headers = unauthorized
			if appID == "" || timestamp == "" || signature == "" {
				logger.Error("auth failure: missing authentication headers",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("app_id", appID),
					zap.String("timestamp", timestamp),
					zap.String("signature", signature),
				)
				http.Error(w, `{"error": {"status_code": 401, "status": "Unauthorized"}}`, http.StatusUnauthorized)
				return
			}

			// Check if app is registered
			secret, exists := validCredentials[appID]
			if !exists {
				logger.Error("auth failure: unknown app id",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("received_app_id", appID),
					zap.Strings("registered_app_ids", getRegisteredAppIDs(validCredentials)),
				)
				http.Error(w, `{"error": {"status_code": 401, "status": "Unauthorized"}}`, http.StatusUnauthorized)
				return
			}

			// Read and preserve request body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("auth failure: failed to read request body",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("app_id", appID),
					zap.Error(err),
				)
				http.Error(w, `{"error": {"status_code": 400, "status": "Bad Request"}}`, http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			fullURL := reconstructURL(r)
			expectedSig := GenerateSignature(timestamp, r.Method, fullURL, string(bodyBytes), secret)
			if !bytes.Equal([]byte(signature), []byte(expectedSig)) {
				pathOnlyURL := r.URL.Path
				if r.URL.RawQuery != "" {
					pathOnlyURL = r.URL.Path + "?" + r.URL.RawQuery
				}
				expectedPathSig := GenerateSignature(timestamp, r.Method, pathOnlyURL, string(bodyBytes), secret)
				logger.Error("auth failure: signature mismatch",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("app_id", appID),
					zap.String("received_signature", signature),
					zap.String("expected_signature_full", expectedSig),
					zap.String("expected_signature_path_only", expectedPathSig),
					zap.String("secret", secret),
					zap.String("timestamp", timestamp),
					zap.String("body", string(bodyBytes)),
				)
				http.Error(w, `{"error": {"status_code": 401, "status": "Unauthorized"}}`, http.StatusUnauthorized)
				return
			}

			logger.Info("auth success",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("app_id", appID),
			)
			next.ServeHTTP(w, r)
		})
	}
}

// getRegisteredAppIDs returns a list of registered app IDs (for logging purposes)
func getRegisteredAppIDs(validCredentials map[string]string) []string {
	appIDs := make([]string, 0, len(validCredentials))
	for appID := range validCredentials {
		appIDs = append(appIDs, appID)
	}
	return appIDs
}

func reconstructURL(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}

	fullURL := scheme + "://" + host + r.URL.Path
	if r.URL.RawQuery != "" {
		fullURL += "?" + r.URL.RawQuery
	}
	return fullURL
}
