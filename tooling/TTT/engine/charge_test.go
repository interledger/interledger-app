package engine_test

import (
	"strings"
	"testing"

	"ttt/engine"
	"ttt/engine/memory"
)

// setupCrossProviderWithFX returns an engine wired with a scripted FX service
// (EUR/ZAR = 15/1) and the two-provider topology from setupCrossProvider.
func setupCrossProviderWithFX(t *testing.T, dir engine.DirectionSource) (*engine.Engine, engine.Account) {
	t.Helper()
	store := memory.New()
	fx := engine.NewFXService(dir)
	if err := fx.Set("EUR", "ZAR", 15, 1); err != nil {
		t.Fatalf("fx.Set: %v", err)
	}
	e := engine.New(store).WithFX(fx)

	setupProvider(t, e, "gh", "GateHub")
	setupProvider(t, e, "xg", "Xago")

	setupSystemAccount(t, e, "gh", eur)
	setupLiquidityAccount(t, e, "gh", eur)
	setupSystemAccount(t, e, "gh", zar)
	setupLiquidityAccount(t, e, "gh", zar)
	setupLiquidityAccount(t, e, "xg", eur)
	setupSystemAccount(t, e, "xg", zar)
	setupLiquidityAccount(t, e, "xg", zar)

	sender := setupUserAccount(t, e, "alice", "gh", eur)
	if _, err := e.UserOnboard("alice", "gh", eur, 50000); err != nil {
		t.Fatalf("UserOnboard: %v", err)
	}
	return e, sender
}

func TestChargeRate_ChargeAmount(t *testing.T) {
	tests := []struct {
		name     string
		rate     *engine.ChargeRate
		dispatch int64
		want     int64
	}{
		{"nil charge", nil, 10000, 0},
		{"zero percent", &engine.ChargeRate{Num: 0, Den: 10000}, 10000, 0},
		{"2 percent", &engine.ChargeRate{Num: 200, Den: 10000}, 10000, 200},
		{"2.5 percent", &engine.ChargeRate{Num: 250, Den: 10000}, 10000, 250},
		{"floor division", &engine.ChargeRate{Num: 1, Den: 3}, 10, 3},
		{"100 percent", &engine.ChargeRate{Num: 1, Den: 1}, 500, 500},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.rate.ChargeAmount(tc.dispatch)
			if got != tc.want {
				t.Errorf("ChargeAmount(%d) = %d, want %d", tc.dispatch, got, tc.want)
			}
		})
	}
}

func TestEngine_SetGetCharge(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	setupProvider(t, e, "xg", "Xago")

	t.Run("nil by default", func(t *testing.T) {
		c, err := e.GetCharge("gh", "xg")
		if err != nil {
			t.Fatalf("GetCharge: %v", err)
		}
		if c != nil {
			t.Errorf("expected nil charge, got %+v", c)
		}
	})

	t.Run("set and retrieve", func(t *testing.T) {
		charge := &engine.ChargeRate{Num: 200, Den: 10000}
		if err := e.SetCharge("gh", "xg", charge); err != nil {
			t.Fatalf("SetCharge: %v", err)
		}
		got, err := e.GetCharge("gh", "xg")
		if err != nil {
			t.Fatalf("GetCharge: %v", err)
		}
		if got == nil || got.Num != 200 || got.Den != 10000 {
			t.Errorf("expected {200, 10000}, got %+v", got)
		}
	})

	t.Run("other direction independent", func(t *testing.T) {
		c, err := e.GetCharge("xg", "gh")
		if err != nil {
			t.Fatalf("GetCharge: %v", err)
		}
		if c != nil {
			t.Errorf("xg→gh should still be nil, got %+v", c)
		}
	})

	t.Run("clear with nil", func(t *testing.T) {
		if err := e.SetCharge("gh", "xg", nil); err != nil {
			t.Fatalf("SetCharge nil: %v", err)
		}
		c, err := e.GetCharge("gh", "xg")
		if err != nil {
			t.Fatalf("GetCharge: %v", err)
		}
		if c != nil {
			t.Errorf("expected nil after clear, got %+v", c)
		}
	})

	t.Run("zero percent is valid", func(t *testing.T) {
		if err := e.SetCharge("gh", "xg", &engine.ChargeRate{Num: 0, Den: 10000}); err != nil {
			t.Fatalf("SetCharge 0%%: %v", err)
		}
		c, err := e.GetCharge("gh", "xg")
		if err != nil {
			t.Fatalf("GetCharge: %v", err)
		}
		if c == nil {
			t.Error("expected non-nil charge for 0%")
		}
	})

	t.Run("negative numerator rejected", func(t *testing.T) {
		err := e.SetCharge("gh", "xg", &engine.ChargeRate{Num: -1, Den: 100})
		if err == nil {
			t.Error("expected error for negative numerator")
		}
	})

	t.Run("non-positive denominator rejected", func(t *testing.T) {
		err := e.SetCharge("gh", "xg", &engine.ChargeRate{Num: 1, Den: 0})
		if err == nil {
			t.Error("expected error for zero denominator")
		}
	})
}

func TestCrossProviderTransferAuto_WithCharge(t *testing.T) {
	t.Run("charge deducted from sender, stays in liquidity", func(t *testing.T) {
		e, sender := setupCrossProviderWithFX(t, engine.NewScriptedDirection(true))

		// Set 2% charge on GH→XG.
		if err := e.SetCharge("gh", "xg", &engine.ChargeRate{Num: 200, Den: 10000}); err != nil {
			t.Fatalf("SetCharge: %v", err)
		}

		// Send 100 EUR (10000 base units) from alice (GH EUR) to carlos (XG ZAR).
		// Charge = 2% of 10000 = 200 base units (2.00 EUR).
		// Sender total debit = 10200 EUR base units.
		lines, _, err := e.CrossProviderTransferAutoLines(
			"alice", "gh", eur,
			"carlos", "xg", zar,
			10000,
		)
		if err != nil {
			t.Fatalf("CrossProviderTransferAutoLines: %v", err)
		}
		if len(lines) != 6 {
			t.Fatalf("expected 6 lines, got %d", len(lines))
		}

		// Sender balance: started with 50000, debited 10000 + 200 = 10200.
		if bal, _ := e.Balance(sender.ID); bal != 39800 {
			t.Errorf("sender balance: want 39800, got %d", bal)
		}

		// Charge metadata must be present on all lines.
		for _, l := range lines {
			if l.Metadata[engine.MetaChargeRateNum] != "200" {
				t.Errorf("line %s missing charge.rate_num metadata: %v", l.ID, l.Metadata)
			}
			if l.Metadata[engine.MetaChargeRateDen] != "10000" {
				t.Errorf("line %s missing charge.rate_den metadata: %v", l.ID, l.Metadata)
			}
			if l.Metadata[engine.MetaChargeAmount] != "200" {
				t.Errorf("line %s missing charge.amount metadata: %v", l.ID, l.Metadata)
			}
		}

		// First line: sender debited 10200, liqSrc credited 10200.
		if lines[0].Amount != 10200 {
			t.Errorf("first line amount: want 10200, got %d", lines[0].Amount)
		}
		// Second line: only the dispatch amount moves from liq to system.
		if lines[1].Amount != 10000 {
			t.Errorf("second line amount: want 10000, got %d", lines[1].Amount)
		}

		// Global and per-event balance checks must pass.
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

	t.Run("nil charge applies no deduction", func(t *testing.T) {
		e, sender := setupCrossProviderWithFX(t, engine.NewScriptedDirection(true))
		// No charge set.
		lines, _, err := e.CrossProviderTransferAutoLines(
			"alice", "gh", eur, "carlos", "xg", zar, 10000,
		)
		if err != nil {
			t.Fatalf("CrossProviderTransferAutoLines: %v", err)
		}
		// First line amount == dispatch only.
		if lines[0].Amount != 10000 {
			t.Errorf("first line amount: want 10000, got %d", lines[0].Amount)
		}
		// Sender debited only dispatch amount.
		if bal, _ := e.Balance(sender.ID); bal != 40000 {
			t.Errorf("sender balance: want 40000, got %d", bal)
		}
		// No charge metadata.
		if lines[0].Metadata[engine.MetaChargeAmount] != "" {
			t.Errorf("unexpected charge.amount metadata on no-charge transfer")
		}
	})

	t.Run("zero percent charge has no financial effect", func(t *testing.T) {
		e, sender := setupCrossProviderWithFX(t, engine.NewScriptedDirection(true))
		if err := e.SetCharge("gh", "xg", &engine.ChargeRate{Num: 0, Den: 10000}); err != nil {
			t.Fatalf("SetCharge: %v", err)
		}
		lines, _, err := e.CrossProviderTransferAutoLines(
			"alice", "gh", eur, "carlos", "xg", zar, 10000,
		)
		if err != nil {
			t.Fatalf("CrossProviderTransferAutoLines: %v", err)
		}
		if lines[0].Amount != 10000 {
			t.Errorf("first line amount: want 10000, got %d", lines[0].Amount)
		}
		if bal, _ := e.Balance(sender.ID); bal != 40000 {
			t.Errorf("sender balance: want 40000, got %d", bal)
		}
	})

	t.Run("insufficient balance covers dispatch+charge", func(t *testing.T) {
		e, _ := setupCrossProviderWithFX(t, engine.NewScriptedDirection(true))
		if err := e.SetCharge("gh", "xg", &engine.ChargeRate{Num: 200, Den: 10000}); err != nil {
			t.Fatalf("SetCharge: %v", err)
		}
		// sender has 50000; try to send 49900 with 2% charge → total = 49900+998 = 50898 > 50000.
		_, _, err := e.CrossProviderTransferAutoLines(
			"alice", "gh", eur, "carlos", "xg", zar, 49900,
		)
		if err == nil {
			t.Error("expected insufficient balance error")
		}
		if err != nil && !strings.Contains(err.Error(), "insufficient balance") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("charge does not affect FX mutation on success", func(t *testing.T) {
		dir := engine.NewScriptedDirection(true) // up
		e, _ := setupCrossProviderWithFX(t, dir)
		if err := e.SetCharge("gh", "xg", &engine.ChargeRate{Num: 200, Den: 10000}); err != nil {
			t.Fatalf("SetCharge: %v", err)
		}
		_, rate, err := e.CrossProviderTransferAutoLines(
			"alice", "gh", eur, "carlos", "xg", zar, 10000,
		)
		if err != nil {
			t.Fatalf("CrossProviderTransferAutoLines: %v", err)
		}
		// Applied rate should be the initial 15/1.
		if rate.Num != 15 || rate.Den != 1 {
			t.Errorf("applied rate: want 15/1, got %d/%d", rate.Num, rate.Den)
		}
		// FX mutated after success — next rate should be +5% = 63/4.
		next, err := e.FX().Rate("EUR", "ZAR")
		if err != nil {
			t.Fatalf("Rate: %v", err)
		}
		if next.Num != 63 || next.Den != 4 {
			t.Errorf("next rate after mutation: want 63/4, got %d/%d", next.Num, next.Den)
		}
	})

	t.Run("failed transfer does not mutate FX", func(t *testing.T) {
		e, _ := setupCrossProviderWithFX(t, engine.NewScriptedDirection(true))
		if err := e.SetCharge("gh", "xg", &engine.ChargeRate{Num: 200, Den: 10000}); err != nil {
			t.Fatalf("SetCharge: %v", err)
		}
		before, _ := e.FX().Rate("EUR", "ZAR")
		_, _, err := e.CrossProviderTransferAutoLines(
			"alice", "gh", eur, "carlos", "xg", zar, 99999999,
		)
		if err == nil {
			t.Error("expected failure")
		}
		after, _ := e.FX().Rate("EUR", "ZAR")
		if before.Num != after.Num || before.Den != after.Den {
			t.Errorf("FX mutated on failure: before=%d/%d after=%d/%d", before.Num, before.Den, after.Num, after.Den)
		}
	})

	t.Run("charge on opposite direction is independent", func(t *testing.T) {
		e, sender := setupCrossProviderWithFX(t, engine.NewScriptedDirection(true))
		// Set charge only on XG→GH (opposite direction).
		if err := e.SetCharge("xg", "gh", &engine.ChargeRate{Num: 500, Den: 10000}); err != nil {
			t.Fatalf("SetCharge xg→gh: %v", err)
		}
		// GH→XG should still have nil charge.
		lines, _, err := e.CrossProviderTransferAutoLines(
			"alice", "gh", eur, "carlos", "xg", zar, 10000,
		)
		if err != nil {
			t.Fatalf("CrossProviderTransferAutoLines: %v", err)
		}
		if lines[0].Amount != 10000 {
			t.Errorf("first line amount: want 10000, got %d", lines[0].Amount)
		}
		if bal, _ := e.Balance(sender.ID); bal != 40000 {
			t.Errorf("sender balance: want 40000, got %d", bal)
		}
	})
}
