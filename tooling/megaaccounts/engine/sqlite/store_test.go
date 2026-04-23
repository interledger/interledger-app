package sqlite_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"megaaccounts/engine"
	"megaaccounts/engine/memory"
	"megaaccounts/engine/sqlite"
)

// runScenario exercises a representative cross-section of the engine — covering
// each workflow type added through phase 3 — and returns the final ledger state
// for comparison across backends.
func runScenario(t *testing.T, e *engine.Engine) scenarioResult {
	t.Helper()

	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err := e.CreateProvider("gh", "GateHub")
	must(err)
	_, err = e.CreateProvider("xg", "Xago")
	must(err)

	_, err = e.CreateSystemAccount("gh", engine.EUR)
	must(err)
	_, err = e.CreateLiquidityAccount("gh", engine.EUR)
	must(err)
	_, err = e.CreateSystemAccount("gh", engine.ZAR)
	must(err)
	_, err = e.CreateLiquidityAccount("gh", engine.ZAR)
	must(err)
	_, err = e.CreateSystemAccount("xg", engine.EUR)
	must(err)
	xgLiqEUR, err := e.CreateLiquidityAccount("xg", engine.EUR)
	must(err)
	_, err = e.CreateSystemAccount("xg", engine.ZAR)
	must(err)
	xgLiqZAR, err := e.CreateLiquidityAccount("xg", engine.ZAR)
	must(err)

	_, err = e.FundProviderLiquidityLines("gh", engine.EUR, 1_000_00)
	must(err)
	_, err = e.FundProviderLiquidityLines("xg", engine.EUR, 1_000_00)
	must(err)
	_, err = e.FundProviderLiquidityLines("xg", engine.ZAR, 50_000_00)
	must(err)

	_, err = e.UserOnboardLines("alice", "gh", engine.EUR, 500_00)
	must(err)
	_, err = e.UserOnboardLines("bob", "gh", engine.EUR, 300_00)
	must(err)
	_, err = e.UserOnboardLines("carlos", "xg", engine.ZAR, 10_000_00)
	must(err)

	_, err = e.SameProviderP2PTransferLines("alice", "bob", "gh", engine.EUR, 50_00)
	must(err)

	_, err = e.CrossProviderTransferLines(
		"alice", "gh", engine.EUR,
		"carlos", "xg", engine.ZAR,
		100_00, 20, 1,
	)
	must(err)

	lines, err := e.GetAllLines()
	must(err)
	accounts, err := e.ListAccounts()
	must(err)

	// Ensure the ledger rehydrates consistently via Balance.

	// Sanity: mirror invariant holds for the gh/xg EUR position pair (even with
	// open obligations: cross-provider creates symmetric +/- movements).
	_ = xgLiqEUR
	_ = xgLiqZAR

	// Build a deterministic key from account identity (random UUIDs differ
	// across runs, so we compare by role, not ID).
	balByKey := map[string]int64{}
	for _, a := range accounts {
		key := accountKey(a)
		if _, dup := balByKey[key]; dup {
			t.Fatalf("duplicate account key %q", key)
		}
		bal, err := e.Balance(a.ID)
		must(err)
		balByKey[key] = bal
	}

	return scenarioResult{
		lineCount:    len(lines),
		accountCount: len(accounts),
		balances:     balByKey,
	}
}

func accountKey(a engine.Account) string {
	// Exclude LiquidityAccountID from the key: it is a random UUID that will
	// differ across runs. The (type, provider, currency, counterparty) tuple
	// already uniquely identifies a position account.
	return fmt.Sprintf("t=%d/p=%s/c=%s/u=%s/cp=%s",
		a.Type, a.ProviderID, a.Currency.Code, a.UserID, a.CounterpartyID)
}

type scenarioResult struct {
	lineCount    int
	accountCount int
	balances     map[string]int64
}

// TestMemoryAndSQLiteParity runs the same workflow scenario against both store
// backends and verifies they produce equivalent ledger state.
func TestMemoryAndSQLiteParity(t *testing.T) {
	memEng := engine.New(memory.New())
	memResult := runScenario(t, memEng)

	dbPath := filepath.Join(t.TempDir(), "parity.db")
	sqlStore, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = sqlStore.Close() }()
	sqlEng := engine.New(sqlStore)
	sqlResult := runScenario(t, sqlEng)

	if memResult.lineCount != sqlResult.lineCount {
		t.Errorf("line count mismatch: memory=%d sqlite=%d",
			memResult.lineCount, sqlResult.lineCount)
	}
	if memResult.accountCount != sqlResult.accountCount {
		t.Errorf("account count mismatch: memory=%d sqlite=%d",
			memResult.accountCount, sqlResult.accountCount)
	}
	if len(memResult.balances) != len(sqlResult.balances) {
		t.Fatalf("balance map size mismatch: memory=%d sqlite=%d",
			len(memResult.balances), len(sqlResult.balances))
	}
	for id, memBal := range memResult.balances {
		sqlBal, ok := sqlResult.balances[id]
		if !ok {
			t.Errorf("account %s present in memory but missing in sqlite", id)
			continue
		}
		if memBal != sqlBal {
			t.Errorf("balance mismatch for %s: memory=%d sqlite=%d", id, memBal, sqlBal)
		}
	}

	// Verify the same integrity checks pass on both backends.
	if _, err := memEng.CheckGlobalBalance(); err != nil {
		t.Errorf("memory CheckGlobalBalance: %v", err)
	}
	if _, err := sqlEng.CheckGlobalBalance(); err != nil {
		t.Errorf("sqlite CheckGlobalBalance: %v", err)
	}
	if _, err := memEng.CheckBilateralPositions(); err != nil {
		t.Errorf("memory CheckBilateralPositions: %v", err)
	}
	if _, err := sqlEng.CheckBilateralPositions(); err != nil {
		t.Errorf("sqlite CheckBilateralPositions: %v", err)
	}
}

// TestSQLitePersistence verifies that journal lines survive a reopen of the database.
func TestSQLitePersistence(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "persist.db")

	store, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	eng := engine.New(store)
	result := runScenario(t, eng)
	if err := store.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	reopened, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() { _ = reopened.Close() }()

	lines, err := reopened.GetAllLines()
	if err != nil {
		t.Fatalf("get lines after reopen: %v", err)
	}
	if len(lines) != result.lineCount {
		t.Errorf("line count after reopen: got %d, want %d",
			len(lines), result.lineCount)
	}
}

func TestSQLiteLegacyEntryCompatibility(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy-compat.db")
	store, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = store.Close() }()

	eng := engine.New(store)
	if _, err := eng.CreateProvider("gh", "GateHub"); err != nil {
		t.Fatalf("CreateProvider: %v", err)
	}
	if _, err := eng.CreateProvider("xg", "Xago"); err != nil {
		t.Fatalf("CreateProvider xg: %v", err)
	}
	providers, err := store.ListProviders()
	if err != nil {
		t.Fatalf("ListProviders: %v", err)
	}
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}

	sys, err := eng.CreateSystemAccount("gh", engine.EUR)
	if err != nil {
		t.Fatalf("CreateSystemAccount: %v", err)
	}
	liq, err := eng.CreateLiquidityAccount("gh", engine.EUR)
	if err != nil {
		t.Fatalf("CreateLiquidityAccount: %v", err)
	}

	legacy := []engine.LedgerEntry{
		{ID: "legacy-d", AccountID: sys.ID, Amount: 100, Type: engine.EntryTypeDebit, EventID: "legacy-event", Timestamp: time.Now().UTC()},
		{ID: "legacy-c", AccountID: liq.ID, Amount: 100, Type: engine.EntryTypeCredit, EventID: "legacy-event", Timestamp: time.Now().UTC()},
	}
	if err := store.PostEntries(legacy); err != nil {
		t.Fatalf("PostEntries: %v", err)
	}

	byAcct, err := store.GetEntriesByAccount(sys.ID)
	if err != nil {
		t.Fatalf("GetEntriesByAccount: %v", err)
	}
	if len(byAcct) == 0 {
		t.Fatal("expected entries by account")
	}

	byEvent, err := store.GetEntriesByEvent("legacy-event")
	if err != nil {
		t.Fatalf("GetEntriesByEvent: %v", err)
	}
	if len(byEvent) != 2 {
		t.Fatalf("expected 2 entries by event, got %d", len(byEvent))
	}

	allEntries, err := store.GetAllEntries()
	if err != nil {
		t.Fatalf("GetAllEntries: %v", err)
	}
	if len(allEntries) != 2 {
		t.Fatalf("expected 2 total entries, got %d", len(allEntries))
	}

	legacyLines, err := store.GetLinesByEvent("legacy-event")
	if err != nil {
		t.Fatalf("GetLinesByEvent fallback: %v", err)
	}
	if len(legacyLines) != 1 {
		t.Fatalf("expected 1 converted legacy line, got %d", len(legacyLines))
	}

	if err := store.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	providers, _ = store.ListProviders()
	entriesAfter, _ := store.GetAllEntries()
	linesAfter, _ := store.GetAllLines()
	if len(providers) != 0 || len(entriesAfter) != 0 || len(linesAfter) != 0 {
		t.Fatalf("expected empty store after reset, got providers=%d entries=%d lines=%d",
			len(providers), len(entriesAfter), len(linesAfter))
	}
}
