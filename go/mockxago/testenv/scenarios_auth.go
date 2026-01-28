package main

import "fmt"

func scenarioLogin() scenario {
	return scenario{
		name: "Login returns token",
		run: func(h *harness) (string, error) {
			if err := h.login(); err != nil {
				return "", err
			}
			return "token issued", nil
		},
	}
}

func scenarioBalanceAuthRequired() scenario {
	return scenario{
		name: "Balance endpoint requires auth",
		run: func(h *harness) (string, error) {
			// Intentionally avoid login
			_, err := h.getBalance("00000000-0000-0000-0000-000000000000")
			if err == nil {
				return "", fmt.Errorf("expected unauthorized error")
			}
			return "unauthorized rejected", nil
		},
	}
}
