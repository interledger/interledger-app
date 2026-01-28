package main

import (
	"fmt"
	"time"
)

func scenarioGetBalance() scenario {
	return scenario{
		name: "Balance returns ZAR and USD",
		run: func(h *harness) (string, error) {
			if err := h.login(); err != nil {
				return "", err
			}
			walletID := fmt.Sprintf("wallet-bal-%d", time.Now().UnixNano())
			created, err := h.createSubAccount(walletID)
			if err != nil {
				return "", err
			}
			bal, err := h.getBalance(created.AccountID)
			if err != nil {
				return "", err
			}
			if bal.AccountID != created.AccountID {
				return "", fmt.Errorf("accountId mismatch in balance")
			}
			zar, err := h.expectBalanceCurrency(bal.Balances, "ZAR")
			if err != nil {
				return "", err
			}
			usd, err := h.expectBalanceCurrency(bal.Balances, "USD")
			if err != nil {
				return "", err
			}
			if zar.Available != 0 || usd.Available != 0 {
				return "", fmt.Errorf("expected zero balances, got zar %.2f usd %.2f", zar.Available, usd.Available)
			}
			return fmt.Sprintf("balance ok for %s", bal.AccountID), nil
		},
	}
}
