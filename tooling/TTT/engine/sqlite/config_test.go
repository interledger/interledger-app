package sqlite_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"ttt/engine"
	"ttt/engine/sqlite"
)

func TestSQLiteConfigParadigmLifecycle(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "config.db")
	store, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = store.Close() }()

	set, err := store.IsParadigmSet()
	if err != nil {
		t.Fatalf("IsParadigmSet: %v", err)
	}
	if set {
		t.Fatalf("new DB should not have paradigm set")
	}

	if _, err := store.GetParadigm(); err == nil {
		t.Fatalf("expected error when config paradigm is missing")
	}

	if err := store.SetParadigm(engine.Paradigm(999)); err == nil {
		t.Fatalf("expected error for invalid paradigm in SetParadigm")
	}

	if err := store.SetParadigm(engine.ParadigmPOSTwo); err != nil {
		t.Fatalf("SetParadigm: %v", err)
	}

	set, err = store.IsParadigmSet()
	if err != nil {
		t.Fatalf("IsParadigmSet after set: %v", err)
	}
	if !set {
		t.Fatalf("expected paradigm to be marked as set")
	}

	p, err := store.GetParadigm()
	if err != nil {
		t.Fatalf("GetParadigm: %v", err)
	}
	if p != engine.ParadigmPOSTwo {
		t.Fatalf("expected ParadigmPOSTwo, got %v", p)
	}

	if err := store.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	// Reset should keep config choice intact.
	p, err = store.GetParadigm()
	if err != nil {
		t.Fatalf("GetParadigm after reset: %v", err)
	}
	if p != engine.ParadigmPOSTwo {
		t.Fatalf("expected ParadigmPOSTwo after reset, got %v", p)
	}
}

func TestSQLiteConfigParadigmInvalidStoredValue(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "config-invalid.db")
	store, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = store.Close() }()

	raw, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open raw sqlite: %v", err)
	}
	defer func() { _ = raw.Close() }()

	if _, err := raw.Exec(`INSERT INTO config (key, value) VALUES ('paradigm', 'not-an-int')`); err != nil {
		t.Fatalf("insert invalid config: %v", err)
	}
	if _, err := store.GetParadigm(); err == nil {
		t.Fatalf("expected parse error for non-int paradigm value")
	}

	if _, err := raw.Exec(`UPDATE config SET value = '999' WHERE key = 'paradigm'`); err != nil {
		t.Fatalf("update invalid numeric config: %v", err)
	}
	if _, err := store.GetParadigm(); err == nil {
		t.Fatalf("expected validation error for unknown paradigm value")
	}
}
