package main

import (
	"fmt"
	"time"

	"github.com/cucumber/godog"
)

func (tc *TestContext) obtainValidAccessToken() error {
	if err := tc.loginWithDefaults(); err != nil {
		return err
	}
	return tc.responseContainsValidAccessToken()
}

func (tc *TestContext) obtainExpiringAccessToken() error {
	if err := tc.loginWithDefaults(); err != nil {
		return err
	}
	if err := tc.responseContainsValidAccessToken(); err != nil {
		return err
	}
	tc.expiredToken = tc.token
	return nil
}

func (tc *TestContext) clearAccessToken() error {
	tc.token = ""
	return nil
}

func (tc *TestContext) useInvalidAccessToken(token string) error {
	tc.token = token
	return nil
}

func (tc *TestContext) requestLoginTokenWithValidCredentials(table *godog.Table) error {
	return tc.loginWithTable(table)
}

func (tc *TestContext) requestLoginTokenWithInvalidCredentials(table *godog.Table) error {
	return tc.loginWithTable(table)
}

func (tc *TestContext) requestLoginTokenWithMissingFields(table *godog.Table) error {
	return tc.loginWithTable(table)
}

func (tc *TestContext) requestNewLoginTokenWithValidCredentials() error {
	if err := tc.loginWithDefaults(); err != nil {
		return err
	}
	return tc.responseContainsValidAccessToken()
}

func (tc *TestContext) attemptToUseExpiredToken() error {
	tc.token = "expired_token_" + fmt.Sprint(time.Now().UnixNano())
	_, _ = tc.request("GET", "/v1/example-route", nil, true, nil)
	return nil
}

func (tc *TestContext) responseContainsValidAccessToken() error {
	var resp struct {
		TokenValue string `json:"tokenValue"`
	}
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	if resp.TokenValue == "" {
		return fmt.Errorf("tokenValue missing in response")
	}
	tc.lastLoginToken = resp.TokenValue
	tc.token = resp.TokenValue
	return nil
}

func (tc *TestContext) tokenExpiresIn55Minutes() error {
	if tc.lastLoginToken == "" {
		return fmt.Errorf("no token available to validate")
	}
	return nil
}

func (tc *TestContext) newTokenIsDifferentFromExpired() error {
	if tc.lastLoginToken == "" || tc.expiredToken == "" {
		return fmt.Errorf("missing token values for comparison")
	}
	if tc.lastLoginToken == tc.expiredToken {
		return fmt.Errorf("expected new token to differ from expired token")
	}
	return nil
}

func (tc *TestContext) newTokenIsValid() error {
	_, err := tc.request("GET", "/v1/example-route", nil, true, nil)
	if err != nil {
		return err
	}
	return tc.responseStatusIs(200)
}

func (tc *TestContext) callProtectedRouteWithToken() error {
	_, err := tc.request("GET", "/v1/example-route", nil, true, nil)
	return err
}

func (tc *TestContext) loginWithDefaults() error {
	payload := map[string]interface{}{
		"policyId": tc.policy,
		"fields": []map[string]string{
			{"fieldName": "publicKey", "fieldValue": tc.pubKey},
			{"fieldName": "secret", "fieldValue": tc.secret},
		},
	}
	_, err := tc.request("POST", "/v1/login", payload, false, nil)
	return err
}
func (tc *TestContext) loginWithTable(table *godog.Table) error {
	values := tableToMap(table)
	policyID := values["policyId"]
	publicKey := values["publicKey"]
	if publicKey == "" {
		publicKey = values["apiPublicKey"]
	}
	secretKey := values["secretKey"]
	if secretKey == "" {
		secretKey = values["apiSecretKey"]
	}

	payload := map[string]interface{}{}
	if policyID != "" {
		payload["policyId"] = policyID
	}

	fields := []map[string]string{}
	if publicKey != "" {
		fields = append(fields, map[string]string{"fieldName": "publicKey", "fieldValue": publicKey})
	}
	if secretKey != "" {
		fields = append(fields, map[string]string{"fieldName": "secret", "fieldValue": secretKey})
	}
	if len(fields) > 0 {
		payload["fields"] = fields
	}

	_, err := tc.request("POST", "/v1/login", payload, false, nil)
	return err
}
