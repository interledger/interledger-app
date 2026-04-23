package memory_test

import (
	"testing"
	"time"

	"megaaccounts/engine"
	"megaaccounts/engine/memory"
)

var (
	eur = engine.Currency{Code: "EUR", AssetScale: 2}
)

func mustProvider(t *testing.T, s *memory.Store, id, name string) {
	t.Helper()
	if err := s.SaveProvider(engine.Provider{ID: id, Name: name}); err != nil {
		t.Fatalf("SaveProvider: %v", err)
	}
}

func mustAccount(t *testing.T, s *memory.Store, a engine.Account) {
	t.Helper()
	if err := s.SaveAccount(a); err != nil {
		t.Fatalf("SaveAccount: %v", err)
	}
}

func setupAccounts(t *testing.T, s *memory.Store) (sys, liq, pos, user engine.Account) {
	t.Helper()
	mustProvider(t, s, "gh", "GateHub")
	mustProvider(t, s, "xg", "Xago")

	sys = engine.Account{ID: "sys-gh-eur", Type: engine.AccountTypeSystem, ProviderID: "gh", Currency: eur}
	liq = engine.Account{ID: "liq-gh-eur", Type: engine.AccountTypeLiquidity, ProviderID: "gh", Currency: eur}
	pos = engine.Account{ID: "pos-gh-xg-eur", Type: engine.AccountTypePosition, ProviderID: "gh", Currency: eur, LiquidityAccountID: liq.ID, CounterpartyID: "xg"}
	user = engine.Account{ID: "user-alice-gh-eur", Type: engine.AccountTypeUser, ProviderID: "gh", Currency: eur, UserID: "alice"}

	mustAccount(t, s, sys)
	mustAccount(t, s, liq)
	mustAccount(t, s, pos)
	mustAccount(t, s, user)
	return sys, liq, pos, user
}

func TestMemoryStoreProviderAccountLookups(t *testing.T) {
	s := memory.New()
	sys, liq, pos, user := setupAccounts(t, s)

	if _, err := s.GetProvider("gh"); err != nil {
		t.Fatalf("GetProvider: %v", err)
	}
	providers, err := s.ListProviders()
	if err != nil {
		t.Fatalf("ListProviders: %v", err)
	}
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}

	if _, err := s.GetAccount(sys.ID); err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	accounts, err := s.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts: %v", err)
	}
	if len(accounts) != 4 {
		t.Fatalf("expected 4 accounts, got %d", len(accounts))
	}

	if got, ok, err := s.FindSystemAccount("gh", eur); err != nil || !ok || got.ID != sys.ID {
		t.Fatalf("FindSystemAccount failed: got=%+v ok=%v err=%v", got, ok, err)
	}
	if got, ok, err := s.FindLiquidityAccount("gh", eur); err != nil || !ok || got.ID != liq.ID {
		t.Fatalf("FindLiquidityAccount failed: got=%+v ok=%v err=%v", got, ok, err)
	}
	if got, ok, err := s.FindPositionAccount(liq.ID, "xg"); err != nil || !ok || got.ID != pos.ID {
		t.Fatalf("FindPositionAccount failed: got=%+v ok=%v err=%v", got, ok, err)
	}
	if got, ok, err := s.FindUserAccount("alice", "gh", eur); err != nil || !ok || got.ID != user.ID {
		t.Fatalf("FindUserAccount failed: got=%+v ok=%v err=%v", got, ok, err)
	}
}

func TestMemoryStoreLinePrimaryWithLegacyCompatibility(t *testing.T) {
	s := memory.New()
	sys, liq, _, user := setupAccounts(t, s)

	line := engine.JournalLine{
		ID:              "line-1",
		EventID:         "event-1",
		Timestamp:       time.Now().UTC(),
		DebitAccountID:  sys.ID,
		CreditAccountID: liq.ID,
		Amount:          1000,
		Metadata:        map[string]string{"workflow": "test"},
	}
	if err := s.PostLines([]engine.JournalLine{line}); err != nil {
		t.Fatalf("PostLines: %v", err)
	}

	lines, err := s.GetAllLines()
	if err != nil {
		t.Fatalf("GetAllLines: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].ID != "line-1" {
		t.Fatalf("unexpected line id %q", lines[0].ID)
	}

	byEvent, err := s.GetLinesByEvent("event-1")
	if err != nil {
		t.Fatalf("GetLinesByEvent: %v", err)
	}
	if len(byEvent) != 1 {
		t.Fatalf("expected 1 line by event, got %d", len(byEvent))
	}

	byAccount, err := s.GetLinesByAccount(sys.ID)
	if err != nil {
		t.Fatalf("GetLinesByAccount: %v", err)
	}
	if len(byAccount) != 1 {
		t.Fatalf("expected 1 line by account, got %d", len(byAccount))
	}

	legacyEntries := []engine.LedgerEntry{
		{ID: "legacy-d", AccountID: user.ID, Amount: 250, Type: engine.EntryTypeDebit, EventID: "event-legacy", Timestamp: time.Now().UTC()},
		{ID: "legacy-c", AccountID: sys.ID, Amount: 250, Type: engine.EntryTypeCredit, EventID: "event-legacy", Timestamp: time.Now().UTC()},
	}
	if err := s.PostEntries(legacyEntries); err != nil {
		t.Fatalf("PostEntries: %v", err)
	}

	allEntries, err := s.GetAllEntries()
	if err != nil {
		t.Fatalf("GetAllEntries: %v", err)
	}
	if len(allEntries) != 4 {
		t.Fatalf("expected 4 total entries (2 expanded + 2 legacy), got %d", len(allEntries))
	}

	bySys, err := s.GetEntriesByAccount(sys.ID)
	if err != nil {
		t.Fatalf("GetEntriesByAccount: %v", err)
	}
	if len(bySys) != 2 {
		t.Fatalf("expected 2 entries for system account, got %d", len(bySys))
	}

	allLines, err := s.GetAllLines()
	if err != nil {
		t.Fatalf("GetAllLines after legacy: %v", err)
	}
	if len(allLines) != 2 {
		t.Fatalf("expected 2 total lines (1 primary + 1 legacy-converted), got %d", len(allLines))
	}

	legacyByEvent, err := s.GetLinesByEvent("event-legacy")
	if err != nil {
		t.Fatalf("GetLinesByEvent legacy: %v", err)
	}
	if len(legacyByEvent) != 1 {
		t.Fatalf("expected 1 converted legacy line, got %d", len(legacyByEvent))
	}
}

func TestMemoryStoreReset(t *testing.T) {
	s := memory.New()
	sys, liq, _, _ := setupAccounts(t, s)
	if err := s.PostLines([]engine.JournalLine{{
		ID:              "line-r",
		EventID:         "event-r",
		Timestamp:       time.Now().UTC(),
		DebitAccountID:  sys.ID,
		CreditAccountID: liq.ID,
		Amount:          1,
	}}); err != nil {
		t.Fatalf("PostLines: %v", err)
	}
	if err := s.PostEntries([]engine.LedgerEntry{{
		ID: "legacy-r", AccountID: sys.ID, Amount: 1, Type: engine.EntryTypeDebit, EventID: "event-r-legacy", Timestamp: time.Now().UTC(),
	}}); err != nil {
		t.Fatalf("PostEntries: %v", err)
	}

	if err := s.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	providers, _ := s.ListProviders()
	accounts, _ := s.ListAccounts()
	lines, _ := s.GetAllLines()
	entries, _ := s.GetAllEntries()
	if len(providers) != 0 || len(accounts) != 0 || len(lines) != 0 || len(entries) != 0 {
		t.Fatalf("expected fully reset store, got providers=%d accounts=%d lines=%d entries=%d",
			len(providers), len(accounts), len(lines), len(entries))
	}
}
