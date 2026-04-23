// Package sqlite provides a SQLite-backed implementation of engine.Store.
// Uses the pure-Go driver modernc.org/sqlite (no CGO required).
package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"ttt/engine"
)

// Store is a SQLite-backed engine.Store.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) a SQLite database at the given DSN and ensures the
// schema is up to date.
//
// Use "file::memory:?cache=shared" for an in-memory database (useful in tests).
func Open(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite performs better with foreign keys and WAL; single-connection
	// eliminates locking surprises for our use case.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

// Close closes the underlying database handle.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS providers (
			id   TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS accounts (
			id                   TEXT PRIMARY KEY,
			type                 INTEGER NOT NULL,
			provider_id          TEXT NOT NULL,
			currency_code        TEXT NOT NULL,
			currency_scale       INTEGER NOT NULL,
			user_id              TEXT NOT NULL DEFAULT '',
			counterparty_id      TEXT NOT NULL DEFAULT '',
			liquidity_account_id TEXT NOT NULL DEFAULT ''
		);`,
		`CREATE INDEX IF NOT EXISTS idx_accounts_lookup
			ON accounts (type, provider_id, currency_code, user_id, counterparty_id, liquidity_account_id);`,
		`CREATE TABLE IF NOT EXISTS ledger_entries (
			seq        INTEGER PRIMARY KEY AUTOINCREMENT,
			id         TEXT NOT NULL UNIQUE,
			account_id TEXT NOT NULL,
			amount     INTEGER NOT NULL,
			type       INTEGER NOT NULL,
			event_id   TEXT NOT NULL,
			timestamp  TEXT NOT NULL,
			metadata   TEXT NOT NULL DEFAULT '{}'
		);`,
		`CREATE INDEX IF NOT EXISTS idx_entries_account ON ledger_entries (account_id);`,
		`CREATE INDEX IF NOT EXISTS idx_entries_event ON ledger_entries (event_id);`,
		`CREATE TABLE IF NOT EXISTS journal_lines (
			seq              INTEGER PRIMARY KEY AUTOINCREMENT,
			id               TEXT NOT NULL UNIQUE,
			event_id         TEXT NOT NULL,
			timestamp        TEXT NOT NULL,
			debit_account_id TEXT NOT NULL,
			credit_account_id TEXT NOT NULL,
			amount           INTEGER NOT NULL,
			metadata         TEXT NOT NULL DEFAULT '{}',
			debit_metadata   TEXT NOT NULL DEFAULT '{}',
			credit_metadata  TEXT NOT NULL DEFAULT '{}'
		);`,
		`CREATE INDEX IF NOT EXISTS idx_lines_event ON journal_lines (event_id);`,
		`CREATE INDEX IF NOT EXISTS idx_lines_debit_account ON journal_lines (debit_account_id);`,
		`CREATE INDEX IF NOT EXISTS idx_lines_credit_account ON journal_lines (credit_account_id);`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// ── Providers ────────────────────────────────────────────────────────────────

func (s *Store) SaveProvider(p engine.Provider) error {
	_, err := s.db.Exec(
		`INSERT INTO providers (id, name) VALUES (?, ?)
		 ON CONFLICT(id) DO UPDATE SET name = excluded.name`,
		p.ID, p.Name,
	)
	return err
}

func (s *Store) GetProvider(id string) (engine.Provider, error) {
	var p engine.Provider
	err := s.db.QueryRow(`SELECT id, name FROM providers WHERE id = ?`, id).Scan(&p.ID, &p.Name)
	if err == sql.ErrNoRows {
		return engine.Provider{}, fmt.Errorf("provider %q not found", id)
	}
	return p, err
}

func (s *Store) ListProviders() ([]engine.Provider, error) {
	rows, err := s.db.Query(`SELECT id, name FROM providers ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []engine.Provider
	for rows.Next() {
		var p engine.Provider
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ── Accounts ─────────────────────────────────────────────────────────────────

func (s *Store) SaveAccount(a engine.Account) error {
	_, err := s.db.Exec(
		`INSERT INTO accounts
			(id, type, provider_id, currency_code, currency_scale,
			 user_id, counterparty_id, liquidity_account_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			provider_id = excluded.provider_id,
			currency_code = excluded.currency_code,
			currency_scale = excluded.currency_scale,
			user_id = excluded.user_id,
			counterparty_id = excluded.counterparty_id,
			liquidity_account_id = excluded.liquidity_account_id`,
		a.ID, int(a.Type), a.ProviderID, a.Currency.Code, a.Currency.AssetScale,
		a.UserID, a.CounterpartyID, a.LiquidityAccountID,
	)
	return err
}

const accountCols = `id, type, provider_id, currency_code, currency_scale,
	user_id, counterparty_id, liquidity_account_id`

func scanAccount(row interface{ Scan(...any) error }) (engine.Account, error) {
	var (
		a     engine.Account
		tp    int
		scale int
	)
	err := row.Scan(&a.ID, &tp, &a.ProviderID, &a.Currency.Code, &scale,
		&a.UserID, &a.CounterpartyID, &a.LiquidityAccountID)
	if err != nil {
		return engine.Account{}, err
	}
	a.Type = engine.AccountType(tp)
	a.Currency.AssetScale = scale
	return a, nil
}

func (s *Store) GetAccount(id string) (engine.Account, error) {
	row := s.db.QueryRow(`SELECT `+accountCols+` FROM accounts WHERE id = ?`, id)
	a, err := scanAccount(row)
	if err == sql.ErrNoRows {
		return engine.Account{}, fmt.Errorf("account %q not found", id)
	}
	return a, err
}

func (s *Store) ListAccounts() ([]engine.Account, error) {
	rows, err := s.db.Query(`SELECT ` + accountCols + ` FROM accounts ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []engine.Account
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) findAccount(where string, args ...any) (engine.Account, bool, error) {
	row := s.db.QueryRow(`SELECT `+accountCols+` FROM accounts WHERE `+where+` LIMIT 1`, args...)
	a, err := scanAccount(row)
	if err == sql.ErrNoRows {
		return engine.Account{}, false, nil
	}
	if err != nil {
		return engine.Account{}, false, err
	}
	return a, true, nil
}

func (s *Store) FindSystemAccount(providerID string, currency engine.Currency) (engine.Account, bool, error) {
	return s.findAccount(
		`type = ? AND provider_id = ? AND currency_code = ?`,
		int(engine.AccountTypeSystem), providerID, currency.Code,
	)
}

func (s *Store) FindLiquidityAccount(providerID string, currency engine.Currency) (engine.Account, bool, error) {
	return s.findAccount(
		`type = ? AND provider_id = ? AND currency_code = ?`,
		int(engine.AccountTypeLiquidity), providerID, currency.Code,
	)
}

func (s *Store) FindPositionAccount(liquidityAccountID, counterpartyID string) (engine.Account, bool, error) {
	return s.findAccount(
		`type = ? AND liquidity_account_id = ? AND counterparty_id = ?`,
		int(engine.AccountTypePosition), liquidityAccountID, counterpartyID,
	)
}

func (s *Store) FindUserAccount(userID, providerID string, currency engine.Currency) (engine.Account, bool, error) {
	return s.findAccount(
		`type = ? AND user_id = ? AND provider_id = ? AND currency_code = ?`,
		int(engine.AccountTypeUser), userID, providerID, currency.Code,
	)
}

// ── Ledger ───────────────────────────────────────────────────────────────────

func (s *Store) PostEntries(entries []engine.LedgerEntry) error {
	if len(entries) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO ledger_entries
		(id, account_id, amount, type, event_id, timestamp, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer func() { _ = stmt.Close() }()
	for _, e := range entries {
		meta, err := marshalMetadata(e.Metadata)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if _, err := stmt.Exec(
			e.ID, e.AccountID, e.Amount, int(e.Type), e.EventID,
			e.Timestamp.UTC().Format(time.RFC3339Nano), meta,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

const entryCols = `id, account_id, amount, type, event_id, timestamp, metadata`

func scanEntry(row interface{ Scan(...any) error }) (engine.LedgerEntry, error) {
	var (
		e        engine.LedgerEntry
		tp       int
		ts, meta string
	)
	if err := row.Scan(&e.ID, &e.AccountID, &e.Amount, &tp, &e.EventID, &ts, &meta); err != nil {
		return engine.LedgerEntry{}, err
	}
	e.Type = engine.EntryType(tp)
	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return engine.LedgerEntry{}, fmt.Errorf("parse timestamp: %w", err)
	}
	e.Timestamp = parsed
	e.Metadata, err = unmarshalMetadata(meta)
	if err != nil {
		return engine.LedgerEntry{}, err
	}
	return e, nil
}

func (s *Store) queryEntries(where string, args ...any) ([]engine.LedgerEntry, error) {
	q := `SELECT ` + entryCols + ` FROM ledger_entries`
	if where != "" {
		q += ` WHERE ` + where
	}
	q += ` ORDER BY seq`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []engine.LedgerEntry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

const lineCols = `id, event_id, timestamp, debit_account_id, credit_account_id,
	amount, metadata, debit_metadata, credit_metadata`

func scanLine(row interface{ Scan(...any) error }) (engine.JournalLine, error) {
	var (
		l                               engine.JournalLine
		ts, meta, debitMeta, creditMeta string
	)
	if err := row.Scan(
		&l.ID, &l.EventID, &ts, &l.DebitAccountID, &l.CreditAccountID,
		&l.Amount, &meta, &debitMeta, &creditMeta,
	); err != nil {
		return engine.JournalLine{}, err
	}
	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return engine.JournalLine{}, fmt.Errorf("parse timestamp: %w", err)
	}
	l.Timestamp = parsed
	l.Metadata, err = unmarshalMetadata(meta)
	if err != nil {
		return engine.JournalLine{}, err
	}
	l.DebitMetadata, err = unmarshalMetadata(debitMeta)
	if err != nil {
		return engine.JournalLine{}, err
	}
	l.CreditMetadata, err = unmarshalMetadata(creditMeta)
	if err != nil {
		return engine.JournalLine{}, err
	}
	return l, nil
}

func (s *Store) queryLines(where string, args ...any) ([]engine.JournalLine, error) {
	q := `SELECT ` + lineCols + ` FROM journal_lines`
	if where != "" {
		q += ` WHERE ` + where
	}
	q += ` ORDER BY seq`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []engine.JournalLine
	for rows.Next() {
		l, err := scanLine(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) GetEntriesByAccount(accountID string) ([]engine.LedgerEntry, error) {
	lines, err := s.queryLines(`debit_account_id = ? OR credit_account_id = ?`, accountID, accountID)
	if err != nil {
		return nil, err
	}
	legacy, err := s.queryEntries(`account_id = ?`, accountID)
	if err != nil {
		return nil, err
	}
	out := make([]engine.LedgerEntry, 0, len(lines)*2+len(legacy))
	out = append(out, engine.ExpandJournalLines(lines, accountID)...)
	out = append(out, legacy...)
	return out, nil
}

func (s *Store) GetEntriesByEvent(eventID string) ([]engine.LedgerEntry, error) {
	lines, err := s.queryLines(`event_id = ?`, eventID)
	if err != nil {
		return nil, err
	}
	legacy, err := s.queryEntries(`event_id = ?`, eventID)
	if err != nil {
		return nil, err
	}
	out := make([]engine.LedgerEntry, 0, len(lines)*2+len(legacy))
	out = append(out, engine.ExpandJournalLines(lines, "")...)
	out = append(out, legacy...)
	return out, nil
}

func (s *Store) GetAllEntries() ([]engine.LedgerEntry, error) {
	lines, err := s.queryLines("")
	if err != nil {
		return nil, err
	}
	legacy, err := s.queryEntries("")
	if err != nil {
		return nil, err
	}
	out := make([]engine.LedgerEntry, 0, len(lines)*2+len(legacy))
	out = append(out, engine.ExpandJournalLines(lines, "")...)
	out = append(out, legacy...)
	return out, nil
}

// Reset deletes every provider, account, and ledger entry. Used by the
// "Clear Everything" UI action.
func (s *Store) Reset() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, stmt := range []string{
		`DELETE FROM journal_lines`,
		`DELETE FROM ledger_entries`,
		`DELETE FROM accounts`,
		`DELETE FROM providers`,
		`DELETE FROM sqlite_sequence WHERE name='journal_lines'`,
		`DELETE FROM sqlite_sequence WHERE name='ledger_entries'`,
	} {
		if _, err := tx.Exec(stmt); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func marshalMetadata(m map[string]string) (string, error) {
	if len(m) == 0 {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func unmarshalMetadata(s string) (map[string]string, error) {
	if s == "" || s == "{}" {
		return nil, nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, err
	}
	return m, nil
}
