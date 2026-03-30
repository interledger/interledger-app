//go:build e2e
// +build e2e

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// ptiRequest sends a request to the mock PTI service optionally including the
// x-pti-client-id header so the auth middleware allows it through.
// extraHeaders are merged in after the standard headers.
func (tc *TestContext) ptiRequest(method, path string, body interface{}, withAuth bool, extraHeaders ...map[string]string) (*http.Response, error) {
	path = tc.resolvePath(path)

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

	if withAuth {
		req.Header.Set("x-pti-client-id", tc.clientID)
	}

	for _, h := range extraHeaders {
		for k, v := range h {
			req.Header.Set(k, v)
		}
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

// resolvePath replaces placeholders like {userId} in URL paths before dispatch.
func (tc *TestContext) resolvePath(path string) string {
	path = strings.ReplaceAll(path, "{userId}", tc.lastUserID)
	path = strings.ReplaceAll(path, "{walletId}", tc.lastWalletID)
	path = strings.ReplaceAll(path, "{paymentInformationId}", tc.lastPaymentInformationID)
	path = strings.ReplaceAll(path, "{requestId}", tc.lastTransactionRequestID)
	return path
}
