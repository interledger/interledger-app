package main

import (
	"fmt"
	"net/http"
	"time"
)

func newHarness() *harness {
	return &harness{client: &http.Client{Timeout: 10 * time.Second}}
}

func (h *harness) login() error {
	payload := map[string]interface{}{
		"policyId": "mock-policy",
		"fields": []map[string]string{
			{"fieldName": "publicKey", "fieldValue": testPublicKey},
			{"fieldName": "secret", "fieldValue": testSecret},
		},
	}
	var resp loginResponse
	if err := h.postJSON("/xago/v1/login", payload, false, &resp); err != nil {
		return err
	}
	if resp.TokenValue == "" {
		return fmt.Errorf("tokenValue missing in login response")
	}
	h.token = resp.TokenValue
	return nil
}

func (h *harness) createSubAccount(walletID string) (*createSubAccountResponse, error) {
	req := createSubAccountRequest{
		WalletID:                  walletID,
		FirstName:                 "Jane",
		LastName:                  "Doe",
		Email:                     fmt.Sprintf("%s@example.com", walletID),
		MobileNumber:              "1234567890",
		IdentityType:              "passport",
		IDNumber:                  "A1234567",
		PhysicalAddress:           "123 Test Street",
		ThirdPartyVerificationURL: "https://example.com/verify",
	}
	var resp createSubAccountResponse
	if err := h.postJSON("/xago/v1/company/accounts", req, true, &resp); err != nil {
		return nil, err
	}
	if resp.AccountID == "" {
		return nil, fmt.Errorf("accountId missing in response")
	}
	return &resp, nil
}

func (h *harness) getSubAccountByWallet(walletID string) (*createSubAccountResponse, error) {
	var resp createSubAccountResponse
	if err := h.getJSON(fmt.Sprintf("/xago/v1/company/accounts?walletId=%s", walletID), true, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (h *harness) updateSubAccount(accountID string) error {
	payload := map[string]string{
		"thirdPartyVerificationUrl": "https://example.com/updated",
		"idNumber":                  "B7654321",
		"physicalAddress":           "456 Updated Ave",
	}
	return h.putJSON(fmt.Sprintf("/xago/v1/company/accounts/%s", accountID), payload, true, &struct{}{})
}

func (h *harness) getBalance(accountID string) (*balanceResponse, error) {
	var resp balanceResponse
	if err := h.getJSON(fmt.Sprintf("/xago/v1/accounts/%s/balance", accountID), true, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (h *harness) listCurrencies() ([]currencyResponse, error) {
	var resp []currencyResponse
	if err := h.getJSON("/xago/v1/currencies", false, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (h *harness) expectBalanceCurrency(balances []balanceItem, code string) (balanceItem, error) {
	for _, b := range balances {
		if b.CurrencyCode == code {
			return b, nil
		}
	}
	return balanceItem{}, fmt.Errorf("currency %s not found in balance", code)
}
