package gui

import (
	"github.com/charmbracelet/lipgloss"
)

type formInput interface {
	insert(text string)
	backspace()
	cycleOption(delta int)
	val() string
	render(focused bool) string
	isLocked() bool
}

// ── simpleInput: free-text field ──────────────────────────────────────────────

type simpleInput struct {
	label       string
	placeholder string
	value       string
}

func (s *simpleInput) insert(text string)    { s.value += text }
func (s *simpleInput) cycleOption(delta int) {}
func (s *simpleInput) isLocked() bool        { return false }

func (s *simpleInput) backspace() {
	if len(s.value) == 0 {
		return
	}
	runes := []rune(s.value)
	s.value = string(runes[:len(runes)-1])
}

func (s *simpleInput) val() string { return s.value }

func (s *simpleInput) render(focused bool) string {
	content := s.value
	if focused {
		content += "█"
	} else if content == "" {
		content = s.placeholder
	}

	borderCol := lipgloss.Color("240")
	if focused {
		borderCol = lipgloss.Color("86")
	}

	labelStyle := lipgloss.NewStyle()
	if focused {
		labelStyle = labelStyle.Foreground(lipgloss.Color("86"))
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Padding(0, 1).
		Width(28)

	return labelStyle.Render(s.label+":") + "\n" + boxStyle.Render(content)
}

// ── selectInput: cycle through fixed options ──────────────────────────────────

type selectInput struct {
	label   string
	options []string
	idx     int
	locked  bool // preset from menu; skipped in navigation
}

func (s *selectInput) insert(text string) {}
func (s *selectInput) backspace()         {}
func (s *selectInput) isLocked() bool     { return s.locked }

func (s *selectInput) cycleOption(delta int) {
	if s.locked || len(s.options) == 0 {
		return
	}
	n := len(s.options)
	s.idx = ((s.idx+delta)%n + n) % n
}

func (s *selectInput) val() string {
	if len(s.options) == 0 {
		return ""
	}
	return s.options[s.idx]
}

func (s *selectInput) render(focused bool) string {
	if s.locked {
		lockedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		return lockedStyle.Render(s.label + ": " + s.val() + "  (from menu)")
	}

	content := "◂ " + s.val() + " ▸"

	borderCol := lipgloss.Color("240")
	if focused {
		borderCol = lipgloss.Color("86")
	}

	labelStyle := lipgloss.NewStyle()
	if focused {
		labelStyle = labelStyle.Foreground(lipgloss.Color("86"))
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Padding(0, 1).
		Width(28)

	return labelStyle.Render(s.label+":") + "\n" + boxStyle.Render(content)
}
