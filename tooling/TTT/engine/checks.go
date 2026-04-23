package engine

import "fmt"

// CurrencyBalance holds the net signed ledger amount for a currency.
// A healthy ledger has Net == 0 for every currency.
type CurrencyBalance struct {
	Currency string
	Net      int64
}

// CheckPerEventBalance verifies that a single workflow event has equal total debits and credits.
// Returns an error if the event is missing, has no lines, or is unbalanced.
func (e *Engine) CheckPerEventBalance(eventID string) error {
	lines, err := e.store.GetLinesByEvent(eventID)
	if err == nil {
		if len(lines) == 0 {
			return fmt.Errorf("event %q has no lines", eventID)
		}
		nets, err := e.currencyNetsFromLines(lines)
		if err != nil {
			return err
		}
		for code, net := range nets {
			if net != 0 {
				return fmt.Errorf("event %q is unbalanced for %q: net signed amount = %d (expected 0)", eventID, code, net)
			}
		}
		return nil
	}

	// Legacy fallback for intentionally malformed direct entry posts in tests
	// and diagnostics where lines cannot be reconstructed.
	entries, legacyErr := e.store.GetEntriesByEvent(eventID)
	if legacyErr != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("event %q has no lines", eventID)
	}
	nets, legacyErr := e.currencyNetsFromEntries(entries)
	if legacyErr != nil {
		return legacyErr
	}
	for code, net := range nets {
		if net != 0 {
			return fmt.Errorf("event %q is unbalanced for %q: net signed amount = %d (expected 0)", eventID, code, net)
		}
	}
	return nil
}

// CheckGlobalBalance verifies that across all journal lines the net signed amount is zero
// for each currency. Returns the per-currency net values and a non-nil error if any are non-zero.
func (e *Engine) CheckGlobalBalance() ([]CurrencyBalance, error) {
	lines, err := e.store.GetAllLines()

	var nets map[string]int64
	if err == nil {
		nets, err = e.currencyNetsFromLines(lines)
		if err != nil {
			return nil, err
		}
	} else {
		// Legacy fallback for intentionally malformed direct entry posts in tests
		// and diagnostics where lines cannot be reconstructed.
		entries, legacyErr := e.store.GetAllEntries()
		if legacyErr != nil {
			return nil, err
		}
		nets, legacyErr = e.currencyNetsFromEntries(entries)
		if legacyErr != nil {
			return nil, legacyErr
		}
	}

	results := make([]CurrencyBalance, 0, len(nets))
	for code, net := range nets {
		results = append(results, CurrencyBalance{Currency: code, Net: net})
	}

	for _, r := range results {
		if r.Net != 0 {
			return results, fmt.Errorf("global balance check failed for currency %q: net = %d (expected 0)", r.Currency, r.Net)
		}
	}
	return results, nil
}

func (e *Engine) currencyNetsFromLines(lines []JournalLine) (map[string]int64, error) {
	nets := make(map[string]int64)
	for _, line := range lines {
		debitAcct, err := e.store.GetAccount(line.DebitAccountID)
		if err != nil {
			return nil, fmt.Errorf("line %q references missing debit account %q: %w", line.ID, line.DebitAccountID, err)
		}
		creditAcct, err := e.store.GetAccount(line.CreditAccountID)
		if err != nil {
			return nil, fmt.Errorf("line %q references missing credit account %q: %w", line.ID, line.CreditAccountID, err)
		}
		nets[debitAcct.Currency.Code] -= line.Amount
		nets[creditAcct.Currency.Code] += line.Amount
	}
	return nets, nil
}

func (e *Engine) currencyNetsFromEntries(entries []LedgerEntry) (map[string]int64, error) {
	nets := make(map[string]int64)
	for _, entry := range entries {
		acct, err := e.store.GetAccount(entry.AccountID)
		if err != nil {
			return nil, fmt.Errorf("entry %q references missing account %q: %w", entry.ID, entry.AccountID, err)
		}
		nets[acct.Currency.Code] += SignedAmount(entry)
	}
	return nets, nil
}
