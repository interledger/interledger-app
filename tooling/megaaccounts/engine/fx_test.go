package engine_test

import (
	"strings"
	"testing"

	"megaaccounts/engine"
)

func TestFXService_SetRateMutate(t *testing.T) {
	dir := engine.NewScriptedDirection(true, false, true) // up, down, up
	fx := engine.NewFXService(dir)

	if err := fx.Set("EUR", "ZAR", 15, 1); err != nil {
		t.Fatalf("Set: %v", err)
	}

	r, err := fx.Rate("EUR", "ZAR")
	if err != nil {
		t.Fatalf("Rate: %v", err)
	}
	if r.Num != 15 || r.Den != 1 {
		t.Errorf("initial rate: want 15/1, got %d/%d", r.Num, r.Den)
	}

	// First mutation: +5% → 15 * 21/20 = 315/20 = 63/4.
	next, err := fx.Mutate("EUR", "ZAR")
	if err != nil {
		t.Fatalf("Mutate: %v", err)
	}
	if next.Num != 63 || next.Den != 4 {
		t.Errorf("after +5%%: want 63/4, got %d/%d", next.Num, next.Den)
	}

	// Second mutation: -5% → 63/4 * 19/20 = 1197/80.
	next, err = fx.Mutate("EUR", "ZAR")
	if err != nil {
		t.Fatalf("Mutate: %v", err)
	}
	if next.Num != 1197 || next.Den != 80 {
		t.Errorf("after -5%%: want 1197/80, got %d/%d", next.Num, next.Den)
	}
}

func TestFXService_InverseLookup(t *testing.T) {
	fx := engine.NewFXService(engine.NewScriptedDirection(true))
	if err := fx.Set("EUR", "ZAR", 15, 1); err != nil {
		t.Fatalf("Set: %v", err)
	}
	inv, err := fx.Rate("ZAR", "EUR")
	if err != nil {
		t.Fatalf("Rate inverse: %v", err)
	}
	if inv.Num != 1 || inv.Den != 15 {
		t.Errorf("inverse: want 1/15, got %d/%d", inv.Num, inv.Den)
	}
}

func TestFXService_UnknownPair(t *testing.T) {
	fx := engine.NewFXService(engine.NewScriptedDirection(true))
	if _, err := fx.Rate("USD", "JPY"); err == nil {
		t.Error("expected error for unknown pair")
	}
	if _, err := fx.Mutate("USD", "JPY"); err == nil {
		t.Error("expected error mutating unknown pair")
	}
}

func TestFXService_InvalidInput(t *testing.T) {
	fx := engine.NewFXService(engine.NewScriptedDirection(true))
	if err := fx.Set("", "ZAR", 15, 1); err == nil {
		t.Error("expected error for empty base")
	}
	if err := fx.Set("EUR", "ZAR", 0, 1); err == nil {
		t.Error("expected error for zero numerator")
	}
	if err := fx.Set("EUR", "ZAR", -1, 1); err == nil {
		t.Error("expected error for negative numerator")
	}
}

func TestCrossProviderTransferAuto(t *testing.T) {
	t.Run("happy path uses fx rate then mutates", func(t *testing.T) {
		e, sender, recipient := setupCrossProvider(t)
		fx := engine.NewFXService(engine.NewScriptedDirection(true)) // +5% after
		if err := fx.Set("EUR", "ZAR", 15, 1); err != nil {
			t.Fatalf("Set: %v", err)
		}
		e.WithFX(fx)

		// 100 EUR at 15/1 → 1500 ZAR.
		entries, rate, err := e.CrossProviderTransferAuto(
			"user1", "gh", eur, "userA", "xg", zar, 10000,
		)
		if err != nil {
			t.Fatalf("CrossProviderTransferAuto: %v", err)
		}
		if rate.Num != 15 || rate.Den != 1 {
			t.Errorf("applied rate: want 15/1, got %d/%d", rate.Num, rate.Den)
		}
		// Event metadata must record the applied (pre-mutation) rate.
		for _, en := range entries {
			if en.Metadata[engine.MetaFXRateNum] != "15" || en.Metadata[engine.MetaFXRateDen] != "1" {
				t.Errorf("entry metadata rate: %v", en.Metadata)
			}
		}
		if bal, _ := e.Balance(sender.ID); bal != 40000 {
			t.Errorf("sender balance: want 40000, got %d", bal)
		}
		if bal, _ := e.Balance(recipient.ID); bal != 150000 {
			t.Errorf("recipient balance: want 150000, got %d", bal)
		}

		// Rate must have mutated to 63/4.
		next, err := fx.Rate("EUR", "ZAR")
		if err != nil {
			t.Fatal(err)
		}
		if next.Num != 63 || next.Den != 4 {
			t.Errorf("post-mutation: want 63/4, got %d/%d", next.Num, next.Den)
		}
	})

	t.Run("sequential conversions use chained rate", func(t *testing.T) {
		e, _, _ := setupCrossProvider(t)
		fx := engine.NewFXService(engine.NewScriptedDirection(true, true))
		if err := fx.Set("EUR", "ZAR", 15, 1); err != nil {
			t.Fatalf("fx.Set: %v", err)
		}
		e.WithFX(fx)

		// First event — 100 EUR at 15/1.
		ents1, r1, err := e.CrossProviderTransferAuto("user1", "gh", eur, "userA", "xg", zar, 10000)
		if err != nil {
			t.Fatalf("first: %v", err)
		}
		if r1.Num != 15 || r1.Den != 1 {
			t.Errorf("r1: %v", r1)
		}

		// Second event — 100 EUR at mutated rate 63/4 → 157 ZAR base units (floor).
		// Wait — 10000 * 63 / 4 = 157500. So 1575 ZAR.
		ents2, r2, err := e.CrossProviderTransferAuto("user1", "gh", eur, "userA", "xg", zar, 10000)
		if err != nil {
			t.Fatalf("second: %v", err)
		}
		if r2.Num != 63 || r2.Den != 4 {
			t.Errorf("r2: want 63/4, got %d/%d", r2.Num, r2.Den)
		}

		// Historical event (ents1) metadata must still reflect 15/1.
		found := false
		for _, en := range ents1 {
			if en.Metadata[engine.MetaFXRateNum] == "15" && en.Metadata[engine.MetaFXRateDen] == "1" {
				found = true
				break
			}
		}
		if !found {
			t.Error("first event lost its original rate metadata")
		}
		// Second event metadata must reflect 63/4.
		for _, en := range ents2 {
			if en.Metadata[engine.MetaFXRateNum] != "63" || en.Metadata[engine.MetaFXRateDen] != "4" {
				t.Errorf("second event metadata rate: %v", en.Metadata)
				break
			}
		}
	})

	t.Run("failed conversion leaves rate unchanged", func(t *testing.T) {
		e, _, _ := setupCrossProvider(t)
		fx := engine.NewFXService(engine.NewScriptedDirection(true))
		if err := fx.Set("EUR", "ZAR", 15, 1); err != nil {
			t.Fatalf("fx.Set: %v", err)
		}
		e.WithFX(fx)

		// Insufficient balance → failure.
		if _, _, err := e.CrossProviderTransferAuto(
			"user1", "gh", eur, "userA", "xg", zar, 99999999,
		); err == nil {
			t.Fatal("expected failure")
		}
		r, _ := fx.Rate("EUR", "ZAR")
		if r.Num != 15 || r.Den != 1 {
			t.Errorf("rate mutated despite failure: %d/%d", r.Num, r.Den)
		}
	})

	t.Run("no fx service attached", func(t *testing.T) {
		e, _, _ := setupCrossProvider(t)
		_, _, err := e.CrossProviderTransferAuto(
			"user1", "gh", eur, "userA", "xg", zar, 10000,
		)
		if err == nil || !strings.Contains(err.Error(), "no FX service") {
			t.Errorf("expected no-FX error, got %v", err)
		}
	})
}

func TestCrossProviderTransferAutoLines(t *testing.T) {
	e, sender, recipient := setupCrossProvider(t)
	fx := engine.NewFXService(engine.NewScriptedDirection(true))
	if err := fx.Set("EUR", "ZAR", 15, 1); err != nil {
		t.Fatalf("Set: %v", err)
	}
	e.WithFX(fx)

	lines, rate, err := e.CrossProviderTransferAutoLines(
		"user1", "gh", eur, "userA", "xg", zar, 10000,
	)
	if err != nil {
		t.Fatalf("CrossProviderTransferAutoLines: %v", err)
	}
	if rate.Num != 15 || rate.Den != 1 {
		t.Errorf("applied rate: want 15/1, got %d/%d", rate.Num, rate.Den)
	}
	if len(lines) != 6 {
		t.Fatalf("expected 6 lines, got %d", len(lines))
	}
	for _, line := range lines {
		if line.Metadata[engine.MetaFXRateNum] != "15" || line.Metadata[engine.MetaFXRateDen] != "1" {
			t.Errorf("line metadata rate: %v", line.Metadata)
		}
	}

	if bal, _ := e.Balance(sender.ID); bal != 40000 {
		t.Errorf("sender balance: want 40000, got %d", bal)
	}
	if bal, _ := e.Balance(recipient.ID); bal != 150000 {
		t.Errorf("recipient balance: want 150000, got %d", bal)
	}
}
