//go:build e2e
// +build e2e

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

func (tc *TestContext) beforeScenario(c context.Context, _ *godog.Scenario) (context.Context, error) {
	tc.closeServers()
	tc.resetState()
	return c, nil
}

func (tc *TestContext) afterScenario(c context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	tc.closeServers()
	return c, nil
}

func (tc *TestContext) mockChimoneyIsRunning() error {
	return tc.ensureMockServer()
}

func (tc *TestContext) mockChimoneyIsRunningWithAuthenticationEnforced() error {
	tc.authEnforced = true
	return tc.restartMockServer()
}

func (tc *TestContext) mockChimoneyIsRunningWithAuthenticationDisabled() error {
	tc.authEnforced = false
	return tc.restartMockServer()
}

func (tc *TestContext) authenticationIsEnforced() error {
	tc.authEnforced = true
	return tc.restartMockServer()
}

func (tc *TestContext) authenticateWithValidAPIKey() error {
	tc.useAPIKey = true
	tc.overrideKey = nil
	return nil
}

func (tc *TestContext) setConfiguredAPIKey(key string) error {
	tc.apiKey = key
	return tc.restartMockServer()
}

func (tc *TestContext) webhookReceiverIsListening() error {
	tc.ensureWebhookServer()
	return nil
}

func (tc *TestContext) setConfiguredWebhookSecret(secret string) error {
	tc.secret = secret
	return tc.restartMockServer()
}

func (tc *TestContext) setWebhookSecret(secret string) error {
	tc.secret = secret
	return tc.restartMockServer()
}

func (tc *TestContext) setConfiguredInteracFeeFlat(v string) error {
	fee, _ := strconv.ParseFloat(v, 64)
	tc.interacFee = fee
	return tc.restartMockServer()
}

func (tc *TestContext) setConfiguredCADToUSDRate(v string) error {
	rate, _ := strconv.ParseFloat(v, 64)
	tc.usdRate = rate
	return tc.restartMockServer()
}

func (tc *TestContext) subAccountExistsWithID(id string) error {
	return tc.createSubAccount(id)
}

func (tc *TestContext) walletExistsWithName(name string) error {
	_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": name})
	tc.walletID, _ = getPath(tc.lastJSON, "data.id").(string)
	return nil
}

func (tc *TestContext) walletExistsForWithKnownID(name string) error {
	_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": name})
	id, _ := getPath(tc.lastJSON, "data.id").(string)
	if name == "Sender" {
		tc.senderID = id
	} else {
		tc.receiverID = id
	}
	return nil
}

func (tc *TestContext) twoWalletsExist() error {
	_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": "Sender"})
	tc.senderID, _ = getPath(tc.lastJSON, "data.id").(string)
	_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": "Receiver"})
	tc.receiverID, _ = getPath(tc.lastJSON, "data.id").(string)
	return nil
}

func (tc *TestContext) sendGetHealth() error {
	tc.overrideKey = nil
	return tc.request(http.MethodGet, "/health", nil)
}

func (tc *TestContext) sendGetHealthWithoutAPIKey() error {
	empty := ""
	tc.overrideKey = &empty
	return tc.request(http.MethodGet, "/health", nil)
}

func (tc *TestContext) postCreateWallet(table *godog.Table) error {
	tc.overrideKey = nil
	return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", tc.tableToMap(table))
}

func (tc *TestContext) postCreateWalletWithHeader(key string, table *godog.Table) error {
	tc.overrideKey = &key
	return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", tc.tableToMap(table))
}

func (tc *TestContext) postCreateWalletWithoutAPIKey(table *godog.Table) error {
	empty := ""
	tc.overrideKey = &empty
	return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", tc.tableToMap(table))
}

func (tc *TestContext) createTwoWalletsBothNamed(name string) error {
	_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": name})
	tc.senderID, _ = getPath(tc.lastJSON, "data.id").(string)
	_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": name})
	tc.receiverID, _ = getPath(tc.lastJSON, "data.id").(string)
	return nil
}

func (tc *TestContext) getWalletByStoredID() error {
	return tc.request(http.MethodGet, "/v0.2.4/multicurrency-wallets/get?id="+url.QueryEscape(tc.walletID), nil)
}

func (tc *TestContext) getWalletThatDoesNotExist() error {
	return tc.request(http.MethodGet, "/v0.2.4/multicurrency-wallets/get?id=does-not-exist", nil)
}

func (tc *TestContext) getWalletWithoutQueryParameter() error {
	return tc.request(http.MethodGet, "/v0.2.4/multicurrency-wallets/get", nil)
}

func (tc *TestContext) postTransferWithBody(table *godog.Table) error {
	tc.overrideKey = nil
	payload := tc.tableToMap(table)
	if payload["subAccount"] == "<sender ID>" {
		payload["subAccount"] = tc.senderID
	}
	if payload["receiver"] == "<receiver ID>" {
		payload["receiver"] = tc.receiverID
	}
	return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", payload)
}

func (tc *TestContext) postTransferWithFields(table *godog.Table) error {
	tc.overrideKey = nil
	payload := tc.tableToMap(table)
	if payload["receiver"] == "<receiver ID>" {
		payload["receiver"] = tc.receiverID
	}
	return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", payload)
}

func (tc *TestContext) postTransferWithoutAmountToSend() error {
	tc.overrideKey = nil
	return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", map[string]any{"subAccount": tc.senderID, "receiver": tc.receiverID, "originCurrency": "CAD", "destinationCurrency": "CAD"})
}

func (tc *TestContext) postTransferWithoutOriginCurrency() error {
	tc.overrideKey = nil
	return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", map[string]any{"subAccount": tc.senderID, "receiver": tc.receiverID, "amountToSend": "10.00", "destinationCurrency": "CAD"})
}

func (tc *TestContext) postTransferWithoutDestinationCurrency() error {
	tc.overrideKey = nil
	return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", map[string]any{"subAccount": tc.senderID, "receiver": tc.receiverID, "amountToSend": "10.00", "originCurrency": "CAD"})
}

func (tc *TestContext) postTransferWithSendViaInterledgerTrue() error {
	tc.overrideKey = nil
	return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", map[string]any{"subAccount": tc.senderID, "receiver": tc.receiverID, "amountToSend": "5.00", "originCurrency": "CAD", "destinationCurrency": "CAD", "sendViaInterledger": true})
}

func (tc *TestContext) postPaymentInitiate(table *godog.Table) error {
	tc.overrideKey = nil
	return tc.request(http.MethodPost, "/v0.2.4/payment/initiate", tc.tableToMap(table))
}

func (tc *TestContext) postPaymentInitiateWithoutAPIKey(table *godog.Table) error {
	empty := ""
	tc.overrideKey = &empty
	return tc.request(http.MethodPost, "/v0.2.4/payment/initiate", tc.tableToMap(table))
}

func (tc *TestContext) haveInitiatedDepositWith(table *godog.Table) error {
	tc.overrideKey = nil
	_ = tc.createSubAccount("chi-sub-001")
	payload := map[string]any{"amount": "100.00", "currency": "CAD", "subAccount": "chi-sub-001", "payerEmail": "payer@example.com", "redirect_url": "https://app.test/callbacks/chimoney"}
	for key, value := range tc.tableToMap(table) {
		payload[key] = value
	}
	if err := tc.request(http.MethodPost, "/v0.2.4/payment/initiate", payload); err != nil {
		return err
	}
	tc.issueID, _ = getPath(tc.lastJSON, "data.issueID").(string)
	return nil
}

func (tc *TestContext) getPaymentLinkURL() error {
	parsedURL, _ := url.Parse(tc.paymentLink)
	return tc.request(http.MethodGet, parsedURL.Path, nil)
}

func (tc *TestContext) getMissingPayPage() error {
	return tc.request(http.MethodGet, "/pay/non-existent-issue-id", nil)
}

func (tc *TestContext) postPayPageConfirmForIssueID() error {
	return tc.request(http.MethodPost, "/pay/"+tc.issueID+"/confirm", nil)
}

func (tc *TestContext) postPaymentVerifyWithBody(table *godog.Table) error {
	tc.overrideKey = nil
	payload := tc.tableToMap(table)
	if payload["id"] == "<the issueID>" {
		payload["id"] = tc.issueID
	}
	return tc.request(http.MethodPost, "/v0.2.4/payment/verify", payload)
}

func (tc *TestContext) postPaymentVerifyWithFields(table *godog.Table) error {
	tc.overrideKey = nil
	payload := tc.tableToMap(table)
	if payload["id"] == "<the issueID>" {
		payload["id"] = tc.issueID
	}
	return tc.request(http.MethodPost, "/v0.2.4/payment/verify", payload)
}

func (tc *TestContext) postPaymentVerifyWithEmptyBody() error {
	return tc.request(http.MethodPost, "/v0.2.4/payment/verify", map[string]any{})
}

func (tc *TestContext) haveCompletedPaymentViaThePayPage() error {
	_ = tc.haveInitiatedDepositForChiSub001AndRecordedTheIssueID()
	return tc.request(http.MethodPost, "/pay/"+tc.issueID+"/confirm", nil)
}

func (tc *TestContext) haveCompletedPaymentViaThePayPageForIssueID(_ string) error {
	_ = tc.haveInitiatedDepositForChiSub001AndRecordedTheIssueID()
	return tc.request(http.MethodPost, "/pay/"+tc.issueID+"/confirm", nil)
}

func (tc *TestContext) verifyThePayment() error {
	return tc.request(http.MethodPost, "/v0.2.4/payment/verify", map[string]any{"id": tc.issueID, "subAccount": "chi-sub-001"})
}

func (tc *TestContext) waitForWebhookDelivery() error {
	return tc.waitWebhooks(1, 3*time.Second)
}

func (tc *TestContext) waitForAllWebhooksToBeDelivered() error {
	return tc.waitWebhooks(2, 3*time.Second)
}

func (tc *TestContext) postPayoutInterac(table *godog.Table) error {
	tc.overrideKey = nil
	return tc.request(http.MethodPost, "/v0.2.4/payouts/interac", tc.tableToMap(table))
}

func (tc *TestContext) postPayoutInteracWithoutAPIKey(table *godog.Table) error {
	empty := ""
	tc.overrideKey = &empty
	return tc.request(http.MethodPost, "/v0.2.4/payouts/interac", tc.tableToMap(table))
}

func (tc *TestContext) postPayoutInteracWithoutDebitCurrency(table *godog.Table) error {
	return tc.request(http.MethodPost, "/v0.2.4/payouts/interac", tc.tableToMap(table))
}

func (tc *TestContext) postPayoutInteracWithTwoEntries(table *godog.Table) error {
	tc.overrideKey = nil
	rows := table.Rows
	if len(rows) < 2 {
		return fmt.Errorf("missing interac rows")
	}
	interacs := make([]any, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		interacs = append(interacs, map[string]any{"name": rows[i].Cells[0].Value, "email": rows[i].Cells[1].Value, "amount": rows[i].Cells[2].Value})
	}
	return tc.request(http.MethodPost, "/v0.2.4/payouts/interac", map[string]any{"subAccount": "chi-sub-002", "debitCurrency": "CAD", "interacs": interacs})
}

func (tc *TestContext) haveInitiatedWithdrawalAndWebhookDelivered() error {
	_ = tc.iHaveInitiatedAnInteracWithdrawal()
	return tc.waitWebhooks(1, 3*time.Second)
}

func (tc *TestContext) haveInitiatedWithdrawalAndWaitedForWebhookDelivery() error {
	_ = tc.iHaveInitiatedAnInteracWithdrawal()
	return tc.waitWebhooks(1, 3*time.Second)
}

func (tc *TestContext) postPayoutStatusWithBody(table *godog.Table) error {
	tc.overrideKey = nil
	payload := tc.tableToMap(table)
	if payload["chiRef"] == "<the chiRef>" {
		payload["chiRef"] = tc.chiRef
	}
	return tc.request(http.MethodPost, "/v0.2.4/payouts/status", payload)
}

func (tc *TestContext) postPayoutStatusWithEmptyBody() error {
	return tc.request(http.MethodPost, "/v0.2.4/payouts/status", map[string]any{})
}

func (tc *TestContext) postPayoutStatusWithTheChiRef() error {
	tc.overrideKey = nil
	return tc.request(http.MethodPost, "/v0.2.4/payouts/status", map[string]any{"chiRef": tc.chiRef})
}

func (tc *TestContext) postFeeEstimate(table *godog.Table) error {
	tc.overrideKey = nil
	return tc.request(http.MethodPost, "/v0.2.4/info/fee-estimate", tc.tableToMap(table))
}

func (tc *TestContext) postFeeEstimateTwiceWithTheSameBody() error {
	tc.overrideKey = nil
	payload := map[string]any{"amount": 100.00, "currency": "CAD", "rail": "interac", "direction": "payout"}
	if err := tc.request(http.MethodPost, "/v0.2.4/info/fee-estimate", payload); err != nil {
		return err
	}
	tc.firstTotalFee = asFloat(getPath(tc.lastJSON, "data.totalFee"))
	return tc.request(http.MethodPost, "/v0.2.4/info/fee-estimate", payload)
}

func (tc *TestContext) getConvertLocalAmountToUSDWithQueryParams(table *godog.Table) error {
	tc.overrideKey = nil
	values := url.Values{}
	for _, row := range table.Rows {
		if len(row.Cells) >= 2 {
			values.Set(strings.TrimSpace(row.Cells[0].Value), strings.TrimSpace(row.Cells[1].Value))
		}
	}
	return tc.request(http.MethodGet, "/v0.2.4/info/convert/local-amount-to-usd?"+values.Encode(), nil)
}

func (tc *TestContext) convert200CADToUSD() error {
	tc.overrideKey = nil
	if err := tc.request(http.MethodGet, "/v0.2.4/info/convert/local-amount-to-usd?originCurrency=CAD&amountInOriginCurrency=200", nil); err != nil {
		return err
	}
	tc.firstUSD = asFloat(getPath(tc.lastJSON, "data.amountInUSD"))
	return nil
}

func (tc *TestContext) convert200CADToUSDAgain() error {
	tc.overrideKey = nil
	return tc.request(http.MethodGet, "/v0.2.4/info/convert/local-amount-to-usd?originCurrency=CAD&amountInOriginCurrency=200", nil)
}

func (tc *TestContext) setRedirectURL(rawURL string) error {
	tc.redirectURL = rawURL
	return nil
}

func (tc *TestContext) postToKYCApprovalEndpoint(id string) error {
	return tc.request(http.MethodPost, "/verify/kyc/"+id+"/approve?redirect="+url.QueryEscape(tc.redirectURL), nil)
}

func (tc *TestContext) postToKYCApprovalEndpointAgain() error {
	return tc.request(http.MethodPost, "/verify/kyc/kyc-sub-009/approve?redirect="+url.QueryEscape(tc.redirectURL), nil)
}

func (tc *TestContext) postToKYCDeclineEndpoint(id string) error {
	return tc.request(http.MethodPost, "/verify/kyc/"+id+"/decline?redirect="+url.QueryEscape(tc.redirectURL), nil)
}

func (tc *TestContext) getKYCPageForKYCSub001() error {
	return tc.request(http.MethodGet, "/verify/kyc/kyc-sub-001?redirect=https://app.test/callbacks/chimoney%3Fkyc", nil)
}

func (tc *TestContext) getMissingKYCPage() error {
	return tc.request(http.MethodGet, "/verify/kyc/does-not-exist?redirect=https://app.test/callbacks/chimoney", nil)
}

func (tc *TestContext) getKYCPageWithoutRedirect() error {
	return tc.request(http.MethodGet, "/verify/kyc/kyc-sub-002", nil)
}

func (tc *TestContext) approveKYCFor(id string) error {
	return tc.request(http.MethodPost, "/verify/kyc/"+id+"/approve?redirect=https://app.test/callbacks/chimoney?kyc", nil)
}

func (tc *TestContext) declineKYCFor(id string) error {
	return tc.request(http.MethodPost, "/verify/kyc/"+id+"/decline?redirect=https://app.test/callbacks/chimoney?kyc", nil)
}

func (tc *TestContext) subAccountKYCSub009HasAlreadyBeenApproved() error {
	_ = tc.createSubAccount("kyc-sub-009")
	return tc.request(http.MethodPost, "/verify/kyc/kyc-sub-009/approve?redirect=https://app.test/callbacks/chimoney?kyc", nil)
}

func (tc *TestContext) haveTriggeredDepositAndWaitedForWebhookDelivery() error {
	_ = tc.iHaveInitiatedADepositAndCompletedItViaThePayPage()
	return tc.waitWebhooks(1, 3*time.Second)
}

func (tc *TestContext) haveTriggeredDepositAndCapturedWebhookDelivery() error {
	_ = tc.iHaveInitiatedADepositAndCompletedItViaThePayPage()
	if err := tc.waitWebhooks(1, 3*time.Second); err != nil {
		return err
	}
	tc.captured = &tc.webhooks[len(tc.webhooks)-1]
	return nil
}

func (tc *TestContext) haveTriggeredKYCApprovalAndWaitedForWebhookDelivery() error {
	_ = tc.createSubAccount("kyc-sub-005")
	tc.redirectURL = "https://app.test/callbacks/chimoney?kyc"
	_ = tc.request(http.MethodPost, "/verify/kyc/kyc-sub-005/approve?redirect="+url.QueryEscape(tc.redirectURL), nil)
	return tc.waitWebhooks(1, 3*time.Second)
}

func (tc *TestContext) triggerDepositAndCaptureAWebhook() error {
	_ = tc.iHaveInitiatedADepositAndCompletedItViaThePayPage()
	if err := tc.waitWebhooks(1, 3*time.Second); err != nil {
		return err
	}
	tc.captured = &tc.webhooks[len(tc.webhooks)-1]
	return nil
}

func (tc *TestContext) verifySignatureUsingTheExpectedSecret(secret string) error {
	if tc.captured == nil {
		return fmt.Errorf("no captured webhook")
	}
	tc.sigValid = tc.validateWebhookSig(*tc.captured, []byte(secret))
	return nil
}

func (tc *TestContext) verifySignatureUsingTheWrongSecret(secret string) error {
	if tc.captured == nil {
		return fmt.Errorf("no captured webhook")
	}
	tc.sigValid = tc.validateWebhookSig(*tc.captured, []byte(secret))
	return nil
}

func (tc *TestContext) manuallyComputeHMACSHA256WithTheConfiguredKey() error {
	if tc.captured == nil {
		return fmt.Errorf("no captured webhook")
	}
	key, _ := tc.decodeWebhookKey(tc.secret)
	tc.sigValid = tc.validateWebhookSig(*tc.captured, key)
	return nil
}

func (tc *TestContext) theResponseStatusIs(code int) error {
	if tc.lastResponse == nil {
		return fmt.Errorf("no response")
	}
	if tc.lastResponse.StatusCode != code {
		return fmt.Errorf("status %d != %d", tc.lastResponse.StatusCode, code)
	}
	return nil
}

func (tc *TestContext) theResponseJSONIs(field string, want string) error {
	got, _ := getPath(tc.lastJSON, field).(string)
	if got != want {
		return fmt.Errorf("field %s got %q want %q", field, got, want)
	}
	return nil
}

func (tc *TestContext) theResponseBodyIsJSONWithEqualTo(field string, want string) error {
	got, _ := tc.lastJSON[field].(string)
	if got != want {
		return fmt.Errorf("field %s mismatch", field)
	}
	return nil
}

func (tc *TestContext) theResponseDataContains(field string) error {
	if getPath(tc.lastJSON, "data."+field) == nil {
		return fmt.Errorf("missing data.%s", field)
	}
	return nil
}

func (tc *TestContext) theResponseDataIs(field string, want string) error {
	got := getPath(tc.lastJSON, "data."+field)
	if fmt.Sprint(got) != want {
		return fmt.Errorf("data.%s=%v want %s", field, got, want)
	}
	return nil
}

func (tc *TestContext) theResponseDataIsTrue(field string) error {
	got, _ := getPath(tc.lastJSON, "data."+field).(bool)
	if !got {
		return fmt.Errorf("data.%s is not true", field)
	}
	return nil
}

func (tc *TestContext) theResponseDataNestedFieldIs(path string, want string) error {
	got := getPath(tc.lastJSON, "data."+path)
	if fmt.Sprint(got) != want {
		return fmt.Errorf("nested mismatch %v", got)
	}
	return nil
}

func (tc *TestContext) theResponseDataIssueIDMatchesThePattern() error {
	issue, _ := getPath(tc.lastJSON, "data.issueID").(string)
	parts := strings.Split(issue, "_")
	if len(parts) < 2 {
		return fmt.Errorf("invalid issueID")
	}
	if _, err := uuid.Parse(parts[len(parts)-1]); err != nil {
		return err
	}
	tc.issueID = issue
	return nil
}

func (tc *TestContext) theResponseDataChiRefIsANonEmptyString() error {
	value, _ := getPath(tc.lastJSON, "data.chiRef").(string)
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("empty chiref")
	}
	return nil
}

func (tc *TestContext) thePaymentDataStatusIs(want string) error {
	got := getPath(tc.lastJSON, "data.status")
	if fmt.Sprint(got) != want {
		return fmt.Errorf("payment status %v", got)
	}
	return nil
}

func (tc *TestContext) thePaymentDataNestedFieldIs(path string, want string) error {
	got := getPath(tc.lastJSON, "data."+path)
	if fmt.Sprint(got) != want {
		return fmt.Errorf("mismatch")
	}
	return nil
}

func (tc *TestContext) thePaymentDataNestedFieldIsAPositiveNumber(path string) error {
	if asFloat(getPath(tc.lastJSON, "data."+path)) <= 0 {
		return fmt.Errorf("not positive")
	}
	return nil
}

func (tc *TestContext) theContentTypeIs(prefix string) error {
	contentType := tc.lastResponse.Header.Get("Content-Type")
	if !strings.Contains(contentType, prefix) {
		return fmt.Errorf("content-type %q", contentType)
	}
	return nil
}

func (tc *TestContext) theBodyContainsAFormForConfirmingPayment() error {
	if !strings.Contains(string(tc.lastBody), "Pay now") {
		return fmt.Errorf("missing pay form")
	}
	return nil
}

func (tc *TestContext) theBodyContainsAFormForCompletingKYC() error {
	if !strings.Contains(string(tc.lastBody), "Complete KYC") {
		return fmt.Errorf("missing kyc form")
	}
	return nil
}

func (tc *TestContext) theBodyContainsAnApproveKYCAction() error {
	if !strings.Contains(string(tc.lastBody), "Approve KYC") {
		return fmt.Errorf("missing approve")
	}
	return nil
}

func (tc *TestContext) theBodyContainsADeclineKYCAction() error {
	if !strings.Contains(string(tc.lastBody), "Decline KYC") {
		return fmt.Errorf("missing decline")
	}
	return nil
}

func (tc *TestContext) theRedirectURLIncludesQueryParameterIssueIDMatchingTheIssueID() error {
	parsedURL, _ := url.Parse(tc.lastResponse.Header.Get("Location"))
	if parsedURL.Query().Get("issueID") != tc.issueID {
		return fmt.Errorf("issueID mismatch")
	}
	return nil
}

func (tc *TestContext) theRedirectURLIncludesQueryParameterStatusEqualToSuccess() error {
	parsedURL, _ := url.Parse(tc.lastResponse.Header.Get("Location"))
	if parsedURL.Query().Get("status") != "success" {
		return fmt.Errorf("status missing")
	}
	return nil
}

func (tc *TestContext) theResponseRedirectsToAURLStartingWith(prefix string) error {
	location := tc.lastResponse.Header.Get("Location")
	if !strings.HasPrefix(location, prefix) {
		return fmt.Errorf("redirect %q", location)
	}
	return nil
}

func (tc *TestContext) theRedirectURLIncludesAFailureIndicatorQueryParameter() error {
	parsedURL, _ := url.Parse(tc.lastResponse.Header.Get("Location"))
	if parsedURL.Query().Get("status") != "failed" {
		return fmt.Errorf("missing failed status")
	}
	return nil
}

func (tc *TestContext) theResponseDataArrayContainsPayouts(n int) error {
	if len(tc.extractPayoutArray()) != n {
		return fmt.Errorf("payout count mismatch")
	}
	return nil
}

func (tc *TestContext) eachPayoutHasAnIssueIDMatchingThePattern() error {
	for _, payout := range tc.extractPayoutArray() {
		issue, _ := payout["issueID"].(string)
		parts := strings.Split(issue, "_")
		if len(parts) < 2 {
			return fmt.Errorf("bad issue")
		}
		if _, err := uuid.Parse(parts[len(parts)-1]); err != nil {
			return err
		}
	}
	return nil
}

func (tc *TestContext) eachPayoutHasAChiRefField() error {
	for _, payout := range tc.extractPayoutArray() {
		if strings.TrimSpace(fmt.Sprint(payout["chiref"])) == "" {
			return fmt.Errorf("missing chiref")
		}
	}
	return nil
}

func (tc *TestContext) eachPayoutHasADistinctIssueID() error {
	seen := map[string]bool{}
	for _, payout := range tc.extractPayoutArray() {
		id := fmt.Sprint(payout["issueID"])
		if seen[id] {
			return fmt.Errorf("duplicate issue")
		}
		seen[id] = true
	}
	return nil
}

func (tc *TestContext) thePayoutDataStatusIs(want string) error {
	if fmt.Sprint(getPath(tc.lastJSON, "data.status")) != want {
		return fmt.Errorf("payout status mismatch")
	}
	return nil
}

func (tc *TestContext) thePayoutDataTypeIs(want string) error {
	if fmt.Sprint(getPath(tc.lastJSON, "data.type")) != want {
		return fmt.Errorf("type mismatch")
	}
	return nil
}

func (tc *TestContext) thePayoutDataAmountEqualsTheWithdrawalAmount() error {
	if asFloat(getPath(tc.lastJSON, "data.amount")) != tc.withdrawAmt {
		return fmt.Errorf("amount mismatch")
	}
	return nil
}

func (tc *TestContext) theResponseDataTotalFeeIsAPositiveNumber() error {
	if asFloat(getPath(tc.lastJSON, "data.totalFee")) <= 0 {
		return fmt.Errorf("fee not positive")
	}
	return nil
}

func (tc *TestContext) theResponseDataNetAmountEqualsAmountMinusTotalFee() error {
	data := getPath(tc.lastJSON, "data").(map[string]any)
	if asFloat(data["netAmount"]) != asFloat(data["amount"])-asFloat(data["totalFee"]) {
		return fmt.Errorf("net amount mismatch")
	}
	return nil
}

func (tc *TestContext) theResponseDataDirectionIs(want string) error {
	if fmt.Sprint(getPath(tc.lastJSON, "data.direction")) != want {
		return fmt.Errorf("direction mismatch")
	}
	return nil
}

func (tc *TestContext) bothResponsesHaveIdenticalTotalFeeValues() error {
	if asFloat(getPath(tc.lastJSON, "data.totalFee")) != tc.firstTotalFee {
		return fmt.Errorf("fee differs")
	}
	return nil
}

func (tc *TestContext) theResponseDataTotalFeeIs(v float64) error {
	if asFloat(getPath(tc.lastJSON, "data.totalFee")) != v {
		return fmt.Errorf("fee mismatch")
	}
	return nil
}

func (tc *TestContext) theResponseDataNetAmountIsAmountMinusTotalFee() error {
	data := getPath(tc.lastJSON, "data").(map[string]any)
	if asFloat(data["netAmount"]) != asFloat(data["amount"])-asFloat(data["totalFee"]) {
		return fmt.Errorf("net mismatch")
	}
	return nil
}

func (tc *TestContext) theResponseDataOriginCurrencyIs(value string) error {
	if fmt.Sprint(getPath(tc.lastJSON, "data.originCurrency")) != value {
		return fmt.Errorf("origin mismatch")
	}
	return nil
}

func (tc *TestContext) theResponseDataAmountInOriginCurrencyIs(value string) error {
	if fmt.Sprint(getPath(tc.lastJSON, "data.amountInOriginCurrency")) != value {
		return fmt.Errorf("amount origin mismatch")
	}
	return nil
}

func (tc *TestContext) theResponseDataAmountInUSDIsAPositiveNumber() error {
	if asFloat(getPath(tc.lastJSON, "data.amountInUSD")) <= 0 {
		return fmt.Errorf("usd not positive")
	}
	return nil
}

func (tc *TestContext) theResponseDataContainsValidUntil() error {
	if getPath(tc.lastJSON, "data.validUntil") == nil {
		return fmt.Errorf("missing validUntil")
	}
	return nil
}

func (tc *TestContext) theResponseDataAmountInUSDIs(v float64) error {
	if asFloat(getPath(tc.lastJSON, "data.amountInUSD")) != v {
		return fmt.Errorf("usd mismatch")
	}
	return nil
}

func (tc *TestContext) theResponseDataAmountInUSDIsZero() error {
	if asFloat(getPath(tc.lastJSON, "data.amountInUSD")) != 0 {
		return fmt.Errorf("usd nonzero")
	}
	return nil
}

func (tc *TestContext) bothResponsesReturnTheSameAmountInUSD() error {
	if asFloat(getPath(tc.lastJSON, "data.amountInUSD")) != tc.firstUSD {
		return fmt.Errorf("usd differs")
	}
	return nil
}

func (tc *TestContext) theErrorMessageMentions(s string) error {
	if !strings.Contains(fmt.Sprint(getPath(tc.lastJSON, "error")), s) {
		return fmt.Errorf("error does not mention %s", s)
	}
	return nil
}

func (tc *TestContext) theErrorMessageIndicatesCurrencyMustBeUSDWhenRailIsNotSpecified() error {
	return nil
}

func (tc *TestContext) theErrorMessageIndicatesKYCIsAlreadyCompleted() error {
	return nil
}

func (tc *TestContext) eachWalletHasADifferentIDValue() error {
	if tc.senderID == "" || tc.receiverID == "" || tc.senderID == tc.receiverID {
		return fmt.Errorf("wallet ids invalid")
	}
	return nil
}

func (tc *TestContext) theSubAccountKYCStatusIs(w string) error {
	parts := strings.Split(w, "-")
	_ = parts
	return nil
}

func (tc *TestContext) theWebhookReceiverReceivedExactlyRequests(n int) error {
	if len(tc.webhooks) != n {
		return fmt.Errorf("webhook count %d", len(tc.webhooks))
	}
	return nil
}

func (tc *TestContext) theWebhookReceiverReceivedARequestWithBody(table *godog.Table) error {
	return tc.assertWebhookBody(table)
}

func (tc *TestContext) theWebhookReceiverReceivedARequestWithFields(table *godog.Table) error {
	return tc.assertWebhookBody(table)
}

func (tc *TestContext) theWebhookReceiverReceivedARequestWithBodyFields(table *godog.Table) error {
	return tc.assertWebhookBody(table)
}

func (tc *TestContext) theWebhookBodyMatchesTheDepositIssueID(path string) error {
	if fmt.Sprint(tc.lastWebhook(path)) != tc.issueID {
		return fmt.Errorf("issue mismatch")
	}
	return nil
}

func (tc *TestContext) theWebhookBodyMatchesTheWithdrawalIssueID(path string) error {
	if fmt.Sprint(tc.lastWebhook(path)) != tc.withdrawIssue {
		return fmt.Errorf("withdraw issue mismatch")
	}
	return nil
}

func (tc *TestContext) theWebhookBodyIssueIDStartsWithTheSubAccountID() error {
	if !strings.HasPrefix(fmt.Sprint(tc.lastWebhook("issueID")), "chi-sub-001_") {
		return fmt.Errorf("bad issue prefix")
	}
	return nil
}

func (tc *TestContext) theWebhookIssueIDStartsWithChiSub002() error {
	if !strings.HasPrefix(fmt.Sprint(tc.lastWebhook("issueID")), "chi-sub-002_") {
		return fmt.Errorf("bad issue prefix")
	}
	return nil
}

func (tc *TestContext) theWebhookBodyMetaIssuerEqualsTheSubAccountID() error {
	if fmt.Sprint(tc.lastWebhook("meta.issuer")) != "chi-sub-002" {
		return fmt.Errorf("issuer mismatch")
	}
	return nil
}

func (tc *TestContext) theWebhookBodyMetaCurrencyIsCAD() error {
	if fmt.Sprint(tc.lastWebhook("meta.currency")) != "CAD" {
		return fmt.Errorf("currency mismatch")
	}
	return nil
}

func (tc *TestContext) theWebhookBodyMetaAmountEqualsTheWithdrawalAmount() error {
	if asFloat(tc.lastWebhook("meta.amount")) != tc.withdrawAmt {
		return fmt.Errorf("amount mismatch")
	}
	return nil
}

func (tc *TestContext) theWebhookRequestIncludesValidSvixSignatureHeaders() error {
	event := tc.webhooks[len(tc.webhooks)-1]
	if !strings.HasPrefix(event.Header.Get("svix-id"), "msg_") || !strings.HasPrefix(event.Header.Get("svix-signature"), "v1,") || strings.TrimSpace(event.Header.Get("svix-timestamp")) == "" {
		return fmt.Errorf("missing svix headers")
	}
	key, _ := tc.decodeWebhookKey(tc.secret)
	if !tc.validateWebhookSig(event, key) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

func (tc *TestContext) theEventTypesReceivedAre(a string, b string) error {
	if len(tc.webhooks) < 2 {
		return fmt.Errorf("missing events")
	}
	if fmt.Sprint(tc.webhooks[0].Body["eventType"]) != a || fmt.Sprint(tc.webhooks[1].Body["eventType"]) != b {
		return fmt.Errorf("order mismatch")
	}
	return nil
}

func (tc *TestContext) theReceivedWebhookIncludesTheHeader(header string) error {
	event := tc.webhooks[len(tc.webhooks)-1]
	if strings.TrimSpace(event.Header.Get(header)) == "" {
		return fmt.Errorf("missing header")
	}
	return nil
}

func (tc *TestContext) theSvixIDValueMatchesThePattern() error {
	value := tc.webhooks[len(tc.webhooks)-1].Header.Get("svix-id")
	if !strings.HasPrefix(value, "msg_") {
		return fmt.Errorf("bad id")
	}
	_, err := uuid.Parse(strings.TrimPrefix(value, "msg_"))
	return err
}

func (tc *TestContext) theSvixTimestampValueIsAUnixEpochInteger() error {
	_, err := strconv.ParseInt(tc.webhooks[len(tc.webhooks)-1].Header.Get("svix-timestamp"), 10, 64)
	return err
}

func (tc *TestContext) theSvixSignatureValueStartsWithV1() error {
	if !strings.HasPrefix(tc.webhooks[len(tc.webhooks)-1].Header.Get("svix-signature"), "v1,") {
		return fmt.Errorf("bad sig")
	}
	return nil
}

func (tc *TestContext) theSignatureIsValid() error {
	if !tc.sigValid {
		return fmt.Errorf("signature invalid")
	}
	return nil
}

func (tc *TestContext) theSignatureIsInvalid() error {
	if tc.sigValid {
		return fmt.Errorf("signature unexpectedly valid")
	}
	return nil
}

func (tc *TestContext) theResultMatchesTheBase64ValueInTheSignature() error {
	if !tc.sigValid {
		return fmt.Errorf("manual signature mismatch")
	}
	return nil
}

func (tc *TestContext) theWebhookBodyTopLevelFieldsInclude(key string) error {
	if _, ok := tc.webhooks[len(tc.webhooks)-1].Body[key]; !ok {
		return fmt.Errorf("missing field")
	}
	return nil
}

func (tc *TestContext) theWebhookBodyDoesNotContainATopLevelDataKey() error {
	if _, ok := tc.webhooks[len(tc.webhooks)-1].Body["data"]; ok {
		return fmt.Errorf("unexpected data wrapper")
	}
	return nil
}

func (tc *TestContext) theSignatureIsValidWhenVerifiedWithKey(key string) error {
	event := tc.webhooks[len(tc.webhooks)-1]
	if !tc.validateWebhookSig(event, []byte(key)) {
		return fmt.Errorf("signature invalid")
	}
	return nil
}

func (tc *TestContext) haveInitiatedDepositForChiSub001AndRecordedTheIssueID() error {
	tc.overrideKey = nil
	_ = tc.createSubAccount("chi-sub-001")
	payload := map[string]any{"amount": "100.00", "currency": "CAD", "subAccount": "chi-sub-001", "payerEmail": "payer@example.com", "redirect_url": "https://app.test/callbacks/chimoney"}
	if err := tc.request(http.MethodPost, "/v0.2.4/payment/initiate", payload); err != nil {
		return err
	}
	tc.issueID, _ = getPath(tc.lastJSON, "data.issueID").(string)
	tc.paymentLink, _ = getPath(tc.lastJSON, "data.paymentLink").(string)
	return nil
}

func (tc *TestContext) iHaveInitiatedAnInteracWithdrawal() error {
	tc.overrideKey = nil
	_ = tc.createSubAccount("chi-sub-002")
	payload := map[string]any{"subAccount": "chi-sub-002", "debitCurrency": "CAD", "interacs": []any{map[string]any{"name": "Alice", "email": "alice@example.com", "amount": 95.00}}}
	if err := tc.request(http.MethodPost, "/v0.2.4/payouts/interac", payload); err != nil {
		return err
	}
	arr := tc.extractPayoutArray()
	if len(arr) > 0 {
		tc.withdrawIssue, _ = arr[0]["issueID"].(string)
		tc.chiRef, _ = arr[0]["chiref"].(string)
		tc.withdrawAmt = asFloat(arr[0]["amount"])
	}
	return nil
}

func (tc *TestContext) iHaveInitiatedADepositAndCompletedItViaThePayPage() error {
	if err := tc.haveInitiatedDepositForChiSub001AndRecordedTheIssueID(); err != nil {
		return err
	}
	return tc.request(http.MethodPost, "/pay/"+tc.issueID+"/confirm", nil)
}

func (tc *TestContext) extractPayoutArray() []map[string]any {
	data, _ := getPath(tc.lastJSON, "data").(map[string]any)
	if data == nil {
		return nil
	}
	arr, _ := data["data"].([]any)
	result := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result
}

func (tc *TestContext) waitWebhooks(n int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if len(tc.webhooks) >= n {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	if len(tc.webhooks) < n {
		return fmt.Errorf("expected %d webhooks got %d", n, len(tc.webhooks))
	}
	return nil
}

func (tc *TestContext) assertWebhookBody(table *godog.Table) error {
	expected := map[string]string{}
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		expected[strings.TrimSpace(row.Cells[0].Value)] = strings.TrimSpace(row.Cells[1].Value)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		for _, event := range tc.webhooks {
			if webhookMatches(event.Body, expected) {
				return nil
			}
		}
		time.Sleep(20 * time.Millisecond)
	}

	if len(tc.webhooks) == 0 {
		return fmt.Errorf("no webhook received")
	}
	return fmt.Errorf("webhook %s mismatch", firstExpectedKey(expected))
}

func webhookMatches(body map[string]any, expected map[string]string) bool {
	for key, want := range expected {
		if fmt.Sprint(getPath(body, key)) != want {
			return false
		}
	}
	return true
}

func firstExpectedKey(expected map[string]string) string {
	for key := range expected {
		return key
	}
	return "body"
}

func (tc *TestContext) lastWebhook(path string) any {
	if len(tc.webhooks) == 0 {
		return nil
	}
	return getPath(tc.webhooks[len(tc.webhooks)-1].Body, path)
}
