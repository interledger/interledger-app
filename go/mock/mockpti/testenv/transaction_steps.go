//go:build e2e
// +build e2e

package main

import "fmt"

// anExistingPTIUserWithUSDWalletAndBankAccount ensures a user, wallet, and
// bank account payment information exist before a transaction scenario runs.
func (tc *TestContext) anExistingPTIUserWithUSDWalletAndBankAccount() error {
	if err := tc.anExistingPTIUser(); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	if err := tc.postWithUSDWalletPayload("/users/{userId}/wallets"); err != nil {
		return fmt.Errorf("create wallet: %w", err)
	}
	if err := tc.responseStatusShouldBe(200); err != nil {
		return err
	}
	if err := tc.responseShouldIncludeWalletID(); err != nil {
		return err
	}
	if err := tc.postWithBankAccountPayload("/users/{userId}/payment-information"); err != nil {
		return fmt.Errorf("create payment information: %w", err)
	}
	if err := tc.responseStatusShouldBe(200); err != nil {
		return err
	}
	return tc.responseShouldIncludePaymentInformationID()
}

// twoPTIUsersEachWithUSDWallet creates two users each with a wallet for transfer scenarios.
func (tc *TestContext) twoPTIUsersEachWithUSDWallet() error {
	// Create first user + wallet (stored in tc.lastUserID and tc.lastWalletID).
	if err := tc.anExistingPTIUser(); err != nil {
		return fmt.Errorf("create first user: %w", err)
	}
	if err := tc.postWithUSDWalletPayload("/users/{userId}/wallets"); err != nil {
		return fmt.Errorf("create first wallet: %w", err)
	}
	if err := tc.responseStatusShouldBe(200); err != nil {
		return err
	}
	return tc.responseShouldIncludeWalletID()
}

// postWithValidDepositPayload creates a deposit transaction.
func (tc *TestContext) postWithValidDepositPayload(path string) error {
	payload := map[string]interface{}{
		"initiator": map[string]interface{}{
			"id":   tc.lastUserID,
			"type": "PERSON",
		},
		"sourceMethod": map[string]interface{}{
			"currency":          "USD",
			"paymentMethodType": "ACH",
			"paymentInformation": map[string]interface{}{
				"type": "BANK_ACCOUNT",
				"id":   tc.lastPaymentInformationID,
			},
		},
		"destinationMethod": map[string]interface{}{
			"paymentMethodType": "WALLET",
			"paymentInformation": map[string]interface{}{
				"type": "WALLET",
				"id":   tc.lastWalletID,
			},
		},
		"amount":   100.00,
		"usdValue": 100.00,
		"type":     "DEPOSIT",
	}
	_, err := tc.ptiRequest("POST", path, payload, true)
	return err
}

// postWithValidWithdrawalPayload creates a withdrawal transaction.
func (tc *TestContext) postWithValidWithdrawalPayload(path string) error {
	payload := map[string]interface{}{
		"initiator": map[string]interface{}{
			"id":   tc.lastUserID,
			"type": "PERSON",
		},
		"sourceMethod": map[string]interface{}{
			"paymentMethodType": "WALLET",
			"paymentInformation": map[string]interface{}{
				"type": "WALLET",
				"id":   tc.lastWalletID,
			},
		},
		"destinationMethod": map[string]interface{}{
			"currency":          "USD",
			"paymentMethodType": "ACH",
			"paymentInformation": map[string]interface{}{
				"type": "BANK_ACCOUNT",
				"id":   tc.lastPaymentInformationID,
			},
		},
		"amount":   50.00,
		"usdValue": 50.00,
		"type":     "WITHDRAWAL",
	}
	_, err := tc.ptiRequest("POST", path, payload, true)
	return err
}

// postWithValidTransferPayload creates a transfer transaction.
func (tc *TestContext) postWithValidTransferPayload(path string) error {
	payload := map[string]interface{}{
		"initiator": map[string]interface{}{
			"id":   tc.lastUserID,
			"type": "PERSON",
		},
		"sourceTransferMethod": map[string]interface{}{
			"paymentMethodType": "WALLET",
			"paymentInformation": map[string]interface{}{
				"type": "WALLET",
				"id":   tc.lastWalletID,
			},
		},
		"destinationTransferMethod": map[string]interface{}{
			"paymentMethodType": "WALLET",
			"paymentInformation": map[string]interface{}{
				"type": "WALLET",
				"id":   "dest-wallet-id",
			},
		},
		"amount":   75.00,
		"usdValue": 75.00,
		"type":     "TRANSFER",
	}
	_, err := tc.ptiRequest("POST", path, payload, true)
	return err
}

// responseShouldIncludeTransactionRequestID verifies the response includes a
// transaction request id and stores it for subsequent steps.
func (tc *TestContext) responseShouldIncludeTransactionRequestID() error {
	var resp idResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return fmt.Errorf("failed to decode transaction response: %w", err)
	}
	if resp.ID == "" {
		return fmt.Errorf("response does not include a transaction request id. Body: %s", string(tc.lastResponseBody))
	}
	tc.lastTransactionRequestID = resp.ID
	return nil
}

// anExistingPTITransactionRequestID creates a deposit and stores its request id.
func (tc *TestContext) anExistingPTITransactionRequestID() error {
	if err := tc.anExistingPTIUser(); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	payload := map[string]interface{}{
		"initiator": map[string]interface{}{
			"id":   tc.lastUserID,
			"type": "PERSON",
		},
		"sourceMethod": map[string]interface{}{
			"currency":          "USD",
			"paymentMethodType": "ACH",
			"paymentInformation": map[string]interface{}{
				"type": "BANK_ACCOUNT",
				"id":   "pi-seed",
			},
		},
		"destinationMethod": map[string]interface{}{
			"paymentMethodType": "WALLET",
			"paymentInformation": map[string]interface{}{
				"type": "WALLET",
				"id":   "w-seed",
			},
		},
		"amount":   10.00,
		"usdValue": 10.00,
		"type":     "DEPOSIT",
	}
	if _, err := tc.ptiRequest("POST", "/transactions/deposits", payload, true); err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}
	if err := tc.responseStatusShouldBe(200); err != nil {
		return err
	}
	return tc.responseShouldIncludeTransactionRequestID()
}

// postWithFeedbackPayload sends update feedback to a transaction.
func (tc *TestContext) postWithFeedbackPayload(path string) error {
	payload := map[string]interface{}{
		"transactionId": tc.lastTransactionRequestID,
		"feedback":      "SETTLED",
		"providerName":  "test-provider",
		"payload":       `{"status":"SETTLED"}`,
	}
	_, err := tc.ptiRequest("POST", path, payload, true)
	return err
}

// responseShouldIncludeUpdateID verifies the response includes an update id.
func (tc *TestContext) responseShouldIncludeUpdateID() error {
	var resp idResponse
	if err := tc.decodeLastResponse(&resp); err != nil {
		return fmt.Errorf("failed to decode update response: %w", err)
	}
	if resp.ID == "" {
		return fmt.Errorf("response does not include an update id. Body: %s", string(tc.lastResponseBody))
	}
	tc.lastUpdateID = resp.ID
	return nil
}
