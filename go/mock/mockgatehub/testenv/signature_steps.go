package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ============ SIGNATURE AUTH STEPS ============

func (tc *TestContext) cleanMockGatehubInstanceWithAuthenticationEnforced() error {
	tc.Reset()
	return nil
}

func (tc *TestContext) validCredentialsWithAppIdAndSecret(appID, secret string) error {
	tc.appID = appID
	tc.appSecret = secret
	return nil
}

func (tc *TestContext) currentTimestampInMilliseconds() error {
	tc.signatureTime = fmt.Sprintf("%d", time.Now().UnixMilli())
	return nil
}

func (tc *TestContext) fixedTimestampValue(value string) error {
	tc.signatureTime = value
	return nil
}

func (tc *TestContext) bodyIs(body string) error {
	tc.signatureBody = body
	return nil
}

func (tc *TestContext) baseStringIs(template string) error {
	tc.signatureBase = template
	return nil
}

func (tc *TestContext) baseStringWithNoBodyComponent(template string) error {
	tc.signatureBase = template
	return nil
}

func (tc *TestContext) baseStringUsingPathOnly(template string) error {
	tc.signatureBase = template
	return nil
}

func (tc *TestContext) baseStringUsingOldSimpleFormat(template string) error {
	tc.signatureBase = template
	return nil
}

func (tc *TestContext) baseStringIncludesQueryString() error {
	return nil
}

func (tc *TestContext) baseStringUsesURL(url string) error {
	tc.signatureURL = url
	return nil
}

func (tc *TestContext) signatureComputedFrom(baseTemplate string) error {
	base := tc.buildBaseString(baseTemplate)
	tc.signatureOverride = computeHMAC(tc.appSecret, base)
	return nil
}

func (tc *TestContext) signatureComputedUsingSeconds(seconds string) error {
	base := fmt.Sprintf("%s|POST|http://localhost:25151/auth/v1/users/managed|%s", seconds, tc.signatureBody)
	base = strings.Trim(base, "|")
	tc.signatureOverride = computeHMAC(tc.appSecret, base)
	return nil
}

func (tc *TestContext) buildBaseString(template string) string {
	base := template
	if tc.signatureTime != "" {
		base = strings.ReplaceAll(base, "timestamp_ms", tc.signatureTime)
	}
	if tc.signatureBody != "" {
		base = strings.ReplaceAll(base, "{body}", tc.signatureBody)
	}
	return tc.replacePlaceholders(base)
}

func (tc *TestContext) requestIncludesHeaderXForwardedProto(value string) error {
	if tc.signatureHeaders == nil {
		tc.signatureHeaders = map[string]string{}
	}
	tc.signatureHeaders["X-Forwarded-Proto"] = value
	return tc.maybeSendPendingRequest()
}

func (tc *TestContext) requestIncludesHeaderXForwardedHost(value string) error {
	if tc.signatureHeaders == nil {
		tc.signatureHeaders = map[string]string{}
	}
	tc.signatureHeaders["X-Forwarded-Host"] = value
	return tc.maybeSendPendingRequest()
}

func (tc *TestContext) maybeSendPendingRequest() error {
	if tc.pendingMethod == "" {
		return nil
	}
	if tc.signatureHeaders == nil {
		return nil
	}
	if tc.signatureHeaders["X-Forwarded-Proto"] == "" || tc.signatureHeaders["X-Forwarded-Host"] == "" {
		return nil
	}
	err := tc.sendSignedRequest(tc.pendingMethod, tc.pendingPath, tc.pendingBody, tc.pendingMode, "", "", "", false, false, false)
	if err != nil {
		return err
	}
	tc.pendingMethod = ""
	tc.pendingPath = ""
	tc.pendingBody = ""
	tc.pendingMode = ""
	return nil
}

func (tc *TestContext) responseContainsUserID() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}
	if id, ok := result["id"].(string); ok && id != "" {
		return nil
	}
	if id, ok := result["user_id"].(string); ok && id != "" {
		return nil
	}
	return fmt.Errorf("no user id in response")
}

func (tc *TestContext) responseStatusIsNot(status int) error {
	if tc.lastResponse == nil {
		return fmt.Errorf("no response")
	}
	if tc.lastResponse.StatusCode == status {
		return fmt.Errorf("expected status not %d, got %d", status, tc.lastResponse.StatusCode)
	}
	return nil
}

func (tc *TestContext) sendSignedRequest(method, path, bodyStr, mode, appIDOverride, secretOverride, timestampOverride string, omitAppID, omitTimestamp, omitSignature bool) error {
	path = tc.replacePlaceholders(path)
	url := tc.baseURL + path

	if timestampOverride == "" {
		if tc.signatureTime == "" {
			tc.signatureTime = fmt.Sprintf("%d", time.Now().UnixMilli())
		}
		timestampOverride = tc.signatureTime
	}

	appID := tc.appID
	if appIDOverride != "" {
		appID = appIDOverride
	}
	secret := tc.appSecret
	if secretOverride != "" {
		secret = secretOverride
	}

	signURL := url
	if tc.signatureURL != "" {
		signURL = tc.signatureURL
	}
	if mode == "path" {
		signURL = path
	}

	signature := tc.signatureOverride
	if signature == "" {
		if mode == "simple" {
			base := timestampOverride + method + path + bodyStr
			signature = computeHMAC(secret, base)
		} else {
			base := fmt.Sprintf("%s|%s|%s|%s", timestampOverride, method, signURL, bodyStr)
			base = strings.Trim(base, "|")
			signature = computeHMAC(secret, base)
		}
	}

	headers := map[string]string{}
	for k, v := range tc.signatureHeaders {
		headers[k] = v
	}
	if !omitAppID {
		headers["x-gatehub-app-id"] = appID
	}
	if !omitTimestamp {
		headers["x-gatehub-timestamp"] = timestampOverride
	}
	if !omitSignature {
		headers["x-gatehub-signature"] = signature
	}

	contentType := ""
	if bodyStr != "" {
		contentType = "application/json"
	}

	_, err := tc.requestRaw(method, path, bodyStr, contentType, headers)

	tc.signatureOverride = ""
	tc.signatureURL = ""
	tc.signatureHeaders = nil

	return err
}

func (tc *TestContext) postSignedUsingFullURL(path, body string) error {
	return tc.sendSignedRequest("POST", path, body, "full", "", "", "", false, false, false)
}

func (tc *TestContext) getSignedUsingFullURL(path string) error {
	return tc.sendSignedRequest("GET", path, "", "full", "", "", "", false, false, false)
}

func (tc *TestContext) postSignedUsingFullURLWithQuery(path, body string) error {
	return tc.sendSignedRequest("POST", path, body, "full", "", "", "", false, false, false)
}

func (tc *TestContext) postSignedUsingProxiedURL(path, body string) error {
	tc.pendingMethod = "POST"
	tc.pendingPath = path
	tc.pendingBody = body
	tc.pendingMode = "full"
	return tc.maybeSendPendingRequest()
}

func (tc *TestContext) postSignedUsingPathOnly(path, body string) error {
	return tc.sendSignedRequest("POST", path, body, "path", "", "", "", false, false, false)
}

func (tc *TestContext) postSignedUsingSimpleFormat(path, body string) error {
	return tc.sendSignedRequest("POST", path, body, "simple", "", "", "", false, false, false)
}

func (tc *TestContext) postWithoutSignatureHeader(path, body string) error {
	return tc.sendSignedRequest("POST", path, body, "full", "", "", "", false, false, true)
}

func (tc *TestContext) postWithoutTimestampHeader(path, body string) error {
	return tc.sendSignedRequest("POST", path, body, "full", "", "", "", false, true, false)
}

func (tc *TestContext) postWithoutAppIDHeader(path, body string) error {
	return tc.sendSignedRequest("POST", path, body, "full", "", "", "", true, false, false)
}

func (tc *TestContext) postWithUnknownAppID(path, body, appID string) error {
	return tc.sendSignedRequest("POST", path, body, "full", appID, "", "", false, false, false)
}

func (tc *TestContext) postSignedWithSecret(path, body, secret string) error {
	return tc.sendSignedRequest("POST", path, body, "full", "", secret, "", false, false, false)
}

func (tc *TestContext) postSignedWithTimestamp(path, timestamp string) error {
	return tc.sendSignedRequest("POST", path, tc.signatureBody, "full", "", "", timestamp, false, false, false)
}

func (tc *TestContext) getHealthWithoutHMAC() error {
	_, err := tc.requestRaw("GET", "/health", "", "", nil)
	return err
}

func (tc *TestContext) getRootWithoutHMAC() error {
	if tc.userID == "" {
		if err := tc.existingManagedUserGeneric(); err != nil {
			return err
		}
	}
	body := map[string]interface{}{"scope": []string{"auth"}}
	headers := map[string]string{"x-gatehub-managed-user-uuid": tc.userID}
	_, err := tc.request("POST", "/auth/v1/tokens?clientId=test-client", body, headers)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}
	token, _ := result["token"].(string)
	if token == "" {
		return fmt.Errorf("missing token for root request")
	}

	path := fmt.Sprintf("/?paymentType=onboarding&bearer=%s", token)
	_, err = tc.requestRaw("GET", path, "", "", nil)
	return err
}

func (tc *TestContext) postIframeSubmitWithoutHMAC() error {
	if tc.userID == "" {
		if err := tc.existingManagedUserGeneric(); err != nil {
			return err
		}
	}

	body := fmt.Sprintf("user_id=%s&first_name=Test&last_name=User&dob=1990-01-01&address=123+Main+St&city=NY&country=USA&risk_level=low", tc.userID)
	_, err := tc.requestRaw("POST", "/iframe/submit", body, "application/x-www-form-urlencoded", nil)
	return err
}
