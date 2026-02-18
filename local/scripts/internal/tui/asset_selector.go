package tui

import (
	"fmt"
	"strings"

	"local-dev-tool/internal/rafiki"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	normalStyle   = lipgloss.NewStyle()
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Toggle  key.Binding
	All     key.Binding
	None    key.Binding
	Confirm key.Binding
	Quit    key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" ", "enter"),
		key.WithHelp("space/enter", "toggle"),
	),
	All: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "select all"),
	),
	None: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "select none"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("ctrl+s", "ctrl+d"),
		key.WithHelp("ctrl+s", "confirm"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type AssetSelectorModel struct {
	assets    []rafiki.Asset
	selected  map[int]bool
	cursor    int
	Cancelled bool
}

func NewAssetSelectorModel(assets []rafiki.Asset) AssetSelectorModel {
	selected := make(map[int]bool)
	// Select all by default
	for i := range assets {
		selected[i] = true
	}

	return AssetSelectorModel{
		assets:   assets,
		selected: selected,
		cursor:   0,
	}
}

func (m AssetSelectorModel) Init() tea.Cmd {
	return nil
}

func (m AssetSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			m.Cancelled = true
			return m, tea.Quit

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.assets)-1 {
				m.cursor++
			}

		case key.Matches(msg, keys.Toggle):
			m.selected[m.cursor] = !m.selected[m.cursor]

		case key.Matches(msg, keys.All):
			for i := range m.assets {
				m.selected[i] = true
			}

		case key.Matches(msg, keys.None):
			for i := range m.assets {
				m.selected[i] = false
			}

		case key.Matches(msg, keys.Confirm):
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m AssetSelectorModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Select assets to create in Rafiki"))
	b.WriteString("\n\n")

	for i, asset := range m.assets {
		cursor := "  "
		if m.cursor == i {
			cursor = selectedStyle.Render("→ ")
		}

		checked := "[ ]"
		if m.selected[i] {
			checked = selectedStyle.Render("[✓]")
		}

		line := fmt.Sprintf("%s %s %s (scale: %d)", cursor, checked, asset.Code, asset.Scale)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("space: toggle • a: all • n: none • ctrl+s: confirm • q: quit"))
	b.WriteString("\n")

	selectedCount := 0
	for _, sel := range m.selected {
		if sel {
			selectedCount++
		}
	}
	b.WriteString(helpStyle.Render(fmt.Sprintf("\nSelected: %d/%d assets", selectedCount, len(m.assets))))

	return b.String()
}

func (m AssetSelectorModel) SelectedAssets() []rafiki.Asset {
	var result []rafiki.Asset
	for i, asset := range m.assets {
		if m.selected[i] {
			result = append(result, asset)
		}
	}
	return result
}
