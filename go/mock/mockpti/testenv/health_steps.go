//go:build e2e
// +build e2e

package main

import (
	"fmt"
	"net/http"
)

// mockptiIsRunning is a no-op pre-condition: if the suite reached this step the
// service is already healthy (waitForServices checks the health endpoint).
func (tc *TestContext) mockptiIsRunning() error {
	resp, err := http.Get(tc.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned %d", resp.StatusCode)
	}
	return nil
}

// sendGETRequest sends an unauthenticated GET to a plain literal path.
func (tc *TestContext) sendGETRequest(path string) error {
	_, err := tc.ptiRequest("GET", path, nil, false)
	return err
}
