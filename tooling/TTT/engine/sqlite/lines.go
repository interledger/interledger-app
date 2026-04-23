package sqlite

import (
	"time"

	"ttt/engine"
)

func (s *Store) PostLines(lines []engine.JournalLine) error {
	if len(lines) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO journal_lines
		(id, event_id, timestamp, debit_account_id, credit_account_id, amount, metadata, debit_metadata, credit_metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer func() { _ = stmt.Close() }()
	for _, l := range lines {
		meta, err := marshalMetadata(l.Metadata)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		debitMeta, err := marshalMetadata(l.DebitMetadata)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		creditMeta, err := marshalMetadata(l.CreditMetadata)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if _, err := stmt.Exec(
			l.ID, l.EventID, l.Timestamp.UTC().Format(time.RFC3339Nano), l.DebitAccountID, l.CreditAccountID, l.Amount,
			meta, debitMeta, creditMeta,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) GetLinesByAccount(accountID string) ([]engine.JournalLine, error) {
	lines, err := s.queryLines(`debit_account_id = ? OR credit_account_id = ?`, accountID, accountID)
	if err != nil {
		return nil, err
	}
	if len(lines) > 0 {
		return lines, nil
	}
	legacy, err := s.queryEntries(`account_id = ?`, accountID)
	if err != nil {
		return nil, err
	}
	if len(legacy) == 0 {
		return nil, nil
	}
	return engine.LinesFromEntries(legacy, s.GetAccount)
}

func (s *Store) GetLinesByEvent(eventID string) ([]engine.JournalLine, error) {
	lines, err := s.queryLines(`event_id = ?`, eventID)
	if err != nil {
		return nil, err
	}
	if len(lines) > 0 {
		return lines, nil
	}
	legacy, err := s.queryEntries(`event_id = ?`, eventID)
	if err != nil {
		return nil, err
	}
	if len(legacy) == 0 {
		return nil, nil
	}
	return engine.LinesFromEntries(legacy, s.GetAccount)
}

func (s *Store) GetAllLines() ([]engine.JournalLine, error) {
	lines, err := s.queryLines("")
	if err != nil {
		return nil, err
	}
	legacy, err := s.queryEntries("")
	if err != nil {
		return nil, err
	}
	if len(legacy) == 0 {
		return lines, nil
	}
	legacyLines, err := engine.LinesFromEntries(legacy, s.GetAccount)
	if err != nil {
		return nil, err
	}
	out := make([]engine.JournalLine, 0, len(lines)+len(legacyLines))
	out = append(out, lines...)
	out = append(out, legacyLines...)
	return out, nil
}
