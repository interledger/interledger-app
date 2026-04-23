package sqlite

import (
	"database/sql"
	"fmt"
	"strconv"

	"ttt/engine"
)

const configKeyParadigm = "paradigm"

// IsParadigmSet reports whether a paradigm has been stored in the config table.
func (s *Store) IsParadigmSet() (bool, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM config WHERE key = ?`, configKeyParadigm).Scan(&v)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking paradigm in config: %w", err)
	}
	return true, nil
}

// GetParadigm reads and validates the stored paradigm. Returns an error if the
// config table is missing the paradigm entry or if the stored value is invalid.
func (s *Store) GetParadigm() (engine.Paradigm, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM config WHERE key = ?`, configKeyParadigm).Scan(&v)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("no paradigm set in config table; reinitialise or delete the database")
	}
	if err != nil {
		return 0, fmt.Errorf("reading paradigm from config: %w", err)
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid paradigm value %q in config table: %w", v, err)
	}
	p := engine.Paradigm(n)
	if !p.IsValid() {
		return 0, fmt.Errorf("invalid paradigm %d in config table; reinitialise or delete the database", n)
	}
	return p, nil
}

// SetParadigm writes the paradigm to the config table.
func (s *Store) SetParadigm(p engine.Paradigm) error {
	if !p.IsValid() {
		return fmt.Errorf("SetParadigm: unknown paradigm %d", int(p))
	}
	_, err := s.db.Exec(
		`INSERT INTO config (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		configKeyParadigm, strconv.Itoa(int(p)),
	)
	return err
}
