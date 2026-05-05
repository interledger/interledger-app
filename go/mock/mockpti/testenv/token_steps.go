//go:build e2e
// +build e2e

package main

// validPTIHeadersArePresent is a no-op pre-condition: the clientID stored in
// TestContext is always attached for authenticated requests.
func (tc *TestContext) validPTIHeadersArePresent() error {
	return nil
}

// postWithURLAndMethodPayload sends POST /auth/jwt with a minimal payload.
func (tc *TestContext) postWithURLAndMethodPayload(path string) error {
	payload := map[string]string{
		"url":    "https://example.com/some-form",
		"method": "GET",
	}
	_, err := tc.ptiRequest("POST", path, payload, true)
	return err
}
