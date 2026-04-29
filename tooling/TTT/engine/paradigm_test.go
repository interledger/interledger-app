package engine_test

import (
	"testing"

	"ttt/engine"
)

func hasAccount(accounts []engine.Account, typ engine.AccountType, providerID, currency string, userID string) bool {
	for _, a := range accounts {
		if a.Type != typ {
			continue
		}
		if a.ProviderID != providerID {
			continue
		}
		if a.Currency.Code != currency {
			continue
		}
		if userID != "" && a.UserID != userID {
			continue
		}
		return true
	}
	return false
}

func TestParadigm_Metadata(t *testing.T) {
	if engine.ParadigmStandard != engine.ParadigmPOSTwo {
		t.Fatalf("ParadigmStandard must alias ParadigmPOSTwo")
	}

	if !engine.ParadigmPOSTwo.IsValid() {
		t.Fatalf("ParadigmPOSTwo should be valid")
	}
	if !engine.ParadigmSingleGHEUR.IsValid() {
		t.Fatalf("ParadigmSingleGHEUR should be valid")
	}
	if engine.Paradigm(999).IsValid() {
		t.Fatalf("unknown paradigm should be invalid")
	}

	if got := engine.ParadigmPOSTwo.Name(); got == "" {
		t.Fatalf("Name() should be non-empty")
	}
	if got := engine.ParadigmSingleGHEUR.Name(); got == "" {
		t.Fatalf("Name() should be non-empty")
	}
	if got := engine.Paradigm(999).Name(); got == "" {
		t.Fatalf("Name() for unknown paradigm should still be non-empty")
	}
}

func TestSeedParadigm_POSTwo(t *testing.T) {
	e := newEngine()

	if err := engine.SeedParadigm(engine.ParadigmPOSTwo, e); err != nil {
		t.Fatalf("SeedParadigm POSTwo: %v", err)
	}
	// Idempotency: should tolerate already-existing records.
	if err := engine.SeedParadigm(engine.ParadigmPOSTwo, e); err != nil {
		t.Fatalf("SeedParadigm POSTwo second run: %v", err)
	}

	providers, err := e.ListProviders()
	if err != nil {
		t.Fatalf("ListProviders: %v", err)
	}
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}

	accounts, err := e.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts: %v", err)
	}

	checks := []struct {
		typ      engine.AccountType
		provider string
		currency string
		userID   string
	}{
		{engine.AccountTypeSystem, "gatehub", "EUR", ""},
		{engine.AccountTypeLiquidity, "gatehub", "EUR", ""},
		{engine.AccountTypeSystem, "xago", "EUR", ""},
		{engine.AccountTypeLiquidity, "xago", "EUR", ""},
		{engine.AccountTypeSystem, "xago", "ZAR", ""},
		{engine.AccountTypeLiquidity, "xago", "ZAR", ""},
		{engine.AccountTypeUser, "gatehub", "EUR", "alice"},
		{engine.AccountTypeUser, "gatehub", "EUR", "bob"},
		{engine.AccountTypeUser, "xago", "ZAR", "carlos"},
	}
	for _, c := range checks {
		if !hasAccount(accounts, c.typ, c.provider, c.currency, c.userID) {
			t.Fatalf("missing account: type=%v provider=%s currency=%s user=%s", c.typ, c.provider, c.currency, c.userID)
		}
	}
}

func TestSeedParadigm_SingleGHEUR(t *testing.T) {
	e := newEngine()

	if err := engine.SeedParadigm(engine.ParadigmSingleGHEUR, e); err != nil {
		t.Fatalf("SeedParadigm SingleGHEUR: %v", err)
	}
	if err := engine.SeedParadigm(engine.ParadigmSingleGHEUR, e); err != nil {
		t.Fatalf("SeedParadigm SingleGHEUR second run: %v", err)
	}

	accounts, err := e.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts: %v", err)
	}

	if !hasAccount(accounts, engine.AccountTypeSystem, "gatehub", "EUR", "") {
		t.Fatalf("expected gatehub EUR system account")
	}
	if !hasAccount(accounts, engine.AccountTypeLiquidity, "gatehub", "EUR", "") {
		t.Fatalf("expected gatehub EUR liquidity account")
	}
	if !hasAccount(accounts, engine.AccountTypeSystem, "xago", "ZAR", "") {
		t.Fatalf("expected xago ZAR system account")
	}
	if !hasAccount(accounts, engine.AccountTypeLiquidity, "xago", "ZAR", "") {
		t.Fatalf("expected xago ZAR liquidity account")
	}

	if hasAccount(accounts, engine.AccountTypeSystem, "xago", "EUR", "") {
		t.Fatalf("did not expect xago EUR system account in single GateHub paradigm")
	}
	if hasAccount(accounts, engine.AccountTypeLiquidity, "xago", "EUR", "") {
		t.Fatalf("did not expect xago EUR liquidity account in single GateHub paradigm")
	}
}

func TestSeedParadigm_SelfExchange(t *testing.T) {
	e := newEngine()

	if err := engine.SeedParadigm(engine.ParadigmSelfExchange, e); err != nil {
		t.Fatalf("SeedParadigm SelfExchange: %v", err)
	}
	// Idempotency.
	if err := engine.SeedParadigm(engine.ParadigmSelfExchange, e); err != nil {
		t.Fatalf("SeedParadigm SelfExchange second run: %v", err)
	}

	accounts, err := e.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts: %v", err)
	}

	checks := []struct {
		typ      engine.AccountType
		provider string
		currency string
		userID   string
	}{
		{engine.AccountTypeSystem, "gatehub", "EUR", ""},
		{engine.AccountTypeLiquidity, "gatehub", "EUR", ""},
		{engine.AccountTypeSystem, "xago", "ZAR", ""},
		{engine.AccountTypeLiquidity, "xago", "ZAR", ""},
		{engine.AccountTypeFX, "xago", "ZAR", ""},
		{engine.AccountTypeSystem, "xago", "EUR", ""},
		{engine.AccountTypeLiquidity, "xago", "EUR", ""},
		{engine.AccountTypeUser, "gatehub", "EUR", "alice"},
		{engine.AccountTypeUser, "gatehub", "EUR", "bob"},
		{engine.AccountTypeUser, "xago", "ZAR", "carlos"},
	}
	for _, c := range checks {
		if !hasAccount(accounts, c.typ, c.provider, c.currency, c.userID) {
			t.Fatalf("missing account: type=%v provider=%s currency=%s user=%s", c.typ, c.provider, c.currency, c.userID)
		}
	}
}

func TestSeedParadigm_Unknown(t *testing.T) {
	e := newEngine()
	if err := engine.SeedParadigm(engine.Paradigm(777), e); err == nil {
		t.Fatalf("expected error for unknown paradigm")
	}
}
