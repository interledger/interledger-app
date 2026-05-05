package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSignature_FullURL(t *testing.T) {
	timestamp := "1686040166173"
	method := "POST"
	url := "http://mockgatehub:8080/auth/v1/tokens?clientId=abc"
	body := `{"scope":["auth"]}`
	secret := "local-test-app-secret"

	message := fmt.Sprintf("%s|%s|%s|%s", timestamp, method, url, body)
	message = strings.Trim(message, "|")
	want := hmacSHA256Hex(message, secret)

	got := GenerateSignature(timestamp, method, url, body, secret)

	assert.Equal(t, want, got)
	assert.Len(t, got, 64)
}

func TestGenerateSignature_EmptyBody_TrimsPipe(t *testing.T) {
	timestamp := "1686040166173"
	method := "GET"
	url := "http://mockgatehub:8080/id/v1/users/123"
	body := ""
	secret := "local-test-app-secret"

	message := fmt.Sprintf("%s|%s|%s", timestamp, method, url)
	want := hmacSHA256Hex(message, secret)

	got := GenerateSignature(timestamp, method, url, body, secret)

	assert.Equal(t, want, got)
	assert.Len(t, got, 64)
}

func TestGenerateSignature_Deterministic(t *testing.T) {
	timestamp := "1686040166173"
	method := "POST"
	url := "http://mockgatehub:8080/auth/v1/tokens"
	body := `{"scope":["auth"]}`
	secret := "local-test-app-secret"

	sig1 := GenerateSignature(timestamp, method, url, body, secret)
	sig2 := GenerateSignature(timestamp, method, url, body, secret)

	assert.Equal(t, sig1, sig2)
}

func hmacSHA256Hex(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestSignRequest_SetsHeaders(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com/auth/v1/tokens", nil)
	body := []byte(`{"key":"value"}`)

	SignRequest(req, "my-secret", body)

	assert.NotEmpty(t, req.Header.Get("x-gatehub-timestamp"))
	assert.NotEmpty(t, req.Header.Get("x-gatehub-signature"))
	assert.Equal(t, "gitlab.com/fynbos/mock/mockgatehub", req.Header.Get("x-gatehub-app-id"))

	// Signature should be valid
	ts := req.Header.Get("x-gatehub-timestamp")
	sig := req.Header.Get("x-gatehub-signature")
	expected := GenerateSignature(ts, "POST", "http://example.com/auth/v1/tokens", string(body), "my-secret")
	assert.Equal(t, expected, sig)
}

func TestBuildOutgoingURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"full url", "http://example.com/path", "http://example.com/path"},
		{"with query", "http://example.com/path?q=1", "http://example.com/path?q=1"},
		{"https", "https://example.com/path", "https://example.com/path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			got := buildOutgoingURL(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateGateHubWebhookSignature_HexSecret(t *testing.T) {
	body := `{"event":"test"}`
	secret := hex.EncodeToString([]byte("webhook-key"))

	sig := GenerateGateHubWebhookSignature(body, secret)
	assert.Len(t, sig, 64, "should be hex-encoded SHA256")

	// Verify manually
	key, _ := hex.DecodeString(secret)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(body))
	expected := hex.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expected, sig)
}

func TestGenerateGateHubWebhookSignature_NonHexFallback(t *testing.T) {
	body := `{"event":"test"}`
	secret := "not-hex-at-all"

	sig := GenerateGateHubWebhookSignature(body, secret)
	assert.Len(t, sig, 64)

	// When hex decode fails, uses raw secret
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	expected := hex.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expected, sig)
}
