package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

func (tc *TestContext) requestCurrencyList() error {
	_, err := tc.request("GET", "/v1/currencies", nil, false, nil)
	if err != nil {
		return err
	}
	var resp []currencyResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	tc.lastCurrencies = resp
	return nil
}

func (tc *TestContext) requestCurrencyListAgain() error {
	return tc.requestCurrencyList()
}

func (tc *TestContext) requestCurrencyListWithoutAuth() error {
	_, err := tc.request("GET", "/v1/currencies", nil, false, nil)
	if err != nil {
		return err
	}
	var resp []currencyResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	tc.lastCurrencies = resp
	return nil
}

func (tc *TestContext) responseIncludesAtLeastCurrencies(count int, table *godog.Table) error {
	if len(tc.lastCurrencies) < count {
		return fmt.Errorf("expected at least %d currencies, got %d", count, len(tc.lastCurrencies))
	}
	pairs := tableToPairs(table)
	for _, pair := range pairs {
		code := pair[0]
		name := pair[1]
		if !tc.currencyExists(code, name) {
			return fmt.Errorf("expected currency %s (%s) in response", code, name)
		}
	}
	return nil
}

func (tc *TestContext) zarCurrencyIncludes(table *godog.Table) error {
	return tc.currencyIncludes("ZAR", table)
}

func (tc *TestContext) usdCurrencyIncludes(table *godog.Table) error {
	return tc.currencyIncludes("USD", table)
}

func (tc *TestContext) storeCurrencyList() error {
	_, err := tc.request("GET", "/v1/currencies", nil, false, nil)
	if err != nil {
		return err
	}
	if tc.lastResponseBody == nil {
		return fmt.Errorf("no response body")
	}
	tc.previousCurrencies = string(tc.lastResponseBody)
	var resp []currencyResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	tc.lastCurrencies = resp
	return nil
}

func (tc *TestContext) responseIdenticalToPrevious() error {
	if tc.previousCurrencies == "" {
		return fmt.Errorf("no previous currency response stored")
	}
	if string(tc.lastResponseBody) != tc.previousCurrencies {
		return fmt.Errorf("currency response changed between calls")
	}
	return nil
}

func (tc *TestContext) accountNumbersAndBankCodesSame() error {
	if tc.previousCurrencies == "" {
		return fmt.Errorf("no previous currency response stored")
	}
	var prev []currencyResponse
	if err := json.Unmarshal([]byte(tc.previousCurrencies), &prev); err != nil {
		return err
	}
	if len(prev) != len(tc.lastCurrencies) {
		return fmt.Errorf("currency list lengths differ between calls")
	}
	for _, curr := range prev {
		current := tc.findCurrency(curr.CurrencyID)
		if current == nil {
			return fmt.Errorf("currency %s missing in latest response", curr.CurrencyID)
		}
		if current.AccountNumber != curr.AccountNumber || current.BranchCode != curr.BranchCode {
			return fmt.Errorf("currency %s account details changed", curr.CurrencyID)
		}
	}
	return nil
}

func (tc *TestContext) responseIncludesAvailableCurrencies() error {
	if len(tc.lastCurrencies) == 0 {
		return fmt.Errorf("no currencies returned")
	}
	return nil
}

func (tc *TestContext) bankDetailsMatchCurrenciesEndpoint() error {
	if len(tc.lastCurrencies) == 0 {
		if err := tc.requestCurrencyList(); err != nil {
			return err
		}
	}
	return nil
}

func (tc *TestContext) zarBankDetailsMatchExactly() error {
	return tc.bankDetailsMatch("ZAR")
}

func (tc *TestContext) usdBankDetailsMatchExactly() error {
	return tc.bankDetailsMatch("USD")
}

func (tc *TestContext) bankDetailsMatch(currency string) error {
	bankDetails, ok := tc.lastSubAccount.BankDepositDetails[currency]
	if !ok || len(bankDetails) == 0 {
		return fmt.Errorf("missing bankDepositDetails for %s", currency)
	}
	curr := tc.findCurrency(currency)
	if curr == nil {
		return fmt.Errorf("currency %s not found in currencies response", currency)
	}
	if bankDetails[0].BankName != curr.BankName || bankDetails[0].AccountName != curr.AccountName || bankDetails[0].AccountNumber != curr.AccountNumber || bankDetails[0].BranchCode != curr.BranchCode || bankDetails[0].SwiftBIC != curr.SwiftBIC {
		return fmt.Errorf("bank details for %s do not match currencies endpoint", currency)
	}
	return nil
}

func (tc *TestContext) zarDepositReferenceContains(value string) error {
	return tc.depositReferenceContains("ZAR", value)
}

func (tc *TestContext) usdDepositReferenceContains(value string) error {
	return tc.depositReferenceContains("USD", value)
}

func (tc *TestContext) depositReferenceContains(currency, value string) error {
	for _, ben := range tc.lastSubAccount.Beneficiaries {
		if ben.CurrencyID == currency {
			if !strings.Contains(ben.DepositReference, value) {
				return fmt.Errorf("expected %s deposit reference to contain %s", currency, value)
			}
			return nil
		}
	}
	return fmt.Errorf("no beneficiary for currency %s", currency)
}

func (tc *TestContext) currencyExists(code, name string) bool {
	for _, curr := range tc.lastCurrencies {
		if curr.CurrencyID == code && curr.CurrencyName == name {
			return true
		}
	}
	return false
}

func (tc *TestContext) currencyIncludes(code string, table *godog.Table) error {
	curr := tc.findCurrency(code)
	if curr == nil {
		return fmt.Errorf("currency %s not found", code)
	}

	values := tableToMap(table)
	for key, expected := range values {
		switch key {
		case "currencyId":
			if curr.CurrencyID != expected {
				return fmt.Errorf("expected currencyId %s, got %s", expected, curr.CurrencyID)
			}
		case "bankName":
			if curr.BankName != expected {
				return fmt.Errorf("expected bankName %s, got %s", expected, curr.BankName)
			}
		case "accountNumber":
			if strings.Contains(expected, "valid") {
				if !regexp.MustCompile(`^[0-9]+$`).MatchString(curr.AccountNumber) {
					return fmt.Errorf("expected numeric accountNumber, got %s", curr.AccountNumber)
				}
			} else if curr.AccountNumber != expected {
				return fmt.Errorf("expected accountNumber %s, got %s", expected, curr.AccountNumber)
			}
		case "branchCode":
			if curr.BranchCode != expected {
				return fmt.Errorf("expected branchCode %s, got %s", expected, curr.BranchCode)
			}
		case "swiftBIC":
			if curr.SwiftBIC != expected {
				return fmt.Errorf("expected swiftBIC %s, got %s", expected, curr.SwiftBIC)
			}
		}
	}
	return nil
}

func (tc *TestContext) findCurrency(code string) *currencyResponse {
	for i, curr := range tc.lastCurrencies {
		if curr.CurrencyID == code {
			return &tc.lastCurrencies[i]
		}
	}
	return nil
}
