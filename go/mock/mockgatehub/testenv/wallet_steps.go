package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ============ WALLET STEPS ============

func (tc *TestContext) walletsArrayReturned() error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	if wallets, ok := result["wallets"].([]interface{}); !ok || len(wallets) == 0 {
		return fmt.Errorf("expected wallets array with at least one wallet, got %v", result)
	}

	return nil
}

func (tc *TestContext) firstWalletStartsWith(prefix string) error {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return err
	}

	wallets, ok := result["wallets"].([]interface{})
	if !ok || len(wallets) == 0 {
		return fmt.Errorf("no wallets found")
	}

	wallet, ok := wallets[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("wallet is not an object")
	}

	address, ok := wallet["address"].(string)
	if !ok {
		return fmt.Errorf("no address in wallet")
	}

	if !strings.HasPrefix(address, prefix) {
		return fmt.Errorf("expected address to start with %s, got %s", prefix, address)
	}

	tc.walletAddress = address
	return nil
}

func (tc *TestContext) managedUserWithWalletAddress() error {
	// Create a managed user
	if err := tc.existingManagedUserGeneric(); err != nil {
		return err
	}

	// GET the user — GateHub auto-provisions a wallet for managed users,
	// so we just retrieve it rather than explicitly creating one.
	resp, err := tc.request("GET", "/core/v1/users/{userId}", nil, nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &result); err != nil {
		return fmt.Errorf("failed to unmarshal user wallets response: %w. Response: %s. Status: %d", err, string(tc.lastResponseBody), resp.StatusCode)
	}

	// Extract wallet address from the first wallet in the array
	if wallets, ok := result["wallets"].([]interface{}); ok && len(wallets) > 0 {
		if wallet, ok := wallets[0].(map[string]interface{}); ok {
			if address, ok := wallet["address"].(string); ok {
				tc.walletAddress = address
			}
		}
	}

	if tc.walletAddress == "" {
		return fmt.Errorf("no wallet address found in user response: %s", string(tc.lastResponseBody))
	}

	tc.lastResponse = resp
	return nil
}

func (tc *TestContext) getWalletBalance() error {
	// Wait for async balance update to complete
	// (deposits are processed asynchronously with 2s delay when webhook URL is configured)
	time.Sleep(3 * time.Second)

	// Use placeholder pattern - will be replaced by replacePlaceholders
	path := "/core/v1/wallets/{walletAddress}/balances"

	resp, err := tc.request("GET", path, nil, nil)
	if err != nil {
		return err
	}

	tc.lastResponse = resp
	return nil
}

func (tc *TestContext) currencyBalanceIs(currency, expectedBalance string) error {
	// Parse the balance response (an array of balance objects)
	var balances []map[string]interface{}
	if err := json.Unmarshal(tc.lastResponseBody, &balances); err != nil {
		return fmt.Errorf("failed to unmarshal balance response: %w. Response: %s", err, string(tc.lastResponseBody))
	}

	// Find the balance for the specified currency
	for _, balance := range balances {
		if vault, ok := balance["vault"].(map[string]interface{}); ok {
			if assetCode, ok := vault["asset_code"].(string); ok && assetCode == currency {
				// Found the currency, check the available balance
				if available, ok := balance["available"].(string); ok {
					if available != expectedBalance {
						return fmt.Errorf("expected %s balance to be %s, got %s", currency, expectedBalance, available)
					}
					return nil
				}
			}
		}
	}

	return fmt.Errorf("currency %s not found in balance response", currency)
}
