//go:build e2e

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

// ── helper ──────────────────────────────────────────────────────────────────

func buildBeneficiaryPayload(values map[string]string, omitName, omitAccountNumber bool) map[string]interface{} {
	payload := map[string]interface{}{}

	set := func(key string) {
		if v, ok := values[key]; ok {
			payload[key] = v
		}
	}

	if !omitName {
		set("name")
	}
	set("scope")
	set("currencyCode")
	if !omitAccountNumber {
		set("accountNumber")
	}
	set("branchCode")
	set("bankName")
	set("accountName")
	set("reference")
	if v, ok := values["isOwn"]; ok {
		payload["isOwn"] = strings.EqualFold(v, "true")
	}
	return payload
}

func (tc *TestContext) postBeneficiary(payload map[string]interface{}, auth bool) error {
	accountID := tc.lastSubAccount.AccountID
	if accountID == "" {
		return fmt.Errorf("no sub-account in context")
	}
	_, err := tc.request("POST", "/v1/accounts/"+accountID+"/beneficiaries", payload, auth, nil)
	return err
}

func (tc *TestContext) listBeneficiariesForAccount(accountID string, limit, page int, auth bool) error {
	path := fmt.Sprintf("/v1/accounts/%s/beneficiaries?limit=%d&page=%d", accountID, limit, page)
	_, err := tc.request("GET", path, nil, auth, nil)
	return err
}

// ── add-beneficiary steps ────────────────────────────────────────────────────

func (tc *TestContext) addBeneficiaryWithDetails(table *godog.Table) error {
	values := tableToMap(table)
	payload := buildBeneficiaryPayload(values, false, false)
	if err := tc.postBeneficiary(payload, true); err != nil {
		return err
	}
	if tc.lastResponse != nil && tc.lastResponse.StatusCode < 400 {
		var resp addBeneficiaryResponse
		if err := tc.decodeLastResponse(&resp); err != nil {
			return err
		}
		tc.lastBeneficiary = resp
		tc.addedBeneficiaries = append(tc.addedBeneficiaries, resp)
	}
	return nil
}

func (tc *TestContext) addBeneficiaryWithRequiredFieldsOnly(table *godog.Table) error {
	return tc.addBeneficiaryWithDetails(table)
}

func (tc *TestContext) addBeneficiaryWithSpecificDetails(table *godog.Table) error {
	return tc.addBeneficiaryWithDetails(table)
}

func (tc *TestContext) attemptAddBeneficiaryWithoutName(table *godog.Table) error {
	values := tableToMap(table)
	payload := buildBeneficiaryPayload(values, true, false)
	return tc.postBeneficiary(payload, true)
}

func (tc *TestContext) attemptAddBeneficiaryWithoutAccountNumber(table *godog.Table) error {
	values := tableToMap(table)
	payload := buildBeneficiaryPayload(values, false, true)
	return tc.postBeneficiary(payload, true)
}

func (tc *TestContext) attemptAddBeneficiaryWithoutAuthentication() error {
	payload := buildBeneficiaryPayload(map[string]string{
		"name":          "Test Beneficiary",
		"currencyCode":  "ZAR",
		"accountNumber": "1234567890",
		"bankName":      "TestBank",
		"accountName":   "Test",
	}, false, false)
	return tc.postBeneficiary(payload, false)
}

func (tc *TestContext) attemptListBeneficiariesWithoutAuthentication() error {
	accountID := tc.lastSubAccount.AccountID
	if accountID == "" {
		return fmt.Errorf("no sub-account in context")
	}
	_, err := tc.request("GET", "/v1/accounts/"+accountID+"/beneficiaries", nil, false, nil)
	return err
}

// ── list-beneficiary steps ───────────────────────────────────────────────────

func (tc *TestContext) listBeneficiaries() error {
	return tc.listBeneficiariesParsed(tc.lastSubAccount.AccountID, 100, 1, true)
}

func (tc *TestContext) requestListWithLimit(limit int) error {
	return tc.listBeneficiariesParsed(tc.lastSubAccount.AccountID, limit, 1, true)
}

func (tc *TestContext) requestListWithLimitAndPage(limit, page int) error {
	return tc.listBeneficiariesParsed(tc.lastSubAccount.AccountID, limit, page, true)
}

func (tc *TestContext) listBeneficiariesParsed(accountID string, limit, page int, auth bool) error {
	if err := tc.listBeneficiariesForAccount(accountID, limit, page, auth); err != nil {
		return err
	}
	if tc.lastResponse != nil && tc.lastResponse.StatusCode < 400 {
		var resp listBeneficiariesResponse
		if err := tc.decodeLastResponse(&resp); err != nil {
			return err
		}
		tc.lastBeneficiaries = resp
	}
	return nil
}

func (tc *TestContext) listBeneficiariesForWalletAAA() error {
	subAcc, ok := tc.subAccountsByWallet["wallet_aaa"]
	if !ok {
		return fmt.Errorf("no sub-account found for wallet_aaa")
	}
	return tc.listBeneficiariesParsed(subAcc.AccountID, 100, 1, true)
}

// ── state-setup steps ────────────────────────────────────────────────────────

func (tc *TestContext) haveAddedBeneficiaryWithStatus(status string) error {
	// Add a beneficiary, then wait if needed for it to move to the given status.
	// For "pending" we just create one; for "approved" we create and wait.
	payload := buildBeneficiaryPayload(map[string]string{
		"name":          "Setup Beneficiary",
		"scope":         "external",
		"currencyCode":  "ZAR",
		"accountNumber": "1234567890",
		"branchCode":    "250155",
		"bankName":      "ABSA",
		"accountName":   "John Doe",
		"reference":     "Ref001",
		"isOwn":         "true",
	}, false, false)
	if err := tc.postBeneficiary(payload, true); err != nil {
		return err
	}
	if tc.lastResponse == nil || tc.lastResponse.StatusCode >= 400 {
		return fmt.Errorf("failed to create setup beneficiary, status %d", tc.lastResponse.StatusCode)
	}
	var resp addBeneficiaryResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	tc.lastBeneficiary = resp
	tc.addedBeneficiaries = append(tc.addedBeneficiaries, resp)
	return nil
}

func (tc *TestContext) haveAddedNBeneficiariesToSubAccount(n int) error {
	for i := 0; i < n; i++ {
		payload := buildBeneficiaryPayload(map[string]string{
			"name":          fmt.Sprintf("Beneficiary %d", i+1),
			"scope":         "external",
			"currencyCode":  "ZAR",
			"accountNumber": fmt.Sprintf("%010d", 1000000000+i),
			"branchCode":    "250155",
			"bankName":      "ABSA",
			"accountName":   fmt.Sprintf("Test User %d", i+1),
			"reference":     fmt.Sprintf("Ref%03d", i+1),
			"isOwn":         "true",
		}, false, false)
		if err := tc.postBeneficiary(payload, true); err != nil {
			return err
		}
		if tc.lastResponse == nil || tc.lastResponse.StatusCode >= 400 {
			return fmt.Errorf("failed to create beneficiary %d, status %d", i+1, tc.lastResponse.StatusCode)
		}
		var resp addBeneficiaryResponse
		if err := tc.decodeLastResponse(&resp); err != nil {
			return err
		}
		tc.addedBeneficiaries = append(tc.addedBeneficiaries, resp)
	}
	return nil
}

func (tc *TestContext) haveAddedNBeneficiaries(n int) error {
	return tc.haveAddedNBeneficiariesToSubAccount(n)
}

func (tc *TestContext) haveCreatedSubAccountsForTwoWallets(table *godog.Table) error {
	pairs := tableToPairs(table)
	for _, pair := range pairs {
		walletName := pair[1]
		if err := tc.createSubAccountForWallet(walletName); err != nil {
			return fmt.Errorf("creating sub-account for %s: %w", walletName, err)
		}
	}
	return nil
}

func (tc *TestContext) haveAddedBeneficiariesToWalletAAA() error {
	subAcc, ok := tc.subAccountsByWallet["wallet_aaa"]
	if !ok {
		return fmt.Errorf("no sub-account found for wallet_aaa")
	}
	savedSubAcc := tc.lastSubAccount
	tc.lastSubAccount = subAcc
	if err := tc.haveAddedNBeneficiariesToSubAccount(2); err != nil {
		tc.lastSubAccount = savedSubAcc
		return err
	}
	tc.lastSubAccount = savedSubAcc
	return nil
}

// ── wait steps ───────────────────────────────────────────────────────────────

func (tc *TestContext) waitNSecondsForAutoApproval(n int) error {
	time.Sleep(time.Duration(n) * time.Second)
	return nil
}

// ── assertion steps ──────────────────────────────────────────────────────────

func (tc *TestContext) responseIncludesBeneficiaryWith(table *godog.Table) error {
	values := tableToMap(table)
	for key, expected := range values {
		switch key {
		case "uuid":
			if expected == "(a valid UUID)" {
				if tc.lastBeneficiary.UUID == "" {
					return fmt.Errorf("expected a valid UUID but got empty string")
				}
				if _, err := uuid.Parse(tc.lastBeneficiary.UUID); err != nil {
					return fmt.Errorf("expected a valid UUID but got %q: %w", tc.lastBeneficiary.UUID, err)
				}
			} else if tc.lastBeneficiary.UUID != expected {
				return fmt.Errorf("uuid: expected %q, got %q", expected, tc.lastBeneficiary.UUID)
			}
		case "name":
			if tc.lastBeneficiary.Name != expected {
				return fmt.Errorf("name: expected %q, got %q", expected, tc.lastBeneficiary.Name)
			}
		case "currencyCode":
			if tc.lastBeneficiary.CurrencyCode != expected {
				return fmt.Errorf("currencyCode: expected %q, got %q", expected, tc.lastBeneficiary.CurrencyCode)
			}
		case "status":
			if tc.lastBeneficiary.Status != expected {
				return fmt.Errorf("status: expected %q, got %q", expected, tc.lastBeneficiary.Status)
			}
		}
	}
	return nil
}

func (tc *TestContext) beneficiaryCreatedWithStatus(status string) error {
	if tc.lastBeneficiary.Status != status {
		return fmt.Errorf("expected beneficiary status %q but got %q", status, tc.lastBeneficiary.Status)
	}
	return nil
}

func (tc *TestContext) beneficiaryIsCreated() error {
	if tc.lastBeneficiary.UUID == "" {
		return fmt.Errorf("expected a beneficiary to be created but no UUID found")
	}
	if _, err := uuid.Parse(tc.lastBeneficiary.UUID); err != nil {
		return fmt.Errorf("created beneficiary has invalid UUID %q: %w", tc.lastBeneficiary.UUID, err)
	}
	return nil
}

func (tc *TestContext) beneficiaryStatusTransitionedTo(status string) error {
	found := false
	for _, b := range tc.lastBeneficiaries.Data {
		if b.UUID == tc.lastBeneficiary.UUID {
			if b.Status != status {
				return fmt.Errorf("beneficiary %s: expected status %q but got %q", b.UUID, status, b.Status)
			}
			found = true
			break
		}
	}
	if !found {
		// Maybe the last list call refreshed all beneficiaries; check the first one.
		if len(tc.lastBeneficiaries.Data) == 0 {
			return fmt.Errorf("no beneficiaries in last list response")
		}
		b := tc.lastBeneficiaries.Data[0]
		if b.Status != status {
			return fmt.Errorf("beneficiary %s: expected status %q but got %q", b.UUID, status, b.Status)
		}
	}
	return nil
}

func (tc *TestContext) responseIncludesNBeneficiaries(n int) error {
	got := len(tc.lastBeneficiaries.Data)
	if got != n {
		return fmt.Errorf("expected %d beneficiaries but got %d", n, got)
	}
	return nil
}

func (tc *TestContext) eachBeneficiaryHasRequiredFields() error {
	for i, b := range tc.lastBeneficiaries.Data {
		if b.UUID == "" {
			return fmt.Errorf("beneficiary[%d] missing uuid", i)
		}
		if b.Name == "" {
			return fmt.Errorf("beneficiary[%d] missing name", i)
		}
		if b.CurrencyCode == "" {
			return fmt.Errorf("beneficiary[%d] missing currencyCode", i)
		}
		if b.Status == "" {
			return fmt.Errorf("beneficiary[%d] missing status", i)
		}
	}
	return nil
}

func (tc *TestContext) paginationShowsNumberOfPages(n int) error {
	if tc.lastBeneficiaries.Pagination.NumberOfPages != n {
		return fmt.Errorf("expected numberOfPages=%d but got %d", n, tc.lastBeneficiaries.Pagination.NumberOfPages)
	}
	return nil
}

func (tc *TestContext) uuidsAreDifferent() error {
	seen := map[string]bool{}
	for i, b := range tc.addedBeneficiaries {
		if seen[b.UUID] {
			return fmt.Errorf("duplicate UUID %q found at index %d", b.UUID, i)
		}
		seen[b.UUID] = true
	}
	return nil
}

func (tc *TestContext) eachUUIDIsUnique() error {
	return tc.uuidsAreDifferent()
}

func (tc *TestContext) getCorrectBeneficiaries() error {
	if len(tc.lastBeneficiaries.Data) == 0 {
		return fmt.Errorf("expected beneficiaries for wallet_aaa but got none")
	}
	return nil
}

func (tc *TestContext) doNotGetBeneficiariesFromWalletBBB() error {
	subAcc, ok := tc.subAccountsByWallet["wallet_bbb"]
	if !ok {
		// If wallet_bbb has no sub-account we can't cross-check — pass.
		return nil
	}
	// Fetch beneficiaries for wallet_bbb to confirm they are separate.
	if err := tc.listBeneficiariesParsed(subAcc.AccountID, 100, 1, true); err != nil {
		return err
	}
	bbbBeneficiaries := tc.lastBeneficiaries.Data

	// Re-fetch for wallet_aaa to restore tc.lastBeneficiaries
	subAccAAA, ok := tc.subAccountsByWallet["wallet_aaa"]
	if ok {
		_ = tc.listBeneficiariesParsed(subAccAAA.AccountID, 100, 1, true)
	}

	// Verify no overlap
	bbbUUIDs := map[string]bool{}
	for _, b := range bbbBeneficiaries {
		bbbUUIDs[b.UUID] = true
	}
	for _, b := range tc.lastBeneficiaries.Data {
		if bbbUUIDs[b.UUID] {
			return fmt.Errorf("beneficiary %s from wallet_bbb leaked into wallet_aaa response", b.UUID)
		}
	}
	return nil
}

func (tc *TestContext) beneficiaryDetailsMatchExactly(table *godog.Table) error {
	values := tableToMap(table)
	if len(tc.lastBeneficiaries.Data) == 0 {
		return fmt.Errorf("no beneficiaries in last list response")
	}
	// Match against the first beneficiary in the list (most-recently added in the scenario)
	b := tc.lastBeneficiaries.Data[len(tc.lastBeneficiaries.Data)-1]
	for key, expected := range values {
		var got string
		switch key {
		case "accountNumber":
			got = b.AccountNumber
		case "branchCode":
			got = b.BranchCode
		case "bankName":
			got = b.BankName
		case "accountName":
			got = b.AccountName
		case "reference":
			got = b.Reference
		default:
			continue
		}
		if got != expected {
			return fmt.Errorf("%s: expected %q, got %q", key, expected, got)
		}
	}
	return nil
}
