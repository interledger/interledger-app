package engine

import (
	"fmt"
	"sort"
)

// LiquidityDecomposition describes the split of a liquidity account between
// its available float and the sum of its position balances.
type LiquidityDecomposition struct {
	LiquidityAccountID string
	ProviderID         string
	Currency           string
	LiquidityBalance   int64
	PositionTotal      int64
	Float              int64 // LiquidityBalance - PositionTotal
}

// CheckLiquidityDecomposition returns the per-liquidity-account split and
// a non-nil error if any account's available float is negative (meaning the
// position obligations exceed the reserve pool).
func (e *Engine) CheckLiquidityDecomposition() ([]LiquidityDecomposition, error) {
	accounts, err := e.store.ListAccounts()
	if err != nil {
		return nil, err
	}

	byLiq := make(map[string][]Account)
	var liqs []Account
	for _, a := range accounts {
		switch a.Type {
		case AccountTypeLiquidity:
			liqs = append(liqs, a)
		case AccountTypePosition:
			byLiq[a.LiquidityAccountID] = append(byLiq[a.LiquidityAccountID], a)
		}
	}
	sort.Slice(liqs, func(i, j int) bool {
		if liqs[i].ProviderID != liqs[j].ProviderID {
			return liqs[i].ProviderID < liqs[j].ProviderID
		}
		return liqs[i].Currency.Code < liqs[j].Currency.Code
	})

	results := make([]LiquidityDecomposition, 0, len(liqs))
	var firstErr error
	for _, liq := range liqs {
		liqBal, err := e.Balance(liq.ID)
		if err != nil {
			return results, err
		}
		var posTotal int64
		for _, p := range byLiq[liq.ID] {
			if p.Currency != liq.Currency {
				return results, fmt.Errorf("position %q currency %q does not match liquidity %q currency %q",
					p.ID, p.Currency.Code, liq.ID, liq.Currency.Code)
			}
			pbal, err := e.Balance(p.ID)
			if err != nil {
				return results, err
			}
			posTotal += pbal
		}
		d := LiquidityDecomposition{
			LiquidityAccountID: liq.ID,
			ProviderID:         liq.ProviderID,
			Currency:           liq.Currency.Code,
			LiquidityBalance:   liqBal,
			PositionTotal:      posTotal,
			Float:              liqBal - posTotal,
		}
		results = append(results, d)
		if d.Float < 0 && firstErr == nil {
			firstErr = fmt.Errorf("liquidity %s/%s over-committed: float=%d (balance=%d, positions=%d)",
				d.ProviderID, d.Currency, d.Float, d.LiquidityBalance, d.PositionTotal)
		}
	}
	return results, firstErr
}

// BilateralPair represents a mirrored pair of position accounts (A→B, B→A)
// for a given currency and the sum of their credit-normal balances.
// Sum == 0 means the mirror invariant holds.
type BilateralPair struct {
	ProviderA string
	ProviderB string
	Currency  string
	BalanceA  int64 // balance of A's position account for B
	BalanceB  int64 // balance of B's position account for A
	MirrorSum int64 // BalanceA + BalanceB; 0 when healthy
}

// CheckBilateralPositions returns one entry per mirrored position pair and
// a non-nil error if any pair fails the mirror invariant (BalanceA+BalanceB != 0).
// Position accounts that have no mirror on the counterparty's side are reported
// as an error.
func (e *Engine) CheckBilateralPositions() ([]BilateralPair, error) {
	accounts, err := e.store.ListAccounts()
	if err != nil {
		return nil, err
	}

	liqByID := make(map[string]Account)
	for _, a := range accounts {
		if a.Type == AccountTypeLiquidity {
			liqByID[a.ID] = a
		}
	}

	type posKey struct {
		provider, counterparty, currency string
	}
	positions := make(map[posKey]Account)
	for _, a := range accounts {
		if a.Type != AccountTypePosition {
			continue
		}
		liq, ok := liqByID[a.LiquidityAccountID]
		if !ok {
			return nil, fmt.Errorf("position %q references missing liquidity %q", a.ID, a.LiquidityAccountID)
		}
		positions[posKey{liq.ProviderID, a.CounterpartyID, liq.Currency.Code}] = a
	}

	var pairs []BilateralPair
	seen := make(map[posKey]bool)
	var firstErr error
	for k, posA := range positions {
		mirror := posKey{k.counterparty, k.provider, k.currency}
		if seen[k] || seen[mirror] {
			continue
		}
		seen[k] = true
		seen[mirror] = true

		posB, ok := positions[mirror]
		if !ok {
			if firstErr == nil {
				firstErr = fmt.Errorf("position %s→%s (%s) has no mirror on %s",
					k.provider, k.counterparty, k.currency, k.counterparty)
			}
			continue
		}
		balA, err := e.Balance(posA.ID)
		if err != nil {
			return pairs, err
		}
		balB, err := e.Balance(posB.ID)
		if err != nil {
			return pairs, err
		}
		p := BilateralPair{
			ProviderA: k.provider, ProviderB: k.counterparty, Currency: k.currency,
			BalanceA: balA, BalanceB: balB, MirrorSum: balA + balB,
		}
		// Normalise ordering for stable output.
		if p.ProviderA > p.ProviderB {
			p.ProviderA, p.ProviderB = p.ProviderB, p.ProviderA
			p.BalanceA, p.BalanceB = p.BalanceB, p.BalanceA
		}
		pairs = append(pairs, p)
		if p.MirrorSum != 0 && firstErr == nil {
			firstErr = fmt.Errorf("bilateral mirror broken for %s↔%s (%s): A=%d B=%d sum=%d",
				p.ProviderA, p.ProviderB, p.Currency, p.BalanceA, p.BalanceB, p.MirrorSum)
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Currency != pairs[j].Currency {
			return pairs[i].Currency < pairs[j].Currency
		}
		if pairs[i].ProviderA != pairs[j].ProviderA {
			return pairs[i].ProviderA < pairs[j].ProviderA
		}
		return pairs[i].ProviderB < pairs[j].ProviderB
	})
	return pairs, firstErr
}
