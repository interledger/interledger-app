//go:build e2e
// +build e2e

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func (tc *TestContext) request(method, path string, body interface{}, auth bool, headers map[string]string) (*http.Response, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	url := tc.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if auth && tc.token != "" {
		req.Header.Set("Authorization", "Bearer "+tc.token)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return tc.doRequest(req)
}

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
	tc.lastResponseBody = respBody
	return resp, nil
}
