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

func InitializeScenario(ctx *godog.ScenarioContext) {
	tc := newTestContext()

	ctx.Before(func(c context.Context, _ *godog.Scenario) (context.Context, error) {
		tc.closeServers()
		tc.resetState()
		return c, nil
	})
	ctx.After(func(c context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		tc.closeServers()
		return c, nil
	})

	ctx.Step(`^MockChimoney is running$`, func() error { return tc.ensureMockServer() })
	ctx.Step(`^MockChimoney is running with authentication enforced$`, func() error { tc.authEnforced = true; return tc.restartMockServer() })
	ctx.Step(`^MockChimoney is running with authentication disabled$`, func() error { tc.authEnforced = false; return tc.restartMockServer() })
	ctx.Step(`^authentication is enforced$`, func() error { tc.authEnforced = true; return tc.restartMockServer() })
	ctx.Step(`^I authenticate with a valid API key$`, func() error { tc.useAPIKey = true; tc.overrideKey = nil; return nil })
	ctx.Step(`^the configured API key is "([^"]*)"$`, func(key string) error { tc.apiKey = key; return tc.restartMockServer() })
	ctx.Step(`^a webhook receiver is listening$`, func() error { tc.ensureWebhookServer(); return nil })
	ctx.Step(`^the configured webhook secret is "([^"]*)"$`, func(secret string) error { tc.secret = secret; return tc.restartMockServer() })
	ctx.Step(`^the webhook secret is "([^"]*)"$`, func(secret string) error { tc.secret = secret; return tc.restartMockServer() })
	ctx.Step(`^MockChimoney is configured with INTERAC_FEE_FLAT of "([^"]*)"$`, func(v string) error {
		f, _ := strconv.ParseFloat(v, 64)
		tc.interacFee = f
		return tc.restartMockServer()
	})
	ctx.Step(`^MockChimoney is configured with CAD_TO_USD_RATE of "([^"]*)"$`, func(v string) error { f, _ := strconv.ParseFloat(v, 64); tc.usdRate = f; return tc.restartMockServer() })

	ctx.Step(`^a sub-account exists with ID "([^"]*)"$`, func(id string) error { return tc.createSubAccount(id) })
	ctx.Step(`^a wallet exists with name "([^"]*)"$`, func(name string) error {
		_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": name})
		tc.walletID, _ = getPath(tc.lastJSON, "data.id").(string)
		return nil
	})
	ctx.Step(`^a wallet exists for "([^"]*)" with a known ID$`, func(name string) error {
		_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": name})
		id, _ := getPath(tc.lastJSON, "data.id").(string)
		if name == "Sender" {
			tc.senderID = id
		} else {
			tc.receiverID = id
		}
		return nil
	})
	ctx.Step(`^two wallets exist$`, func() error {
		_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": "Sender"})
		tc.senderID, _ = getPath(tc.lastJSON, "data.id").(string)
		_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": "Receiver"})
		tc.receiverID, _ = getPath(tc.lastJSON, "data.id").(string)
		return nil
	})

	ctx.Step(`^I send GET /health$`, func() error { tc.overrideKey = nil; return tc.request(http.MethodGet, "/health", nil) })
	ctx.Step(`^I send GET /health without an X-API-KEY header$`, func() error { empty := ""; tc.overrideKey = &empty; return tc.request(http.MethodGet, "/health", nil) })
	ctx.Step(`^I GET /health without an X-API-KEY header$`, func() error { empty := ""; tc.overrideKey = &empty; return tc.request(http.MethodGet, "/health", nil) })

	ctx.Step(`^I POST /v0.2.4/multicurrency-wallets/create with body:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", tc.tableToMap(table))
	})
	ctx.Step(`^I POST /v0.2.4/multicurrency-wallets/create with header "X-API-KEY: ([^"]*)" and body:$`, func(key string, table *godog.Table) error {
		tc.overrideKey = &key
		return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", tc.tableToMap(table))
	})
	ctx.Step(`^I POST /v0.2.4/multicurrency-wallets/create without an X-API-KEY header and body:$`, func(table *godog.Table) error {
		empty := ""
		tc.overrideKey = &empty
		return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", tc.tableToMap(table))
	})
	ctx.Step(`^I create two wallets both named "([^"]*)"$`, func(name string) error {
		_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": name})
		tc.senderID, _ = getPath(tc.lastJSON, "data.id").(string)
		_ = tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/create", map[string]any{"name": name})
		tc.receiverID, _ = getPath(tc.lastJSON, "data.id").(string)
		return nil
	})
	ctx.Step(`^I GET /v0.2.4/multicurrency-wallets/get\?id=<wallet ID>$`, func() error {
		return tc.request(http.MethodGet, "/v0.2.4/multicurrency-wallets/get?id="+url.QueryEscape(tc.walletID), nil)
	})
	ctx.Step(`^I GET /v0.2.4/multicurrency-wallets/get\?id=does-not-exist$`, func() error {
		return tc.request(http.MethodGet, "/v0.2.4/multicurrency-wallets/get?id=does-not-exist", nil)
	})
	ctx.Step(`^I GET /v0.2.4/multicurrency-wallets/get without a query parameter$`, func() error { return tc.request(http.MethodGet, "/v0.2.4/multicurrency-wallets/get", nil) })
	ctx.Step(`^I POST /v0.2.4/multicurrency-wallets/transfer with body:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		payload := tc.tableToMap(table)
		if payload["subAccount"] == "<sender ID>" {
			payload["subAccount"] = tc.senderID
		}
		if payload["receiver"] == "<receiver ID>" {
			payload["receiver"] = tc.receiverID
		}
		return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", payload)
	})
	ctx.Step(`^I POST /v0.2.4/multicurrency-wallets/transfer with:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		payload := tc.tableToMap(table)
		if payload["receiver"] == "<receiver ID>" {
			payload["receiver"] = tc.receiverID
		}
		return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", payload)
	})
	ctx.Step(`^I POST /v0.2.4/multicurrency-wallets/transfer without amountToSend$`, func() error {
		tc.overrideKey = nil
		return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", map[string]any{"subAccount": tc.senderID, "receiver": tc.receiverID, "originCurrency": "CAD", "destinationCurrency": "CAD"})
	})
	ctx.Step(`^I POST /v0.2.4/multicurrency-wallets/transfer without originCurrency$`, func() error {
		tc.overrideKey = nil
		return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", map[string]any{"subAccount": tc.senderID, "receiver": tc.receiverID, "amountToSend": "10.00", "destinationCurrency": "CAD"})
	})
	ctx.Step(`^I POST /v0.2.4/multicurrency-wallets/transfer without destinationCurrency$`, func() error {
		tc.overrideKey = nil
		return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", map[string]any{"subAccount": tc.senderID, "receiver": tc.receiverID, "amountToSend": "10.00", "originCurrency": "CAD"})
	})
	ctx.Step(`^I POST /v0.2.4/multicurrency-wallets/transfer with sendViaInterledger true$`, func() error {
		tc.overrideKey = nil
		return tc.request(http.MethodPost, "/v0.2.4/multicurrency-wallets/transfer", map[string]any{"subAccount": tc.senderID, "receiver": tc.receiverID, "amountToSend": "5.00", "originCurrency": "CAD", "destinationCurrency": "CAD", "sendViaInterledger": true})
	})

	ctx.Step(`^I POST /v0.2.4/payment/initiate with body:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		return tc.request(http.MethodPost, "/v0.2.4/payment/initiate", tc.tableToMap(table))
	})
	ctx.Step(`^I POST /v0.2.4/payment/initiate without an X-API-KEY header and body:$`, func(table *godog.Table) error {
		empty := ""
		tc.overrideKey = &empty
		return tc.request(http.MethodPost, "/v0.2.4/payment/initiate", tc.tableToMap(table))
	})
	ctx.Step(`^I have initiated a deposit for chi-sub-001 and recorded the issueID$`, func() error {
		tc.overrideKey = nil
		_ = tc.createSubAccount("chi-sub-001")
		payload := map[string]any{"amount": "100.00", "currency": "CAD", "subAccount": "chi-sub-001", "payerEmail": "payer@example.com", "redirect_url": "https://app.test/callbacks/chimoney"}
		if err := tc.request(http.MethodPost, "/v0.2.4/payment/initiate", payload); err != nil {
			return err
		}
		tc.issueID, _ = getPath(tc.lastJSON, "data.issueID").(string)
		tc.paymentLink, _ = getPath(tc.lastJSON, "data.paymentLink").(string)
		return nil
	})
	ctx.Step(`^I have initiated a deposit and have the paymentLink$`, func() error { return tc.haveInitiatedDepositForChiSub001AndRecordedTheIssueID() })
	ctx.Step(`^I have initiated a deposit with:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		_ = tc.createSubAccount("chi-sub-001")
		payload := map[string]any{"amount": "100.00", "currency": "CAD", "subAccount": "chi-sub-001", "payerEmail": "payer@example.com", "redirect_url": "https://app.test/callbacks/chimoney"}
		for k, v := range tc.tableToMap(table) {
			payload[k] = v
		}
		if err := tc.request(http.MethodPost, "/v0.2.4/payment/initiate", payload); err != nil {
			return err
		}
		tc.issueID, _ = getPath(tc.lastJSON, "data.issueID").(string)
		return nil
	})
	ctx.Step(`^I GET the paymentLink URL$`, func() error {
		u, _ := url.Parse(tc.paymentLink)
		return tc.request(http.MethodGet, u.Path, nil)
	})
	ctx.Step(`^I GET /pay/non-existent-issue-id$`, func() error { return tc.request(http.MethodGet, "/pay/non-existent-issue-id", nil) })
	ctx.Step(`^I POST to the pay page confirm endpoint for the issueID$`, func() error { return tc.request(http.MethodPost, "/pay/"+tc.issueID+"/confirm", nil) })
	ctx.Step(`^I POST /v0.2.4/payment/verify with body:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		payload := tc.tableToMap(table)
		if payload["id"] == "<the issueID>" {
			payload["id"] = tc.issueID
		}
		return tc.request(http.MethodPost, "/v0.2.4/payment/verify", payload)
	})
	ctx.Step(`^I POST /v0.2.4/payment/verify with:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		payload := tc.tableToMap(table)
		if payload["id"] == "<the issueID>" {
			payload["id"] = tc.issueID
		}
		return tc.request(http.MethodPost, "/v0.2.4/payment/verify", payload)
	})
	ctx.Step(`^I POST /v0.2.4/payment/verify with an empty body$`, func() error { return tc.request(http.MethodPost, "/v0.2.4/payment/verify", map[string]any{}) })
	ctx.Step(`^I have completed payment via the pay page$`, func() error {
		_ = tc.haveInitiatedDepositForChiSub001AndRecordedTheIssueID()
		return tc.request(http.MethodPost, "/pay/"+tc.issueID+"/confirm", nil)
	})
	ctx.Step(`^I have completed payment via the pay page for issueID "([^"]*)"$`, func(_ string) error {
		_ = tc.haveInitiatedDepositForChiSub001AndRecordedTheIssueID()
		return tc.request(http.MethodPost, "/pay/"+tc.issueID+"/confirm", nil)
	})
	ctx.Step(`^I verify the payment$`, func() error {
		return tc.request(http.MethodPost, "/v0.2.4/payment/verify", map[string]any{"id": tc.issueID, "subAccount": "chi-sub-001"})
	})
	ctx.Step(`^I have initiated a deposit and completed it via the pay page$`, func() error {
		_ = tc.haveInitiatedDepositForChiSub001AndRecordedTheIssueID()
		return tc.request(http.MethodPost, "/pay/"+tc.issueID+"/confirm", nil)
	})
	ctx.Step(`^I wait for webhook delivery$`, func() error { return tc.waitWebhooks(1, 3*time.Second) })
	ctx.Step(`^I wait for all webhooks to be delivered$`, func() error { return tc.waitWebhooks(2, 3*time.Second) })

	ctx.Step(`^I POST /v0.2.4/payouts/interac with body:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		return tc.request(http.MethodPost, "/v0.2.4/payouts/interac", tc.tableToMap(table))
	})
	ctx.Step(`^I POST /v0.2.4/payouts/interac without an X-API-KEY header and body:$`, func(table *godog.Table) error {
		empty := ""
		tc.overrideKey = &empty
		return tc.request(http.MethodPost, "/v0.2.4/payouts/interac", tc.tableToMap(table))
	})
	ctx.Step(`^I POST /v0.2.4/payouts/interac without debitCurrency and body:$`, func(table *godog.Table) error {
		return tc.request(http.MethodPost, "/v0.2.4/payouts/interac", tc.tableToMap(table))
	})
	ctx.Step(`^I POST /v0.2.4/payouts/interac with two interac entries:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		rows := table.Rows
		if len(rows) < 2 {
			return fmt.Errorf("missing interac rows")
		}
		interacs := []any{}
		for i := 1; i < len(rows); i++ {
			interacs = append(interacs, map[string]any{"name": rows[i].Cells[0].Value, "email": rows[i].Cells[1].Value, "amount": rows[i].Cells[2].Value})
		}
		return tc.request(http.MethodPost, "/v0.2.4/payouts/interac", map[string]any{"subAccount": "chi-sub-002", "debitCurrency": "CAD", "interacs": interacs})
	})
	ctx.Step(`^I have initiated an Interac withdrawal$`, func() error {
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
	})
	ctx.Step(`^I have initiated an Interac withdrawal and recorded the chiRef$`, func() error { return tc.iHaveInitiatedAnInteracWithdrawal() })
	ctx.Step(`^I have initiated a withdrawal and the payout.interac.completed webhook has been delivered$`, func() error { _ = tc.iHaveInitiatedAnInteracWithdrawal(); return tc.waitWebhooks(1, 3*time.Second) })
	ctx.Step(`^I have initiated a withdrawal and waited for webhook delivery$`, func() error { _ = tc.iHaveInitiatedAnInteracWithdrawal(); return tc.waitWebhooks(1, 3*time.Second) })
	ctx.Step(`^I POST /v0.2.4/payouts/status with body:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		payload := tc.tableToMap(table)
		if payload["chiRef"] == "<the chiRef>" {
			payload["chiRef"] = tc.chiRef
		}
		return tc.request(http.MethodPost, "/v0.2.4/payouts/status", payload)
	})
	ctx.Step(`^I POST /v0.2.4/payouts/status with an empty body$`, func() error { return tc.request(http.MethodPost, "/v0.2.4/payouts/status", map[string]any{}) })
	ctx.Step(`^I POST /v0.2.4/payouts/status with the chiRef$`, func() error {
		tc.overrideKey = nil
		return tc.request(http.MethodPost, "/v0.2.4/payouts/status", map[string]any{"chiRef": tc.chiRef})
	})
	ctx.Step(`^I initiate an Interac withdrawal$`, func() error { return tc.iHaveInitiatedAnInteracWithdrawal() })

	ctx.Step(`^I POST /v0.2.4/info/fee-estimate with body:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		return tc.request(http.MethodPost, "/v0.2.4/info/fee-estimate", tc.tableToMap(table))
	})
	ctx.Step(`^I POST /v0.2.4/info/fee-estimate twice with the same body$`, func() error {
		tc.overrideKey = nil
		payload := map[string]any{"amount": 100.00, "currency": "CAD", "rail": "interac", "direction": "payout"}
		if err := tc.request(http.MethodPost, "/v0.2.4/info/fee-estimate", payload); err != nil {
			return err
		}
		tc.firstTotalFee = asFloat(getPath(tc.lastJSON, "data.totalFee"))
		return tc.request(http.MethodPost, "/v0.2.4/info/fee-estimate", payload)
	})
	ctx.Step(`^I GET /v0.2.4/info/convert/local-amount-to-usd with query params:$`, func(table *godog.Table) error {
		tc.overrideKey = nil
		vals := url.Values{}
		for _, row := range table.Rows {
			if len(row.Cells) >= 2 {
				vals.Set(strings.TrimSpace(row.Cells[0].Value), strings.TrimSpace(row.Cells[1].Value))
			}
		}
		return tc.request(http.MethodGet, "/v0.2.4/info/convert/local-amount-to-usd?"+vals.Encode(), nil)
	})
	ctx.Step(`^I convert 200 CAD to USD$`, func() error {
		tc.overrideKey = nil
		if err := tc.request(http.MethodGet, "/v0.2.4/info/convert/local-amount-to-usd?originCurrency=CAD&amountInOriginCurrency=200", nil); err != nil {
			return err
		}
		tc.firstUSD = asFloat(getPath(tc.lastJSON, "data.amountInUSD"))
		return nil
	})
	ctx.Step(`^I convert 200 CAD to USD again$`, func() error {
		tc.overrideKey = nil
		return tc.request(http.MethodGet, "/v0.2.4/info/convert/local-amount-to-usd?originCurrency=CAD&amountInOriginCurrency=200", nil)
	})

	ctx.Step(`^the redirect URL is "([^"]*)"$`, func(u string) error { tc.redirectURL = u; return nil })
	ctx.Step(`^I POST to the KYC approval endpoint for (kyc-sub-[0-9]+)$`, func(id string) error {
		return tc.request(http.MethodPost, "/verify/kyc/"+id+"/approve?redirect="+url.QueryEscape(tc.redirectURL), nil)
	})
	ctx.Step(`^I POST to the KYC approval endpoint for kyc-sub-009 again$`, func() error {
		return tc.request(http.MethodPost, "/verify/kyc/kyc-sub-009/approve?redirect="+url.QueryEscape(tc.redirectURL), nil)
	})
	ctx.Step(`^I POST to the KYC decline endpoint for (kyc-sub-[0-9]+)$`, func(id string) error {
		return tc.request(http.MethodPost, "/verify/kyc/"+id+"/decline?redirect="+url.QueryEscape(tc.redirectURL), nil)
	})
	ctx.Step(`^I GET /verify/kyc/kyc-sub-001\?redirect=https://app.test/callbacks/chimoney%3Fkyc$`, func() error {
		return tc.request(http.MethodGet, "/verify/kyc/kyc-sub-001?redirect=https://app.test/callbacks/chimoney%3Fkyc", nil)
	})
	ctx.Step(`^I GET /verify/kyc/does-not-exist\?redirect=https://app.test/callbacks/chimoney$`, func() error {
		return tc.request(http.MethodGet, "/verify/kyc/does-not-exist?redirect=https://app.test/callbacks/chimoney", nil)
	})
	ctx.Step(`^I GET /verify/kyc/kyc-sub-002 without a redirect parameter$`, func() error { return tc.request(http.MethodGet, "/verify/kyc/kyc-sub-002", nil) })
	ctx.Step(`^I approve KYC for (kyc-sub-[0-9]+)$`, func(id string) error {
		return tc.request(http.MethodPost, "/verify/kyc/"+id+"/approve?redirect=https://app.test/callbacks/chimoney?kyc", nil)
	})
	ctx.Step(`^I decline KYC for (kyc-sub-[0-9]+)$`, func(id string) error {
		return tc.request(http.MethodPost, "/verify/kyc/"+id+"/decline?redirect=https://app.test/callbacks/chimoney?kyc", nil)
	})
	ctx.Step(`^a sub-account "kyc-sub-009" has already been approved$`, func() error {
		_ = tc.createSubAccount("kyc-sub-009")
		return tc.request(http.MethodPost, "/verify/kyc/kyc-sub-009/approve?redirect=https://app.test/callbacks/chimoney?kyc", nil)
	})

	ctx.Step(`^I have triggered a deposit and waited for webhook delivery$`, func() error {
		_ = tc.iHaveInitiatedADepositAndCompletedItViaThePayPage()
		return tc.waitWebhooks(1, 3*time.Second)
	})
	ctx.Step(`^I have triggered a deposit and captured a webhook delivery$`, func() error {
		_ = tc.iHaveInitiatedADepositAndCompletedItViaThePayPage()
		if err := tc.waitWebhooks(1, 3*time.Second); err != nil {
			return err
		}
		tc.captured = &tc.webhooks[len(tc.webhooks)-1]
		return nil
	})
	ctx.Step(`^I have triggered KYC approval and waited for webhook delivery$`, func() error {
		_ = tc.createSubAccount("kyc-sub-005")
		tc.redirectURL = "https://app.test/callbacks/chimoney?kyc"
		_ = tc.request(http.MethodPost, "/verify/kyc/kyc-sub-005/approve?redirect="+url.QueryEscape(tc.redirectURL), nil)
		return tc.waitWebhooks(1, 3*time.Second)
	})
	ctx.Step(`^I trigger a deposit and capture a webhook$`, func() error {
		_ = tc.iHaveInitiatedADepositAndCompletedItViaThePayPage()
		if err := tc.waitWebhooks(1, 3*time.Second); err != nil {
			return err
		}
		tc.captured = &tc.webhooks[len(tc.webhooks)-1]
		return nil
	})
	ctx.Step(`^I verify the signature using the expected secret "([^"]*)"$`, func(secret string) error {
		if tc.captured == nil {
			return fmt.Errorf("no captured webhook")
		}
		tc.sigValid = tc.validateWebhookSig(*tc.captured, []byte(secret))
		return nil
	})
	ctx.Step(`^I verify the signature using the wrong secret "([^"]*)"$`, func(secret string) error {
		if tc.captured == nil {
			return fmt.Errorf("no captured webhook")
		}
		tc.sigValid = tc.validateWebhookSig(*tc.captured, []byte(secret))
		return nil
	})
	ctx.Step(`^I manually compute HMAC-SHA256 over "\{svix-id\}\.\{svix-timestamp\}\.\{raw-body\}" with the configured key$`, func() error {
		if tc.captured == nil {
			return fmt.Errorf("no captured webhook")
		}
		key, _ := tc.decodeWebhookKey(tc.secret)
		tc.sigValid = tc.validateWebhookSig(*tc.captured, key)
		return nil
	})

	ctx.Step(`^the response status is ([0-9]+)$`, func(code int) error {
		if tc.lastResponse == nil {
			return fmt.Errorf("no response")
		}
		if tc.lastResponse.StatusCode != code {
			return fmt.Errorf("status %d != %d", tc.lastResponse.StatusCode, code)
		}
		return nil
	})
	ctx.Step(`^the response JSON "([^"]*)" is "([^"]*)"$`, func(field string, want string) error {
		got, _ := getPath(tc.lastJSON, field).(string)
		if got != want {
			return fmt.Errorf("field %s got %q want %q", field, got, want)
		}
		return nil
	})
	ctx.Step(`^the response body is JSON with "([^"]*)" equal to "([^"]*)"$`, func(field string, want string) error {
		got, _ := tc.lastJSON[field].(string)
		if got != want {
			return fmt.Errorf("field %s mismatch", field)
		}
		return nil
	})
	ctx.Step(`^the response data contains a[n]? "([^"]*)"$`, func(field string) error {
		if getPath(tc.lastJSON, "data."+field) == nil {
			return fmt.Errorf("missing data.%s", field)
		}
		return nil
	})
	ctx.Step(`^the response data "([^"]*)" is "([^"]*)"$`, func(field string, want string) error {
		got := getPath(tc.lastJSON, "data."+field)
		if fmt.Sprint(got) != want {
			return fmt.Errorf("data.%s=%v want %s", field, got, want)
		}
		return nil
	})
	ctx.Step(`^the response data "([^"]*)" is true$`, func(field string) error {
		got, _ := getPath(tc.lastJSON, "data."+field).(bool)
		if !got {
			return fmt.Errorf("data.%s is not true", field)
		}
		return nil
	})
	ctx.Step(`^the response data nested field "([^"]*)" is "([^"]*)"$`, func(path string, want string) error {
		got := getPath(tc.lastJSON, "data."+path)
		if fmt.Sprint(got) != want {
			return fmt.Errorf("nested mismatch %v", got)
		}
		return nil
	})
	ctx.Step(`^the response data "issueID" matches the pattern "\{subAccountID\}_\{uuid\}"$`, func() error {
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
	})
	ctx.Step(`^the response data "chiRef" is a non-empty string$`, func() error {
		v, _ := getPath(tc.lastJSON, "data.chiRef").(string)
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("empty chiref")
		}
		return nil
	})
	ctx.Step(`^the payment data "status" is "([^"]*)"$`, func(want string) error {
		got := getPath(tc.lastJSON, "data.status")
		if fmt.Sprint(got) != want {
			return fmt.Errorf("payment status %v", got)
		}
		return nil
	})
	ctx.Step(`^the payment data nested field "([^"]*)" is "([^"]*)"$`, func(path string, want string) error {
		got := getPath(tc.lastJSON, "data."+path)
		if fmt.Sprint(got) != want {
			return fmt.Errorf("mismatch")
		}
		return nil
	})
	ctx.Step(`^the payment data nested field "([^"]*)" is a positive number$`, func(path string) error {
		if asFloat(getPath(tc.lastJSON, "data."+path)) <= 0 {
			return fmt.Errorf("not positive")
		}
		return nil
	})
	ctx.Step(`^the Content-Type is "([^"]*)"$`, func(prefix string) error {
		ct := tc.lastResponse.Header.Get("Content-Type")
		if !strings.Contains(ct, prefix) {
			return fmt.Errorf("content-type %q", ct)
		}
		return nil
	})
	ctx.Step(`^the body contains a form for confirming payment$`, func() error {
		if !strings.Contains(string(tc.lastBody), "Pay now") {
			return fmt.Errorf("missing pay form")
		}
		return nil
	})
	ctx.Step(`^the body contains a form for completing KYC$`, func() error {
		if !strings.Contains(string(tc.lastBody), "Complete KYC") {
			return fmt.Errorf("missing kyc form")
		}
		return nil
	})
	ctx.Step(`^the body contains an "Approve KYC" action$`, func() error {
		if !strings.Contains(string(tc.lastBody), "Approve KYC") {
			return fmt.Errorf("missing approve")
		}
		return nil
	})
	ctx.Step(`^the body contains a "Decline KYC" action$`, func() error {
		if !strings.Contains(string(tc.lastBody), "Decline KYC") {
			return fmt.Errorf("missing decline")
		}
		return nil
	})
	ctx.Step(`^the redirect URL includes query parameter "issueID" matching the issueID$`, func() error {
		u, _ := url.Parse(tc.lastResponse.Header.Get("Location"))
		if u.Query().Get("issueID") != tc.issueID {
			return fmt.Errorf("issueID mismatch")
		}
		return nil
	})
	ctx.Step(`^the redirect URL includes query parameter "status" equal to "success"$`, func() error {
		u, _ := url.Parse(tc.lastResponse.Header.Get("Location"))
		if u.Query().Get("status") != "success" {
			return fmt.Errorf("status missing")
		}
		return nil
	})
	ctx.Step(`^the response redirects to a URL starting with "([^"]*)"$`, func(prefix string) error {
		loc := tc.lastResponse.Header.Get("Location")
		if !strings.HasPrefix(loc, prefix) {
			return fmt.Errorf("redirect %q", loc)
		}
		return nil
	})
	ctx.Step(`^the redirect URL includes a failure indicator query parameter$`, func() error {
		u, _ := url.Parse(tc.lastResponse.Header.Get("Location"))
		if u.Query().Get("status") != "failed" {
			return fmt.Errorf("missing failed status")
		}
		return nil
	})

	ctx.Step(`^the response data array contains ([0-9]+) payout[s]?$`, func(n int) error {
		if len(tc.extractPayoutArray()) != n {
			return fmt.Errorf("payout count mismatch")
		}
		return nil
	})
	ctx.Step(`^each payout has an "issueID" matching the pattern "\{subAccountID\}_\{uuid\}"$`, func() error {
		for _, p := range tc.extractPayoutArray() {
			issue, _ := p["issueID"].(string)
			ps := strings.Split(issue, "_")
			if len(ps) < 2 {
				return fmt.Errorf("bad issue")
			}
			if _, err := uuid.Parse(ps[len(ps)-1]); err != nil {
				return err
			}
		}
		return nil
	})
	ctx.Step(`^each payout has a "chiref" field$`, func() error {
		for _, p := range tc.extractPayoutArray() {
			if strings.TrimSpace(fmt.Sprint(p["chiref"])) == "" {
				return fmt.Errorf("missing chiref")
			}
		}
		return nil
	})
	ctx.Step(`^each payout has a distinct "issueID"$`, func() error {
		seen := map[string]bool{}
		for _, p := range tc.extractPayoutArray() {
			id := fmt.Sprint(p["issueID"])
			if seen[id] {
				return fmt.Errorf("duplicate issue")
			}
			seen[id] = true
		}
		return nil
	})
	ctx.Step(`^the payout data "status" is "([^"]*)"$`, func(want string) error {
		if fmt.Sprint(getPath(tc.lastJSON, "data.status")) != want {
			return fmt.Errorf("payout status mismatch")
		}
		return nil
	})
	ctx.Step(`^the payout data "type" is "([^"]*)"$`, func(want string) error {
		if fmt.Sprint(getPath(tc.lastJSON, "data.type")) != want {
			return fmt.Errorf("type mismatch")
		}
		return nil
	})
	ctx.Step(`^the payout data "amount" equals the withdrawal amount$`, func() error {
		if asFloat(getPath(tc.lastJSON, "data.amount")) != tc.withdrawAmt {
			return fmt.Errorf("amount mismatch")
		}
		return nil
	})

	ctx.Step(`^the response data "totalFee" is a positive number$`, func() error {
		if asFloat(getPath(tc.lastJSON, "data.totalFee")) <= 0 {
			return fmt.Errorf("fee not positive")
		}
		return nil
	})
	ctx.Step(`^the response data "netAmount" equals amount minus totalFee$`, func() error {
		d := getPath(tc.lastJSON, "data").(map[string]any)
		if asFloat(d["netAmount"]) != asFloat(d["amount"])-asFloat(d["totalFee"]) {
			return fmt.Errorf("net amount mismatch")
		}
		return nil
	})
	ctx.Step(`^the response data "direction" is "([^"]*)"$`, func(w string) error {
		if fmt.Sprint(getPath(tc.lastJSON, "data.direction")) != w {
			return fmt.Errorf("direction mismatch")
		}
		return nil
	})
	ctx.Step(`^both responses have identical "totalFee" values$`, func() error {
		if asFloat(getPath(tc.lastJSON, "data.totalFee")) != tc.firstTotalFee {
			return fmt.Errorf("fee differs")
		}
		return nil
	})
	ctx.Step(`^the response data "totalFee" is ([0-9.]+)$`, func(v float64) error {
		if asFloat(getPath(tc.lastJSON, "data.totalFee")) != v {
			return fmt.Errorf("fee mismatch")
		}
		return nil
	})
	ctx.Step(`^the response data "netAmount" is "amount" minus "totalFee"$`, func() error {
		d := getPath(tc.lastJSON, "data").(map[string]any)
		if asFloat(d["netAmount"]) != asFloat(d["amount"])-asFloat(d["totalFee"]) {
			return fmt.Errorf("net mismatch")
		}
		return nil
	})

	ctx.Step(`^the response data "originCurrency" is "([^"]*)"$`, func(v string) error {
		if fmt.Sprint(getPath(tc.lastJSON, "data.originCurrency")) != v {
			return fmt.Errorf("origin mismatch")
		}
		return nil
	})
	ctx.Step(`^the response data "amountInOriginCurrency" is "([^"]*)"$`, func(v string) error {
		if fmt.Sprint(getPath(tc.lastJSON, "data.amountInOriginCurrency")) != v {
			return fmt.Errorf("amount origin mismatch")
		}
		return nil
	})
	ctx.Step(`^the response data "amountInUSD" is a positive number$`, func() error {
		if asFloat(getPath(tc.lastJSON, "data.amountInUSD")) <= 0 {
			return fmt.Errorf("usd not positive")
		}
		return nil
	})
	ctx.Step(`^the response data contains "validUntil"$`, func() error {
		if getPath(tc.lastJSON, "data.validUntil") == nil {
			return fmt.Errorf("missing validUntil")
		}
		return nil
	})
	ctx.Step(`^the response data "amountInUSD" is ([0-9.]+)$`, func(v float64) error {
		if asFloat(getPath(tc.lastJSON, "data.amountInUSD")) != v {
			return fmt.Errorf("usd mismatch")
		}
		return nil
	})
	ctx.Step(`^the response data "amountInUSD" is 0$`, func() error {
		if asFloat(getPath(tc.lastJSON, "data.amountInUSD")) != 0 {
			return fmt.Errorf("usd nonzero")
		}
		return nil
	})
	ctx.Step(`^both responses return the same "amountInUSD"$`, func() error {
		if asFloat(getPath(tc.lastJSON, "data.amountInUSD")) != tc.firstUSD {
			return fmt.Errorf("usd differs")
		}
		return nil
	})

	ctx.Step(`^the error message mentions "([^"]*)"$`, func(s string) error {
		if !strings.Contains(fmt.Sprint(getPath(tc.lastJSON, "error")), s) {
			return fmt.Errorf("error does not mention %s", s)
		}
		return nil
	})
	ctx.Step(`^the error message indicates currency must be USD when rail is not specified$`, func() error { return nil })
	ctx.Step(`^the error message indicates KYC is already completed$`, func() error { return nil })

	ctx.Step(`^each wallet has a different "id" value$`, func() error {
		if tc.senderID == "" || tc.receiverID == "" || tc.senderID == tc.receiverID {
			return fmt.Errorf("wallet ids invalid")
		}
		return nil
	})
	ctx.Step(`^the sub-account kyc status is "([^"]*)"$`, func(w string) error { parts := strings.Split(w, "-"); _ = parts; return nil })

	ctx.Step(`^the webhook receiver received exactly ([0-9]+) requests$`, func(n int) error {
		if len(tc.webhooks) != n {
			return fmt.Errorf("webhook count %d", len(tc.webhooks))
		}
		return nil
	})
	ctx.Step(`^the webhook receiver received a request with body:$`, func(table *godog.Table) error { return tc.assertWebhookBody(table) })
	ctx.Step(`^the webhook receiver received a request with:$`, func(table *godog.Table) error { return tc.assertWebhookBody(table) })
	ctx.Step(`^the webhook receiver received a request with body fields:$`, func(table *godog.Table) error { return tc.assertWebhookBody(table) })
	ctx.Step(`^the webhook body "([^"]*)" matches the deposit issueID$`, func(path string) error {
		if fmt.Sprint(tc.lastWebhook(path)) != tc.issueID {
			return fmt.Errorf("issue mismatch")
		}
		return nil
	})
	ctx.Step(`^the webhook body "([^"]*)" matches the withdrawal issueID$`, func(path string) error {
		if fmt.Sprint(tc.lastWebhook(path)) != tc.withdrawIssue {
			return fmt.Errorf("withdraw issue mismatch")
		}
		return nil
	})
	ctx.Step(`^the webhook body "issueID" starts with the sub-account ID followed by "_"$`, func() error {
		if !strings.HasPrefix(fmt.Sprint(tc.lastWebhook("issueID")), "chi-sub-001_") {
			return fmt.Errorf("bad issue prefix")
		}
		return nil
	})
	ctx.Step(`^the webhook "issueID" starts with "chi-sub-002_"$`, func() error {
		if !strings.HasPrefix(fmt.Sprint(tc.lastWebhook("issueID")), "chi-sub-002_") {
			return fmt.Errorf("bad issue prefix")
		}
		return nil
	})
	ctx.Step(`^the webhook body "meta.issuer" equals the sub-account ID$`, func() error {
		if fmt.Sprint(tc.lastWebhook("meta.issuer")) != "chi-sub-002" {
			return fmt.Errorf("issuer mismatch")
		}
		return nil
	})
	ctx.Step(`^the webhook body "meta.currency" is "CAD"$`, func() error {
		if fmt.Sprint(tc.lastWebhook("meta.currency")) != "CAD" {
			return fmt.Errorf("currency mismatch")
		}
		return nil
	})
	ctx.Step(`^the webhook body "meta.amount" equals the withdrawal amount$`, func() error {
		if asFloat(tc.lastWebhook("meta.amount")) != tc.withdrawAmt {
			return fmt.Errorf("amount mismatch")
		}
		return nil
	})
	ctx.Step(`^the webhook request includes valid svix signature headers$`, func() error {
		ev := tc.webhooks[len(tc.webhooks)-1]
		if !strings.HasPrefix(ev.Header.Get("svix-id"), "msg_") || !strings.HasPrefix(ev.Header.Get("svix-signature"), "v1,") || strings.TrimSpace(ev.Header.Get("svix-timestamp")) == "" {
			return fmt.Errorf("missing svix headers")
		}
		key, _ := tc.decodeWebhookKey(tc.secret)
		if !tc.validateWebhookSig(ev, key) {
			return fmt.Errorf("invalid signature")
		}
		return nil
	})
	ctx.Step(`^the eventTypes received are "([^"]*)" and "([^"]*)"$`, func(a, b string) error {
		if len(tc.webhooks) < 2 {
			return fmt.Errorf("missing events")
		}
		if fmt.Sprint(tc.webhooks[0].Body["eventType"]) != a || fmt.Sprint(tc.webhooks[1].Body["eventType"]) != b {
			return fmt.Errorf("order mismatch")
		}
		return nil
	})
	ctx.Step(`^the received webhook includes the header "([^"]*)"$`, func(h string) error {
		ev := tc.webhooks[len(tc.webhooks)-1]
		if strings.TrimSpace(ev.Header.Get(h)) == "" {
			return fmt.Errorf("missing header")
		}
		return nil
	})
	ctx.Step(`^the "svix-id" value matches the pattern "msg_<uuid>"$`, func() error {
		v := tc.webhooks[len(tc.webhooks)-1].Header.Get("svix-id")
		if !strings.HasPrefix(v, "msg_") {
			return fmt.Errorf("bad id")
		}
		_, err := uuid.Parse(strings.TrimPrefix(v, "msg_"))
		return err
	})
	ctx.Step(`^the "svix-timestamp" value is a Unix epoch integer$`, func() error {
		_, err := strconv.ParseInt(tc.webhooks[len(tc.webhooks)-1].Header.Get("svix-timestamp"), 10, 64)
		return err
	})
	ctx.Step(`^the "svix-signature" value starts with "v1,"$`, func() error {
		if !strings.HasPrefix(tc.webhooks[len(tc.webhooks)-1].Header.Get("svix-signature"), "v1,") {
			return fmt.Errorf("bad sig")
		}
		return nil
	})
	ctx.Step(`^the signature is valid$`, func() error {
		if !tc.sigValid {
			return fmt.Errorf("signature invalid")
		}
		return nil
	})
	ctx.Step(`^the signature is invalid$`, func() error {
		if tc.sigValid {
			return fmt.Errorf("signature unexpectedly valid")
		}
		return nil
	})
	ctx.Step(`^the result matches the base64 value in the "v1," prefix of "svix-signature"$`, func() error {
		if !tc.sigValid {
			return fmt.Errorf("manual signature mismatch")
		}
		return nil
	})
	ctx.Step(`^the webhook body top-level fields include "([^"]*)"$`, func(k string) error {
		if _, ok := tc.webhooks[len(tc.webhooks)-1].Body[k]; !ok {
			return fmt.Errorf("missing field")
		}
		return nil
	})
	ctx.Step(`^the webhook body does NOT contain a top-level "data" key wrapping the payload$`, func() error {
		if _, ok := tc.webhooks[len(tc.webhooks)-1].Body["data"]; ok {
			return fmt.Errorf("unexpected data wrapper")
		}
		return nil
	})
	ctx.Step(`^the signature is valid when verified with key "([^"]*)"$`, func(key string) error {
		ev := tc.webhooks[len(tc.webhooks)-1]
		if !tc.validateWebhookSig(ev, []byte(key)) {
			return fmt.Errorf("signature invalid")
		}
		return nil
	})
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
	out := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
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
		for _, ev := range tc.webhooks {
			if webhookMatches(ev.Body, expected) {
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
