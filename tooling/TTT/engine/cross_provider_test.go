package engine_test

import (
	"strings"
	"testing"
	"time"

	"ttt/engine"
)

// setupCrossProvider bootstraps two providers and all accounts required for
// a GateHub EUR -> Xago ZAR cross-provider transfer under sender-side FX
// conversion. Returns the sender + recipient user accounts.
func setupCrossProvider(t *testing.T) (*engine.Engine, engine.Account, engine.Account) {
	t.Helper()
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	setupProvider(t, e, "xg", "Xago")

	setupSystemAccount(t, e, "gh", eur)
	setupLiquidityAccount(t, e, "gh", eur)
	setupSystemAccount(t, e, "gh", zar)
	setupLiquidityAccount(t, e, "gh", zar)
	setupLiquidityAccount(t, e, "xg", eur) // hosts mirror position
	setupSystemAccount(t, e, "xg", zar)
	setupLiquidityAccount(t, e, "xg", zar)

	sender := setupUserAccount(t, e, "user1", "gh", eur)
	recipient := setupUserAccount(t, e, "userA", "xg", zar)

	// Onboard sender with 500 EUR (50000 base units).
	if _, err := e.UserOnboard("user1", "gh", eur, 50000); err != nil {
		t.Fatalf("UserOnboard: %v", err)
	}
	return e, sender, recipient
}

func TestCrossProviderTransfer(t *testing.T) {
	t.Run("gatehub to xago works without gatehub zar", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupProvider(t, e, "xg", "Xago")

		// GateHub stays EUR-only.
		setupSystemAccount(t, e, "gh", eur)
		setupLiquidityAccount(t, e, "gh", eur)

		// Xago can convert between EUR and ZAR.
		setupSystemAccount(t, e, "xg", eur)
		setupLiquidityAccount(t, e, "xg", eur)
		setupSystemAccount(t, e, "xg", zar)
		setupLiquidityAccount(t, e, "xg", zar)

		sender := setupUserAccount(t, e, "alice", "gh", eur)
		if _, err := e.UserOnboard("alice", "gh", eur, 50000); err != nil {
			t.Fatalf("UserOnboard: %v", err)
		}

		entries, err := e.CrossProviderTransfer(
			"alice", "gh", eur,
			"carlos", "xg", zar,
			10000, 20, 1,
		)
		if err != nil {
			t.Fatalf("CrossProviderTransfer: %v", err)
		}
		if len(entries) != 12 {
			t.Fatalf("expected 12 entries, got %d", len(entries))
		}

		if bal, _ := e.Balance(sender.ID); bal != 40000 {
			t.Errorf("sender balance: want 40000, got %d", bal)
		}
		accounts, err := e.ListAccounts()
		if err != nil {
			t.Fatalf("ListAccounts: %v", err)
		}
		recipientID := ""
		for _, a := range accounts {
			if a.Type == engine.AccountTypeUser && a.UserID == "carlos" && a.ProviderID == "xg" && a.Currency.Code == "ZAR" {
				recipientID = a.ID
				break
			}
		}
		if recipientID == "" {
			t.Fatal("expected recipient account to be created")
		}
		if bal, _ := e.Balance(recipientID); bal != 200000 {
			t.Errorf("recipient balance: want 200000, got %d", bal)
		}
		if _, err := e.CheckBilateralPositions(); err != nil {
			t.Errorf("bilateral mirror failed: %v", err)
		}
	})

	t.Run("happy path", func(t *testing.T) {
		e, sender, recipient := setupCrossProvider(t)

		// 100 EUR → 2000 ZAR at rate 20/1.
		entries, err := e.CrossProviderTransfer(
			"user1", "gh", eur,
			"userA", "xg", zar,
			10000, 20, 1,
		)
		if err != nil {
			t.Fatalf("CrossProviderTransfer: %v", err)
		}
		if len(entries) != 12 {
			t.Fatalf("expected 12 entries, got %d", len(entries))
		}
		// Every entry carries the FX metadata.
		for _, en := range entries {
			if en.Metadata[engine.MetaFXRateNum] != "20" || en.Metadata[engine.MetaFXRateDen] != "1" {
				t.Errorf("entry %s missing or wrong FX metadata: %v", en.ID, en.Metadata)
			}
			if en.Metadata[engine.MetaFXBase] != "EUR" || en.Metadata[engine.MetaFXQuote] != "ZAR" {
				t.Errorf("entry %s missing FX base/quote: %v", en.ID, en.Metadata)
			}
		}

		bal, err := e.Balance(sender.ID)
		if err != nil {
			t.Fatal(err)
		}
		if bal != 40000 {
			t.Errorf("sender balance: want 40000, got %d", bal)
		}
		rbal, err := e.Balance(recipient.ID)
		if err != nil {
			t.Fatal(err)
		}
		if rbal != 200000 {
			t.Errorf("recipient balance: want 200000, got %d", rbal)
		}

		// Per-event balance (grouped by currency).
		if err := e.CheckPerEventBalance(entries[0].EventID); err != nil {
			t.Errorf("per-event balance failed: %v", err)
		}
		// Global balance.
		if _, err := e.CheckGlobalBalance(); err != nil {
			t.Errorf("global balance failed: %v", err)
		}
		// Bilateral mirror should hold (credit on gh pos, debit on xg pos).
		if _, err := e.CheckBilateralPositions(); err != nil {
			t.Errorf("bilateral mirror failed: %v", err)
		}
		// Liquidity decomposition: Xago ZAR float should be 0 (reserved for
		// recipient user). GateHub EUR float = reserves unchanged - 100 EUR
		// position commitment = -10000 → over-committed (not yet funded).
		if _, err := e.CheckLiquidityDecomposition(); err == nil {
			t.Error("expected over-committed liquidity error before provider liquidity funding")
		}
	})

	t.Run("insufficient balance", func(t *testing.T) {
		e, _, _ := setupCrossProvider(t)
		if _, err := e.CrossProviderTransfer("user1", "gh", eur, "userA", "xg", zar, 99999999, 20, 1); err == nil {
			t.Error("expected insufficient balance")
		}
	})

	t.Run("fractional rate uses floor division", func(t *testing.T) {
		e, sender, recipient := setupCrossProvider(t)
		// 100 EUR * 3 / 7 = 42.857… → floors to 42 ZAR base units.
		entries, err := e.CrossProviderTransfer("user1", "gh", eur, "userA", "xg", zar, 10000, 3, 7)
		if err != nil {
			t.Fatalf("CrossProviderTransfer: %v", err)
		}
		if len(entries) != 12 {
			t.Fatalf("expected 12 entries, got %d", len(entries))
		}
		senderBal, _ := e.Balance(sender.ID)
		if senderBal != 40000 {
			t.Errorf("sender balance: want 40000, got %d", senderBal)
		}
		recvBal, _ := e.Balance(recipient.ID)
		// 10000 * 3 / 7 = 4285 (floor).
		if recvBal != 4285 {
			t.Errorf("recipient balance: want 4285, got %d", recvBal)
		}
	})

	t.Run("same provider rejected", func(t *testing.T) {
		e, _, _ := setupCrossProvider(t)
		if _, err := e.CrossProviderTransfer("user1", "gh", eur, "userA", "gh", zar, 10000, 1, 1); err == nil {
			t.Error("expected same-provider rejection")
		}
	})

	t.Run("non-positive amount rejected", func(t *testing.T) {
		e, _, _ := setupCrossProvider(t)
		if _, err := e.CrossProviderTransfer("user1", "gh", eur, "userA", "xg", zar, 0, 20, 1); err == nil {
			t.Error("expected zero-amount rejection")
		}
	})

	t.Run("non-positive rate rejected", func(t *testing.T) {
		e, _, _ := setupCrossProvider(t)
		if _, err := e.CrossProviderTransfer("user1", "gh", eur, "userA", "xg", zar, 10000, 0, 1); err == nil {
			t.Error("expected zero numerator rejection")
		}
	})
}

func TestSettleBilateral(t *testing.T) {
	setup := func(t *testing.T) *engine.Engine {
		t.Helper()
		e, _, _ := setupCrossProvider(t)
		// Fund GateHub ZAR liquidity so settlement can move reserves safely.
		if _, err := e.FundProviderLiquidity("gh", zar, 100000); err != nil {
			t.Fatalf("FundProviderLiquidity: %v", err)
		}
		// Also fund Xago ZAR liquidity so it can settle outflows.
		// Note: amount 0 should fail.
		if _, err := e.FundProviderLiquidity("xg", zar, 0); err == nil {
			t.Error("expected failure for zero amount")
		}
		if _, err := e.FundProviderLiquidity("xg", zar, 300000); err != nil {
			t.Fatalf("FundProviderLiquidity xg ZAR: %v", err)
		}
		// Perform a 100 EUR → 2000 ZAR cross-provider transfer.
		if _, err := e.CrossProviderTransfer("user1", "gh", eur, "userA", "xg", zar, 10000, 20, 1); err != nil {
			t.Fatalf("CrossProviderTransfer: %v", err)
		}
		return e
	}

	t.Run("happy path", func(t *testing.T) {
		e := setup(t)
		future := time.Now().UTC().Add(time.Hour)

		entries, err := e.SettleBilateral("gh", "xg", zar, future)
		if err != nil {
			t.Fatalf("SettleBilateral: %v", err)
		}
		if len(entries) != 4 {
			t.Fatalf("expected 4 entries, got %d", len(entries))
		}

		// After settlement, both positions should be zero.
		if _, err := e.CheckBilateralPositions(); err != nil {
			t.Errorf("bilateral mirror after settlement: %v", err)
		}
		// Liquidity decomposition must be healthy.
		if _, err := e.CheckLiquidityDecomposition(); err != nil {
			t.Errorf("liquidity decomposition after settlement: %v", err)
		}
	})

	t.Run("nothing to settle", func(t *testing.T) {
		e, _, _ := setupCrossProvider(t)
		future := time.Now().UTC().Add(time.Hour)
		if _, err := e.SettleBilateral("gh", "xg", zar, future); err == nil {
			t.Error("expected nothing-to-settle error (no positions yet)")
		}
	})

	t.Run("same provider rejected", func(t *testing.T) {
		e := setup(t)
		if _, err := e.SettleBilateral("gh", "gh", zar, time.Now()); err == nil {
			t.Error("expected same-provider rejection")
		}
	})

	t.Run("cutoff before transfer excludes entries", func(t *testing.T) {
		e := setup(t)
		past := time.Now().UTC().Add(-time.Hour)
		if _, err := e.SettleBilateral("gh", "xg", zar, past); err == nil {
			t.Error("expected nothing-to-settle for past cutoff")
		}
	})
}

func TestLineNativeCrossProviderAndSettlement(t *testing.T) {
	e, sender, recipient := setupCrossProvider(t)

	if _, err := e.FundProviderLiquidityLines("gh", zar, 100000); err != nil {
		t.Fatalf("FundProviderLiquidityLines gh: %v", err)
	}
	if _, err := e.FundProviderLiquidityLines("xg", zar, 300000); err != nil {
		t.Fatalf("FundProviderLiquidityLines xg: %v", err)
	}

	lines, err := e.CrossProviderTransferLines(
		"user1", "gh", eur,
		"userA", "xg", zar,
		10000, 20, 1,
	)
	if err != nil {
		t.Fatalf("CrossProviderTransferLines: %v", err)
	}
	if len(lines) != 6 {
		t.Fatalf("expected 6 lines, got %d", len(lines))
	}
	for _, line := range lines {
		if line.Metadata[engine.MetaFXRateNum] != "20" || line.Metadata[engine.MetaFXRateDen] != "1" {
			t.Errorf("line %s missing fx rate metadata: %v", line.ID, line.Metadata)
		}
	}

	if bal, _ := e.Balance(sender.ID); bal != 40000 {
		t.Errorf("sender balance: want 40000, got %d", bal)
	}
	if bal, _ := e.Balance(recipient.ID); bal != 200000 {
		t.Errorf("recipient balance: want 200000, got %d", bal)
	}

	future := time.Now().UTC().Add(time.Hour)
	settled, err := e.SettleBilateralLines("gh", "xg", zar, future)
	if err != nil {
		t.Fatalf("SettleBilateralLines: %v", err)
	}
	if len(settled) != 2 {
		t.Fatalf("expected 2 settlement lines, got %d", len(settled))
	}
	if _, err := e.CheckBilateralPositions(); err != nil {
		t.Errorf("bilateral mirror after line settlement: %v", err)
	}
}

func TestCheckLiquidityDecomposition(t *testing.T) {
	t.Run("healthy after funding", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupSystemAccount(t, e, "gh", eur)
		setupLiquidityAccount(t, e, "gh", eur)
		if _, err := e.FundProviderLiquidity("gh", eur, 50000); err != nil {
			t.Fatal(err)
		}
		decomp, err := e.CheckLiquidityDecomposition()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(decomp) != 1 {
			t.Fatalf("expected 1 liquidity account, got %d", len(decomp))
		}
		if decomp[0].Float != 50000 {
			t.Errorf("float: want 50000, got %d", decomp[0].Float)
		}
	})

	t.Run("over-commitment detected", func(t *testing.T) {
		e, _, _ := setupCrossProvider(t)
		// No liquidity funding → cross-provider transfer leaves GH EUR float negative.
		if _, err := e.CrossProviderTransfer("user1", "gh", eur, "userA", "xg", zar, 10000, 20, 1); err != nil {
			t.Fatal(err)
		}
		_, err := e.CheckLiquidityDecomposition()
		if err == nil {
			t.Error("expected over-commitment error")
		}
		if err != nil && !strings.Contains(err.Error(), "over-committed") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// setupSelfExchange bootstraps the self-exchange paradigm accounts and
// pre-funds Xago ZAR liquidity. Returns the sender and recipient user accounts.
func setupSelfExchange(t *testing.T) (*engine.Engine, engine.Account, engine.Account) {
	t.Helper()
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	setupProvider(t, e, "xg", "Xago")

	setupSystemAccount(t, e, "gh", eur)
	setupLiquidityAccount(t, e, "gh", eur)
	setupSystemAccount(t, e, "xg", eur)
	setupLiquidityAccount(t, e, "xg", eur)
	setupSystemAccount(t, e, "xg", zar)
	setupLiquidityAccount(t, e, "xg", zar)
	setupFXAccount(t, e, "xg", zar)

	sender := setupUserAccount(t, e, "alice", "gh", eur)
	recipient := setupUserAccount(t, e, "carlos", "xg", zar)

	// Onboard sender with 500 EUR.
	if _, err := e.UserOnboard("alice", "gh", eur, 50000); err != nil {
		t.Fatalf("UserOnboard: %v", err)
	}
	// Pre-fund Xago ZAR pool with 150 000 ZAR.
	if _, err := e.FundProviderLiquidity("xg", zar, 15000000); err != nil {
		t.Fatalf("FundProviderLiquidity xg ZAR: %v", err)
	}
	return e, sender, recipient
}

func TestSelfExchangeTransfer(t *testing.T) {
	t.Run("happy path uses 5 lines", func(t *testing.T) {
		e, sender, recipient := setupSelfExchange(t)

		lines, err := e.CrossProviderTransferLines(
			"alice", "gh", eur,
			"carlos", "xg", zar,
			10000, 15, 1, // 100 EUR → 1500 ZAR
		)
		if err != nil {
			t.Fatalf("CrossProviderTransferLines: %v", err)
		}
		if len(lines) != 5 {
			t.Fatalf("expected 5 journal lines for self-exchange, got %d", len(lines))
		}

		// Sender loses 100 EUR.
		if bal, _ := e.Balance(sender.ID); bal != 40000 {
			t.Errorf("sender balance: want 40000, got %d", bal)
		}
		// Recipient receives 1500 ZAR.
		if bal, _ := e.Balance(recipient.ID); bal != 150000 {
			t.Errorf("recipient balance: want 150000, got %d", bal)
		}

		// FX metadata present on all lines.
		for _, l := range lines {
			if l.Metadata[engine.MetaFXRateNum] != "15" {
				t.Errorf("line missing fx.rate_num: %v", l.Metadata)
			}
			if l.Metadata[engine.MetaSelfExchange] != "true" {
				t.Errorf("line missing fx.self_exchange flag: %v", l.Metadata)
			}
		}

		// Integrity checks.
		if err := e.CheckPerEventBalance(lines[0].EventID); err != nil {
			t.Errorf("per-event balance: %v", err)
		}
		if _, err := e.CheckGlobalBalance(); err != nil {
			t.Errorf("global balance: %v", err)
		}
		if _, err := e.CheckBilateralPositions(); err != nil {
			t.Errorf("bilateral mirror: %v", err)
		}
	})

	t.Run("FX account passes through to zero balance", func(t *testing.T) {
		e, _, _ := setupSelfExchange(t)

		if _, err := e.CrossProviderTransferLines(
			"alice", "gh", eur,
			"carlos", "xg", zar,
			10000, 15, 1,
		); err != nil {
			t.Fatalf("CrossProviderTransferLines: %v", err)
		}

		fxAcct, ok, err := e.FindFXAccount("xg", zar)
		if err != nil || !ok {
			t.Fatalf("FindFXAccount: ok=%v err=%v", ok, err)
		}
		if bal, _ := e.Balance(fxAcct.ID); bal != 0 {
			t.Errorf("FX account balance: want 0 (transient), got %d", bal)
		}
	})

	t.Run("insufficient ZAR liquidity rejected", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		setupProvider(t, e, "xg", "Xago")
		setupSystemAccount(t, e, "gh", eur)
		setupLiquidityAccount(t, e, "gh", eur)
		setupSystemAccount(t, e, "xg", eur)
		setupLiquidityAccount(t, e, "xg", eur)
		setupSystemAccount(t, e, "xg", zar)
		setupLiquidityAccount(t, e, "xg", zar) // zero balance
		setupFXAccount(t, e, "xg", zar)
		if _, err := e.UserOnboard("alice", "gh", eur, 50000); err != nil {
			t.Fatalf("UserOnboard: %v", err)
		}
		// ZAR liquidity is empty — transfer must fail.
		if _, err := e.CrossProviderTransferLines(
			"alice", "gh", eur,
			"carlos", "xg", zar,
			10000, 15, 1,
		); err == nil {
			t.Error("expected insufficient ZAR liquidity error")
		}
	})

	t.Run("bilateral settlement still works after self-exchange transfers", func(t *testing.T) {
		e, _, _ := setupSelfExchange(t)

		if _, err := e.CrossProviderTransferLines(
			"alice", "gh", eur,
			"carlos", "xg", zar,
			10000, 15, 1,
		); err != nil {
			t.Fatalf("CrossProviderTransferLines: %v", err)
		}

		// Fund GH EUR liquidity so settlement can proceed.
		if _, err := e.FundProviderLiquidity("gh", eur, 50000); err != nil {
			t.Fatalf("FundProviderLiquidity gh EUR: %v", err)
		}

		// EUR bilateral settlement closes the position created during self-exchange.
		future := time.Now().UTC().Add(time.Hour)
		if _, err := e.SettleBilateralLines("gh", "xg", eur, future); err != nil {
			t.Fatalf("SettleBilateralLines EUR: %v", err)
		}
		if _, err := e.CheckBilateralPositions(); err != nil {
			t.Errorf("bilateral mirror after settlement: %v", err)
		}
	})
}

func TestCheckBilateralPositions(t *testing.T) {
	t.Run("healthy after cross-provider transfer", func(t *testing.T) {
		e, _, _ := setupCrossProvider(t)
		if _, err := e.CrossProviderTransfer("user1", "gh", eur, "userA", "xg", zar, 10000, 20, 1); err != nil {
			t.Fatal(err)
		}
		pairs, err := e.CheckBilateralPositions()
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
		if len(pairs) != 1 {
			t.Fatalf("expected 1 pair, got %d", len(pairs))
		}
		if pairs[0].MirrorSum != 0 {
			t.Errorf("mirror sum: want 0, got %d", pairs[0].MirrorSum)
		}
	})

	t.Run("no positions yet", func(t *testing.T) {
		e := newEngine()
		setupProvider(t, e, "gh", "GateHub")
		pairs, err := e.CheckBilateralPositions()
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
		if len(pairs) != 0 {
			t.Errorf("expected 0 pairs, got %d", len(pairs))
		}
	})
}
