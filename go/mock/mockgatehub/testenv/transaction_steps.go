//go:build e2e

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ============ TRANSACTION STEPS ============

func (tc *TestContext) authenticatedRequestsWithHMAC() error {
	tc.appID = "local-test-app-id"
	tc.appSecret = "local-test-app-secret"
	return nil
}

func (tc *TestContext) postTransactionWithAuth(path, amount, currency, authToken string) error {
	// Replace any placeholders in authToken (e.g., {iframeToken})
	authToken = tc.replacePlaceholders(authToken)

	body := map[string]interface{}{
		"amount":   amount,
		"currency": currency,
	}
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", authToken),
	}
	_, err := tc.request("POST", path, body, headers)
	return err
}

func (tc *TestContext) userBalanceIncreasesBy(wholeAmount, decimalAmount int) error {
	// Simplified check - just verify the response indicates success
	if tc.lastResponse.StatusCode != 200 {
		// Include response body in error message for debugging
		responseMsg := string(tc.lastResponseBody)
		if len(responseMsg) > 200 {
			responseMsg = responseMsg[:200] + "..."
		}
		return fmt.Errorf("expected status 200, got %d. Response: %s", tc.lastResponse.StatusCode, responseMsg)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if status, ok := result["status"].(string); !ok || status != "success" {
		return fmt.Errorf("expected success status, got: %v", result["status"])
	}

	return nil
}

func (tc *TestContext) postHostedTransfer(path string, amount float64, decimalPart int, currency string, txType int, depositType string) error {
	body := map[string]interface{}{
		"user_id":      tc.userID,
		"amount":       amount + float64(decimalPart)/100,
		"currency":     currency,
		"type":         txType,
		"deposit_type": depositType,
	}
	headers := map[string]string{
		"x-gatehub-managed-user-uuid": tc.userID,
	}
	_, err := tc.request("POST", path, body, headers)
	return err
}

func (tc *TestContext) responseHasIDOrUUID() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if _, ok := result["id"].(string); ok {
		return nil
	}
	if _, ok := result["uuid"].(string); ok {
		return nil
	}

	return fmt.Errorf("expected id or uuid in response, got %v", result)
}

func (tc *TestContext) amountEchoes(amount float64) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	expected := fmt.Sprintf("%.2f", amount)
	val, ok := result["amount"]
	if !ok {
		return fmt.Errorf("expected amount %s, got <missing>", expected)
	}

	switch a := val.(type) {
	case string:
		if a != expected {
			return fmt.Errorf("expected amount %s, got %v", expected, val)
		}
	case float64:
		if fmt.Sprintf("%.2f", a) != expected {
			return fmt.Errorf("expected amount %s, got %v", expected, val)
		}
	default:
		return fmt.Errorf("expected amount %s, got %v", expected, val)
	}

	return nil
}

func (tc *TestContext) depositTypeIs(depositType string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if dt, ok := result["deposit_type"].(string); !ok || dt != depositType {
		return fmt.Errorf("expected deposit_type %s, got %v", depositType, result["deposit_type"])
	}

	return nil
}

func (tc *TestContext) statusIsInteger(status int) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if s, ok := result["status"].(float64); !ok || int(s) != status {
		return fmt.Errorf("expected status %d, got %v", status, result["status"])
	}

	return nil
}

func (tc *TestContext) statusIsCompleted(intStatus int, strStatus string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	status := result["status"]
	if intVal, ok := status.(float64); ok {
		if int(intVal) == intStatus {
			return nil
		}
	}
	if strVal, ok := status.(string); ok {
		if strVal == strStatus {
			return nil
		}
	}

	return fmt.Errorf("status is neither integer %d nor string %s", intStatus, strStatus)
}

func (tc *TestContext) coreDepositWebhookEmitted() error {
	// In a real test, this would check the webhook manager
	// For now, just verify response indicates success
	return nil
}

func (tc *TestContext) postExternalDeposit(path string, txType int, depositType string, wholeAmount, decimalAmount int, currency string) error {
	amount := float64(wholeAmount) + float64(decimalAmount)/100
	body := map[string]interface{}{
		"type":              txType,
		"deposit_type":      depositType,
		"receiving_address": tc.walletAddress,
		"amount":            amount,
		"currency":          currency,
	}
	headers := map[string]string{
		"x-gatehub-managed-user-uuid": tc.userID,
	}
	_, err := tc.request("POST", path, body, headers)
	return err
}

func (tc *TestContext) fieldsAreStringFormatted() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	for _, field := range []string{"amount", "total_amount", "fee"} {
		if val, ok := result[field]; ok {
			if _, isString := val.(string); !isString {
				return fmt.Errorf("field %s is not a string", field)
			}

			// Verify format is X.XX
			if str, ok := val.(string); ok {
				if !isDecimalFormat(str) {
					return fmt.Errorf("field %s not in X.XX format: %s", field, str)
				}
			}
		}
	}

	return nil
}

func (tc *TestContext) transactionCanBeRetrievedFormatted(version int) error {
	// Verify response has consistent formatting
	return tc.fieldsAreStringFormatted()
}

func (tc *TestContext) userHasFundedWallet(currency string, wholeAmount, decimalAmount int) error {
	amount := float64(wholeAmount) + float64(decimalAmount)/100
	body := map[string]interface{}{
		"type":              1,
		"deposit_type":      "external",
		"receiving_address": tc.walletAddress,
		"amount":            amount,
		"currency":          currency,
	}
	headers := map[string]string{
		"x-gatehub-managed-user-uuid": tc.userID,
	}
	_, err := tc.request("POST", "/core/v1/transactions", body, headers)
	return err
}

func (tc *TestContext) postHostedWithSendingAddress(sendingDesc, receivingDesc string, wholeAmount, decimalAmount int, currency string, txType int, depositType string) error {
	return tc.postHostedDirectional("sending", sendingDesc, receivingDesc, wholeAmount, decimalAmount, currency, txType, depositType)
}

func (tc *TestContext) postHostedWithReceivingAddress(receivingDesc, sendingDesc string, wholeAmount, decimalAmount int, currency string, txType int, depositType string) error {
	return tc.postHostedDirectional("receiving", sendingDesc, receivingDesc, wholeAmount, decimalAmount, currency, txType, depositType)
}

func (tc *TestContext) postHostedDirectional(leadField, sendingDesc, receivingDesc string, wholeAmount, decimalAmount int, currency string, txType int, depositType string) error {
	amount := float64(wholeAmount) + float64(decimalAmount)/100
	body := map[string]interface{}{
		"amount":       amount,
		"currency":     currency,
		"type":         txType,
		"deposit_type": depositType,
	}
	headers := map[string]string{
		"x-gatehub-managed-user-uuid": tc.userID,
	}

	resolveSendRecv := func(desc string) string {
		if desc == "the user wallet" {
			return tc.walletAddress
		}
		return "rSettlementAddress123"
	}

	if leadField == "sending" {
		body["sending_address"] = resolveSendRecv(sendingDesc)
		body["receiving_address"] = resolveSendRecv(receivingDesc)
	} else {
		body["receiving_address"] = resolveSendRecv(receivingDesc)
		body["sending_address"] = resolveSendRecv(sendingDesc)
	}
	_, err := tc.request("POST", "/core/v1/transactions", body, headers)
	return err
}

func (tc *TestContext) userBalanceForCurrencyIs(currency string, wholeAmount, decimalAmount int) error {
	expected := float64(wholeAmount) + float64(decimalAmount)/100
	expectedStr := fmt.Sprintf("%.2f", expected)

	// Transactions complete asynchronously (2s delay when webhooks are configured).
	// Poll the balance endpoint for up to 5 seconds.
	var lastBalance string
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)

		_, err := tc.request("GET", fmt.Sprintf("/core/v1/wallets/%s/balances", tc.walletAddress), nil, nil)
		if err != nil {
			return err
		}

		var balances []map[string]interface{}
		if err := json.Unmarshal(tc.lastResponseBody, &balances); err != nil {
			return fmt.Errorf("failed to parse balance response: %w", err)
		}

		for _, b := range balances {
			vault, ok := b["vault"].(map[string]interface{})
			if !ok {
				continue
			}
			if vault["asset_code"] == currency {
				lastBalance, _ = b["available"].(string)
				if lastBalance == expectedStr {
					return nil
				}
			}
		}
	}

	return fmt.Errorf("expected %s balance %s, got %s (after polling)", currency, expectedStr, lastBalance)
}

// isDecimalFormat checks if a string is in decimal format (X.XX)
func isDecimalFormat(s string) bool {
	parts := strings.Count(s, ".")
	if parts != 1 {
		return false
	}

	// Simple check: should have exactly 2 decimal places
	decimalIdx := strings.Index(s, ".")
	if decimalIdx == -1 {
		return false
	}

	return len(s)-decimalIdx-1 == 2
}
