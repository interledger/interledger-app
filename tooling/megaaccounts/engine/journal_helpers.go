package engine

import "fmt"

// LinesFromEntries converts a pairwise debit/credit entry slice into balanced
// journal lines. During the migration, entry pairs must appear adjacently.
func LinesFromEntries(entries []LedgerEntry, getAccount func(string) (Account, error)) ([]JournalLine, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	if len(entries)%2 != 0 {
		return nil, fmt.Errorf("entry batch must contain an even number of rows to convert to journal lines")
	}
	lines := make([]JournalLine, 0, len(entries)/2)
	for i := 0; i < len(entries); i += 2 {
		left, right := entries[i], entries[i+1]
		if left.EventID != right.EventID {
			return nil, fmt.Errorf("entry pair %d/%d mixes event ids %q and %q", i, i+1, left.EventID, right.EventID)
		}
		if left.Amount != right.Amount {
			return nil, fmt.Errorf("entry pair %d/%d has mismatched amounts %d and %d", i, i+1, left.Amount, right.Amount)
		}

		var debit, credit LedgerEntry
		switch {
		case left.Type == EntryTypeDebit && right.Type == EntryTypeCredit:
			debit, credit = left, right
		case left.Type == EntryTypeCredit && right.Type == EntryTypeDebit:
			debit, credit = right, left
		default:
			return nil, fmt.Errorf("entry pair %d/%d must contain one debit and one credit", i, i+1)
		}

		debitAcct, err := getAccount(debit.AccountID)
		if err != nil {
			return nil, fmt.Errorf("entry pair %d debit account: %w", i, err)
		}
		creditAcct, err := getAccount(credit.AccountID)
		if err != nil {
			return nil, fmt.Errorf("entry pair %d credit account: %w", i, err)
		}
		if debitAcct.Currency.Code != creditAcct.Currency.Code || debitAcct.Currency.AssetScale != creditAcct.Currency.AssetScale {
			return nil, fmt.Errorf("entry pair %d/%d currency mismatch: debit=%s credit=%s", i, i+1, debitAcct.Currency.Code, creditAcct.Currency.Code)
		}

		lines = append(lines, JournalLine{
			ID:              debit.ID,
			EventID:         debit.EventID,
			Timestamp:       debit.Timestamp,
			DebitAccountID:  debit.AccountID,
			CreditAccountID: credit.AccountID,
			Amount:          debit.Amount,
			Metadata:        CommonMetadata(debit.Metadata, credit.Metadata),
			DebitMetadata:   cloneMetadata(debit.Metadata),
			CreditMetadata:  cloneMetadata(credit.Metadata),
		})
	}
	return lines, nil
}

// ExpandJournalLines renders journal lines back into the legacy one-sided
// ledger-entry view used by the current UI and checks.
func ExpandJournalLines(lines []JournalLine, accountFilter string) []LedgerEntry {
	entries := make([]LedgerEntry, 0, len(lines)*2)
	for _, line := range lines {
		if accountFilter == "" || line.DebitAccountID == accountFilter {
			entries = append(entries, LedgerEntry{
				ID:        line.ID + ":debit",
				AccountID: line.DebitAccountID,
				Amount:    line.Amount,
				Type:      EntryTypeDebit,
				EventID:   line.EventID,
				Timestamp: line.Timestamp,
				Metadata:  firstNonEmptyMetadata(line.DebitMetadata, line.Metadata),
			})
		}
		if accountFilter == "" || line.CreditAccountID == accountFilter {
			entries = append(entries, LedgerEntry{
				ID:        line.ID + ":credit",
				AccountID: line.CreditAccountID,
				Amount:    line.Amount,
				Type:      EntryTypeCredit,
				EventID:   line.EventID,
				Timestamp: line.Timestamp,
				Metadata:  firstNonEmptyMetadata(line.CreditMetadata, line.Metadata),
			})
		}
	}
	return entries
}

// CommonMetadata returns the subset of key/value pairs that are identical in
// both metadata maps.
func CommonMetadata(a, b map[string]string) map[string]string {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	out := map[string]string{}
	for k, av := range a {
		if bv, ok := b[k]; ok && av == bv {
			out[k] = av
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func firstNonEmptyMetadata(primary, fallback map[string]string) map[string]string {
	// Merge primary and fallback, with primary taking precedence
	merged := make(map[string]string)
	for k, v := range fallback {
		merged[k] = v
	}
	for k, v := range primary {
		merged[k] = v
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}
