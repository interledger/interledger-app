package main

import (
	"fmt"
	"os"

	"ttt/cli"
	"ttt/engine"
	"ttt/engine/sqlite"
	"ttt/gui"

	tea "charm.land/bubbletea/v2"
)

// defaultDBPath is where the ledger is persisted. Override with TTT_DB.
const defaultDBPath = "ttt.db"

func main() {
	dbPath := defaultDBPath
	if v := os.Getenv("TTT_DB"); v != "" {
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

	// ── CLI mode ─────────────────────────────────────────────────────────
	// If any positional argument is present, run headless CLI and exit.
	if len(os.Args) > 1 {
		os.Exit(cli.Run(store, eng, os.Args[1:]))
	}

	// ── Paradigm selection ───────────────────────────────────────────────
	// On first run, prompt the user to pick an account topology.
	// On subsequent runs, read the stored paradigm (refuse to start if invalid).
	paradigmSet, err := store.IsParadigmSet()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config check:", err)
		os.Exit(1)
	}

	if !paradigmSet {
		paradigm, err := selectParadigm()
		if err != nil {
			fmt.Fprintln(os.Stderr, "paradigm selection:", err)
			os.Exit(1)
		}
		if err := store.SetParadigm(paradigm); err != nil {
			fmt.Fprintln(os.Stderr, "saving paradigm:", err)
			os.Exit(1)
		}
	}

	paradigm, err := store.GetParadigm()
	if err != nil {
		fmt.Fprintln(os.Stderr, "reading paradigm:", err)
		os.Exit(1)
	}

	// Seed engine with the chosen topology (idempotent).
	if err := engine.SeedParadigm(paradigm, eng); err != nil {
		fmt.Fprintln(os.Stderr, "seed engine:", err)
		os.Exit(1)
	}

	// The reset callback re-seeds using the same stored paradigm.
	seed := func(e *engine.Engine) {
		p, err := store.GetParadigm()
		if err != nil {
			fmt.Fprintln(os.Stderr, "seed after reset:", err)
			return
		}
		if err := engine.SeedParadigm(p, e); err != nil {
			fmt.Fprintln(os.Stderr, "seed after reset:", err)
		}
	}

	m := gui.New(eng, seed)
	prog := tea.NewProgram(m)
	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// selectParadigm runs a minimal Bubble Tea program to let the user choose an
// account topology on first run. Returns an error if the user quit without
// making a selection.
func selectParadigm() (engine.Paradigm, error) {
	sel := gui.NewParadigmSelector()
	prog := tea.NewProgram(sel)
	result, err := prog.Run()
	if err != nil {
		return 0, fmt.Errorf("paradigm selector: %w", err)
	}
	chosen := result.(gui.ParadigmSelectorModel).Selected()
	if chosen == 0 {
		return 0, fmt.Errorf("no paradigm selected; exiting")
	}
	return chosen, nil
}
