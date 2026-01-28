package main

import "fmt"

func scenarioListCurrencies() scenario {
	return scenario{
		name: "List currencies returns ZAR and USD",
		run: func(h *harness) (string, error) {
			resp, err := h.listCurrencies()
			if err != nil {
				return "", err
			}
			if len(resp) != 2 {
				return "", fmt.Errorf("expected 2 currencies, got %d", len(resp))
			}
			seen := map[string]bool{}
			for _, c := range resp {
				seen[c.CurrencyID] = true
			}
			if !seen["ZAR"] || !seen["USD"] {
				return "", fmt.Errorf("expected ZAR and USD in response")
			}
			return "currencies ok", nil
		},
	}
}
