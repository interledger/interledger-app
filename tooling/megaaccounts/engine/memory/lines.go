package memory

import "megaaccounts/engine"

func (s *Store) PostLines(lines []engine.JournalLine) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lines = append(s.lines, lines...)
	return nil
}

func (s *Store) GetLinesByAccount(accountID string) ([]engine.JournalLine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	filtered := make([]engine.JournalLine, 0, len(s.lines))
	for _, line := range s.lines {
		if line.DebitAccountID == accountID || line.CreditAccountID == accountID {
			filtered = append(filtered, line)
		}
	}
	return filtered, nil
}

func (s *Store) GetLinesByEvent(eventID string) ([]engine.JournalLine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	filtered := make([]engine.JournalLine, 0, len(s.lines))
	for _, line := range s.lines {
		if line.EventID == eventID {
			filtered = append(filtered, line)
		}
	}
	if len(filtered) > 0 {
		return filtered, nil
	}
	// Compatibility fallback for legacy direct entry posts.
	legacy := make([]engine.LedgerEntry, 0)
	for _, e := range s.legacy {
		if e.EventID == eventID {
			legacy = append(legacy, e)
		}
	}
	if len(legacy) == 0 {
		return nil, nil
	}
	return engine.LinesFromEntries(legacy, s.GetAccount)
}

func (s *Store) GetAllLines() ([]engine.JournalLine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.legacy) == 0 {
		out := make([]engine.JournalLine, len(s.lines))
		copy(out, s.lines)
		return out, nil
	}
	legacyLines, err := engine.LinesFromEntries(s.legacy, s.GetAccount)
	if err != nil {
		return nil, err
	}
	out := make([]engine.JournalLine, 0, len(s.lines)+len(legacyLines))
	out = append(out, s.lines...)
	out = append(out, legacyLines...)
	return out, nil
}
