//go:build e2e
// +build e2e

package main

import (
	"fmt"
	"net/http"
	"time"
)

// TestContext holds state for godog scenarios.
type TestContext struct {
	client  *http.Client
	baseURL string

	lastResponse     *http.Response
	lastResponseBody []byte
	lastError        error

	// flow state threaded across steps
	linkToken       string
	publicToken     string
	accessToken     string
	lastAccountID   string
	prevAccountID   string
	institutionName string
}

func (tc *TestContext) Reset() {
	tc.client = &http.Client{Timeout: 10 * time.Second}
	tc.baseURL = mockPlaidURL

	tc.lastResponse = nil
	tc.lastResponseBody = nil
	tc.lastError = nil

	tc.linkToken = ""
	tc.publicToken = ""
	tc.accessToken = ""
	tc.lastAccountID = ""
	tc.prevAccountID = ""
	tc.institutionName = ""
}

// resetBackend clears mock state between scenarios.
func (tc *TestContext) resetBackend() error {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", tc.baseURL+"/test/reset", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("backend reset failed with status: %d", resp.StatusCode)
	}
	return nil
}
