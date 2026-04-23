package gui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"ttt/engine"
)

func (m *Model) resetCrossWizard() {
	m.crossStep = crossStepSender
	m.crossSenderUserID = ""
	m.crossSenderAcct = nil
	m.crossRecipientUserID = ""
	m.crossRecipientAcct = nil
	m.crossAmountInput = ""
	m.formErr = ""
}

func (m Model) updateCrossProviderWizard(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEsc:
		m.state = stateMenu
		m.resetCrossWizard()
		return m, nil
	case tea.KeyBackspace:
		switch m.crossStep {
		case crossStepSender:
			m.crossSenderUserID = backspaceString(m.crossSenderUserID)
		case crossStepRecipient:
			m.crossRecipientUserID = backspaceString(m.crossRecipientUserID)
		case crossStepAmount:
			m.crossAmountInput = backspaceString(m.crossAmountInput)
		}
		m.formErr = ""
		return m, nil
	case tea.KeyEnter:
		return m.advanceCrossWizard()
	}

	if msg.Text != "" {
		switch m.crossStep {
		case crossStepSender:
			m.crossSenderUserID += msg.Text
		case crossStepRecipient:
			m.crossRecipientUserID += msg.Text
		case crossStepAmount:
			m.crossAmountInput += msg.Text
		}
		m.formErr = ""
	}
	return m, nil
}

func (m Model) advanceCrossWizard() (tea.Model, tea.Cmd) {
	switch m.crossStep {
	case crossStepSender:
		acct, err := lookupUniqueUserAccount(m.eng, trim(m.crossSenderUserID))
		if err != nil {
			m.formErr = err.Error()
			return m, nil
		}
		m.crossSenderAcct = &acct
		m.crossStep = crossStepRecipient
		return m, nil

	case crossStepRecipient:
		acct, err := lookupUniqueUserAccount(m.eng, trim(m.crossRecipientUserID))
		if err != nil {
			m.formErr = err.Error()
			return m, nil
		}
		if m.crossSenderAcct != nil && acct.ProviderID == m.crossSenderAcct.ProviderID {
			m.formErr = "recipient must belong to a different provider for cross-provider transfer"
			return m, nil
		}
		m.crossRecipientAcct = &acct
		m.crossStep = crossStepAmount
		return m, nil

	case crossStepAmount:
		if m.crossSenderAcct == nil || m.crossRecipientAcct == nil {
			m.formErr = "sender and recipient must be resolved first"
			return m, nil
		}
		if _, err := parseAmount(trim(m.crossAmountInput), m.crossSenderAcct.Currency.AssetScale); err != nil {
			m.formErr = err.Error()
			return m, nil
		}
		m.crossStep = crossStepConfirm
		return m, nil

	case crossStepConfirm:
		if strings.ToUpper(trim(m.crossAmountInput)) == "" {
			m.formErr = "amount is required"
			return m, nil
		}
		if m.crossSenderAcct == nil || m.crossRecipientAcct == nil {
			m.formErr = "sender and recipient must be resolved first"
			return m, nil
		}
		amount, err := parseAmount(trim(m.crossAmountInput), m.crossSenderAcct.Currency.AssetScale)
		if err != nil {
			m.formErr = err.Error()
			return m, nil
		}
		// Pre-flight: check sender can cover dispatch + charge before submitting.
		senderBal, err := m.eng.Balance(m.crossSenderAcct.ID)
		if err != nil {
			m.formErr = "balance lookup failed: " + err.Error()
			return m, nil
		}
		charge, _ := m.eng.GetCharge(m.crossSenderAcct.ProviderID, m.crossRecipientAcct.ProviderID)
		chargeAmount := charge.ChargeAmount(amount)
		totalCost := amount + chargeAmount
		if senderBal < totalCost {
			cur := m.crossSenderAcct.Currency
			m.formErr = fmt.Sprintf(
				"insufficient balance: need %s %s (dispatch + charge), have %s %s — adjust the amount or add funds",
				formatMinor(totalCost, cur.AssetScale), cur.Code,
				formatMinor(senderBal, cur.AssetScale), cur.Code,
			)
			return m, nil
		}
		_, rate, err := m.eng.CrossProviderTransferAutoLines(
			trim(m.crossSenderUserID), m.crossSenderAcct.ProviderID, m.crossSenderAcct.Currency,
			trim(m.crossRecipientUserID), m.crossRecipientAcct.ProviderID, m.crossRecipientAcct.Currency,
			amount,
		)
		if err != nil {
			m.formErr = err.Error()
			return m, nil
		}
		_ = rate // rate is displayed pre-submit via renderCrossWizardSummary.

		m.state = stateMain
		m.menuStack = nil
		m.resetCrossWizard()

		lines, _ := m.eng.GetAllLines()
		if n := len(lines); n > 0 {
			m.cursorRow = n - 1
		}
		m = m.refreshState()
		return m, nil
	}

	return m, nil
}

func (m Model) renderCrossProviderWizard() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("Cross-Provider Transfer"))
	content.WriteString("\n\n")
	content.WriteString(subtleStyle.Render("Step-by-step flow with automatic user account lookup."))
	content.WriteString("\n\n")

	sender := trim(m.crossSenderUserID)
	if sender == "" && m.crossStep == crossStepSender {
		sender = "█"
	}
	recipient := trim(m.crossRecipientUserID)
	if recipient == "" && m.crossStep == crossStepRecipient {
		recipient = "█"
	}
	amount := trim(m.crossAmountInput)
	if amount == "" && m.crossStep == crossStepAmount {
		amount = "█"
	}

	content.WriteString(stepLine(m.crossStep == crossStepSender, "1. Sender user ID", sender))
	if m.crossSenderAcct != nil {
		content.WriteString("\n")
		content.WriteString(subtleStyle.Render(fmt.Sprintf("   -> provider=%s currency=%s", m.crossSenderAcct.ProviderID, m.crossSenderAcct.Currency.Code)))
	}
	content.WriteString("\n\n")

	content.WriteString(stepLine(m.crossStep == crossStepRecipient, "2. Recipient user ID", recipient))
	if m.crossRecipientAcct != nil {
		content.WriteString("\n")
		content.WriteString(subtleStyle.Render(fmt.Sprintf("   -> provider=%s currency=%s", m.crossRecipientAcct.ProviderID, m.crossRecipientAcct.Currency.Code)))
	}
	content.WriteString("\n\n")

	content.WriteString(stepLine(m.crossStep == crossStepAmount, "3. Amount (sender currency)", amount))
	content.WriteString("\n\n")

	content.WriteString(stepLine(m.crossStep == crossStepConfirm, "4. Confirm", "Press Enter to submit"))

	if summary := m.renderCrossWizardSummary(); summary != "" {
		content.WriteString("\n\n")
		content.WriteString(summary)
	}

	if m.formErr != "" {
		content.WriteString("\n\n")
		content.WriteString(errorStyle.Render("✗ " + m.formErr))
	}

	content.WriteString("\n\n")
	content.WriteString(subtleStyle.Render("[Enter] continue/confirm  [Backspace] edit  [Esc] back"))

	var b strings.Builder
	b.WriteString(titleStyle.Render("Toy Treasury Time") + "\n\n")
	b.WriteString(panelStyle.Render(content.String()))
	return b.String()
}

func (m Model) renderCrossWizardSummary() string {
	if m.crossSenderAcct == nil || m.crossRecipientAcct == nil {
		return ""
	}
	amountText := trim(m.crossAmountInput)
	if amountText == "" {
		return ""
	}
	amount, err := parseAmount(amountText, m.crossSenderAcct.Currency.AssetScale)
	if err != nil || amount <= 0 {
		return ""
	}
	fx := m.eng.FX()
	if fx == nil {
		return subtleStyle.Render("FX unavailable (no FX service attached).")
	}
	rate, err := fx.Rate(m.crossSenderAcct.Currency.Code, m.crossRecipientAcct.Currency.Code)
	if err != nil {
		return subtleStyle.Render("FX rate unavailable for this currency pair.")
	}

	srcCur := m.crossSenderAcct.Currency
	dstCur := m.crossRecipientAcct.Currency
	destAmount := (amount * rate.Num) / rate.Den

	// Look up configured charge for this direction.
	charge, _ := m.eng.GetCharge(m.crossSenderAcct.ProviderID, m.crossRecipientAcct.ProviderID)
	chargeAmount := charge.ChargeAmount(amount)
	totalCost := amount + chargeAmount

	var lines []string
	lines = append(lines, fmt.Sprintf(
		"Dispatch:  %s %s  →  %s (%s)",
		formatMinor(amount, srcCur.AssetScale), srcCur.Code,
		trim(m.crossRecipientUserID), m.crossRecipientAcct.ProviderID,
	))

	if charge != nil {
		lines = append(lines, fmt.Sprintf(
			"Charge:    %s %s (%s, paid by sender, stays with %s)",
			formatMinor(chargeAmount, srcCur.AssetScale), srcCur.Code,
			formatChargePercent(*charge),
			m.crossSenderAcct.ProviderID,
		))
		lines = append(lines, fmt.Sprintf(
			"Total:     %s %s debited from sender",
			formatMinor(totalCost, srcCur.AssetScale), srcCur.Code,
		))
	}

	lines = append(lines, fmt.Sprintf(
		"FX rate:   1 %s = %s %s  →  recipient gets ≈ %s %s",
		srcCur.Code, formatRate(rate), dstCur.Code,
		formatMinor(destAmount, dstCur.AssetScale), dstCur.Code,
	))

	// Balance warning when charge pushes total above sender's available balance.
	if senderBal, err := m.eng.Balance(m.crossSenderAcct.ID); err == nil && senderBal < totalCost {
		lines = append(lines, fmt.Sprintf(
			"⚠ Insufficient balance: need %s, have %s %s — cannot submit",
			formatMinor(totalCost, srcCur.AssetScale),
			formatMinor(senderBal, srcCur.AssetScale), srcCur.Code,
		))
	}

	return subtleStyle.Render(strings.Join(lines, "\n"))
}

func stepLine(active bool, label, value string) string {
	line := fmt.Sprintf("%s: %s", label, value)
	if active {
		return selectedItemStyle.Render(line)
	}
	return line
}

func lookupUniqueUserAccount(eng *engine.Engine, userID string) (engine.Account, error) {
	if userID == "" {
		return engine.Account{}, fmt.Errorf("user id is required")
	}
	accounts, err := eng.ListAccounts()
	if err != nil {
		return engine.Account{}, err
	}
	var match *engine.Account
	for i := range accounts {
		a := accounts[i]
		if a.Type != engine.AccountTypeUser {
			continue
		}
		if a.UserID != userID {
			continue
		}
		if match != nil {
			return engine.Account{}, fmt.Errorf("user %q exists on multiple providers; expected globally unique user IDs", userID)
		}
		c := a
		match = &c
	}
	if match == nil {
		return engine.Account{}, fmt.Errorf("user %q not found", userID)
	}
	return *match, nil
}

func backspaceString(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	return string(runes[:len(runes)-1])
}

func formatMinor(v int64, scale int) string {
	if scale <= 0 {
		return fmt.Sprintf("%d", v)
	}
	neg := v < 0
	if neg {
		v = -v
	}
	pow := int64(1)
	for i := 0; i < scale; i++ {
		pow *= 10
	}
	whole := v / pow
	frac := v % pow
	if neg {
		return fmt.Sprintf("-%d.%0*d", whole, scale, frac)
	}
	return fmt.Sprintf("%d.%0*d", whole, scale, frac)
}
