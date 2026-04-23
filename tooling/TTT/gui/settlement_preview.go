package gui

import (
	"fmt"

	"ttt/engine"
)

func settlementPreviewText(e *engine.Engine, providerA, providerB string, cur engine.Currency) string {
	liqA, ok, err := findLiquidityAccount(e, providerA, cur)
	if err != nil {
		return "Settlement preview unavailable: " + err.Error()
	}
	if !ok {
		return fmt.Sprintf("Settlement preview unavailable: no %s liquidity account for %s", cur.Code, providerA)
	}
	liqB, ok, err := findLiquidityAccount(e, providerB, cur)
	if err != nil {
		return "Settlement preview unavailable: " + err.Error()
	}
	if !ok {
		return fmt.Sprintf("Settlement preview unavailable: no %s liquidity account for %s", cur.Code, providerB)
	}
	posA, ok, err := findPositionAccount(e, liqA.ID, providerB)
	if err != nil {
		return "Settlement preview unavailable: " + err.Error()
	}
	if !ok {
		return fmt.Sprintf("Settlement preview unavailable: no %s->%s position account in %s", providerA, providerB, cur.Code)
	}
	posB, ok, err := findPositionAccount(e, liqB.ID, providerA)
	if err != nil {
		return "Settlement preview unavailable: " + err.Error()
	}
	if !ok {
		return fmt.Sprintf("Settlement preview unavailable: no %s->%s position account in %s", providerB, providerA, cur.Code)
	}

	balA, err := e.Balance(posA.ID)
	if err != nil {
		return "Settlement preview unavailable: " + err.Error()
	}
	balB, err := e.Balance(posB.ID)
	if err != nil {
		return "Settlement preview unavailable: " + err.Error()
	}
	if balA+balB != 0 {
		return fmt.Sprintf("Settlement preview unavailable: bilateral mirror mismatch (%s=%d, %s=%d)", providerA, balA, providerB, balB)
	}
	if balA == 0 {
		return fmt.Sprintf("No settlement needed now between %s and %s in %s", providerA, providerB, cur.Code)
	}

	amount := balA
	payer := providerB
	receiver := providerA
	if amount < 0 {
		amount = -amount
		payer = providerA
		receiver = providerB
	}
	return fmt.Sprintf("Settlement now: %s should pay %s %s to %s", payer, formatMinor(amount, cur.AssetScale), cur.Code, receiver)
}

func findLiquidityAccount(e *engine.Engine, providerID string, cur engine.Currency) (engine.Account, bool, error) {
	accounts, err := e.ListAccounts()
	if err != nil {
		return engine.Account{}, false, err
	}
	for _, a := range accounts {
		if a.Type == engine.AccountTypeLiquidity && a.ProviderID == providerID && a.Currency.Code == cur.Code {
			return a, true, nil
		}
	}
	return engine.Account{}, false, nil
}

func findPositionAccount(e *engine.Engine, liquidityAccountID, counterpartyID string) (engine.Account, bool, error) {
	accounts, err := e.ListAccounts()
	if err != nil {
		return engine.Account{}, false, err
	}
	for _, a := range accounts {
		if a.Type == engine.AccountTypePosition && a.LiquidityAccountID == liquidityAccountID && a.CounterpartyID == counterpartyID {
			return a, true, nil
		}
	}
	return engine.Account{}, false, nil
}
