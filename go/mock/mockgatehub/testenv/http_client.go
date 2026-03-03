package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// generateHMAC creates an HMAC signature for requests
func (tc *TestContext) generateHMAC(timestamp, method, fullURL, body string) string {
	payload := fmt.Sprintf("%s|%s|%s|%s", timestamp, method, fullURL, body)
	payload = strings.Trim(payload, "|")
	h := hmac.New(sha256.New, []byte(tc.appSecret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// computeHMAC is a standalone HMAC computation utility
func computeHMAC(secret, payload string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// replacePlaceholders replaces template variables in paths
func (tc *TestContext) replacePlaceholders(path string) string {
	replacements := map[string]string{
		"{userId}":        tc.userID,
		"{customerId}":    tc.customerID,
		"{cardId}":        tc.cardID,
		"{cardID}":        tc.cardID,
		"{transactionId}": tc.transactionID,
		"{txId}":          tc.transactionID,
		"{address}":       tc.walletAddress,
		"{walletAddress}": tc.walletAddress,
		"{iframeToken}":   tc.iframeToken,
		"{token}":         tc.kycToken,
	}
	for placeholder, value := range replacements {
		path = strings.ReplaceAll(path, placeholder, value)
	}
	return path
}

// request makes an HTTP request with JSON body and optional HMAC authentication
func (tc *TestContext) request(method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	path = tc.replacePlaceholders(path)

	var bodyStr string
	var bodyBytes []byte

	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyStr = string(bodyBytes)
	}

	url := tc.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Add HMAC headers if appSecret is set
	if tc.appSecret != "" {
		timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
		req.Header.Set("x-gatehub-app-id", tc.appID)
		req.Header.Set("x-gatehub-timestamp", timestamp)
		req.Header.Set("x-gatehub-signature", tc.generateHMAC(timestamp, method, url, bodyStr))
	}

	// Add custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return tc.doRequest(req)
}

// requestRaw makes a raw HTTP request with full control over headers and body
func (tc *TestContext) requestRaw(method, path, bodyStr, contentType string, headers map[string]string) (*http.Response, error) {
	path = tc.replacePlaceholders(path)
	url := tc.baseURL + path

	var bodyBytes []byte
	if bodyStr != "" {
		bodyBytes = []byte(bodyStr)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return tc.doRequest(req)
}

// requestForm makes a form-encoded HTTP request with HMAC
func (tc *TestContext) requestForm(method, path string, formData map[string]string) (*http.Response, error) {
	path = tc.replacePlaceholders(path)

	// Convert form data to URL-encoded body
	vals := make([]string, 0, len(formData))
	for k, v := range formData {
		vals = append(vals, fmt.Sprintf("%s=%s", k, v))
	}
	bodyStr := strings.Join(vals, "&")
	bodyBytes := []byte(bodyStr)

	url := tc.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Add HMAC headers if appSecret is set
	if tc.appSecret != "" {
		timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
		req.Header.Set("x-gatehub-app-id", tc.appID)
		req.Header.Set("x-gatehub-timestamp", timestamp)
		req.Header.Set("x-gatehub-signature", tc.generateHMAC(timestamp, method, url, bodyStr))
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return tc.doRequest(req)
}

// doRequest executes an HTTP request and stores the response
func (tc *TestContext) doRequest(req *http.Request) (*http.Response, error) {
	resp, err := tc.client.Do(req)
	if err != nil {
		tc.lastError = err
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	tc.lastResponse = resp
	tc.lastResponse.Body = io.NopCloser(bytes.NewReader(respBody))
	tc.lastResponseBody = respBody

	return resp, nil
}

// sendWithManagedUserHeader is a helper to send requests with managed user UUID header
func (tc *TestContext) sendWithManagedUserHeader(method, path string, body interface{}) (*http.Response, error) {
	headers := map[string]string{
		"x-gatehub-managed-user-uuid": tc.userID,
	}
	return tc.request(method, path, body, headers)
}
