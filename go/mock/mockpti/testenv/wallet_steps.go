//go:build e2e
// +build e2e

package main

import "fmt"

// postWithUSDWalletPayload creates a USD wallet for the current user.
func (tc *TestContext) postWithUSDWalletPayload(path string) error {
	payload := map[string]interface{}{
		"currency":  "USD",
		"type":      "FIAT",
		"reference": "wallet-ref-1",
	}
	_, err := tc.ptiRequest("POST", path, payload, true)
	return err
}

// responseShouldIncludeWalletID stores the wallet id from the last response.
func (tc *TestContext) responseShouldIncludeWalletID() error {
	var resp walletResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if resp.WalletID == "" {
		return fmt.Errorf("response does not include wallet id. Body: %s", string(tc.lastResponseBody))
	}
	tc.lastWalletID = resp.WalletID
	return nil
}

// anExistingPTIUserWithWallet ensures a user exists and has at least one wallet.
func (tc *TestContext) anExistingPTIUserWithWallet() error {
	if err := tc.anExistingPTIUser(); err != nil {
		return err
	}
	if err := tc.postWithUSDWalletPayload("/users/{userId}/wallets"); err != nil {
		return err
	}
	if err := tc.responseStatusShouldBe(200); err != nil {
		return err
	}
	return tc.responseShouldIncludeWalletID()
}

// responseShouldIncludeAtLeastOneWallet validates list wallets response.
func (tc *TestContext) responseShouldIncludeAtLeastOneWallet() error {
	var wallets []walletResponse
	if err := tc.decodeLastResponse(&wallets); err != nil {
		return fmt.Errorf("failed to decode wallets list: %w", err)
	}
	if len(wallets) == 0 {
		return fmt.Errorf("expected at least one wallet, got none")
	}
	if wallets[0].WalletID != "" {
		tc.lastWalletID = wallets[0].WalletID
	}
	return nil
}

// postWithBankAccountPayload creates payment information for the current user.
func (tc *TestContext) postWithBankAccountPayload(path string) error {
	payload := map[string]interface{}{
		"type":                  "BANK_ACCOUNT",
		"bankAccountNumber":     "1234567890",
		"bankAccountType":       "CHECKING",
		"bankSwiftCode":         "BOFAUS3N",
		"bankRoutingNumber":     "021000021",
		"bankRoutingCheckDigit": "0",
		"accountBankName":       "Test Bank",
	}
	_, err := tc.ptiRequest("POST", path, payload, true)
	return err
}

// theWalletBalanceShouldBeNegative fetches the current wallet and asserts its balance is < 0.
func (tc *TestContext) theWalletBalanceShouldBeNegative() error {
	if _, err := tc.ptiRequest("GET", "/users/{userId}/wallets/{walletId}", nil, true); err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}
	if err := tc.responseStatusShouldBe(200); err != nil {
		return err
	}
	var wallet walletResponse
	if err := tc.decodeLastResponse(&wallet); err != nil {
		return fmt.Errorf("failed to decode wallet response: %w", err)
	}
	if wallet.Balance >= 0 {
		return fmt.Errorf("expected wallet balance to be negative, got %v", wallet.Balance)
	}
	return nil
}

// theWalletBalanceShouldEqualTheDepositedAmount checks that the wallet balance equals the deposit amount.
func (tc *TestContext) theWalletBalanceShouldEqualTheDepositedAmount() error {
	if _, err := tc.ptiRequest("GET", "/users/{userId}/wallets/{walletId}", nil, true); err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}
	if err := tc.responseStatusShouldBe(200); err != nil {
		return err
	}
	var wallet walletResponse
	if err := tc.decodeLastResponse(&wallet); err != nil {
		return fmt.Errorf("failed to decode wallet response: %w", err)
	}
	if wallet.Balance != tc.depositAmount {
		return fmt.Errorf("expected wallet balance %.2f (deposited amount), got %.2f", tc.depositAmount, wallet.Balance)
	}
	return nil
}

// responseShouldIncludePaymentInformationID stores payment information id.
func (tc *TestContext) responseShouldIncludePaymentInformationID() error {
	var resp paymentInformationResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if resp.ID == "" {
		return fmt.Errorf("response does not include payment information id. Body: %s", string(tc.lastResponseBody))
	}
	tc.lastPaymentInformationID = resp.ID
	return nil
}
