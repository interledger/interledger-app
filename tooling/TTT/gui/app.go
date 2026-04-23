// Package gui implements the Bubble Tea TUI for Toy Treasury Time (TTT).
// All business rules live in the engine package; this package handles rendering,
// navigation, and parameter collection only.
package gui

import (
	"ttt/engine"

	tea "charm.land/bubbletea/v2"
	btable "github.com/evertras/bubble-table/table"
)

type appState int

const (
	stateMain appState = iota
	stateMenu
	stateForm
)

type checkResult struct {
	name   string
	passed bool
	reason string
}

// menuFrame is one level of the drill-down menu stack.
type menuFrame struct {
	title  string
	items  []menuItem
	cursor int
}

// Model is the root Bubble Tea v2 model for Toy Treasury Time (TTT).
type Model struct {
	eng    *engine.Engine
	seed   func(*engine.Engine) // optional: invoked after Reset to re-seed
	state  appState
	width  int
	height int

	// Main view
	table     btable.Model
	cursorRow int

	// Menu drill-down stack. Each frame is a list of items being shown plus
	// the current cursor position. len(menuStack)==0 when not in menu state.
	menuStack []menuFrame

	// Form
	wfIndex  int
	inputs   []formInput
	inputIdx int
	formErr  string

	// Cached integrity checks (updated after each workflow)
	checks []checkResult
}

// New creates a Model backed by the given engine. Pass an optional seed
// function that will be invoked after the "Clear Everything" action to
// repopulate providers / accounts.
func New(eng *engine.Engine, seed func(*engine.Engine)) Model {
	m := Model{
		eng:    eng,
		seed:   seed,
		state:  stateMain,
		width:  80,
		height: 24,
	}
	return m.refreshState()
}

// Init satisfies tea.Model; no initial commands needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// refreshState rebuilds the table and integrity checks from the current engine state.
func (m Model) refreshState() Model {
	m.checks = m.runChecks()
	m.table = m.buildTable()
	return m
}
