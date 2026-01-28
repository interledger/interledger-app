package main

import (
	"fmt"
	"time"
)

func scenarioCreateAndFetchSubAccount() scenario {
	return scenario{
		name: "Create and fetch sub-account",
		run: func(h *harness) (string, error) {
			if err := h.login(); err != nil {
				return "", err
			}
			walletID := fmt.Sprintf("wallet-%d", time.Now().UnixNano())
			created, err := h.createSubAccount(walletID)
			if err != nil {
				return "", err
			}
			fetched, err := h.getSubAccountByWallet(walletID)
			if err != nil {
				return "", err
			}
			if fetched.AccountID != created.AccountID {
				return "", fmt.Errorf("accountId mismatch: created %s fetched %s", created.AccountID, fetched.AccountID)
			}
			return fmt.Sprintf("account %s for wallet %s", created.AccountID, walletID), nil
		},
	}
}

func scenarioUpdateSubAccount() scenario {
	return scenario{
		name: "Update sub-account fields",
		run: func(h *harness) (string, error) {
			if err := h.login(); err != nil {
				return "", err
			}
			walletID := fmt.Sprintf("wallet-up-%d", time.Now().UnixNano())
			created, err := h.createSubAccount(walletID)
			if err != nil {
				return "", err
			}
			if err := h.updateSubAccount(created.AccountID); err != nil {
				return "", err
			}
			return fmt.Sprintf("updated %s", created.AccountID), nil
		},
	}
}
