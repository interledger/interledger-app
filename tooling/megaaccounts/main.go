package main

import (
	"fmt"
	"os"
	"strings"

	"megaaccounts/engine"
	"megaaccounts/engine/sqlite"
	"megaaccounts/gui"

	tea "charm.land/bubbletea/v2"
)

// defaultDBPath is where the ledger is persisted. Override with MEGAACCOUNTS_DB.
const defaultDBPath = "megaaccounts.db"

func main() {
	dbPath := defaultDBPath
	if v := os.Getenv("MEGAACCOUNTS_DB"); v != "" {
		dbPath = v
	}

	store, err := sqlite.Open(dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open db:", err)
		os.Exit(1)
	}
	defer func() { _ = store.Close() }()

	eng := engine.New(store)

	// FX simulator — start with 1 EUR = 15 ZAR; post-conversion mutation
	// direction is drawn from a time-seeded random source.
	fx := engine.NewFXService(nil)
	if err := fx.Set("EUR", "ZAR", 15, 1); err != nil {
		fmt.Fprintln(os.Stderr, "fx setup:", err)
		os.Exit(1)
	}
	eng.WithFX(fx)

	setupEngine(eng)

	m := gui.New(eng, setupEngine)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// setupEngine seeds the simulation with providers, liquidity accounts, and
// user accounts. Safe to call on an already-seeded database — "already
// exists" errors from the engine are ignored so re-runs are idempotent.
func setupEngine(eng *engine.Engine) {
	ignoreExists(eng.CreateProvider("gatehub", "GateHub"))
	ignoreExists(eng.CreateProvider("xago", "Xago"))

	// GateHub — EUR only.
	ignoreExistsAcct(eng.CreateSystemAccount("gatehub", engine.EUR))
	ignoreExistsAcct(eng.CreateLiquidityAccount("gatehub", engine.EUR))

	// Xago — ZAR (primary) and EUR (needed to host the mirror EUR position
	// account for cross-provider settlement with GateHub).
	ignoreExistsAcct(eng.CreateSystemAccount("xago", engine.ZAR))
	ignoreExistsAcct(eng.CreateLiquidityAccount("xago", engine.ZAR))
	ignoreExistsAcct(eng.CreateSystemAccount("xago", engine.EUR))
	ignoreExistsAcct(eng.CreateLiquidityAccount("xago", engine.EUR))

	ignoreExistsAcct(eng.CreateUserAccount("alice", "gatehub", engine.EUR))
	ignoreExistsAcct(eng.CreateUserAccount("bob", "gatehub", engine.EUR))
	ignoreExistsAcct(eng.CreateUserAccount("carlos", "xago", engine.ZAR))
}

// ignoreExists panics on unexpected errors but tolerates "already exists"
// failures produced by the engine when seeding a previously-populated store.
func ignoreExists(_ engine.Provider, err error) {
	if err == nil || strings.Contains(err.Error(), "already exists") {
		return
	}
	panic(fmt.Sprintf("engine setup: %v", err))
}

func ignoreExistsAcct(_ engine.Account, err error) {
	if err == nil || strings.Contains(err.Error(), "already exists") {
		return
	}
	panic(fmt.Sprintf("engine setup: %v", err))
}
