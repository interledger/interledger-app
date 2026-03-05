//go:build e2e
// +build e2e

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

func (tc *TestContext) requestBalanceForSubAccount() error {
	if tc.lastSubAccount.AccountID == "" {
		return fmt.Errorf("no accountId available")
	}
	_, err := tc.request("GET", "/v1/accounts/"+tc.lastSubAccount.AccountID+"/balance", nil, true, nil)
	if err != nil {
		return err
	}
	var resp balanceResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	tc.lastBalanceResponse = resp
	return nil
}

func (tc *TestContext) setBalanceForSubAccount(table *godog.Table) error {
	values := tableToMap(table)
	currency := values["currency"]
	if currency == "" {
		currency = values["currencyCode"]
	}
	available := values["available"]
	if available == "" {
		available = values["amount"]
	}
	reserved := values["reserved"]

	if tc.lastSubAccount.AccountID == "" {
		return fmt.Errorf("no accountId available")
	}
	if strings.TrimSpace(currency) == "" {
		return fmt.Errorf("currency is required")
	}

	availableValue, err := strconv.ParseFloat(available, 64)
	if err != nil {
		return fmt.Errorf("invalid available amount %q", available)
	}
	reservedValue := 0.0
	if reserved != "" {
		reservedValue, err = strconv.ParseFloat(reserved, 64)
		if err != nil {
			return fmt.Errorf("invalid reserved amount %q", reserved)
		}
	}

	payload := map[string]interface{}{
		"accountId":    tc.lastSubAccount.AccountID,
		"currencyCode": currency,
		"available":    availableValue,
		"reserved":     reservedValue,
	}
	_, err = tc.request("POST", "/v1/test/balances/set", payload, true, nil)
	return err
}

func (tc *TestContext) subAccountStartsWithZeroBalance() error {
	if tc.lastSubAccount.AccountID == "" {
		return fmt.Errorf("no accountId available")
	}

	for _, currency := range []string{"ZAR", "USD"} {
		payload := map[string]interface{}{
			"accountId":    tc.lastSubAccount.AccountID,
			"currencyCode": currency,
			"available":    0.0,
			"reserved":     0.0,
		}
		if _, err := tc.request("POST", "/v1/test/balances/set", payload, true, nil); err != nil {
			return err
		}
	}
	return nil
}

func (tc *TestContext) depositReceivedAndProcessed(amount string, currency string) error {
	if tc.lastSubAccount.AccountID == "" {
		return fmt.Errorf("no accountId available")
	}
	amountValue, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("invalid deposit amount %q", amount)
	}

	payload := map[string]interface{}{
		"accountId":    tc.lastSubAccount.AccountID,
		"currencyCode": currency,
		"amount":       amountValue,
	}
	_, err = tc.request("POST", "/v1/test/balances/deposit", payload, true, nil)
	return err
}

func (tc *TestContext) transferInitiatedAndCompleted(amount string, currency string) error {
	if tc.lastSubAccount.AccountID == "" {
		return fmt.Errorf("no accountId available")
	}
	amountValue, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("invalid transfer amount %q", amount)
	}

	payload := map[string]interface{}{
		"accountId":    tc.lastSubAccount.AccountID,
		"currencyCode": currency,
		"amount":       amountValue,
	}
	_, err = tc.request("POST", "/v1/test/balances/transfer", payload, true, nil)
	return err
}

func (tc *TestContext) depositToWallet(amount string, currency string, walletID string) error {
	account, ok := tc.subAccountsByWallet[walletID]
	if !ok || account.AccountID == "" {
		return fmt.Errorf("no sub-account for wallet %s", walletID)
	}
	amountValue, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("invalid deposit amount %q", amount)
	}

	payload := map[string]interface{}{
		"accountId":    account.AccountID,
		"currencyCode": currency,
		"amount":       amountValue,
	}
	_, err = tc.request("POST", "/v1/test/balances/deposit", payload, true, nil)
	return err
}

func (tc *TestContext) requestBalanceForInvalidAccount(accountID string) error {
	_, err := tc.request("GET", "/v1/accounts/"+accountID+"/balance", nil, true, nil)
	return err
}

func (tc *TestContext) requestBalanceWithoutAuth() error {
	token := tc.token
	tc.token = ""
	_, err := tc.request("GET", "/v1/accounts/"+tc.lastSubAccount.AccountID+"/balance", nil, true, nil)
	tc.token = token
	return err
}

func (tc *TestContext) balanceResponseIncludes(table *godog.Table) error {
	values := tableToMap(table)
	if accountID, ok := values["accountId"]; ok {
		if accountID == "(the created accountId)" && tc.lastSubAccount.AccountID != "" {
			accountID = tc.lastSubAccount.AccountID
		}
		if tc.lastBalanceResponse.AccountID != accountID {
			return fmt.Errorf("expected accountId %s, got %s", accountID, tc.lastBalanceResponse.AccountID)
		}
	}
	return nil
}

func (tc *TestContext) balanceIncludesZAR(table *godog.Table) error {
	return tc.balanceIncludesCurrency("ZAR", table)
}

func (tc *TestContext) balanceIncludesUSD(table *godog.Table) error {
	return tc.balanceIncludesCurrency("USD", table)
}

func (tc *TestContext) balanceIncludesCurrency(currency string, table *godog.Table) error {
	item, ok := tc.findBalance(currency)
	if !ok {
		return fmt.Errorf("currency %s not found in balance response", currency)
	}
	values := tableToMap(table)
	for key, expected := range values {
		switch key {
		case "currencyCode":
			if item.CurrencyCode != expected {
				return fmt.Errorf("expected currencyCode %s, got %s", expected, item.CurrencyCode)
			}
		case "available":
			if !floatEquals(item.Available, expected) {
				return fmt.Errorf("expected available %s, got %.2f", expected, item.Available)
			}
		case "reserved":
			if !floatEquals(item.Reserved, expected) {
				return fmt.Errorf("expected reserved %s, got %.2f", expected, item.Reserved)
			}
		case "total":
			if !floatEquals(item.Total, expected) {
				return fmt.Errorf("expected total %s, got %.2f", expected, item.Total)
			}
		}
	}
	return nil
}

func (tc *TestContext) balanceResponseShows(table *godog.Table) error {
	values := tableToMap(table)
	currency := values["currencyCode"]
	if currency == "" {
		return fmt.Errorf("currencyCode is required")
	}
	return tc.balanceIncludesCurrency(currency, table)
}

func (tc *TestContext) availableBalanceIs(currency string, expected string) error {
	// Fetch balance if not already fetched
	if tc.lastBalanceResponse.AccountID == "" || len(tc.lastBalanceResponse.Balances) == 0 {
		if err := tc.requestBalanceForSubAccount(); err != nil {
			return err
		}
	}

	item, ok := tc.findBalance(currency)
	if !ok {
		return fmt.Errorf("currency %s not found in balance response", currency)
	}
	if !floatEquals(item.Available, expected) {
		return fmt.Errorf("expected available %s, got %.2f", expected, item.Available)
	}
	return nil
}

func (tc *TestContext) totalBalanceIs(currency string, expected string) error {
	// Fetch balance if not already fetched
	if tc.lastBalanceResponse.AccountID == "" || len(tc.lastBalanceResponse.Balances) == 0 {
		if err := tc.requestBalanceForSubAccount(); err != nil {
			return err
		}
	}

	item, ok := tc.findBalance(currency)
	if !ok {
		return fmt.Errorf("currency %s not found in balance response", currency)
	}
	if !floatEquals(item.Total, expected) {
		return fmt.Errorf("expected total %s, got %.2f", expected, item.Total)
	}
	return nil
}

func (tc *TestContext) balancesAreIndependent() error {
	// Fetch balance if not already fetched
	if tc.lastBalanceResponse.AccountID == "" || len(tc.lastBalanceResponse.Balances) == 0 {
		if err := tc.requestBalanceForSubAccount(); err != nil {
			return err
		}
	}

	_, zarOk := tc.findBalance("ZAR")
	_, usdOk := tc.findBalance("USD")
	if !zarOk || !usdOk {
		return fmt.Errorf("expected both ZAR and USD balances to exist")
	}
	return nil
}

func (tc *TestContext) balanceForWalletIs(walletID string, expected string) error {
	account, ok := tc.subAccountsByWallet[walletID]
	if !ok || account.AccountID == "" {
		return fmt.Errorf("no sub-account for wallet %s", walletID)
	}

	_, err := tc.request("GET", "/v1/accounts/"+account.AccountID+"/balance", nil, true, nil)
	if err != nil {
		return err
	}
	var resp balanceResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	for _, bal := range resp.Balances {
		if bal.CurrencyCode == "ZAR" {
			if !floatEquals(bal.Total, expected) {
				return fmt.Errorf("expected balance %s for %s, got %.2f", expected, walletID, bal.Total)
			}
			return nil
		}
	}
	return fmt.Errorf("ZAR balance not found for wallet %s", walletID)
}

func (tc *TestContext) findBalance(currency string) (balanceItem, bool) {
	for _, bal := range tc.lastBalanceResponse.Balances {
		if bal.CurrencyCode == currency {
			return bal, true
		}
	}
	return balanceItem{}, false
}

func floatEquals(value float64, expected string) bool {
	parsed, err := strconv.ParseFloat(expected, 64)
	if err != nil {
		return false
	}
	diff := value - parsed
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.001
}
