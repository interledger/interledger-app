package gui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"ttt/engine"
)

// ParadigmSelectorModel is a standalone Bubble Tea model used on first run to
// let the user choose an account topology (paradigm) before the main TUI starts.
type ParadigmSelectorModel struct {
	cursor   int
	selected engine.Paradigm // non-zero once the user confirms a choice
}

// NewParadigmSelector creates the paradigm-selection model.
func NewParadigmSelector() ParadigmSelectorModel {
	return ParadigmSelectorModel{}
}

// Selected returns the chosen paradigm. Call after the program exits.
func (m ParadigmSelectorModel) Selected() engine.Paradigm {
	return m.selected
}

// Init satisfies tea.Model.
func (m ParadigmSelectorModel) Init() tea.Cmd { return nil }

// Update handles navigation and selection.
func (m ParadigmSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(engine.ValidParadigms)-1 {
				m.cursor++
			}
		case tea.KeyEnter:
			m.selected = engine.ValidParadigms[m.cursor]
			return m, tea.Quit
		case tea.KeyEsc:
			// treat Escape as "quit without selecting" — caller will detect selected==0
			return m, tea.Quit
		default:
			switch msg.Text {
			case "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "j":
				if m.cursor < len(engine.ValidParadigms)-1 {
					m.cursor++
				}
			case "q":
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

// View renders the paradigm selection screen.
func (m ParadigmSelectorModel) View() tea.View {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Toy Treasury Time — First Run Setup"))
	b.WriteString("\n\n")

	intro := subtleStyle.Render(
		"No database found. Choose an account topology for this session.\n" +
			"This choice is permanent until you clear or delete the database.",
	)
	b.WriteString(intro)
	b.WriteString("\n\n")

	for i, p := range engine.ValidParadigms {
		line := fmt.Sprintf("  %d. %s", i+1, p.Name())
		if i == m.cursor {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("62")).
				Bold(true).
				Render("> " + line[2:]))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(subtleStyle.Render("[↑↓ / jk] select  [Enter] confirm  [q] quit"))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}
