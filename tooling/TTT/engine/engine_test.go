package engine_test

import (
	"strings"
	"testing"

	"ttt/engine"
	"ttt/engine/memory"
)

// newEngine returns a fresh Engine backed by an empty in-memory store.
func newEngine() *engine.Engine {
	return engine.New(memory.New())
}

// eur is a convenience Currency for tests.
var eur = engine.Currency{Code: "EUR", AssetScale: 2}
var zar = engine.Currency{Code: "ZAR", AssetScale: 2}

// setupProvider creates a provider and returns it; fatals on error.
func setupProvider(t *testing.T, e *engine.Engine, id, name string) engine.Provider {
	t.Helper()
	p, err := e.CreateProvider(id, name)
	if err != nil {
		t.Fatalf("CreateProvider: %v", err)
	}
	return p
}

// setupSystemAccount creates a system account; fatals on error.
func setupSystemAccount(t *testing.T, e *engine.Engine, providerID string, currency engine.Currency) engine.Account {
	t.Helper()
	a, err := e.CreateSystemAccount(providerID, currency)
	if err != nil {
		t.Fatalf("CreateSystemAccount: %v", err)
	}
	return a
}

// setupLiquidityAccount creates a liquidity account; fatals on error.
func setupLiquidityAccount(t *testing.T, e *engine.Engine, providerID string, currency engine.Currency) engine.Account {
	t.Helper()
	a, err := e.CreateLiquidityAccount(providerID, currency)
	if err != nil {
		t.Fatalf("CreateLiquidityAccount: %v", err)
	}
	return a
}

// setupUserAccount creates a user account; fatals on error.
func setupUserAccount(t *testing.T, e *engine.Engine, userID, providerID string, currency engine.Currency) engine.Account {
	t.Helper()
	a, err := e.CreateUserAccount(userID, providerID, currency)
	if err != nil {
		t.Fatalf("CreateUserAccount: %v", err)
	}
	return a
}

// ----- Provider ----------------------------------------------------------------

func TestCreateProvider(t *testing.T) {
	e := newEngine()

	t.Run("valid", func(t *testing.T) {
		p, err := e.CreateProvider("gh", "GateHub")
		if err != nil {
			t.Fatal(err)
		}
		if p.ID != "gh" || p.Name != "GateHub" {
			t.Errorf("unexpected provider %+v", p)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		if _, err := e.CreateProvider("", "GateHub"); err == nil {
			t.Error("expected error for empty id")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		if _, err := e.CreateProvider("p1", ""); err == nil {
			t.Error("expected error for empty name")
		}
	})

	t.Run("duplicate id", func(t *testing.T) {
		if _, err := e.CreateProvider("gh", "GateHub2"); err == nil {
			t.Error("expected error for duplicate id")
		}
	})
}

func TestGetProvider(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")

	t.Run("found", func(t *testing.T) {
		p, err := e.GetProvider("gh")
		if err != nil {
			t.Fatal(err)
		}
		if p.ID != "gh" {
			t.Errorf("unexpected provider id %q", p.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		if _, err := e.GetProvider("missing"); err == nil {
			t.Error("expected error for missing provider")
		}
	})
}

func TestListProviders(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	setupProvider(t, e, "xa", "Xago")

	providers, err := e.ListProviders()
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(providers))
	}
}

// ----- Account creation -------------------------------------------------------

func TestCreateSystemAccount(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")

	t.Run("valid", func(t *testing.T) {
		a, err := e.CreateSystemAccount("gh", eur)
		if err != nil {
			t.Fatal(err)
		}
		if a.Type != engine.AccountTypeSystem || a.ProviderID != "gh" {
			t.Errorf("unexpected account %+v", a)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		if _, err := e.CreateSystemAccount("gh", eur); err == nil {
			t.Error("expected error for duplicate system account")
		}
	})

	t.Run("unknown provider", func(t *testing.T) {
		if _, err := e.CreateSystemAccount("nope", eur); err == nil {
			t.Error("expected error for unknown provider")
		}
	})

	t.Run("invalid currency code", func(t *testing.T) {
		if _, err := e.CreateSystemAccount("gh", engine.Currency{Code: "eur", AssetScale: 2}); err == nil {
			t.Error("expected error for lowercase currency code")
		}
	})

	t.Run("empty currency code", func(t *testing.T) {
		if _, err := e.CreateSystemAccount("gh", engine.Currency{Code: "", AssetScale: 2}); err == nil {
			t.Error("expected error for empty currency code")
		}
	})

	t.Run("currency code too long", func(t *testing.T) {
		if _, err := e.CreateSystemAccount("gh", engine.Currency{Code: "EURZZ", AssetScale: 2}); err == nil {
			t.Error("expected error for currency code > 4 chars")
		}
	})
}

func TestCreateLiquidityAccount(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")

	t.Run("valid", func(t *testing.T) {
		a, err := e.CreateLiquidityAccount("gh", eur)
		if err != nil {
			t.Fatal(err)
		}
		if a.Type != engine.AccountTypeLiquidity {
			t.Errorf("expected AccountTypeLiquidity, got %v", a.Type)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		if _, err := e.CreateLiquidityAccount("gh", eur); err == nil {
			t.Error("expected error for duplicate liquidity account")
		}
	})

	t.Run("unknown provider", func(t *testing.T) {
		if _, err := e.CreateLiquidityAccount("nope", eur); err == nil {
			t.Error("expected error for unknown provider")
		}
	})
}

func TestCreatePositionAccount(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	setupProvider(t, e, "xa", "Xago")
	liq := setupLiquidityAccount(t, e, "gh", eur)

	t.Run("valid", func(t *testing.T) {
		a, err := e.CreatePositionAccount(liq.ID, "xa")
		if err != nil {
			t.Fatal(err)
		}
		if a.Type != engine.AccountTypePosition {
			t.Errorf("expected AccountTypePosition, got %v", a.Type)
		}
		if a.LiquidityAccountID != liq.ID {
			t.Errorf("expected LiquidityAccountID %q, got %q", liq.ID, a.LiquidityAccountID)
		}
		if a.CounterpartyID != "xa" {
			t.Errorf("expected CounterpartyID %q, got %q", "xa", a.CounterpartyID)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		if _, err := e.CreatePositionAccount(liq.ID, "xa"); err == nil {
			t.Error("expected error for duplicate position account")
		}
	})

	t.Run("unknown liquidity account", func(t *testing.T) {
		if _, err := e.CreatePositionAccount("nonexistent-id", "xa"); err == nil {
			t.Error("expected error for unknown liquidity account")
		}
	})

	t.Run("non-liquidity account", func(t *testing.T) {
		sys := setupSystemAccount(t, e, "gh", zar)
		if _, err := e.CreatePositionAccount(sys.ID, "xa"); err == nil {
			t.Error("expected error when referencing non-liquidity account")
		}
	})

	t.Run("unknown counterparty", func(t *testing.T) {
		liq2 := setupLiquidityAccount(t, e, "gh", zar)
		if _, err := e.CreatePositionAccount(liq2.ID, "unknown"); err == nil {
			t.Error("expected error for unknown counterparty provider")
		}
	})
}

func TestCreateUserAccount(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")

	t.Run("valid", func(t *testing.T) {
		a, err := e.CreateUserAccount("user1", "gh", eur)
		if err != nil {
			t.Fatal(err)
		}
		if a.Type != engine.AccountTypeUser || a.UserID != "user1" {
			t.Errorf("unexpected account %+v", a)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		if _, err := e.CreateUserAccount("user1", "gh", eur); err == nil {
			t.Error("expected error for duplicate user account")
		}
	})

	t.Run("empty userID", func(t *testing.T) {
		if _, err := e.CreateUserAccount("", "gh", eur); err == nil {
			t.Error("expected error for empty userID")
		}
	})

	t.Run("unknown provider", func(t *testing.T) {
		if _, err := e.CreateUserAccount("user99", "nope", eur); err == nil {
			t.Error("expected error for unknown provider")
		}
	})
}

func TestGetAccount(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	a := setupSystemAccount(t, e, "gh", eur)

	t.Run("found", func(t *testing.T) {
		got, err := e.GetAccount(a.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != a.ID {
			t.Errorf("expected account %q, got %q", a.ID, got.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		if _, err := e.GetAccount("missing"); err == nil {
			t.Error("expected error for missing account")
		}
	})
}

func TestListAccounts(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	setupSystemAccount(t, e, "gh", eur)
	setupLiquidityAccount(t, e, "gh", eur)

	accounts, err := e.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(accounts))
	}
}

// ----- Balance ----------------------------------------------------------------

func TestBalance(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	setupSystemAccount(t, e, "gh", eur)
	setupUserAccount(t, e, "user1", "gh", eur)

	t.Run("zero before activity", func(t *testing.T) {
		sys, _ := e.GetAccount(mustFindSystemAccount(t, e, "gh", eur).ID)
		bal, err := e.Balance(sys.ID)
		if err != nil {
			t.Fatal(err)
		}
		if bal != 0 {
			t.Errorf("expected 0, got %d", bal)
		}
	})

	t.Run("after onboard", func(t *testing.T) {
		_, err := e.UserOnboard("user1", "gh", eur, 30000)
		if err != nil {
			t.Fatal(err)
		}
		user := mustFindUserAccount(t, e, "user1", "gh", eur)
		bal, err := e.Balance(user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if bal != 30000 {
			t.Errorf("expected 30000, got %d", bal)
		}
	})
}

// ----- FundProviderLiquidity --------------------------------------------------

func TestFundProviderLiquidity(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupSystemAccount(t, e, "gh", eur)
		liq := setupLiquidityAccount(t, e, "gh", eur)

		lines, err := e.FundProviderLiquidityLines("gh", eur, 100000)
		if err != nil {
			t.Fatal(err)
		}
		if len(lines) != 1 {
			t.Fatalf("expected 1 line, got %d", len(lines))
		}
		if lines[0].Amount != 100000 {
			t.Errorf("amounts mismatch")
		}
		if lines[0].DebitAccountID == "" || lines[0].CreditAccountID == "" {
			t.Errorf("line sides must be populated")
		}
		if lines[0].EventID == "" {
			t.Errorf("line event ID must be populated")
		}

		// liquidity balance should be 100000
		bal, err := e.Balance(liq.ID)
		if err != nil {
			t.Fatal(err)
		}
		if bal != 100000 {
			t.Errorf("liquidity balance: want 100000, got %d", bal)
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		if _, err := e.FundProviderLiquidity("gh", eur, 0); err == nil {
			t.Error("expected error for zero amount")
		}
	})

	t.Run("negative amount", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		if _, err := e.FundProviderLiquidity("gh", eur, -1); err == nil {
			t.Error("expected error for negative amount")
		}
	})

	t.Run("missing system account", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupLiquidityAccount(t, e, "gh", eur)
		if _, err := e.FundProviderLiquidity("gh", eur, 1000); err == nil {
			t.Error("expected error when system account missing")
		}
	})

	t.Run("missing liquidity account", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupSystemAccount(t, e, "gh", eur)
		if _, err := e.FundProviderLiquidity("gh", eur, 1000); err == nil {
			t.Error("expected error when liquidity account missing")
		}
	})
}

// ----- UserOnboard ------------------------------------------------------------

func TestUserOnboard(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupSystemAccount(t, e, "gh", eur)
		setupUserAccount(t, e, "user1", "gh", eur)

		lines, err := e.UserOnboardLines("user1", "gh", eur, 30000)
		if err != nil {
			t.Fatal(err)
		}
		if len(lines) != 1 {
			t.Fatalf("expected 1 line, got %d", len(lines))
		}

		user := mustFindUserAccount(t, e, "user1", "gh", eur)
		bal, err := e.Balance(user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if bal != 30000 {
			t.Errorf("user balance: want 30000, got %d", bal)
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		if _, err := e.UserOnboard("user1", "gh", eur, 0); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("missing system account", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupUserAccount(t, e, "user1", "gh", eur)
		if _, err := e.UserOnboard("user1", "gh", eur, 1000); err == nil {
			t.Error("expected error when system account missing")
		}
	})

	t.Run("auto-creates user account when missing", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupSystemAccount(t, e, "gh", eur)
		if _, err := e.UserOnboard("newuser", "gh", eur, 1000); err != nil {
			t.Errorf("expected auto-creation of user account, got error: %v", err)
		}
		accounts, _ := e.ListAccounts()
		var userAcct engine.Account
		for _, a := range accounts {
			if a.UserID == "newuser" {
				userAcct = a
			}
		}
		if userAcct.ID == "" {
			t.Fatal("expected user account to have been created")
		}
		bal, _ := e.Balance(userAcct.ID)
		if bal != 1000 {
			t.Errorf("expected balance 1000, got %d", bal)
		}
	})
}

// ----- SameProviderP2PTransfer ------------------------------------------------

func TestSameProviderP2PTransfer(t *testing.T) {
	setup := func(t *testing.T) (*engine.Engine, engine.Account, engine.Account) {
		t.Helper()
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupSystemAccount(t, e, "gh", eur)
		setupUserAccount(t, e, "user1", "gh", eur)
		setupUserAccount(t, e, "user2", "gh", eur)
		// fund user1 with 50000
		if _, err := e.UserOnboard("user1", "gh", eur, 50000); err != nil {
			t.Fatal(err)
		}
		u1 := mustFindUserAccount(t, e, "user1", "gh", eur)
		u2 := mustFindUserAccount(t, e, "user2", "gh", eur)
		return e, u1, u2
	}

	t.Run("happy path", func(t *testing.T) {
		e, u1, u2 := setup(t)
		lines, err := e.SameProviderP2PTransferLines("user1", "user2", "gh", eur, 10000)
		if err != nil {
			t.Fatal(err)
		}
		if len(lines) != 1 {
			t.Fatalf("expected 1 line, got %d", len(lines))
		}

		bal1, _ := e.Balance(u1.ID)
		bal2, _ := e.Balance(u2.ID)
		if bal1 != 40000 {
			t.Errorf("sender balance: want 40000, got %d", bal1)
		}
		if bal2 != 10000 {
			t.Errorf("recipient balance: want 10000, got %d", bal2)
		}
	})

	t.Run("insufficient balance", func(t *testing.T) {
		e, _, _ := setup(t)
		if _, err := e.SameProviderP2PTransfer("user1", "user2", "gh", eur, 99999); err == nil {
			t.Error("expected insufficient balance error")
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		e, _, _ := setup(t)
		if _, err := e.SameProviderP2PTransfer("user1", "user2", "gh", eur, 0); err == nil {
			t.Error("expected error for zero amount")
		}
	})

	t.Run("same user", func(t *testing.T) {
		e, _, _ := setup(t)
		if _, err := e.SameProviderP2PTransfer("user1", "user1", "gh", eur, 100); err == nil {
			t.Error("expected error for same user transfer")
		}
	})

	t.Run("missing sender", func(t *testing.T) {
		e, _, _ := setup(t)
		if _, err := e.SameProviderP2PTransfer("nouser", "user2", "gh", eur, 100); err == nil {
			t.Error("expected error for missing sender")
		}
	})

	t.Run("missing recipient", func(t *testing.T) {
		e, _, _ := setup(t)
		if _, err := e.SameProviderP2PTransfer("user1", "nouser", "gh", eur, 100); err == nil {
			t.Error("expected error for missing recipient")
		}
	})

	t.Run("no state mutation on failure", func(t *testing.T) {
		e, u1, _ := setup(t)
		balBefore, _ := e.Balance(u1.ID)
		_, _ = e.SameProviderP2PTransfer("user1", "user2", "gh", eur, 99999)
		balAfter, _ := e.Balance(u1.ID)
		if balBefore != balAfter {
			t.Errorf("balance mutated on failed transfer: before=%d after=%d", balBefore, balAfter)
		}
	})
}

// ----- UserOffboard -----------------------------------------------------------

func TestUserOffboard(t *testing.T) {
	setup := func(t *testing.T) (*engine.Engine, engine.Account) {
		t.Helper()
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupSystemAccount(t, e, "gh", eur)
		setupUserAccount(t, e, "user1", "gh", eur)
		if _, err := e.UserOnboard("user1", "gh", eur, 30000); err != nil {
			t.Fatal(err)
		}
		u := mustFindUserAccount(t, e, "user1", "gh", eur)
		return e, u
	}

	t.Run("happy path", func(t *testing.T) {
		e, u := setup(t)
		lines, err := e.UserOffboardLines("user1", "gh", eur, 10000)
		if err != nil {
			t.Fatal(err)
		}
		if len(lines) != 1 {
			t.Fatalf("expected 1 line, got %d", len(lines))
		}
		bal, err := e.Balance(u.ID)
		if err != nil {
			t.Fatal(err)
		}
		if bal != 20000 {
			t.Errorf("user balance after offboard: want 20000, got %d", bal)
		}
		if lines[0].DebitAccountID != u.ID {
			t.Error("line should debit user account")
		}
	})

	t.Run("exact balance", func(t *testing.T) {
		e, u := setup(t)
		_, err := e.UserOffboard("user1", "gh", eur, 30000)
		if err != nil {
			t.Fatal(err)
		}
		bal, _ := e.Balance(u.ID)
		if bal != 0 {
			t.Errorf("expected zero balance after full offboard, got %d", bal)
		}
	})

	t.Run("insufficient balance", func(t *testing.T) {
		e, _ := setup(t)
		if _, err := e.UserOffboard("user1", "gh", eur, 30001); err == nil {
			t.Error("expected insufficient balance error")
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		e, _ := setup(t)
		if _, err := e.UserOffboard("user1", "gh", eur, 0); err == nil {
			t.Error("expected error for zero amount")
		}
	})

	t.Run("missing user account", func(t *testing.T) {
		e, _ := setup(t)
		if _, err := e.UserOffboard("nouser", "gh", eur, 100); err == nil {
			t.Error("expected error for missing user account")
		}
	})

	t.Run("missing system account", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		// create user account only (no system account)
		setupUserAccount(t, e, "user1", "gh", eur)
		if _, err := e.UserOffboard("user1", "gh", eur, 100); err == nil {
			t.Error("expected error for missing system account")
		}
	})
}

// ----- CheckPerEventBalance ---------------------------------------------------

func TestCheckPerEventBalance(t *testing.T) {
	t.Run("balanced event", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupSystemAccount(t, e, "gh", eur)
		setupUserAccount(t, e, "user1", "gh", eur)

		lines, err := e.UserOnboardLines("user1", "gh", eur, 1000)
		if err != nil {
			t.Fatal(err)
		}
		if err := e.CheckPerEventBalance(lines[0].EventID); err != nil {
			t.Errorf("expected balanced event, got: %v", err)
		}
	})

	t.Run("empty event", func(t *testing.T) {
		e := newEngine()
		if err := e.CheckPerEventBalance("nonexistent-event"); err == nil {
			t.Error("expected error for empty event")
		}
	})

	t.Run("unbalanced event via direct store post", func(t *testing.T) {
		store := memory.New()
		e := engine.New(store)
		setupProvider(t, e, "gh", "GateHub")
		setupSystemAccount(t, e, "gh", eur)

		// post an orphan debit with no matching credit
		sys := mustFindSystemAccountViaEngine(t, e, "gh", eur)
		err := store.PostEntries([]engine.LedgerEntry{
			{ID: "orphan-1", AccountID: sys.ID, Amount: 500, Type: engine.EntryTypeDebit, EventID: "bad-event"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := e.CheckPerEventBalance("bad-event"); err == nil {
			t.Error("expected unbalanced event error")
		}
	})
}

// ----- CheckGlobalBalance -----------------------------------------------------

func TestCheckGlobalBalance(t *testing.T) {
	t.Run("no entries returns empty slice", func(t *testing.T) {
		e := newEngine()
		results, err := e.CheckGlobalBalance()
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Errorf("expected empty results, got %v", results)
		}
	})

	t.Run("balanced after multiple workflows", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupSystemAccount(t, e, "gh", eur)
		setupLiquidityAccount(t, e, "gh", eur)
		setupUserAccount(t, e, "user1", "gh", eur)
		setupUserAccount(t, e, "user2", "gh", eur)

		if _, err := e.FundProviderLiquidity("gh", eur, 200000); err != nil {
			t.Fatal(err)
		}
		if _, err := e.UserOnboard("user1", "gh", eur, 50000); err != nil {
			t.Fatal(err)
		}
		if _, err := e.SameProviderP2PTransfer("user1", "user2", "gh", eur, 20000); err != nil {
			t.Fatal(err)
		}
		if _, err := e.UserOffboard("user2", "gh", eur, 10000); err != nil {
			t.Fatal(err)
		}

		results, err := e.CheckGlobalBalance()
		if err != nil {
			t.Fatalf("global balance check failed: %v", err)
		}
		for _, r := range results {
			if r.Net != 0 {
				t.Errorf("currency %q has non-zero net %d", r.Currency, r.Net)
			}
		}
	})

	t.Run("unbalanced via direct store post", func(t *testing.T) {
		store := memory.New()
		e := engine.New(store)
		setupProvider(t, e, "gh", "GateHub")
		setupSystemAccount(t, e, "gh", eur)

		sys := mustFindSystemAccountViaEngine(t, e, "gh", eur)
		err := store.PostEntries([]engine.LedgerEntry{
			{ID: "bad-1", AccountID: sys.ID, Amount: 100, Type: engine.EntryTypeDebit, EventID: "e1"},
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = e.CheckGlobalBalance()
		if err == nil {
			t.Error("expected global balance to fail for unbalanced ledger")
		}
		if !strings.Contains(err.Error(), "EUR") {
			t.Errorf("error should name the failing currency, got: %v", err)
		}
	})
}

// ----- Integrity: per-event + global pass together ----------------------------

func TestIntegrityAfterFullScenario(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	setupSystemAccount(t, e, "gh", eur)
	setupLiquidityAccount(t, e, "gh", eur)
	setupUserAccount(t, e, "user1", "gh", eur)
	setupUserAccount(t, e, "user2", "gh", eur)

	e1, _ := e.FundProviderLiquidity("gh", eur, 100000)
	e2, _ := e.UserOnboard("user1", "gh", eur, 30000)
	e3, _ := e.SameProviderP2PTransfer("user1", "user2", "gh", eur, 15000)
	e4, _ := e.UserOffboard("user2", "gh", eur, 5000)

	for _, entries := range [][]engine.LedgerEntry{e1, e2, e3, e4} {
		if err := e.CheckPerEventBalance(entries[0].EventID); err != nil {
			t.Errorf("per-event check failed: %v", err)
		}
	}

	if _, err := e.CheckGlobalBalance(); err != nil {
		t.Errorf("global balance check failed: %v", err)
	}
}

// ----- helpers ----------------------------------------------------------------

func mustFindSystemAccount(t *testing.T, e *engine.Engine, providerID string, currency engine.Currency) engine.Account {
	t.Helper()
	return mustFindSystemAccountViaEngine(t, e, providerID, currency)
}

func mustFindSystemAccountViaEngine(t *testing.T, e *engine.Engine, providerID string, currency engine.Currency) engine.Account {
	t.Helper()
	accounts, err := e.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range accounts {
		if a.Type == engine.AccountTypeSystem && a.ProviderID == providerID && a.Currency == currency {
			return a
		}
	}
	t.Fatalf("system account for provider %q currency %q not found", providerID, currency.Code)
	return engine.Account{}
}

func mustFindUserAccount(t *testing.T, e *engine.Engine, userID, providerID string, currency engine.Currency) engine.Account {
	t.Helper()
	accounts, err := e.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range accounts {
		if a.Type == engine.AccountTypeUser && a.UserID == userID && a.ProviderID == providerID && a.Currency == currency {
			return a
		}
	}
	t.Fatalf("user account for user %q not found", userID)
	return engine.Account{}
}
