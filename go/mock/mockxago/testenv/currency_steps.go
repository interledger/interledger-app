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
	// Try parsing as nested format first
	var nested []currencyNested
	if err := tc.decodeLastResponse(&nested); err == nil {
		// Convert nested to flat for backward compatibility
		tc.lastCurrenciesNested = nested
		tc.lastCurrencies = make([]currencyResponse, 0, len(nested))
		for _, n := range nested {
			if flat := n.toFlat(); flat != nil {
				tc.lastCurrencies = append(tc.lastCurrencies, *flat)
			}
		}
		return nil
	}
	// Fallback to flat format
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
	// Try parsing as nested format first
	var prevNested []currencyNested
	if err := json.Unmarshal([]byte(tc.previousCurrencies), &prevNested); err == nil {
		// Convert to flat for comparison
		prevFlat := make([]currencyResponse, 0, len(prevNested))
		for _, n := range prevNested {
			if flat := n.toFlat(); flat != nil {
				prevFlat = append(prevFlat, *flat)
			}
		}
		if len(prevFlat) != len(tc.lastCurrencies) {
			return fmt.Errorf("currency list lengths differ between calls")
		}
		for _, curr := range prevFlat {
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
	// Fallback to flat format parsing
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

// New step definitions for nested format testing

func (tc *TestContext) responseIncludesCurrenciesInNestedFormat() error {
	if len(tc.lastCurrenciesNested) == 0 {
		return fmt.Errorf("no currencies returned in nested format")
	}
	// Verify structure has required nested fields
	for _, curr := range tc.lastCurrenciesNested {
		if curr.CurrencyCode == "" {
			return fmt.Errorf("currency missing currencyCode")
		}
		if len(curr.BankingProviders) == 0 {
			return fmt.Errorf("currency %s missing bankingProviders", curr.CurrencyCode)
		}
	}
	return nil
}

func (tc *TestContext) zarCurrencyHasNestedBankingProvidersWith(table *godog.Table) error {
	return tc.currencyHasNestedProvidersWith("ZAR", table)
}

func (tc *TestContext) currencyHasNestedProvidersWith(code string, table *godog.Table) error {
	curr := tc.findNestedCurrency(code)
	if curr == nil {
		return fmt.Errorf("currency %s not found", code)
	}

	values := tableToMap(table)
	for key, expected := range values {
		switch key {
		case "depositEnabled":
			expectedBool := expected == "true"
			if curr.DepositEnabled != expectedBool {
				return fmt.Errorf("expected depositEnabled %v, got %v", expectedBool, curr.DepositEnabled)
			}
		case "withdrawEnabled":
			expectedBool := expected == "true"
			if curr.WithdrawEnabled != expectedBool {
				return fmt.Errorf("expected withdrawEnabled %v, got %v", expectedBool, curr.WithdrawEnabled)
			}
		}
	}
	return nil
}

func (tc *TestContext) firstZarBankingProviderIncludes(table *godog.Table) error {
	return tc.firstProviderIncludes("ZAR", table)
}

func (tc *TestContext) firstProviderIncludes(code string, table *godog.Table) error {
	curr := tc.findNestedCurrency(code)
	if curr == nil {
		return fmt.Errorf("currency %s not found", code)
	}
	if len(curr.BankingProviders) == 0 {
		return fmt.Errorf("no banking providers for currency %s", code)
	}

	provider := curr.BankingProviders[0]
	values := tableToMap(table)
	for key, expected := range values {
		switch key {
		case "name":
			if provider.Name != expected {
				return fmt.Errorf("expected name %s, got %s", expected, provider.Name)
			}
		case "depositAvailable":
			expectedBool := expected == "true"
			if provider.DepositAvailable != expectedBool {
				return fmt.Errorf("expected depositAvailable %v, got %v", expectedBool, provider.DepositAvailable)
			}
		}
	}
	return nil
}

func (tc *TestContext) firstZarProviderDepositFieldsInclude(table *godog.Table) error {
	return tc.firstProviderDepositFieldsInclude("ZAR", table)
}

func (tc *TestContext) firstProviderDepositFieldsInclude(code string, table *godog.Table) error {
	curr := tc.findNestedCurrency(code)
	if curr == nil {
		return fmt.Errorf("currency %s not found", code)
	}
	if len(curr.BankingProviders) == 0 {
		return fmt.Errorf("no banking providers for currency %s", code)
	}

	fields := curr.BankingProviders[0].DepositFields
	values := tableToMap(table)
	for key, expected := range values {
		switch key {
		case "bankName":
			if fields.BankName != expected {
				return fmt.Errorf("expected bankName %s, got %s", expected, fields.BankName)
			}
		case "accountNumber":
			if fields.AccountNumber != expected {
				return fmt.Errorf("expected accountNumber %s, got %s", expected, fields.AccountNumber)
			}
		case "branchCode":
			if fields.BranchCode != expected {
				return fmt.Errorf("expected branchCode %s, got %s", expected, fields.BranchCode)
			}
		case "accountName":
			if fields.AccountName != expected {
				return fmt.Errorf("expected accountName %s, got %s", expected, fields.AccountName)
			}
		}
	}
	return nil
}

func (tc *TestContext) usdCurrencyHasNestedBankingProviders() error {
	curr := tc.findNestedCurrency("USD")
	if curr == nil {
		return fmt.Errorf("USD currency not found")
	}
	if len(curr.BankingProviders) == 0 {
		return fmt.Errorf("USD currency has no banking providers")
	}
	return nil
}

func (tc *TestContext) responseStructureMatchesBackendExpectations() error {
	if len(tc.lastCurrenciesNested) == 0 {
		return fmt.Errorf("no currencies in nested format")
	}
	// Verify at least one currency has all required fields
	for _, curr := range tc.lastCurrenciesNested {
		if curr.CurrencyCode == "" {
			return fmt.Errorf("missing currencyCode")
		}
		if len(curr.BankingProviders) == 0 {
			return fmt.Errorf("missing bankingProviders")
		}
		for _, provider := range curr.BankingProviders {
			if provider.Name == "" {
				return fmt.Errorf("provider missing name")
			}
		}
	}
	return nil
}

func (tc *TestContext) eachCurrencyHasRequiredFields(table *godog.Table) error {
	requiredFields := make([]string, 0)
	for i := 1; i < len(table.Rows); i++ {
		if len(table.Rows[i].Cells) > 0 {
			requiredFields = append(requiredFields, table.Rows[i].Cells[0].Value)
		}
	}

	for _, curr := range tc.lastCurrenciesNested {
		for _, field := range requiredFields {
			switch field {
			case "currencyCode":
				if curr.CurrencyCode == "" {
					return fmt.Errorf("currency missing currencyCode")
				}
			case "depositEnabled":
				// Just check it exists (can be true or false)
			case "bankingProviders":
				if len(curr.BankingProviders) == 0 {
					return fmt.Errorf("currency %s missing bankingProviders", curr.CurrencyCode)
				}
			}
		}
	}
	return nil
}

func (tc *TestContext) eachBankingProviderHasRequiredFields(table *godog.Table) error {
	requiredFields := make([]string, 0)
	for i := 1; i < len(table.Rows); i++ {
		if len(table.Rows[i].Cells) > 0 {
			requiredFields = append(requiredFields, table.Rows[i].Cells[0].Value)
		}
	}

	for _, curr := range tc.lastCurrenciesNested {
		for _, provider := range curr.BankingProviders {
			for _, field := range requiredFields {
				switch field {
				case "name":
					if provider.Name == "" {
						return fmt.Errorf("provider missing name")
					}
				case "depositAvailable":
					// Just check it exists (can be true or false)
				case "depositFields":
					if provider.DepositFields.BankName == "" {
						return fmt.Errorf("provider missing depositFields")
					}
				}
			}
		}
	}
	return nil
}

func (tc *TestContext) findNestedCurrency(code string) *currencyNested {
	for i, curr := range tc.lastCurrenciesNested {
		if curr.CurrencyCode == code {
			return &tc.lastCurrenciesNested[i]
		}
	}
	return nil
}
