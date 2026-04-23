package gui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	btable "github.com/evertras/bubble-table/table"

	"ttt/engine"
)

// ── styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	passStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	failStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("62")).
				Bold(true)
)

// ── View ──────────────────────────────────────────────────────────────────────

// View renders the current state to a Bubble Tea v2 View with alt-screen enabled.
func (m Model) View() tea.View {
	var content string
	switch m.state {
	case stateMain:
		content = m.renderMain()
	case stateMenu:
		content = m.renderMenu()
	case stateForm:
		content = m.renderForm()
	case stateCrossProviderWizard:
		content = m.renderCrossProviderWizard()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// ── state renderers ───────────────────────────────────────────────────────────

func (m Model) renderMain() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Toy Treasury Time"))
	b.WriteString("  ")
	b.WriteString(subtleStyle.Render("[m] menu  [q] quit  [← →] scroll  [↑ ↓ / j k] navigate"))
	b.WriteString("\n\n")

	lines, _ := m.eng.GetAllLines()
	if len(lines) == 0 {
		b.WriteString(subtleStyle.Render("No activity yet — press [m] to open the workflow menu."))
	} else {
		b.WriteString(m.renderSelectedInfo(lines))
		b.WriteString("\n\n")
		b.WriteString(m.renderBalancesPanel(lines))
		b.WriteString("\n\n")
		b.WriteString(m.table.View())
	}

	b.WriteString("\n")
	b.WriteString(m.renderChecks())
	if fxLine := m.renderFXRates(); fxLine != "" {
		b.WriteString("\n")
		b.WriteString(fxLine)
	}

	return b.String()
}

func (m Model) renderMenu() string {
	if len(m.menuStack) == 0 {
		return ""
	}
	top := m.menuStack[len(m.menuStack)-1]

	var content strings.Builder
	content.WriteString(titleStyle.Render(top.title))
	if crumb := m.menuBreadcrumb(); crumb != "" {
		content.WriteString("  ")
		content.WriteString(subtleStyle.Render(crumb))
	}
	content.WriteString("\n\n")

	for i, item := range top.items {
		label := item.label
		if len(item.children) > 0 {
			label += "  ▸"
		}
		if i == top.cursor {
			content.WriteString(selectedItemStyle.Render("> " + label))
		} else {
			content.WriteString("  " + label)
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("[↑↓ / jk] select  [Enter] open  [Esc] back"))

	var b strings.Builder
	b.WriteString(titleStyle.Render("Toy Treasury Time") + "\n\n")
	b.WriteString(panelStyle.Render(content.String()))
	return b.String()
}

// menuBreadcrumb renders the path above the current frame (e.g. "Workflows › Gatehub").
func (m Model) menuBreadcrumb() string {
	if len(m.menuStack) <= 1 {
		return ""
	}
	parts := make([]string, 0, len(m.menuStack)-1)
	for _, f := range m.menuStack[:len(m.menuStack)-1] {
		parts = append(parts, f.title)
	}
	return strings.Join(parts, " › ")
}

func (m Model) renderForm() string {
	wf := workflowDefs[m.wfIndex]

	var content strings.Builder
	content.WriteString(titleStyle.Render(wf.name) + "\n\n")

	for i, inp := range m.inputs {
		content.WriteString(inp.render(i == m.inputIdx))
		content.WriteString("\n")
	}

	if wf.name == "Bilateral Settlement" {
		if summary := m.renderSettlementFormSummary(); summary != "" {
			content.WriteString("\n")
			content.WriteString(subtleStyle.Render(summary))
			content.WriteString("\n")
		}
	}

	if m.formErr != "" {
		content.WriteString("\n")
		wrapWidth := m.width - 12
		if wrapWidth < 30 {
			wrapWidth = 30
		}
		content.WriteString(errorStyle.Render("✗ " + wrapWords(m.formErr, wrapWidth)))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("[Tab / Enter] next field  [← →] cycle option  [Enter on last] submit  [Esc] back"))

	var b strings.Builder
	b.WriteString(titleStyle.Render("Toy Treasury Time") + "\n\n")
	b.WriteString(panelStyle.Render(content.String()))
	return b.String()
}

func (m Model) renderSettlementFormSummary() string {
	if len(m.inputs) < 3 {
		return ""
	}
	providerA := trim(m.inputs[0].val())
	providerB := trim(m.inputs[1].val())
	if providerA == "" || providerB == "" {
		return ""
	}
	cur, err := parseCurrency(m.inputs[2].val())
	if err != nil {
		return ""
	}
	return settlementPreviewText(m.eng, providerA, providerB, cur)
}

// ── integrity checks panel ────────────────────────────────────────────────────

func (m Model) renderChecks() string {
	if len(m.checks) == 0 {
		return subtleStyle.Render("No checks yet")
	}

	var parts []string
	for _, c := range m.checks {
		if c.passed {
			parts = append(parts, passStyle.Render("✓ "+c.name))
		} else {
			reason := c.reason
			if len(reason) > 50 {
				reason = reason[:50] + "…"
			}
			parts = append(parts, failStyle.Render("✗ "+c.name+": "+reason))
		}
	}

	return "Integrity  " + strings.Join(parts, "  │  ")
}

// renderFXRates summarises the current simulator forex rates, if any.
func (m Model) renderFXRates() string {
	fx := m.eng.FX()
	if fx == nil {
		return ""
	}
	snap := fx.Snapshot()
	if len(snap) == 0 {
		return ""
	}
	// Stable order.
	keys := make([][2]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		r := snap[k]
		parts = append(parts, fmt.Sprintf("1 %s = %s %s",
			k[0], formatRate(r), k[1]))
	}
	return "FX  " + strings.Join(parts, "  │  ")
}

// formatRate renders a rational rate with up to 4 decimal places for display.
func formatRate(r engine.Rate) string {
	if r.Den == 0 {
		return "∞"
	}
	if r.Num%r.Den == 0 {
		return fmt.Sprintf("%d", r.Num/r.Den)
	}
	// Four decimals is enough resolution for ±5 % chains without visual noise.
	whole := r.Num / r.Den
	rem := r.Num - whole*r.Den
	if rem < 0 {
		rem = -rem
	}
	frac := (rem * 10000) / r.Den
	return fmt.Sprintf("%d.%04d", whole, frac)
}

func (m Model) runChecks() []checkResult {
	_, globalErr := m.eng.CheckGlobalBalance()
	results := []checkResult{{
		name:   "global balance",
		passed: globalErr == nil,
		reason: errString(globalErr),
	}}

	lines, _ := m.eng.GetAllLines()
	if len(lines) > 0 {
		lastEventID := lines[len(lines)-1].EventID
		eventErr := m.eng.CheckPerEventBalance(lastEventID)
		results = append(results, checkResult{
			name:   "last event",
			passed: eventErr == nil,
			reason: errString(eventErr),
		})
	}

	_, decompErr := m.eng.CheckLiquidityDecomposition()
	results = append(results, checkResult{
		name:   "liquidity decomp",
		passed: decompErr == nil,
		reason: errString(decompErr),
	})

	_, mirrorErr := m.eng.CheckBilateralPositions()
	results = append(results, checkResult{
		name:   "bilateral mirror",
		passed: mirrorErr == nil,
		reason: errString(mirrorErr),
	})

	return results
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func compactWorkflowName(workflow string) string {
	switch workflow {
	case engine.WorkflowFundProviderLiquidity:
		return "FUND-LIQ"
	case engine.WorkflowUserOnboard:
		return "ONBOARD"
	case engine.WorkflowP2PSameProvider:
		return "P2P"
	case engine.WorkflowUserOffboard:
		return "OFFBOARD"
	case engine.WorkflowCrossProviderXfer:
		return "XPROV-XFER"
	case engine.WorkflowBilateralSettlement:
		return "SETTLE"
	default:
		return shortStr(workflow, 12)
	}
}

func compactStepText(step string) string {
	if step == "" {
		return ""
	}
	r := strings.NewReplacer(
		"debit", "dr",
		"credit", "cr",
		"recipient", "rcpt",
		"sender", "snd",
		"liquidity", "liq",
		"position", "pos",
		"system", "sys",
		"counterparty", "cp",
		"settlement", "stl",
		"transfer", "xfer",
		"receivable", "recv",
	)
	step = r.Replace(step)
	return shortStr(step, 28)
}

func wrapWords(text string, width int) string {
	if width <= 0 || len(text) <= width {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var out strings.Builder
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) <= width {
			line += " " + w
			continue
		}
		out.WriteString(line)
		out.WriteString("\n")
		line = w
	}
	out.WriteString(line)
	return out.String()
}

func formatStepFull(line engine.JournalLine) string {
	debitStep := line.DebitMetadata[engine.MetaStep]
	creditStep := line.CreditMetadata[engine.MetaStep]
	step := line.Metadata[engine.MetaStep]
	if step == "" {
		switch {
		case debitStep != "" && creditStep != "" && debitStep != creditStep:
			step = debitStep + " -> " + creditStep
		case debitStep != "":
			step = debitStep
		case creditStep != "":
			step = creditStep
		}
	}
	if step == "" {
		return ""
	}
	num := line.Metadata[engine.MetaFXRateNum]
	den := line.Metadata[engine.MetaFXRateDen]
	if num == "" || den == "" {
		return step
	}
	n, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return step
	}
	d, err := strconv.ParseInt(den, 10, 64)
	if err != nil || d == 0 {
		return step
	}
	return step + " @ " + formatRate(engine.Rate{Num: n, Den: d})
}

func (m Model) renderSelectedInfo(lines []engine.JournalLine) string {
	if len(lines) == 0 {
		return ""
	}
	idx := m.cursorRow
	if idx < 0 {
		idx = 0
	}
	if idx >= len(lines) {
		idx = len(lines) - 1
	}
	line := lines[idx]

	wf := line.Metadata[engine.MetaWorkflow]
	if wf == "" {
		wf = "—"
	}
	step := formatStepFull(line)
	if step == "" {
		step = "—"
	}

	wrapWidth := m.width - 18
	if wrapWidth < 30 {
		wrapWidth = 30
	}

	var body strings.Builder
	body.WriteString(titleStyle.Render("Selected Row"))
	body.WriteString("\n")
	body.WriteString(subtleStyle.Render("Workflow: "))
	body.WriteString(wrapWords(wf, wrapWidth))
	body.WriteString("\n")
	body.WriteString(subtleStyle.Render("Step: "))
	body.WriteString(wrapWords(step, wrapWidth))

	return panelStyle.Render(body.String())
}

// ── table builder ─────────────────────────────────────────────────────────────

// formatStep renders a journal line step label. If debit and credit sides have
// distinct step text, both are shown to keep side semantics visible.
func formatStep(line engine.JournalLine) string {
	return compactStepText(formatStepFull(line))
}

// renderBalancesPanel builds a compact "as-of" view of every active
// account's balance at the highlighted journal line row, plus contextual stats.
func (m Model) renderBalancesPanel(lines []engine.JournalLine) string {
	if len(lines) == 0 {
		return ""
	}

	// Clamp cursor to the valid line range.
	idx := m.cursorRow
	if idx < 0 {
		idx = 0
	}
	if idx >= len(lines) {
		idx = len(lines) - 1
	}
	highlighted := lines[idx]

	// Build credit-normal balances for every account up to and including idx.
	balances := make(map[string]int64)
	for i := 0; i <= idx; i++ {
		line := lines[i]
		balances[line.DebitAccountID] -= line.Amount
		balances[line.CreditAccountID] += line.Amount
	}

	accounts, _ := m.eng.ListAccounts()
	accountMap := make(map[string]engine.Account, len(accounts))
	for _, a := range accounts {
		accountMap[a.ID] = a
	}

	// Only show accounts that have had at least one entry up to this point.
	active := make([]engine.Account, 0)
	seen := map[string]bool{}
	for i := 0; i <= idx; i++ {
		debitID := lines[i].DebitAccountID
		if !seen[debitID] {
			seen[debitID] = true
			if a, ok := accountMap[debitID]; ok {
				active = append(active, a)
			}
		}
		creditID := lines[i].CreditAccountID
		if !seen[creditID] {
			seen[creditID] = true
			if a, ok := accountMap[creditID]; ok {
				active = append(active, a)
			}
		}
	}
	sortAccounts(active)

	// Contextual header line.
	wf := highlighted.Metadata[engine.MetaWorkflow]
	if wf == "" {
		wf = "—"
	}

	header := fmt.Sprintf("Balances as of %s  ·  entry %d/%d  ·  workflow: %s  ·  event: %s",
		highlighted.Timestamp.Format("15:04:05"),
		idx+1, len(lines),
		wf,
		shortStr(highlighted.EventID, 8),
	)

	// Build a two-column-ish grid: "LABEL  VALUE".
	type row struct {
		label string
		value string
	}
	rows := make([]row, 0, len(active))
	for _, a := range active {
		bal := balances[a.ID]
		rows = append(rows, row{
			label: accountDisplayLabel(a),
			value: formatAmount(bal, a.Currency.AssetScale) + " " + a.Currency.Code,
		})
	}

	// Pack into columns so the panel doesn't become absurdly tall.
	const maxRowsPerCol = 8
	numCols := (len(rows) + maxRowsPerCol - 1) / maxRowsPerCol
	if numCols < 1 {
		numCols = 1
	}
	rowsPerCol := (len(rows) + numCols - 1) / numCols

	// Compute label width per column for alignment.
	cols := make([][]row, numCols)
	for i, r := range rows {
		c := i / rowsPerCol
		cols[c] = append(cols[c], r)
	}
	labelStyle := subtleStyle
	valueStyle := lipgloss.NewStyle().Bold(true)

	var grid strings.Builder
	for rIdx := 0; rIdx < rowsPerCol; rIdx++ {
		for cIdx, col := range cols {
			if rIdx >= len(col) {
				continue
			}
			r := col[rIdx]
			grid.WriteString(labelStyle.Render(padRight(r.label, 16)))
			grid.WriteString(" ")
			grid.WriteString(valueStyle.Render(padRight(r.value, 18)))
			if cIdx < len(cols)-1 {
				grid.WriteString("   ")
			}
		}
		grid.WriteString("\n")
	}

	var panel strings.Builder
	panel.WriteString(titleStyle.Render("Balances"))
	panel.WriteString("\n")
	panel.WriteString(subtleStyle.Render(header))
	panel.WriteString("\n")
	panel.WriteString(grid.String())
	return panelStyle.Render(panel.String())
}

// accountDisplayLabel renders a human-readable account identifier for the
// balances panel — slightly more descriptive than the ledger column header.
func accountDisplayLabel(a engine.Account) string {
	switch a.Type {
	case engine.AccountTypeSystem:
		return fmt.Sprintf("%s SYS %s", a.ProviderID, a.Currency.Code)
	case engine.AccountTypeLiquidity:
		return fmt.Sprintf("%s LIQ %s", a.ProviderID, a.Currency.Code)
	case engine.AccountTypePosition:
		return fmt.Sprintf("%s POS→%s %s", a.ProviderID, a.CounterpartyID, a.Currency.Code)
	case engine.AccountTypeUser:
		return fmt.Sprintf("%s@%s %s", a.UserID, a.ProviderID, a.Currency.Code)
	default:
		return shortStr(a.ID, 10)
	}
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

const (
	colKeyTime     = "time"
	colKeyWorkflow = "workflow"
	colKeyStep     = "step"
)

func (m Model) buildTable() btable.Model {
	lines, _ := m.eng.GetAllLines()
	accounts, _ := m.eng.ListAccounts()

	// Identify active accounts (those touched by at least one journal line).
	activeIDs := make(map[string]bool, len(lines)*2)
	for _, line := range lines {
		activeIDs[line.DebitAccountID] = true
		activeIDs[line.CreditAccountID] = true
	}

	accountMap := make(map[string]engine.Account, len(accounts))
	activeAccounts := make([]engine.Account, 0, len(activeIDs))
	for _, a := range accounts {
		accountMap[a.ID] = a
		if activeIDs[a.ID] {
			activeAccounts = append(activeAccounts, a)
		}
	}
	sortAccounts(activeAccounts)

	// Build columns: frozen metadata columns, then one column per active account.
	columns := []btable.Column{
		btable.NewColumn(colKeyTime, "Time", 10),
		btable.NewColumn(colKeyWorkflow, "Workflow", 12),
		btable.NewColumn(colKeyStep, "Step", 30),
	}
	for _, a := range activeAccounts {
		columns = append(columns, btable.NewColumn(accountColKey(a), accountColTitle(a), 14))
	}

	// Build rows: one per journal line, projected into account columns with D/C markers.
	rows := make([]btable.Row, 0, len(lines))
	for _, line := range lines {
		data := btable.RowData{
			colKeyTime:     line.Timestamp.Format("15:04:05"),
			colKeyWorkflow: compactWorkflowName(line.Metadata[engine.MetaWorkflow]),
			colKeyStep:     formatStep(line),
		}

		for _, a := range activeAccounts {
			data[accountColKey(a)] = ""
		}

		if debitAcct, ok := accountMap[line.DebitAccountID]; ok {
			data[accountColKey(debitAcct)] = "D " + formatAmount(line.Amount, debitAcct.Currency.AssetScale)
		}
		if creditAcct, ok := accountMap[line.CreditAccountID]; ok {
			data[accountColKey(creditAcct)] = "C " + formatAmount(line.Amount, creditAcct.Currency.AssetScale)
		}
		rows = append(rows, btable.NewRow(data))
	}

	// Clamp cursor row to valid range
	cursor := m.cursorRow
	if n := len(rows); n > 0 {
		if cursor >= n {
			cursor = n - 1
		}
		if cursor < 0 {
			cursor = 0
		}
	}

	pageSize := m.height - 10
	if pageSize < 1 {
		pageSize = 1
	}
	width := m.width
	if width <= 0 {
		width = 80
	}

	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true)

	t := btable.New(columns).
		Focused(true).
		WithBaseStyle(baseStyle).
		HighlightStyle(highlightStyle).
		WithRows(rows).
		WithMaxTotalWidth(width).
		WithHorizontalFreezeColumnCount(3).
		WithPageSize(pageSize)

	if len(rows) > 0 {
		t = t.WithHighlightedRow(cursor)
	}

	return t
}

// ── helpers ───────────────────────────────────────────────────────────────────

func sortAccounts(accounts []engine.Account) {
	sort.Slice(accounts, func(i, j int) bool {
		a, b := accounts[i], accounts[j]
		if a.ProviderID != b.ProviderID {
			return a.ProviderID < b.ProviderID
		}
		if a.Type != b.Type {
			return int(a.Type) < int(b.Type)
		}
		if a.UserID != b.UserID {
			return a.UserID < b.UserID
		}
		return a.Currency.Code < b.Currency.Code
	})
}

func accountColKey(a engine.Account) string {
	return a.ID
}

func accountColTitle(a engine.Account) string {
	prov := strings.ToUpper(a.ProviderID)
	if len(prov) > 3 {
		prov = prov[:3]
	}
	switch a.Type {
	case engine.AccountTypeSystem:
		return fmt.Sprintf("SYS/%s/%s", prov, a.Currency.Code)
	case engine.AccountTypeLiquidity:
		return fmt.Sprintf("LIQ/%s/%s", prov, a.Currency.Code)
	case engine.AccountTypePosition:
		cp := strings.ToUpper(a.CounterpartyID)
		if len(cp) > 3 {
			cp = cp[:3]
		}
		return fmt.Sprintf("POS/%s/%s", cp, a.Currency.Code)
	case engine.AccountTypeUser:
		name := a.UserID
		if len(name) > 8 {
			name = name[:8]
		}
		return fmt.Sprintf("%s/%s", name, a.Currency.Code)
	default:
		return shortStr(a.ID, 8)
	}
}

func formatAmount(amount int64, scale int) string {
	if scale == 0 {
		return strconv.FormatInt(amount, 10)
	}
	pow := int64(1)
	for i := 0; i < scale; i++ {
		pow *= 10
	}
	whole := amount / pow
	frac := amount % pow
	return fmt.Sprintf("%d.%0*d", whole, scale, frac)
}

func shortStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
