package engine_test

import (
	"strings"
	"testing"

	"ttt/engine"
)

func TestPostJournalLines(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	sys := setupSystemAccount(t, e, "gh", eur)
	liq := setupLiquidityAccount(t, e, "gh", eur)

	posted, err := e.PostJournalLines([]engine.JournalLine{{
		DebitAccountID:  sys.ID,
		CreditAccountID: liq.ID,
		Amount:          12345,
		Metadata: map[string]string{
			engine.MetaWorkflow: "test",
			engine.MetaStep:     "seed liquidity",
		},
	}})
	if err != nil {
		t.Fatalf("PostJournalLines: %v", err)
	}
	if len(posted) != 1 {
		t.Fatalf("expected 1 posted line, got %d", len(posted))
	}
	if posted[0].ID == "" {
		t.Error("expected engine-assigned line id")
	}
	if posted[0].EventID == "" {
		t.Error("expected engine-assigned event id")
	}
	if posted[0].Timestamp.IsZero() {
		t.Error("expected engine-assigned timestamp")
	}

	lines, err := e.GetAllLines()
	if err != nil {
		t.Fatalf("GetAllLines: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 journal line, got %d", len(lines))
	}
	if lines[0].EventID != posted[0].EventID {
		t.Error("expected generated event id to propagate to stored line")
	}

	sysBal, _ := e.Balance(sys.ID)
	if sysBal != -12345 {
		t.Errorf("system balance: want -12345, got %d", sysBal)
	}
	liqBal, _ := e.Balance(liq.ID)
	if liqBal != 12345 {
		t.Errorf("liquidity balance: want 12345, got %d", liqBal)
	}
}

func TestPostJournalLinesRejectsInvalidInput(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	setupProvider(t, e, "xg", "Xago")
	ghSys := setupSystemAccount(t, e, "gh", eur)
	ghLiq := setupLiquidityAccount(t, e, "gh", eur)
	xgLiqZAR := setupLiquidityAccount(t, e, "xg", zar)

	tests := []struct {
		name  string
		lines []engine.JournalLine
		want  string
	}{
		{
			name:  "empty batch",
			lines: nil,
			want:  "at least one journal line",
		},
		{
			name: "non-positive amount",
			lines: []engine.JournalLine{{
				DebitAccountID:  ghSys.ID,
				CreditAccountID: ghLiq.ID,
				Amount:          0,
			}},
			want: "amount must be positive",
		},
		{
			name: "same account both sides",
			lines: []engine.JournalLine{{
				DebitAccountID:  ghSys.ID,
				CreditAccountID: ghSys.ID,
				Amount:          10,
			}},
			want: "must differ",
		},
		{
			name: "missing account",
			lines: []engine.JournalLine{{
				DebitAccountID:  ghSys.ID,
				CreditAccountID: "missing",
				Amount:          10,
			}},
			want: "credit account",
		},
		{
			name: "currency mismatch",
			lines: []engine.JournalLine{{
				DebitAccountID:  ghSys.ID,
				CreditAccountID: xgLiqZAR.ID,
				Amount:          10,
			}},
			want: "currency mismatch",
		},
		{
			name: "mixed event ids",
			lines: []engine.JournalLine{
				{EventID: "evt-1", DebitAccountID: ghSys.ID, CreditAccountID: ghLiq.ID, Amount: 10},
				{EventID: "evt-2", DebitAccountID: ghSys.ID, CreditAccountID: ghLiq.ID, Amount: 10},
			},
			want: "mixed event ids",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := e.PostJournalLines(tt.lines)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}

	lines, err := e.GetAllLines()
	if err != nil {
		t.Fatalf("GetAllLines: %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected no lines to be persisted on validation failure, got %d", len(lines))
	}
}

func TestWorkflowPostingGatewayRejectsUnbalancedEntries(t *testing.T) {
	e := newEngine()
	setupProvider(t, e, "gh", "GateHub")
	setupSystemAccount(t, e, "gh", eur)
	liq := setupLiquidityAccount(t, e, "gh", eur)

	err := e.ValidateJournalLines([]engine.JournalLine{{
		DebitAccountID:  liq.ID,
		CreditAccountID: liq.ID,
		Amount:          1,
	}})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("unexpected error: %v", err)
	}
}
