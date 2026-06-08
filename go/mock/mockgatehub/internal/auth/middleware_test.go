package auth

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReconstructURL(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		host    string
		tls     bool
		url     string
		want    string
	}{
		{
			name: "forwarded headers",
			headers: map[string]string{
				"X-Forwarded-Proto": "https",
				"X-Forwarded-Host":  "mockgatehub.interledger.test",
			},
			host: "localhost:25151",
			url:  "http://localhost/auth/v1/users/managed?clientId=abc",
			want: "https://mockgatehub.interledger.test/auth/v1/users/managed?clientId=abc",
		},
		{
			name:    "tls without forwarded headers",
			headers: map[string]string{},
			host:    "localhost:25151",
			tls:     true,
			url:     "http://localhost/id/v1/users/123",
			want:    "https://localhost:25151/id/v1/users/123",
		},
		{
			name:    "http fallback",
			headers: map[string]string{},
			host:    "localhost:25151",
			url:     "http://localhost/health",
			want:    "http://localhost:25151/health",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatalf("failed to build request: %v", err)
			}
			req.URL.Scheme = ""
			req.URL.Host = ""
			req.Host = tt.host
			if tt.tls {
				req.TLS = &tls.ConnectionState{}
			}
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			got := reconstructURL(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{
			name:    "exact match without wildcard",
			pattern: "/admin/fees",
			path:    "/admin/fees",
			want:    true,
		},
		{
			name:    "single wildcard matches user ID",
			pattern: "/admin/users/*/fees",
			path:    "/admin/users/00000000-0000-0000-0000-000000000001/fees",
			want:    true,
		},
		{
			name:    "wildcard does not match multiple segments",
			pattern: "/admin/*/fees",
			path:    "/admin/users/123/fees",
			want:    false,
		},
		{
			name:    "pattern with different path segments does not match",
			pattern: "/admin/users/*/fees",
			path:    "/admin/users/123/config",
			want:    false,
		},
		{
			name:    "pattern with different length does not match",
			pattern: "/admin/users/*/fees",
			path:    "/admin/users",
			want:    false,
		},
		{
			name:    "trailing slash is handled correctly",
			pattern: "/admin/users/*/fees",
			path:    "/admin/users/123/fees/",
			want:    true, // Trailing slash is trimmed, so it matches
		},
		{
			name:    "wildcard at different position",
			pattern: "/api/*/endpoint",
			path:    "/api/v1/endpoint",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPattern(tt.pattern, tt.path)
			assert.Equal(t, tt.want, got, "matchPattern(%q, %q)", tt.pattern, tt.path)
		})
	}
}

func TestMatchesPublicPattern(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "user-specific fee GET endpoint",
			path: "/admin/users/00000000-0000-0000-0000-000000000001/fees",
			want: true,
		},
		{
			name: "user-specific fee PUT endpoint",
			path: "/admin/users/test-user-123/fees",
			want: true,
		},
		{
			name: "user-specific fee DELETE endpoint",
			path: "/admin/users/abc-def-ghi/fees",
			want: true,
		},
		{
			name: "non-public endpoint",
			path: "/core/v1/transactions",
			want: false,
		},
		{
			name: "global fees endpoint (not in patterns, but in PublicEndpoints)",
			path: "/admin/fees",
			want: false, // Should be handled by exact match in PublicEndpoints
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesPublicPattern(tt.path)
			assert.Equal(t, tt.want, got, "matchesPublicPattern(%q)", tt.path)
		})
	}
}

// --- Middleware integration tests ---

var testCreds = map[string]string{"test-app": "test-secret"}
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
})

func signedRequest(method, url, body string) *http.Request {
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sig := GenerateSignature(ts, method, url, body, "test-secret")
	req := httptest.NewRequest(method, url, nil)
	req.Header.Set("x-gatehub-app-id", "test-app")
	req.Header.Set("x-gatehub-timestamp", ts)
	req.Header.Set("x-gatehub-signature", sig)
	return req
}

func TestMiddleware_ValidSignature(t *testing.T) {
	mw := Middleware(testCreds)(okHandler)
	url := "http://example.com/core/v1/wallets"
	req := signedRequest("GET", url, "")
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_MissingHeaders(t *testing.T) {
	mw := Middleware(testCreds)(okHandler)
	req := httptest.NewRequest("GET", "/core/v1/wallets", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_UnknownAppID(t *testing.T) {
	mw := Middleware(testCreds)(okHandler)
	req := httptest.NewRequest("GET", "/core/v1/wallets", nil)
	req.Header.Set("x-gatehub-app-id", "unknown")
	req.Header.Set("x-gatehub-timestamp", "123456")
	req.Header.Set("x-gatehub-signature", "badsig")
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_BadSignature(t *testing.T) {
	mw := Middleware(testCreds)(okHandler)
	req := httptest.NewRequest("GET", "/core/v1/wallets", nil)
	req.Header.Set("x-gatehub-app-id", "test-app")
	req.Header.Set("x-gatehub-timestamp", "123456")
	req.Header.Set("x-gatehub-signature", "definitely-wrong")
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_PublicEndpointSkipsAuth(t *testing.T) {
	mw := Middleware(testCreds)(okHandler)
	for _, path := range []string{"/health", "/", "/iframe/onboarding", "/admin/fees"} {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "public endpoint %s should skip auth", path)
	}
}

func TestMiddleware_PublicPatternSkipsAuth(t *testing.T) {
	mw := Middleware(testCreds)(okHandler)
	req := httptest.NewRequest("GET", "/admin/users/some-uuid/fees", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_UIPathSkipsAuth(t *testing.T) {
	mw := Middleware(testCreds)(okHandler)
	for _, path := range []string{"/ui", "/ui/", "/ui/users/some-uuid", "/ui/actions/kyc", "/ui/actions/card-transaction"} {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "UI path %s should skip auth", path)
	}
}

func TestGetRegisteredAppIDs(t *testing.T) {
	ids := getRegisteredAppIDs(map[string]string{"a": "1", "b": "2"})
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, "a")
	assert.Contains(t, ids, "b")
}
