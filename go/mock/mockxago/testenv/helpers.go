//go:build e2e

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

func tableToMap(table *godog.Table) map[string]string {
	values := make(map[string]string)
	if table == nil {
		return values
	}
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		key := strings.TrimSpace(row.Cells[0].Value)
		value := strings.TrimSpace(row.Cells[1].Value)
		if key != "" {
			values[key] = value
		}
	}
	return values
}

func (tc *TestContext) decodeLastResponse(out interface{}) error {
	if tc.lastResponseBody == nil {
		return fmt.Errorf("no response body")
	}
	return json.Unmarshal(tc.lastResponseBody, out)
}

func (tc *TestContext) parseErrorResponse() (errorResponse, error) {
	var errResp errorResponse
	if err := tc.decodeLastResponse(&errResp); err != nil {
		return errResp, err
	}
	return errResp, nil
}

func (tc *TestContext) responseStatusIs(status int) error {
	if tc.lastResponse == nil {
		return fmt.Errorf("no response")
	}
	if tc.lastResponse.StatusCode != status {
		return fmt.Errorf("expected status %d, got %d. Body: %s", status, tc.lastResponse.StatusCode, string(tc.lastResponseBody))
	}
	return nil
}

func (tc *TestContext) errorMessageIs(expected string) error {
	errResp, err := tc.parseErrorResponse()
	if err != nil {
		return err
	}
	if errResp.Message == expected || errResp.Error == expected {
		return nil
	}
	return fmt.Errorf("expected error/message %q, got error=%q message=%q", expected, errResp.Error, errResp.Message)
}

func (tc *TestContext) errorMessageContains(expected string) error {
	errResp, err := tc.parseErrorResponse()
	if err != nil {
		return err
	}
	if strings.Contains(errResp.Message, expected) || strings.Contains(errResp.Error, expected) {
		return nil
	}
	return fmt.Errorf("expected error/message to contain %q, got error=%q message=%q", expected, errResp.Error, errResp.Message)
}

func (tc *TestContext) xagoMockServiceRunning() error {
	tc.Reset()
	return nil
}

func (tc *TestContext) environmentVariablesAreSet(table *godog.Table) error {
	values := tableToMap(table)
	for key, value := range values {
		_ = os.Setenv(key, value)
		switch key {
		case "XAGO_API_PUBLIC_KEY":
			tc.pubKey = value
		case "XAGO_API_SECRET":
			tc.secret = value
		}
	}
	return nil
}

func parseUUID(value string) error {
	_, err := uuid.Parse(value)
	return err
}
