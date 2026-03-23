//go:build e2e

package main

import (
	"encoding/json"
	"fmt"
)

// ============ CARD STEPS ============

func (tc *TestContext) postCardCustomer(path, nameOnCard, accountCode, currency, cardCode string) error {
	body := map[string]interface{}{
		"nameOnCard": nameOnCard,
		"account": map[string]string{
			"productCode": accountCode,
			"currency":    currency,
		},
		"card": map[string]string{
			"productCode": cardCode,
		},
	}
	return tc.postCardCustomerWithBody(path, body)
}

func (tc *TestContext) postCardCustomerFromString(path, nameOnCard, accountCode, currency, cardCode string) error {
	return tc.postCardCustomer(path, nameOnCard, accountCode, currency, cardCode)
}

func (tc *TestContext) postCardCustomerWithBody(path string, body interface{}) error {
	headers := map[string]string{
		"x-gatehub-managed-user-uuid": tc.userID,
	}
	resp, err := tc.request("POST", path, body, headers)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(tc.lastResponseBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w, body was: %s", err, string(tc.lastResponseBody))
	}

	// Extract customer ID from nested structure
	if customers, ok := result["customers"].(map[string]interface{}); ok {
		if id, ok := customers["id"].(string); ok {
			tc.customerID = id
		}
	}

	tc.lastResponse = resp
	return nil
}

func (tc *TestContext) responseHasCustomer(customerType, kycStatus string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	customers, ok := result["customers"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid customers object")
	}

	if _, ok := customers["id"]; !ok {
		return fmt.Errorf("missing id in customer")
	}

	if t, ok := customers["type"].(string); !ok || t != customerType {
		return fmt.Errorf("expected type %s, got %v", customerType, customers["type"])
	}

	return nil
}

func (tc *TestContext) customerHasAccount(currency, status, accType string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	customers, ok := result["customers"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing customers object")
	}

	accounts, ok := customers["accounts"].([]interface{})
	if !ok || len(accounts) == 0 {
		return fmt.Errorf("no accounts in customer")
	}

	account, ok := accounts[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("account is not an object")
	}

	if c, ok := account["currency"].(string); !ok || c != currency {
		return fmt.Errorf("expected currency %s", currency)
	}

	return nil
}

func (tc *TestContext) accountHasCardHelper(status, nameOnCard string) error {
	return tc.accountHasCard(status, nameOnCard, true)
}

func (tc *TestContext) accountHasCard(status, nameOnCard string, inFuture interface{}) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	customers, ok := result["customers"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing customers object")
	}

	accounts, ok := customers["accounts"].([]interface{})
	if !ok || len(accounts) == 0 {
		return fmt.Errorf("no accounts in customer")
	}

	account, ok := accounts[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("account is not an object")
	}

	cards, ok := account["cards"].([]interface{})
	if !ok || len(cards) == 0 {
		return fmt.Errorf("no cards in account")
	}

	card, ok := cards[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("card is not an object")
	}

	if s, ok := card["status"].(string); !ok || s != status {
		return fmt.Errorf("expected status %s, got %v", status, card["status"])
	}

	if cardID, ok := card["id"].(string); ok {
		tc.cardID = cardID
	}

	return nil
}

func (tc *TestContext) getWithManagedUserHeader(path string) error {
	_, err := tc.sendWithManagedUserHeader("GET", path, nil)
	return err
}

func (tc *TestContext) postWithManagedUserHeader(path string) error {
	_, err := tc.sendWithManagedUserHeader("POST", path, nil)
	return err
}

func (tc *TestContext) deleteWithManagedUserHeader(path string) error {
	_, err := tc.sendWithManagedUserHeader("DELETE", path, nil)
	return err
}

func (tc *TestContext) responseHasCardData() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		return fmt.Errorf("no data array or empty")
	}

	if card, ok := data[0].(map[string]interface{}); ok {
		if cardID, ok := card["id"].(string); ok {
			tc.cardID = cardID
		}
	}

	return nil
}

func (tc *TestContext) eachCardHasFields() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		return fmt.Errorf("no data array")
	}

	requiredFields := []string{"id", "accountId", "customerId", "nameOnCard", "maskedPan", "status", "expiryDate", "productCode"}

	for _, cardInterface := range data {
		card, ok := cardInterface.(map[string]interface{})
		if !ok {
			return fmt.Errorf("card is not an object")
		}

		for _, field := range requiredFields {
			if _, ok := card[field]; !ok {
				return fmt.Errorf("missing field %s", field)
			}
		}
	}

	return nil
}

func (tc *TestContext) responsePaginated() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	pagination, ok := result["pagination"].(map[string]interface{})
	if !ok {
		// Fallback to top-level fields
		if _, ok := result["pageNumber"]; !ok {
			return fmt.Errorf("missing pagination object and missing pageNumber")
		}
		if _, ok := result["pageSize"]; !ok {
			return fmt.Errorf("missing pageSize")
		}
		if _, ok := result["totalPages"]; !ok {
			return fmt.Errorf("missing totalPages")
		}
		return nil
	}

	if _, ok := pagination["pageNumber"]; !ok {
		return fmt.Errorf("missing pageNumber in pagination")
	}
	if _, ok := pagination["pageSize"]; !ok {
		return fmt.Errorf("missing pageSize in pagination")
	}
	if _, ok := pagination["totalPages"]; !ok {
		return fmt.Errorf("missing totalPages in pagination")
	}

	return nil
}

func (tc *TestContext) responseHasCardObject() error {
	requiredFields := []string{"id", "accountId", "customerId", "nameOnCard", "maskedPan", "status", "expiryDate", "productCode"}
	return tc.checkRequiredFields(requiredFields)
}

func (tc *TestContext) cardStatusIs(status string) error {
	return tc.checkFieldValue("status", status)
}

func (tc *TestContext) cardStatusChangedTo(newStatus string) error {
	return tc.cardStatusIs(newStatus)
}

func (tc *TestContext) cardStatusChangedBackTo(status string) error {
	return tc.cardStatusIs(status)
}

func (tc *TestContext) putWithNote(path, note string) error {
	body := map[string]string{"note": note}
	_, err := tc.sendWithManagedUserHeader("PUT", path, body)
	return err
}

func (tc *TestContext) managedCustomerWithCard() error {
	if tc.userID == "" {
		if err := tc.existingManagedUserWithKYC("accepted"); err != nil {
			return err
		}
	}

	body := map[string]interface{}{
		"walletAddress": "https://ilp.link/test",
		"nameOnCard":    "John Doe",
		"account": map[string]interface{}{
			"productCode": "PWSR_DEBP_2404",
			"currency":    "EUR",
			"card": map[string]interface{}{
				"productCode": "PWSR_DEBP_2404",
			},
		},
	}
	if err := tc.postCardCustomerWithBody("/cards/v1/customers/managed", body); err != nil {
		return err
	}

	return tc.extractCardIDFromResponse()
}

func (tc *TestContext) managedCustomerWithLockedCard() error {
	if err := tc.managedCustomerWithCard(); err != nil {
		return err
	}

	body := map[string]string{"note": "Lock for test"}
	path := fmt.Sprintf("/cards/v1/cards/%s/lock?reasonCode=ClientRequestedLock", tc.cardID)
	_, err := tc.sendWithManagedUserHeader("PUT", path, body)
	if err != nil {
		return err
	}

	if tc.lastResponse.StatusCode != 200 {
		return fmt.Errorf("failed to lock card, status: %d, body: %s", tc.lastResponse.StatusCode, string(tc.lastResponseBody))
	}

	return nil
}

func (tc *TestContext) managedCustomerWithCardAndTransaction() error {
	if err := tc.managedCustomerWithCard(); err != nil {
		return err
	}

	body := map[string]interface{}{
		"cardId":       tc.cardID,
		"amount":       "25.00",
		"currency":     "EUR",
		"type":         0,
		"merchantName": "Test Merchant",
	}
	_, err := tc.sendWithManagedUserHeader("POST", "/cards/v1/transactions", body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}
	if txID, ok := result["transactionId"].(string); ok {
		tc.transactionID = txID
	}

	return nil
}

func (tc *TestContext) managedCustomerWithPending3DS() error {
	if err := tc.managedCustomerWithCard(); err != nil {
		return err
	}

	body := map[string]interface{}{
		"cardId":           tc.cardID,
		"userId":           tc.userID,
		"merchantName":     "Test Merchant",
		"purchaseAmount":   "150.00",
		"purchaseCurrency": "EUR",
	}
	_, err := tc.sendWithManagedUserHeader("POST", "/cards/v1/test/3ds/challenge", body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}
	if txID, ok := result["transactionId"].(string); ok {
		tc.transactionID = txID
	}

	return nil
}

func (tc *TestContext) postCardTokenWithManagedUser(path string) error {
	body := map[string]interface{}{
		"cardId": tc.cardID,
	}
	_, err := tc.sendWithManagedUserHeader("POST", path, body)
	return err
}

func (tc *TestContext) responseContainsLinksArray() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	links, ok := result["links"].([]interface{})
	if !ok || len(links) == 0 {
		return fmt.Errorf("missing or empty links array in response: %s", string(tc.lastResponseBody))
	}

	return nil
}

func (tc *TestContext) putCardLimitsWithUpdate(path string, newDailyLimit int) error {
	limits := []map[string]interface{}{
		{"type": "dailyOverall", "limit": float64(newDailyLimit), "currency": "EUR", "isDisabled": false},
		{"type": "perTransaction", "limit": 500.00, "currency": "EUR", "isDisabled": false},
		{"type": "monthlyOverall", "limit": 5000.00, "currency": "EUR", "isDisabled": false},
		{"type": "dailyAtm", "limit": 300.00, "currency": "EUR", "isDisabled": false},
		{"type": "dailyEcomm", "limit": 800.00, "currency": "EUR", "isDisabled": false},
	}
	_, err := tc.sendWithManagedUserHeader("PUT", path, limits)
	return err
}

func (tc *TestContext) postCardTransaction(path, amount, currency string) error {
	body := map[string]interface{}{
		"cardId":       tc.cardID,
		"amount":       amount,
		"currency":     currency,
		"type":         0,
		"merchantName": "Test Merchant",
	}
	_, err := tc.sendWithManagedUserHeader("POST", path, body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}
	if txID, ok := result["transactionId"].(string); ok {
		tc.transactionID = txID
	}

	return nil
}

func (tc *TestContext) responseContainsCardTransactionWithGH(ghCode string) error {
	if err := tc.checkRequiredFields([]string{"transactionId"}); err != nil {
		return err
	}
	return tc.checkFieldValue("ghResponseCode", ghCode)
}

func (tc *TestContext) responseContainsCardTransactionDetails() error {
	return tc.checkRequiredFields([]string{"transactionId", "transactionAmount"})
}

func (tc *TestContext) responseContainsPendingConfirmations() error {
	return tc.responseContainsPending3DSConfirmations(3)
}

func (tc *TestContext) responseStatusWithStatus(status int, statusStr string) error {
	if tc.lastResponse.StatusCode != status {
		return fmt.Errorf("expected status %d, got %d", status, tc.lastResponse.StatusCode)
	}
	return tc.checkFieldValue("status", statusStr)
}

func (tc *TestContext) responseContainsTokenStarting(prefix string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	token, ok := result["token"].(string)
	if !ok {
		return fmt.Errorf("missing token field")
	}

	if len(token) == 0 || len(prefix) > len(token) || token[:len(prefix)] != prefix {
		return fmt.Errorf("token does not start with %s, got: %s", prefix, token)
	}

	return nil
}

func (tc *TestContext) responseContainsPending3DSConfirmations(version int) error {
	var result struct {
		PendingConfirmations []map[string]interface{} `json:"pendingConfirmations"`
	}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return fmt.Errorf("response is not an object with pendingConfirmations: %w", err)
	}

	if len(result.PendingConfirmations) == 0 {
		return fmt.Errorf("pendingConfirmations array is empty")
	}

	return nil
}

func (tc *TestContext) eachConfirmationHasRequiredFields() error {
	var result struct {
		PendingConfirmations []map[string]interface{} `json:"pendingConfirmations"`
	}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	requiredFields := []string{"transactionId", "merchantName", "purchaseAmount", "purchaseCurrency", "timeout"}
	for i, item := range result.PendingConfirmations {
		for _, field := range requiredFields {
			if _, ok := item[field]; !ok {
				return fmt.Errorf("confirmation %d missing %s", i, field)
			}
		}
	}

	return nil
}

func (tc *TestContext) postWith3DSConfirmation(path string, confirmed string, authMethod string) error {
	isConfirmed := confirmed == "true"
	body := map[string]interface{}{
		"confirmed":  isConfirmed,
		"authMethod": authMethod,
	}
	_, err := tc.sendWithManagedUserHeader("POST", path, body)
	return err
}

func (tc *TestContext) responseIndicatesSuccess(status string) error {
	return tc.checkFieldValue("status", status)
}

func (tc *TestContext) responseIndicatesDeclined(status string) error {
	return tc.checkFieldValue("status", status)
}

func (tc *TestContext) responseHasCardProducts() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if data, ok := result["data"].([]interface{}); !ok || len(data) == 0 {
		return fmt.Errorf("missing or empty data array")
	}

	return nil
}

func (tc *TestContext) responseContainsPlasticCardOrder(status, cardType string) error {
	if err := tc.checkRequiredFields([]string{"orderId", "cardId"}); err != nil {
		return err
	}
	if err := tc.checkFieldValue("status", status); err != nil {
		return err
	}
	return tc.checkFieldValue("type", cardType)
}

func (tc *TestContext) responseIsArrayOfCardLimits() error {
	var result []map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return fmt.Errorf("response is not an array: %w", err)
	}

	if len(result) == 0 {
		return fmt.Errorf("limits array is empty")
	}

	return nil
}

func (tc *TestContext) eachLimitHasRequiredFields(currency string) error {
	var result []map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	for i, item := range result {
		if _, ok := item["type"]; !ok {
			return fmt.Errorf("limit %d missing type", i)
		}
		if _, ok := item["limit"]; !ok {
			return fmt.Errorf("limit %d missing limit", i)
		}
		if c, ok := item["currency"].(string); !ok || c != currency {
			return fmt.Errorf("limit %d currency is not %s", i, currency)
		}
		if _, ok := item["isDisabled"]; !ok {
			return fmt.Errorf("limit %d missing isDisabled", i)
		}
	}

	return nil
}

func (tc *TestContext) limitsIncludeRequiredTypes() error {
	var result []map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	requiredTypes := map[string]bool{
		"dailyOverall":   false,
		"perTransaction": false,
		"monthlyOverall": false,
		"dailyAtm":       false,
		"dailyEcomm":     false,
	}

	for _, item := range result {
		if limitType, ok := item["type"].(string); ok {
			if _, required := requiredTypes[limitType]; required {
				requiredTypes[limitType] = true
			}
		}
	}

	for limitType, found := range requiredTypes {
		if !found {
			return fmt.Errorf("missing limit type: %s", limitType)
		}
	}

	return nil
}

func (tc *TestContext) responseReturnsUpdatedLimits() error {
	return tc.responseIsArrayOfCardLimits()
}

func (tc *TestContext) dailyOverallLimitChanged(newLimit int) error {
	var result []map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	for _, item := range result {
		if limitType, ok := item["type"].(string); ok && limitType == "dailyOverall" {
			if limitVal, ok := item["limit"].(float64); ok && int(limitVal) == newLimit {
				return nil
			}
		}
	}

	return fmt.Errorf("dailyOverall limit not changed to %d", newLimit)
}

func (tc *TestContext) cardNotInListCardsQuery() error {
	_, err := tc.sendWithManagedUserHeader("GET", "/cards/v1/cards/{customerId}", nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if data, ok := result["data"].([]interface{}); ok {
		for _, item := range data {
			if card, ok := item.(map[string]interface{}); ok {
				if status, ok := card["status"].(string); ok && status == "SoftDelete" {
					return fmt.Errorf("found card with SoftDelete status in list")
				}
			}
		}
	}

	return nil
}

// Helper to extract card ID from response
func (tc *TestContext) extractCardIDFromResponse() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	customers, ok := result["customers"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing customers in response: %s", string(tc.lastResponseBody))
	}
	if id, ok := customers["id"].(string); ok {
		tc.customerID = id
	}

	accounts, ok := customers["accounts"].([]interface{})
	if !ok || len(accounts) == 0 {
		return fmt.Errorf("no accounts in response")
	}
	account, ok := accounts[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("account is not an object")
	}
	cards, ok := account["cards"].([]interface{})
	if !ok || len(cards) == 0 {
		return fmt.Errorf("no cards in response")
	}
	card, ok := cards[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("card is not an object")
	}
	if id, ok := card["id"].(string); ok {
		tc.cardID = id
	}

	return nil
}
