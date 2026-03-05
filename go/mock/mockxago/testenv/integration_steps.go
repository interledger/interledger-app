//go:build e2e
// +build e2e

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

// walletBackendConfigured is a no-op background step.
// In this testenv all scenarios communicate directly with mockxago, so
// there is no separate backend to configure.
func (tc *TestContext) walletBackendConfigured() error {
	return nil
}

// submitKYCForWallet posts a KYC form to /kyc/submit using the given walletID
// and name, simulating what the browser KYC iframe does after the user fills
// in their personal details.
func (tc *TestContext) submitKYCForWallet(walletID, firstName, lastName string) error {
	formData := url.Values{}
	formData.Set("user_id", walletID)
	formData.Set("first_name", firstName)
	formData.Set("last_name", lastName)
	formData.Set("address", "123 Test Street")
	formData.Set("dob", "1990-01-01")

	req, err := http.NewRequest(
		"POST",
		tc.baseURL+"/kyc/submit",
		bytes.NewBufferString(formData.Encode()),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = tc.doRequest(req)
	return err
}

// kycSubmissionIsAccepted verifies that the last response has HTTP 200 and
// a JSON body with {"status":"accepted",...}.
func (tc *TestContext) kycSubmissionIsAccepted() error {
	if err := tc.responseStatusIs(200); err != nil {
		return err
	}
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	if resp.Status != "accepted" {
		return fmt.Errorf("expected KYC status %q, got %q", "accepted", resp.Status)
	}
	return nil
}

// subAccountForWalletIsRetrievable confirms that GET /v1/company/accounts?walletId=X
// returns HTTP 200, i.e. the sub-account exists after KYC submission.
func (tc *TestContext) subAccountForWalletIsRetrievable(walletID string) error {
	_, err := tc.request("GET", "/v1/company/accounts?walletId="+walletID, nil, true, nil)
	if err != nil {
		return err
	}
	return tc.responseStatusIs(200)
}

// createFollowingTestTransactions reads a 2-column table (amount | currency) and
// calls POST /v1/test/transactions for each row so the transactions appear in the
// GET /v1/company/transactions history.
func (tc *TestContext) createFollowingTestTransactions(table *godog.Table) error {
	if tc.lastSubAccount.AccountID == "" {
		return fmt.Errorf("no sub-account available; create one first")
	}
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		amountStr := strings.TrimSpace(row.Cells[0].Value)
		currency := strings.TrimSpace(row.Cells[1].Value)

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return fmt.Errorf("invalid amount %q: %w", amountStr, err)
		}

		payload := map[string]interface{}{
			"transactionId": uuid.NewString(),
			"accountId":     tc.lastSubAccount.AccountID,
			"amount":        amount,
			"currencyCode":  currency,
		}
		if _, err := tc.request("POST", "/v1/test/transactions", payload, true, nil); err != nil {
			return fmt.Errorf("create transaction %s %s: %w", amountStr, currency, err)
		}
		if err := tc.responseStatusIs(200); err != nil {
			return fmt.Errorf("create transaction %s %s: %w", amountStr, currency, err)
		}
	}
	return nil
}

// transactionHistoryContainsAtLeast fetches GET /v1/company/transactions and
// asserts the returned data array length is at least n.
func (tc *TestContext) transactionHistoryContainsAtLeast(n int) error {
	_, err := tc.request("GET", "/v1/company/transactions", nil, true, nil)
	if err != nil {
		return err
	}
	if err := tc.responseStatusIs(200); err != nil {
		return err
	}

	var resp struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := tc.decodeLastResponse(&resp); err != nil {
		return err
	}
	if len(resp.Data) < n {
		return fmt.Errorf("expected at least %d transactions, got %d", n, len(resp.Data))
	}
	return nil
}
