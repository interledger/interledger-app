package gui

import (
	tea "charm.land/bubbletea/v2"
)

// Update handles all incoming messages and dispatches to the active state handler.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.refreshState()
		return m, nil
	case tea.KeyPressMsg:
		switch m.state {
		case stateMain:
			return m.updateMain(msg)
		case stateMenu:
			return m.updateMenu(msg)
		case stateForm:
			return m.updateForm(msg)
		case stateCrossProviderWizard:
			return m.updateCrossProviderWizard(msg)
		}
	}
	return m, nil
}

func (m Model) updateMain(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	lines, _ := m.eng.GetAllLines()
	rowCount := len(lines)

	switch msg.Code {
	case tea.KeyUp:
		if m.cursorRow > 0 {
			m.cursorRow--
			m.table = m.table.WithHighlightedRow(m.cursorRow)
		}
	case tea.KeyDown:
		if m.cursorRow < rowCount-1 {
			m.cursorRow++
			m.table = m.table.WithHighlightedRow(m.cursorRow)
		}
	case tea.KeyLeft:
		m.table = m.table.ScrollLeft()
	case tea.KeyRight:
		m.table = m.table.ScrollRight()
	default:
		switch msg.Text {
		case "k":
			if m.cursorRow > 0 {
				m.cursorRow--
				m.table = m.table.WithHighlightedRow(m.cursorRow)
			}
		case "j":
			if m.cursorRow < rowCount-1 {
				m.cursorRow++
				m.table = m.table.WithHighlightedRow(m.cursorRow)
			}
		case "m":
			m.state = stateMenu
			m.menuStack = []menuFrame{{
				title: "Workflows",
				items: buildMenuTree(m.eng),
			}}
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) updateMenu(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if len(m.menuStack) == 0 {
		m.state = stateMain
		return m, nil
	}
	top := &m.menuStack[len(m.menuStack)-1]

	move := func(delta int) {
		n := len(top.items)
		if n == 0 {
			return
		}
		next := top.cursor + delta
		if next < 0 {
			next = 0
		}
		if next >= n {
			next = n - 1
		}
		top.cursor = next
	}

	pop := func() {
		m.menuStack = m.menuStack[:len(m.menuStack)-1]
		if len(m.menuStack) == 0 {
			m.state = stateMain
		}
	}

	activate := func() (tea.Model, tea.Cmd) {
		if len(top.items) == 0 {
			return m, nil
		}
		sel := top.items[top.cursor]
		if len(sel.children) > 0 {
			// Group: drill down.
			m.menuStack = append(m.menuStack, menuFrame{
				title: sel.label,
				items: sel.children,
			})
			return m, nil
		}
		if workflowDefs[sel.wfIndex].name == "Cross-Provider Transfer" {
			m.state = stateCrossProviderWizard
			m.wfIndex = sel.wfIndex
			m.resetCrossWizard()
			return m, nil
		}
		// Leaf: open form with presets applied.
		m.state = stateForm
		m.wfIndex = sel.wfIndex
		m.formErr = ""
		m.inputs = makeInputs(workflowDefs[m.wfIndex], sel.presets, sel.optionOverrides)
		m.inputIdx = firstEditableInput(m.inputs)
		return m, nil
	}

	switch msg.Code {
	case tea.KeyEsc:
		pop()
	case tea.KeyUp:
		move(-1)
	case tea.KeyDown:
		move(1)
	case tea.KeyEnter:
		return activate()
	default:
		switch msg.Text {
		case "k":
			move(-1)
		case "j":
			move(1)
		case "q":
			m.state = stateMain
			m.menuStack = nil
		}
	}
	return m, nil
}

func (m Model) updateForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEsc:
		m.state = stateMenu
		m.formErr = ""
		if len(m.menuStack) == 0 {
			m.state = stateMain
		}
	case tea.KeyEnter:
		if m.inputIdx < lastEditableInput(m.inputs) {
			m.inputIdx = nextEditableInput(m.inputs, m.inputIdx)
		} else {
			return m.submitForm()
		}
	case tea.KeyTab:
		m.inputIdx = nextEditableInput(m.inputs, m.inputIdx)
	case tea.KeyLeft:
		m.inputs[m.inputIdx].cycleOption(-1)
	case tea.KeyRight:
		m.inputs[m.inputIdx].cycleOption(1)
	case tea.KeyBackspace:
		m.inputs[m.inputIdx].backspace()
		m.formErr = ""
	default:
		if msg.Text != "" {
			m.inputs[m.inputIdx].insert(msg.Text)
			m.formErr = ""
		}
	}
	return m, nil
}

func (m Model) submitForm() (tea.Model, tea.Cmd) {
	values := make([]string, len(m.inputs))
	for i, inp := range m.inputs {
		values[i] = inp.val()
	}

	if err := workflowDefs[m.wfIndex].run(m.eng, values); err != nil {
		m.formErr = err.Error()
		return m, nil
	}

	if hook := workflowDefs[m.wfIndex].postSubmit; hook != nil {
		hook(&m)
	}

	m.state = stateMain
	m.formErr = ""
	m.menuStack = nil

	// Advance cursor to last row so new entries are visible
	lines, _ := m.eng.GetAllLines()
	if n := len(lines); n > 0 {
		m.cursorRow = n - 1
	}

	m = m.refreshState()
	return m, nil
}
