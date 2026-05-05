//go:build e2e

package main

import (
	"crypto/rsa"
	"net/http"
	"time"
)

// TestContext holds the state for BDD test scenarios
type TestContext struct {
	// HTTP client
	client *http.Client

	// Base configuration
	baseURL string

	// Authentication
	appID     string
	appSecret string

	// Signature authentication fields
	signatureTime     string
	signatureBody     string
	signatureBase     string
	signatureURL      string
	signatureOverride string
	signatureHeaders  map[string]string

	// Pending request (for multi-step signature tests)
	pendingMethod string
	pendingPath   string
	pendingBody   string
	pendingMode   string

	// Response storage
	lastResponse     *http.Response
	lastResponseBody []byte
	lastError        error

	// Test data
	userID          string
	walletAddress   string
	previousAddress string
	accessToken     string
	iframeToken     string
	kycToken        string
	email           string
	customerID      string
	cardID          string
	transactionID   string

	// Card-data token scenario state. Held per-scenario (cleared in Reset)
	// so concurrent scenarios never bleed RSA keys/tokens into each other.
	cardDataPrivateKey *rsa.PrivateKey
	cardDataToken      string
	cardDataLinkHref   string
}

// Reset initializes the test context to a clean state
func (tc *TestContext) Reset() {
	tc.client = &http.Client{Timeout: 10 * time.Second}
	tc.baseURL = "http://localhost:25151"
	tc.appID = ""
	tc.appSecret = ""
	tc.signatureTime = ""
	tc.signatureBody = ""
	tc.signatureBase = ""
	tc.signatureURL = ""
	tc.signatureOverride = ""
	tc.signatureHeaders = nil
	tc.pendingMethod = ""
	tc.pendingPath = ""
	tc.pendingBody = ""
	tc.pendingMode = ""
	tc.userID = ""
	tc.walletAddress = ""
	tc.accessToken = ""
	tc.iframeToken = ""
	tc.kycToken = ""
	tc.customerID = ""
	tc.cardID = ""
	tc.transactionID = ""
	tc.email = ""
	tc.previousAddress = ""
	tc.lastResponse = nil
	tc.lastResponseBody = nil
	tc.lastError = nil
	tc.cardDataPrivateKey = nil
	tc.cardDataToken = ""
	tc.cardDataLinkHref = ""
}
