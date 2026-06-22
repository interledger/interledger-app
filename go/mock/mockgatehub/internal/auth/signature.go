package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// GenerateSignature generates signature in Gatehub backend format
// Format: HMAC-SHA256(timestamp_ms|method|url|body, secret)
func GenerateSignature(timestamp, method, url, body, secret string) string {
	message := fmt.Sprintf("%s|%s|%s|%s", timestamp, method, url, body)
	message = strings.Trim(message, "|")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// SignRequest adds signature headers to an outgoing request
func SignRequest(r *http.Request, secret string, body []byte) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	method := r.Method
	fullURL := buildOutgoingURL(r)

	signature := GenerateSignature(timestamp, method, fullURL, string(body), secret)

	r.Header.Set("x-gatehub-timestamp", timestamp)
	r.Header.Set("x-gatehub-signature", signature)
	r.Header.Set("x-gatehub-app-id", "github.com/interledger/interledger-app/go/mock/mockgatehub")
}

func buildOutgoingURL(r *http.Request) string {
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "http"
	}
	host := r.URL.Host
	if host == "" {
		host = r.Host
	}
	fullURL := scheme + "://" + host + r.URL.Path
	if r.URL.RawQuery != "" {
		fullURL += "?" + r.URL.RawQuery
	}
	return fullURL
}

// GenerateGateHubWebhookSignature generates the signature for webhooks as expected by the backend
// The backend expects: HMAC-SHA256(json_body, hex_decoded_secret)
func GenerateGateHubWebhookSignature(jsonBody, hexSecret string) string {
	// Decode hex secret to bytes (the secret is stored as hex in env)
	key, err := hex.DecodeString(hexSecret)
	if err != nil {
		// If decoding fails, use the secret as-is (it might not be hex encoded)
		key = []byte(hexSecret)
	}

	// Create HMAC-SHA256 of the body
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(jsonBody))
	return hex.EncodeToString(mac.Sum(nil))
}
