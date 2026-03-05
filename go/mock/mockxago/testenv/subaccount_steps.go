//go:build e2e
// +build e2e

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func buildSubAccountPayload(values map[string]string, fillDefaults bool, walletIDOverride string) map[string]string {
	payload := map[string]string{}

	walletID := values["walletId"]
	if walletID == "" {
		walletID = values["wallet_id"]
	}
	if walletIDOverride != "" {
		walletID = walletIDOverride
	}
	if walletID == "" && fillDefaults {
		walletID = fmt.Sprintf("wallet_%d", time.Now().UnixNano())
	}
	if walletID != "" {
		payload["walletId"] = walletID
	}

	setField := func(key, defaultValue string) {
		value := values[key]
		if value == "" && fillDefaults {
			value = defaultValue
		}
		if value != "" {
			payload[key] = value
		}
	}

	setField("firstName", "John")
	setField("lastName", "Doe")
	setField("email", "john@example.com")
	setField("mobileNumber", "+27123456789")
	setField("identityType", "individual")
	setField("idNumber", "9001011234567")
	setField("physicalAddress", "123 Main St")
	setField("thirdPartyVerificationUrl", "https://example.com/verify")

	return payload
}

func (tc *TestContext) createSubAccountWithDetails(table *godog.Table) error {
	payload := buildSubAccountPayload(tableToMap(table), true, "")
	return tc.postSubAccount(payload, true)
}

func (tc *TestContext) createSubAccountWithOnlyRequiredFields(table *godog.Table) error {
	payload := buildSubAccountPayload(tableToMap(table), true, "")
	return tc.postSubAccount(payload, true)
}

func (tc *TestContext) attemptCreateSubAccountWithoutFirstName(table *godog.Table) error {
	payload := buildSubAccountPayload(tableToMap(table), false, "")
	delete(payload, "firstName")
	return tc.postSubAccount(payload, true)
}

func (tc *TestContext) attemptCreateSubAccountWithoutLastName(table *godog.Table) error {
	payload := buildSubAccountPayload(tableToMap(table), false, "")
	delete(payload, "lastName")
	return tc.postSubAccount(payload, true)
}

func (tc *TestContext) attemptCreateSubAccountWithoutEmail(table *godog.Table) error {
	payload := buildSubAccountPayload(tableToMap(table), false, "")
	delete(payload, "email")
	return tc.postSubAccount(payload, true)
}

func (tc *TestContext) attemptCreateSubAccountWithoutToken(table *godog.Table) error {
	payload := buildSubAccountPayload(tableToMap(table), true, "")
	return tc.postSubAccount(payload, false)
}

func (tc *TestContext) attemptCreateSubAccountWithInvalidToken(table *godog.Table) error {
	payload := buildSubAccountPayload(tableToMap(table), true, "")
	return tc.postSubAccount(payload, true)
}

func (tc *TestContext) postSubAccount(payload map[string]string, auth bool) error {
	if tc.lastSubAccount.AccountID != "" {
		tc.previousAccountID = tc.lastSubAccount.AccountID
	}
	_, err := tc.request("POST", "/v1/company/accounts", payload, auth, nil)
	if err != nil {
		return err
	}
	if tc.lastResponse != nil && tc.lastResponse.StatusCode >= 400 {
		return nil
	}
	var resp createSubAccountResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	if walletID := payload["walletId"]; walletID != "" {
		tc.lastWalletID = walletID
		resp.WalletID = walletID // Store wallet ID in response for later use
	}
	if email := payload["email"]; email != "" {
		tc.lastEmail = email
	}
	tc.lastSubAccount = resp
	if tc.lastWalletID != "" {
		tc.subAccountsByWallet[tc.lastWalletID] = resp
	}
	return nil
}

func (tc *TestContext) subAccountIsCreatedWith(table *godog.Table) error {
	if tc.lastSubAccount.AccountID == "" {
		return fmt.Errorf("accountId missing in response")
	}
	if err := parseUUID(tc.lastSubAccount.AccountID); err != nil {
		return fmt.Errorf("accountId is not a valid UUID: %v", err)
	}
	if tc.lastSubAccount.DepositAddress == "" {
		return fmt.Errorf("depositAddress missing in response")
	}
	if tc.lastSubAccount.DepositTag <= 0 {
		return fmt.Errorf("depositTag missing in response")
	}
	return nil
}

func (tc *TestContext) responseIncludesBankDetailsForZAR() error {
	if len(tc.lastSubAccount.BankDepositDetails["ZAR"]) == 0 {
		return fmt.Errorf("missing ZAR bankDepositDetails")
	}
	return nil
}

func (tc *TestContext) responseIncludesBankDetailsForUSD() error {
	if len(tc.lastSubAccount.BankDepositDetails["USD"]) == 0 {
		return fmt.Errorf("missing USD bankDepositDetails")
	}
	return nil
}

func (tc *TestContext) responseIncludesBeneficiariesWithDepositRefs() error {
	if len(tc.lastSubAccount.Beneficiaries) == 0 {
		return fmt.Errorf("missing beneficiaries in response")
	}
	for _, ben := range tc.lastSubAccount.Beneficiaries {
		if strings.TrimSpace(ben.DepositReference) == "" {
			return fmt.Errorf("missing depositReference in beneficiaries")
		}
	}
	return nil
}

func (tc *TestContext) newSubAccountIsCreated() error {
	if tc.lastSubAccount.AccountID == "" {
		return fmt.Errorf("missing accountId")
	}
	if tc.previousAccountID != "" && tc.lastSubAccount.AccountID == tc.previousAccountID {
		return fmt.Errorf("expected a new accountId, got same as previous")
	}
	return nil
}

func (tc *TestContext) subAccountHasProvidedEmail() error {
	if tc.lastEmail == "" {
		return fmt.Errorf("no email captured for sub-account request")
	}
	return nil
}

func (tc *TestContext) beneficiariesInclude(table *godog.Table) error {
	values := tableToMap(table)
	wantType := values["beneficiaryType"]
	if wantType == "" {
		wantType = "rollup"
	}
	found := false
	for _, ben := range tc.lastSubAccount.Beneficiaries {
		if ben.BeneficiaryType != wantType {
			continue
		}
		if tc.lastWalletID != "" && !strings.Contains(ben.DepositReference, tc.lastWalletID) {
			continue
		}
		if !strings.Contains(ben.DepositReference, ben.CurrencyID) {
			continue
		}
		found = true
		break
	}
	if !found {
		return fmt.Errorf("expected beneficiary with type %s and deposit reference containing wallet and currency", wantType)
	}
	return nil
}

func (tc *TestContext) depositReferencesAreUnique() error {
	var zarRef, usdRef string
	for _, ben := range tc.lastSubAccount.Beneficiaries {
		if ben.CurrencyID == "ZAR" {
			zarRef = ben.DepositReference
		}
		if ben.CurrencyID == "USD" {
			usdRef = ben.DepositReference
		}
	}
	if zarRef == "" || usdRef == "" {
		return fmt.Errorf("missing deposit references for ZAR or USD")
	}
	if zarRef == usdRef {
		return fmt.Errorf("expected unique deposit references per currency")
	}
	return nil
}

func (tc *TestContext) createSubAccountForWallet(walletID string) error {
	payload := buildSubAccountPayload(map[string]string{}, true, walletID)
	return tc.postSubAccount(payload, true)
}

func (tc *TestContext) updateSubAccountWithDetails(table *godog.Table) error {
	if tc.lastSubAccount.AccountID == "" {
		return fmt.Errorf("no accountId available for update")
	}
	payload := tableToMap(table)
	_, err := tc.request("PUT", "/v1/company/accounts/"+tc.lastSubAccount.AccountID, payload, true, nil)
	return err
}

func (tc *TestContext) subAccountUpdatedWithVerificationURL() error {
	var resp struct {
		AccountID string `json:"accountId"`
		Status    string `json:"status"`
	}
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	if resp.Status != "updated" {
		return fmt.Errorf("expected status updated, got %s", resp.Status)
	}
	return nil
}

func (tc *TestContext) responseContainsUpdatedStatus() error {
	return tc.subAccountUpdatedWithVerificationURL()
}

func (tc *TestContext) attemptUpdateSubAccountInvalidID(accountID string) error {
	payload := map[string]string{"thirdPartyVerificationUrl": "https://example.com/updated"}
	_, err := tc.request("PUT", "/v1/company/accounts/"+accountID, payload, true, nil)
	return err
}

func (tc *TestContext) createTwoSubAccountsDifferentWallets(table *godog.Table) error {
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		walletID := strings.TrimSpace(row.Cells[1].Value)
		if walletID == "" {
			continue
		}
		if err := tc.createSubAccountForWallet(walletID); err != nil {
			return err
		}
	}
	return nil
}

func (tc *TestContext) retrieveSubAccountInfoForWallet(walletID string) error {
	_, err := tc.request("GET", "/v1/company/accounts?walletId="+walletID, nil, true, nil)
	if err != nil {
		return err
	}
	tc.lastWalletID = walletID
	var resp createSubAccountResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	tc.lastSubAccount = resp
	return nil
}

func (tc *TestContext) correctSubAccountAssociated(walletID string) error {
	expected, ok := tc.subAccountsByWallet[walletID]
	if !ok {
		return fmt.Errorf("no sub-account stored for wallet %s", walletID)
	}
	if tc.lastSubAccount.AccountID != expected.AccountID {
		return fmt.Errorf("expected accountId %s, got %s", expected.AccountID, tc.lastSubAccount.AccountID)
	}
	return nil
}

func (tc *TestContext) subAccountIsolationConfirmed() error {
	for walletID, account := range tc.subAccountsByWallet {
		if tc.lastSubAccount.AccountID == account.AccountID {
			if tc.lastWalletID != "" && tc.lastWalletID != walletID {
				return fmt.Errorf("received account from different wallet")
			}
		}
	}
	return nil
}

func (tc *TestContext) createSubAccount() error {
	payload := buildSubAccountPayload(map[string]string{}, true, "")
	return tc.postSubAccount(payload, true)
}

func (tc *TestContext) createSubAccountForWalletID(walletID string) error {
	return tc.createSubAccountForWallet(walletID)
}

func (tc *TestContext) retrieveCreatedSubAccountDetails() error {
	if tc.lastWalletID == "" {
		return fmt.Errorf("no walletId available for retrieval")
	}
	return tc.retrieveSubAccountInfoForWallet(tc.lastWalletID)
}

func (tc *TestContext) retrieveSubAccountDetails() error {
	return tc.retrieveCreatedSubAccountDetails()
}

func (tc *TestContext) createTwoSubAccounts(table *godog.Table) error {
	return tc.createTwoSubAccountsDifferentWallets(table)
}

func (tc *TestContext) retrieveBothSubAccounts() error {
	if len(tc.subAccountsByWallet) < 2 {
		return fmt.Errorf("expected at least two sub-accounts")
	}
	return nil
}

func (tc *TestContext) depositReferencesAreDifferent() error {
	walletIDs := make([]string, 0, len(tc.subAccountsByWallet))
	for walletID := range tc.subAccountsByWallet {
		walletIDs = append(walletIDs, walletID)
	}
	if len(walletIDs) < 2 {
		return fmt.Errorf("expected two wallets")
	}

	refA := tc.depositReferenceForWalletCurrency(walletIDs[0], "ZAR")
	refB := tc.depositReferenceForWalletCurrency(walletIDs[1], "ZAR")
	if refA == "" || refB == "" {
		return fmt.Errorf("missing deposit references to compare")
	}
	if refA == refB {
		return fmt.Errorf("expected different deposit references for wallets")
	}
	return nil
}

func (tc *TestContext) depositReferenceWalletAAAUnique() error {
	return tc.depositReferenceUnique("wallet_aaa")
}

func (tc *TestContext) depositReferenceWalletBBBUnique() error {
	return tc.depositReferenceUnique("wallet_bbb")
}

func (tc *TestContext) depositReferenceUnique(walletID string) error {
	ref := tc.depositReferenceForWalletCurrency(walletID, "ZAR")
	if ref == "" {
		return fmt.Errorf("missing deposit reference for %s", walletID)
	}
	for otherWallet := range tc.subAccountsByWallet {
		if otherWallet == walletID {
			continue
		}
		if ref == tc.depositReferenceForWalletCurrency(otherWallet, "ZAR") {
			return fmt.Errorf("expected unique deposit reference for %s", walletID)
		}
	}
	return nil
}

func (tc *TestContext) depositReferenceForWalletCurrency(walletID, currency string) string {
	account, ok := tc.subAccountsByWallet[walletID]
	if !ok {
		return ""
	}
	for _, ben := range account.Beneficiaries {
		if ben.CurrencyID == currency {
			return ben.DepositReference
		}
	}
	return ""
}
