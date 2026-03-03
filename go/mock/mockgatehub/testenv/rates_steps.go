package main

import (
	"encoding/json"
	"fmt"
)

// ============ RATES STEPS ============

func (tc *TestContext) mockgatehubRunningWithHeaders() error {
	tc.appID = "local-test-app-id"
	tc.appSecret = "local-test-app-secret"
	return nil
}

func (tc *TestContext) payloadHasCounterCurrency() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	_, ok1 := result["counter_currency"]
	_, ok2 := result["counterCurrency"]
	_, ok3 := result["counter"]
	if !ok1 && !ok2 && !ok3 {
		return fmt.Errorf("missing counter_currency, counterCurrency, or counter in: %s", string(tc.lastResponseBody))
	}

	return nil
}

func (tc *TestContext) rateEntryExists() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	// Count non-counter fields
	count := 0
	for key := range result {
		if key != "counter_currency" && key != "counterCurrency" {
			count++
		}
	}

	if count == 0 {
		return fmt.Errorf("no rate entries")
	}

	return nil
}

func (tc *TestContext) responseHasVaults() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	vaults, ok := result["vaults"].([]interface{})
	if !ok || len(vaults) == 0 {
		return fmt.Errorf("no vaults array or empty")
	}

	return nil
}
