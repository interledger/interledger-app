//go:build e2e
// +build e2e

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// do sends a request and captures the response status + body.
func (tc *TestContext) do(method, path string, body interface{}) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, tc.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// Plaid SDK always sends these; include for fidelity.
	req.Header.Set("PLAID-CLIENT-ID", "test-client-id")
	req.Header.Set("PLAID-SECRET", "test-secret")

	resp, err := tc.client.Do(req)
	if err != nil {
		tc.lastError = err
		return err
	}
	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	tc.lastResponse = resp
	tc.lastResponseBody = respBody
	return nil
}

// field returns a top-level string field from the last JSON response body.
func (tc *TestContext) field(name string) (string, bool) {
	var m map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &m); err != nil {
		return "", false
	}
	v, ok := m[name]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func (tc *TestContext) bodyString() string {
	return fmt.Sprintf("%s", string(tc.lastResponseBody))
}
