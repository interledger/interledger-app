//go:build e2e
// +build e2e

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

func registerDepositSteps(sc *godog.ScenarioContext, tc *TestContext) {
	sc.Step(`^the wallet webhook URL is configured to "([^"]+)"$`, tc.walletWebhookURLConfigured)

	sc.Step(`^I simulate a test deposit with the following details:$`, tc.simulateTestDepositWithDetails)
	sc.Step(`^I have simulated a test deposit of ([0-9.]+) (ZAR|USD)$`, tc.simulateTestDepositOf)
	sc.Step(`^I have simulated a test deposit$`, tc.simulateTestDeposit)
	sc.Step(`^I simulate a test deposit of ([0-9.]+) (ZAR|USD)$`, tc.simulateTestDepositOf)
	sc.Step(`^I simulate a ZAR deposit of ([0-9.]+)$`, tc.simulateTestDepositZAR)
	sc.Step(`^I simulate a USD deposit of ([0-9.]+)$`, tc.simulateTestDepositUSD)
	sc.Step(`^I simulate a test deposit with specific deposit reference$`, tc.simulateTestDepositWithSpecificReference)
	sc.Step(`^I simulate another deposit with the same reference$`, tc.simulateAnotherDepositSameReference)
	sc.Step(`^I have simulated (\d+) test deposits$`, tc.simulateMultipleDeposits)
	sc.Step(`^I have simulated 5 test deposits$`, tc.simulateFiveDeposits)

	sc.Step(`^I wait (\d+) seconds for the webhook to be delivered$`, tc.waitForWebhookSeconds)
	sc.Step(`^I wait for the webhook delivery$`, tc.waitForWebhookDelivery)
	sc.Step(`^I wait for the deposit to complete$`, tc.waitForDepositCompletion)
	sc.Step(`^I wait for completion$`, tc.waitForDepositCompletion)
	sc.Step(`^I wait for all deposits to complete$`, tc.waitForAllDeposits)
	sc.Step(`^I wait for both to complete$`, tc.waitForAllDeposits)

	sc.Step(`^the response status is "([^"]+)"$`, tc.depositResponseStatusIs)
	sc.Step(`^the wallet receives a webhook with:$`, tc.walletReceivesWebhookWith)
	sc.Step(`^the webhook includes:$`, tc.webhookIncludes)
	sc.Step(`^the webhook includes valid headers:$`, tc.webhookIncludesValidHeaders)
	sc.Step(`^the webhook body includes all required fields:$`, tc.webhookBodyIncludesRequiredFields)

	sc.Step(`^the sub-account ZAR balance is ([0-9.]+)$`, tc.subAccountZARBalanceIs)
	sc.Step(`^the total ZAR balance is ([0-9.]+)$`, tc.subAccountZARBalanceIs)
	sc.Step(`^the ZAR balance is ([0-9.]+)$`, tc.subAccountZARBalanceIs)
	sc.Step(`^the USD balance is ([0-9.]+)$`, tc.subAccountUSDBalanceIs)
	sc.Step(`^the sub-account starts with zero ZAR balance$`, tc.subAccountStartsWithZeroZARBalance)
	sc.Step(`^the sub-account starts with zero balance$`, tc.subAccountStartsWithZeroBalance)
	sc.Step(`^the sub-account has zero balance$`, tc.subAccountStartsWithZeroBalance)

	sc.Step(`^I request to list company deposits with limit (\d+)$`, tc.listCompanyDepositsWithLimit)
	sc.Step(`^I request to list deposits with limit (\d+) and page (\d+)$`, tc.listCompanyDepositsWithLimitAndPage)
	sc.Step(`^I request company deposits$`, tc.listCompanyDeposits)
	sc.Step(`^I attempt to list deposits without authentication$`, tc.listCompanyDepositsWithoutAuth)

	sc.Step(`^the response includes (\d+) deposits$`, tc.responseIncludesDeposits)
	sc.Step(`^the response includes (\d+) deposits on page (\d+)$`, tc.responseIncludesDepositsOnPage)
	sc.Step(`^each deposit includes:$`, tc.eachDepositIncludes)
	sc.Step(`^both deposits are recorded$`, tc.bothDepositsRecorded)
	sc.Step(`^each deposit has a unique transaction ID$`, tc.eachDepositHasUniqueID)
	sc.Step(`^each deposit shows as completed$`, tc.eachDepositShowsCompleted)
	sc.Step(`^all (\d+) deposits appear in the list$`, tc.allDepositsAppearInList)

	sc.Step(`^I have created two sub-accounts with different deposit references$`, tc.createTwoSubAccountsWithDepositReferences)
	sc.Step(`^I simulate a ZAR deposit for wallet_(\d+)'s deposit reference$`, tc.simulateDepositForWalletReference)
	sc.Step(`^wallet_(\d+)'s balance is credited$`, tc.walletBalanceCredited)
	sc.Step(`^wallet_(\d+)'s balance remains unchanged$`, tc.walletBalanceUnchanged)

	sc.Step(`^I attempt to simulate a deposit with invalid account ID "([^"]+)":$`, tc.attemptDepositInvalidAccount)
	sc.Step(`^I attempt to simulate a deposit with amount ([\d\.-]+)$`, tc.attemptDepositWithAmount)
	sc.Step(`^I attempt to simulate a deposit without authentication$`, tc.attemptDepositWithoutAuth)
	sc.Step(`^all deposits have completed$`, tc.allDepositsHaveCompleted)
	sc.Step(`^I have simulated a test deposit with specific deposit reference$`, tc.simulateTestDepositWithSpecificReference)
	sc.Step(`^the account balance is credited twice$`, tc.accountBalanceIsCreditedTwice)
	sc.Step(`^the response includes a transaction ID$`, tc.responseIncludesTransactionID)
}

func (tc *TestContext) walletWebhookURLConfigured(url string) error {
	// Clear server-side deposits so list/count tests start from zero
	if err := tc.clearServerDeposits(); err != nil {
		fmt.Printf("WARN: failed to clear server deposits: %v\n", err)
	}
	return tc.startWebhookServer(url)
}

func (tc *TestContext) simulateTestDepositWithDetails(table *godog.Table) error {
	values := tableToMap(table)
	accountID := values["accountId"]
	if accountID == "(the account id)" {
		accountID = tc.lastSubAccount.AccountID
	}
	amountStr := values["amount"]
	currency := values["currencyCode"]
	depositReference := values["depositReference"]
	if depositReference == "(the deposit ref)" {
		ref, err := tc.depositReferenceForCurrency(currency)
		if err != nil {
			return err
		}
		depositReference = ref
	}

	return tc.doSimulateTestDeposit(accountID, amountStr, currency, depositReference, true)
}

func (tc *TestContext) simulateTestDepositOf(amount string, currency string) error {
	accountID := tc.lastSubAccount.AccountID
	depositReference, err := tc.depositReferenceForCurrency(currency)
	if err != nil {
		return err
	}
	return tc.doSimulateTestDeposit(accountID, amount, currency, depositReference, true)
}

func (tc *TestContext) simulateTestDeposit() error {
	return tc.simulateTestDepositOf("1000.00", "ZAR")
}

func (tc *TestContext) simulateTestDepositZAR(amount string) error {
	return tc.simulateTestDepositOf(amount, "ZAR")
}

func (tc *TestContext) simulateTestDepositUSD(amount string) error {
	return tc.simulateTestDepositOf(amount, "USD")
}

func (tc *TestContext) simulateTestDepositWithSpecificReference() error {
	accountID := tc.lastSubAccount.AccountID
	currency := "ZAR"
	depositReference, err := tc.depositReferenceForCurrency(currency)
	if err != nil {
		return err
	}
	tc.lastDepositReference = depositReference
	return tc.doSimulateTestDeposit(accountID, "1000.00", currency, depositReference, true)
}

func (tc *TestContext) simulateAnotherDepositSameReference() error {
	if tc.lastDepositReference == "" {
		return fmt.Errorf("no deposit reference stored")
	}
	accountID := tc.lastSubAccount.AccountID
	return tc.doSimulateTestDeposit(accountID, "1000.00", "ZAR", tc.lastDepositReference, true)
}

func (tc *TestContext) simulateMultipleDeposits(count int) error {
	for i := 0; i < count; i++ {
		amount := fmt.Sprintf("%d.00", 1000+100*i)
		if err := tc.simulateTestDepositOf(amount, "ZAR"); err != nil {
			return err
		}
	}
	return nil
}

func (tc *TestContext) simulateFiveDeposits() error {
	return tc.simulateMultipleDeposits(5)
}

func (tc *TestContext) waitForWebhookSeconds(seconds int) error {
	// Snapshot baseline BEFORE sleeping, then wait for 1 more webhook
	globalWebhookEventsMu.Lock()
	baseline := len(globalWebhookEvents)
	globalWebhookEventsMu.Unlock()

	// Sleep the requested time
	time.Sleep(time.Duration(seconds) * time.Second)

	// Now wait for baseline+1 (the webhook should have arrived during sleep)
	target := baseline + 1
	return tc.waitForWebhookCount(target, 5*time.Second)
}

func (tc *TestContext) waitForWebhookDelivery() error {
	return tc.waitForNextWebhook(5 * time.Second)
}

func (tc *TestContext) waitForDepositCompletion() error {
	return tc.waitForNextWebhook(5 * time.Second)
}

func (tc *TestContext) waitForAllDeposits() error {
	count := len(tc.createdDeposits)
	if count == 0 {
		count = 1
	}
	// Wait for `count` new webhooks beyond current total
	globalWebhookEventsMu.Lock()
	baseline := len(globalWebhookEvents)
	globalWebhookEventsMu.Unlock()

	target := baseline + count
	return tc.waitForWebhookCount(target, 10*time.Second)
}

func (tc *TestContext) depositResponseStatusIs(status string) error {
	if tc.lastDepositResponse.Status != status {
		return fmt.Errorf("expected deposit status %s, got %s", status, tc.lastDepositResponse.Status)
	}
	return nil
}

func (tc *TestContext) walletReceivesWebhookWith(table *godog.Table) error {
	event, err := tc.lastWebhookEvent()
	if err != nil {
		return err
	}

	values := tableToMap(table)
	for field, expected := range values {
		switch field {
		case "accountId":
			if expected == "(the account id)" {
				expected = tc.lastSubAccount.AccountID
			}
			if event.Body.AccountID != expected {
				return fmt.Errorf("expected accountId %s, got %s", expected, event.Body.AccountID)
			}
		case "amount":
			if !floatEquals(event.Body.Amount, expected) {
				return fmt.Errorf("expected amount %s, got %.2f", expected, event.Body.Amount)
			}
		case "currencyCode":
			if event.Body.CurrencyCode != expected {
				return fmt.Errorf("expected currencyCode %s, got %s", expected, event.Body.CurrencyCode)
			}
		case "status":
			if event.Body.Status != expected {
				return fmt.Errorf("expected status %s, got %s", expected, event.Body.Status)
			}
		case "code":
			code, err := strconv.Atoi(expected)
			if err != nil {
				return fmt.Errorf("invalid code %s", expected)
			}
			if event.Body.Code != code {
				return fmt.Errorf("expected code %d, got %d", code, event.Body.Code)
			}
		}
	}

	return nil
}

func (tc *TestContext) webhookIncludes(table *godog.Table) error {
	event, err := tc.lastWebhookEvent()
	if err != nil {
		return err
	}

	values := tableToMap(table)
	for field, expected := range values {
		switch field {
		case "transactionId":
			if expected == "(a valid UUID)" {
				if _, err := uuid.Parse(event.Body.TransactionID); err != nil {
					return fmt.Errorf("invalid transactionId %s", event.Body.TransactionID)
				}
				continue
			}
			if event.Body.TransactionID != expected {
				return fmt.Errorf("expected transactionId %s, got %s", expected, event.Body.TransactionID)
			}
		case "transactionReference":
			if expected == "(the deposit reference)" {
				// Set expected to the last deposit reference
				if tc.lastDepositReference == "" {
					// If no explicit reference was set, use the default pattern
					// which is the deposit reference from the sub-account beneficiaries
					for _, ben := range tc.lastSubAccount.Beneficiaries {
						if ben.CurrencyID == event.Body.CurrencyCode {
							expected = ben.DepositReference
							break
						}
					}
				} else {
					expected = tc.lastDepositReference
				}
			}
			if event.Body.TransactionReference != expected {
				return fmt.Errorf("expected transactionReference %s, got %s", expected, event.Body.TransactionReference)
			}
		}
	}

	return nil
}

func (tc *TestContext) webhookIncludesValidHeaders(table *godog.Table) error {
	event, err := tc.lastWebhookEvent()
	if err != nil {
		return err
	}

	values := tableToMap(table)
	timestamp := event.Headers.Get("x-gatehub-timestamp")
	for field, expected := range values {
		switch field {
		case "x-gatehub-app-id":
			if event.Headers.Get("x-gatehub-app-id") != expected {
				return fmt.Errorf("expected x-gatehub-app-id %s, got %s", expected, event.Headers.Get("x-gatehub-app-id"))
			}
		case "x-gatehub-timestamp":
			if expected == "(current timestamp)" {
				if timestamp == "" {
					return fmt.Errorf("missing x-gatehub-timestamp header")
				}
				if _, err := strconv.ParseInt(timestamp, 10, 64); err != nil {
					return fmt.Errorf("invalid x-gatehub-timestamp %s", timestamp)
				}
			}
		case "x-gatehub-signature":
			if expected == "(valid HMAC-SHA256)" {
				if err := tc.validateWebhookSignature(event, timestamp); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (tc *TestContext) webhookBodyIncludesRequiredFields(table *godog.Table) error {
	event, err := tc.lastWebhookEvent()
	if err != nil {
		return err
	}

	values := tableToMap(table)
	for field := range values {
		switch field {
		case "accountId":
			if event.Body.AccountID == "" {
				return fmt.Errorf("missing accountId")
			}
		case "amount":
			if event.Body.Amount == 0 {
				return fmt.Errorf("missing amount")
			}
		case "currencyCode":
			if event.Body.CurrencyCode == "" {
				return fmt.Errorf("missing currencyCode")
			}
		case "transactionId":
			if _, err := uuid.Parse(event.Body.TransactionID); err != nil {
				return fmt.Errorf("invalid transactionId")
			}
		case "status":
			if event.Body.Status == "" {
				return fmt.Errorf("missing status")
			}
		case "code":
			if event.Body.Code == 0 {
				return fmt.Errorf("missing code")
			}
		case "createdAt":
			if event.Body.CreatedAt == "" {
				return fmt.Errorf("missing createdAt")
			}
		case "settledAt":
			if event.Body.SettledAt == "" {
				return fmt.Errorf("missing settledAt")
			}
		}
	}

	return nil
}

func (tc *TestContext) subAccountZARBalanceIs(expected string) error {
	fmt.Printf("DEBUG: Checking ZAR balance for account %s (wallet: %s)\n", tc.lastSubAccount.AccountID, tc.lastSubAccount.WalletID)
	if err := tc.requestBalanceForSubAccount(); err != nil {
		return err
	}
	fmt.Printf("DEBUG: Balance response contains %d balances\n", len(tc.lastBalanceResponse.Balances))
	for _, bal := range tc.lastBalanceResponse.Balances {
		fmt.Printf("DEBUG: Balance - Currency: %s, Available: %f, Total: %f\n", bal.CurrencyCode, bal.Available, bal.Total)
	}
	return tc.totalBalanceIs("ZAR", expected)
}

func (tc *TestContext) subAccountUSDBalanceIs(expected string) error {
	if err := tc.requestBalanceForSubAccount(); err != nil {
		return err
	}
	return tc.totalBalanceIs("USD", expected)
}

func (tc *TestContext) subAccountStartsWithZeroZARBalance() error {
	if tc.lastSubAccount.AccountID == "" {
		return fmt.Errorf("no accountId available")
	}
	fmt.Printf("DEBUG: Setting zero ZAR balance for account %s (wallet: %s)\n", tc.lastSubAccount.AccountID, tc.lastSubAccount.WalletID)
	payload := map[string]interface{}{
		"accountId":    tc.lastSubAccount.AccountID,
		"currencyCode": "ZAR",
		"available":    0.0,
		"reserved":     0.0,
	}
	_, err := tc.request("POST", "/v1/test/balances/set", payload, true, nil)
	return err
}

func (tc *TestContext) listCompanyDepositsWithLimit(limit int) error {
	path := fmt.Sprintf("/v1/company/deposits?limit=%d", limit)
	_, err := tc.request("GET", path, nil, true, nil)
	return err
}

func (tc *TestContext) listCompanyDepositsWithLimitAndPage(limit, page int) error {
	path := fmt.Sprintf("/v1/company/deposits?limit=%d&page=%d", limit, page)
	_, err := tc.request("GET", path, nil, true, nil)
	return err
}

func (tc *TestContext) listCompanyDeposits() error {
	return tc.listCompanyDepositsWithLimit(10)
}

func (tc *TestContext) listCompanyDepositsWithoutAuth() error {
	_, err := tc.request("GET", "/v1/company/deposits?limit=10", nil, false, nil)
	return err
}

func (tc *TestContext) responseIncludesDeposits(count int) error {
	if tc.lastResponse == nil || tc.lastResponse.StatusCode >= 400 {
		return fmt.Errorf("failed to list deposits, status %d", tc.lastResponse.StatusCode)
	}

	var resp listCompanyDepositsResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}

	if len(resp.Data) != count {
		return fmt.Errorf("expected %d deposits but got %d", count, len(resp.Data))
	}
	return nil
}

func (tc *TestContext) responseIncludesDepositsOnPage(count int, page int) error {
	return tc.responseIncludesDeposits(count)
}

func (tc *TestContext) eachDepositIncludes(table *godog.Table) error {
	if tc.lastResponse == nil || tc.lastResponse.StatusCode >= 400 {
		return fmt.Errorf("failed to list deposits, status %d", tc.lastResponse.StatusCode)
	}

	var resp listCompanyDepositsResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}

	values := tableToMap(table)
	for i, item := range resp.Data {
		for field, expected := range values {
			switch field {
			case "transactionId":
				if expected == "(valid UUID)" {
					if _, err := uuid.Parse(item.TransactionID); err != nil {
						return fmt.Errorf("deposit %d has invalid transactionId", i)
					}
				}
			case "status":
				if item.Status != expected {
					return fmt.Errorf("deposit %d status mismatch: expected %s, got %s", i, expected, item.Status)
				}
			case "code":
				code, err := strconv.Atoi(expected)
				if err != nil {
					return fmt.Errorf("invalid code %s", expected)
				}
				if item.Code != code {
					return fmt.Errorf("deposit %d code mismatch: expected %d, got %d", i, code, item.Code)
				}
			}
		}
	}

	return nil
}

func (tc *TestContext) bothDepositsRecorded() error {
	// List deposits first, then verify count
	if err := tc.listCompanyDepositsWithLimit(10); err != nil {
		return err
	}
	return tc.responseIncludesDeposits(2)
}

func (tc *TestContext) eachDepositHasUniqueID() error {
	if tc.lastResponse == nil || tc.lastResponse.StatusCode >= 400 {
		return fmt.Errorf("failed to list deposits, status %d", tc.lastResponse.StatusCode)
	}

	var resp listCompanyDepositsResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}

	seen := make(map[string]struct{})
	for _, item := range resp.Data {
		if item.TransactionID == "" {
			return fmt.Errorf("missing transactionId")
		}
		if _, ok := seen[item.TransactionID]; ok {
			return fmt.Errorf("duplicate transactionId %s", item.TransactionID)
		}
		seen[item.TransactionID] = struct{}{}
	}
	return nil
}

func (tc *TestContext) eachDepositShowsCompleted() error {
	if tc.lastResponse == nil || tc.lastResponse.StatusCode >= 400 {
		return fmt.Errorf("failed to list deposits, status %d", tc.lastResponse.StatusCode)
	}

	var resp listCompanyDepositsResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}

	for i, item := range resp.Data {
		if item.Status != "completed" {
			return fmt.Errorf("deposit %d not completed: %s", i, item.Status)
		}
	}
	return nil
}

func (tc *TestContext) allDepositsAppearInList(count int) error {
	return tc.responseIncludesDeposits(count)
}

func (tc *TestContext) createTwoSubAccountsWithDepositReferences(table *godog.Table) error {
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		walletID := strings.TrimSpace(row.Cells[1].Value)
		if walletID == "" {
			continue
		}
		payload := buildSubAccountPayload(map[string]string{"walletId": walletID}, true, walletID)
		if err := tc.postSubAccount(payload, true); err != nil {
			return err
		}
		if tc.lastSubAccount.AccountID == "" {
			return fmt.Errorf("missing accountId for wallet %s", walletID)
		}
		tc.accountIDsByWallet[walletID] = tc.lastSubAccount.AccountID

		refs := make(map[string]string)
		for _, ben := range tc.lastSubAccount.Beneficiaries {
			refs[ben.CurrencyID] = ben.DepositReference
		}
		tc.depositRefsByWallet[walletID] = refs
	}
	return nil
}

func (tc *TestContext) simulateDepositForWalletReference(walletSuffix string) error {
	walletID := "wallet_" + walletSuffix
	accountID, ok := tc.accountIDsByWallet[walletID]
	if !ok {
		return fmt.Errorf("no accountId for wallet %s", walletID)
	}
	refs := tc.depositRefsByWallet[walletID]
	ref := refs["ZAR"]
	if ref == "" {
		return fmt.Errorf("no ZAR deposit reference for wallet %s", walletID)
	}
	return tc.doSimulateTestDeposit(accountID, "1000.00", "ZAR", ref, true)
}

func (tc *TestContext) walletBalanceCredited(walletSuffix string) error {
	walletID := "wallet_" + walletSuffix
	return tc.balanceForWalletIs(walletID, "1000.00")
}

func (tc *TestContext) walletBalanceUnchanged(walletSuffix string) error {
	walletID := "wallet_" + walletSuffix
	return tc.balanceForWalletIs(walletID, "0.00")
}

func (tc *TestContext) attemptDepositInvalidAccount(accountID string, table *godog.Table) error {
	values := tableToMap(table)
	amount := values["amount"]
	currency := values["currencyCode"]
	return tc.doSimulateTestDeposit(accountID, amount, currency, "", true)
}

func (tc *TestContext) attemptDepositWithAmount(amount string) error {
	accountID := tc.lastSubAccount.AccountID
	return tc.doSimulateTestDeposit(accountID, amount, "ZAR", "", true)
}

func (tc *TestContext) attemptDepositWithoutAuth() error {
	accountID := tc.lastSubAccount.AccountID
	return tc.doSimulateTestDeposit(accountID, "1000.00", "ZAR", "", false)
}

func (tc *TestContext) doSimulateTestDeposit(accountID, amountStr, currency, depositReference string, auth bool) error {
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %s", amountStr)
	}

	payload := map[string]interface{}{
		"accountId":        accountID,
		"amount":           amount,
		"currencyCode":     currency,
		"depositReference": depositReference,
	}

	_, err = tc.request("POST", "/v1/company/accounts/testdeposit", payload, auth, nil)
	if err != nil {
		return err
	}

	if tc.lastResponse != nil && tc.lastResponse.StatusCode < 400 {
		var resp depositResponse
		if err := tc.decodeLastResponse(&resp); err != nil {
			return err
		}
		tc.lastDepositResponse = resp
		if resp.TransactionID != "" {
			tc.createdDeposits = append(tc.createdDeposits, resp.TransactionID)
		}
	}

	return nil
}

func (tc *TestContext) depositReferenceForCurrency(currency string) (string, error) {
	for _, ben := range tc.lastSubAccount.Beneficiaries {
		if strings.EqualFold(ben.CurrencyID, currency) {
			return ben.DepositReference, nil
		}
	}
	return "", fmt.Errorf("no deposit reference for currency %s", currency)
}

func (tc *TestContext) validateWebhookSignature(event webhookEvent, timestamp string) error {
	if timestamp == "" {
		return fmt.Errorf("missing x-gatehub-timestamp header")
	}
	secret := defaultWebhookSecret

	// Match the format used by the handler: timestamp|method|url|body
	message := fmt.Sprintf("%s|POST|%s|%s", timestamp, tc.webhookURL, string(event.RawBody))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))
	actual := event.Headers.Get("x-gatehub-signature")
	if actual == "" {
		return fmt.Errorf("missing x-gatehub-signature header")
	}
	if !hmac.Equal([]byte(expected), []byte(actual)) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

func (tc *TestContext) allDepositsHaveCompleted() error {
	// Wait for all deposits to process (use webhook count as proxy)
	expectedCount := len(tc.createdDeposits)
	if expectedCount == 0 {
		return fmt.Errorf("no deposits created")
	}
	// Wait for `expectedCount` new webhooks beyond current total
	globalWebhookEventsMu.Lock()
	baseline := len(globalWebhookEvents)
	globalWebhookEventsMu.Unlock()

	target := baseline + expectedCount
	return tc.waitForWebhookCount(target, 10*time.Second)
}

func (tc *TestContext) accountBalanceIsCreditedTwice() error {
	// Check that balance reflects two deposits
	if err := tc.requestBalanceForSubAccount(); err != nil {
		return err
	}
	return tc.totalBalanceIs("ZAR", "2000.00")
}

func (tc *TestContext) responseIncludesTransactionID() error {
	if tc.lastDepositResponse.TransactionID == "" {
		return fmt.Errorf("no transactionID in response")
	}
	if _, err := uuid.Parse(tc.lastDepositResponse.TransactionID); err != nil {
		return fmt.Errorf("invalid transactionID format: %s", tc.lastDepositResponse.TransactionID)
	}
	return nil
}

// clearServerDeposits calls the test reset endpoint to clear all deposit records on the server.
func (tc *TestContext) clearServerDeposits() error {
	_, err := tc.request("POST", "/v1/test/deposits/clear", nil, true, nil)
	return err
}
