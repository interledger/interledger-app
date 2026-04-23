package gui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"ttt/engine"
)

// providerCurrencies declares which currencies each provider supports.
// Drives the drill-down menu (provider → currency → form).
var providerCurrencies = []struct {
	id         string
	label      string
	currencies []string
}{
	{id: "gatehub", label: "Gatehub", currencies: []string{"EUR"}},
	{id: "xago", label: "Xago", currencies: []string{"ZAR", "EUR"}},
}

type workflowParam struct {
	label        string
	placeholder  string
	options      []string // non-empty → renders as selectInput
	dynamicUsers bool     // options populated at menu-build time from engine (user IDs for the chosen provider+currency)
}

type workflowDef struct {
	name   string
	params []workflowParam
	flat   bool // true: top-level leaf (no provider/currency drill-down)
	// danger: require an explicit confirmation step before running.
	danger bool
	// skipProviderCurrencies lists provider/currency pairs to omit from the
	// drill-down menu for this workflow. Map key is provider ID, value is a
	// set of currency codes to suppress.
	skipProviderCurrencies map[string]map[string]bool
	run                    func(e *engine.Engine, values []string) error
	// postSubmit runs after a successful run. Used for side effects that
	// need access to Model state (re-seeding after Clear Everything).
	postSubmit func(m *Model)
}

// Param label constants — used by the menu layer to set presets.
const (
	paramProvider = "Provider ID"
	paramCurrency = "Currency"
)

func providerParam() workflowParam {
	ids := make([]string, 0, len(providerCurrencies))
	for _, p := range providerCurrencies {
		ids = append(ids, p.id)
	}
	return workflowParam{label: paramProvider, options: ids}
}

func currencyParam() workflowParam {
	// Union of all currencies — individual leaves lock this to one option.
	seen := map[string]bool{}
	var codes []string
	for _, p := range providerCurrencies {
		for _, c := range p.currencies {
			if !seen[c] {
				seen[c] = true
				codes = append(codes, c)
			}
		}
	}
	return workflowParam{label: paramCurrency, options: codes}
}

var workflowDefs = []workflowDef{
	{
		name: "Fund Provider Liquidity",
		params: []workflowParam{
			providerParam(),
			currencyParam(),
			{label: "Amount", placeholder: "1000.00"},
		},
		run: func(e *engine.Engine, v []string) error {
			cur, err := parseCurrency(v[1])
			if err != nil {
				return err
			}
			amount, err := parseAmount(v[2], cur.AssetScale)
			if err != nil {
				return err
			}
			_, err = e.FundProviderLiquidityLines(trim(v[0]), cur, amount)
			return err
		},
	},
	{
		name: "User Onboard",
		// Xago users are ZAR-only; EUR is a liquidity/position currency only.
		skipProviderCurrencies: map[string]map[string]bool{
			"xago": {"EUR": true},
		},
		params: []workflowParam{
			{label: "User ID", placeholder: "alice"},
			providerParam(),
			currencyParam(),
			{label: "Amount", placeholder: "300.00"},
		},
		run: func(e *engine.Engine, v []string) error {
			cur, err := parseCurrency(v[2])
			if err != nil {
				return err
			}
			amount, err := parseAmount(v[3], cur.AssetScale)
			if err != nil {
				return err
			}
			_, err = e.UserOnboardLines(trim(v[0]), trim(v[1]), cur, amount)
			return err
		},
	},
	{
		name: "P2P Transfer (Same Provider)",
		params: []workflowParam{
			providerParam(),
			currencyParam(),
			{label: "Sender User ID", dynamicUsers: true},
			{label: "Recipient User ID", dynamicUsers: true},
			{label: "Amount", placeholder: "100.00"},
		},
		run: func(e *engine.Engine, v []string) error {
			cur, err := parseCurrency(v[1])
			if err != nil {
				return err
			}
			amount, err := parseAmount(v[4], cur.AssetScale)
			if err != nil {
				return err
			}
			_, err = e.SameProviderP2PTransferLines(trim(v[2]), trim(v[3]), trim(v[0]), cur, amount)
			return err
		},
	},
	{
		name: "User Offboard",
		params: []workflowParam{
			providerParam(),
			currencyParam(),
			{label: "User ID", dynamicUsers: true},
			{label: "Amount", placeholder: "100.00"},
		},
		run: func(e *engine.Engine, v []string) error {
			cur, err := parseCurrency(v[1])
			if err != nil {
				return err
			}
			amount, err := parseAmount(v[3], cur.AssetScale)
			if err != nil {
				return err
			}
			_, err = e.UserOffboardLines(trim(v[2]), trim(v[0]), cur, amount)
			return err
		},
	},
	{
		name: "Cross-Provider Transfer",
		flat: true,
		params: []workflowParam{
			{label: "Sender Provider", options: providerIDs()},
			{label: "Sender Currency", options: allCurrencies()},
			{label: "Sender User ID", placeholder: "alice"},
			{label: "Recipient Provider", options: providerIDs()},
			{label: "Recipient Currency", options: allCurrencies()},
			{label: "Recipient User ID", placeholder: "userA"},
			{label: "Amount (sender currency)", placeholder: "100.00"},
		},
		run: func(e *engine.Engine, v []string) error {
			srcCur, err := parseCurrency(v[1])
			if err != nil {
				return err
			}
			dstCur, err := parseCurrency(v[4])
			if err != nil {
				return err
			}
			amount, err := parseAmount(v[6], srcCur.AssetScale)
			if err != nil {
				return err
			}
			_, _, err = e.CrossProviderTransferAutoLines(
				trim(v[2]), trim(v[0]), srcCur,
				trim(v[5]), trim(v[3]), dstCur,
				amount,
			)
			return err
		},
	},
	{
		name: "Bilateral Settlement",
		flat: true,
		params: []workflowParam{
			{label: "Provider A", options: providerIDs()},
			{label: "Provider B", options: providerIDs()},
			{label: "Currency", options: allCurrencies()},
			{label: "Cutoff (RFC3339 or 'now')", placeholder: "now"},
		},
		run: func(e *engine.Engine, v []string) error {
			cur, err := parseCurrency(v[2])
			if err != nil {
				return err
			}
			cutoff, err := parseCutoff(v[3])
			if err != nil {
				return err
			}
			_, err = e.SettleBilateralLines(trim(v[0]), trim(v[1]), cur, cutoff)
			return err
		},
	},
	{
		name:   "Clear Everything (DANGER)",
		flat:   true,
		danger: true,
		params: []workflowParam{
			{label: "Type 'CLEAR' to confirm", placeholder: "CLEAR"},
		},
		run: func(e *engine.Engine, v []string) error {
			if trim(v[0]) != "CLEAR" {
				return fmt.Errorf("confirmation text must be exactly 'CLEAR'")
			}
			return e.Reset()
		},
		postSubmit: func(m *Model) {
			if m.seed != nil {
				m.seed(m.eng)
			}
		},
	},
}

func providerIDs() []string {
	ids := make([]string, 0, len(providerCurrencies))
	for _, p := range providerCurrencies {
		ids = append(ids, p.id)
	}
	return ids
}

func allCurrencies() []string {
	seen := map[string]bool{}
	var codes []string
	for _, p := range providerCurrencies {
		for _, c := range p.currencies {
			if !seen[c] {
				seen[c] = true
				codes = append(codes, c)
			}
		}
	}
	return codes
}

func parseCutoff(s string) (time.Time, error) {
	s = trim(s)
	if s == "" || strings.EqualFold(s, "now") {
		return time.Now().UTC(), nil
	}
	return time.Parse(time.RFC3339, s)
}

// ── menu tree ─────────────────────────────────────────────────────────────────

// menuItem is one row in the menu. Groups have children; leaves have a
// wfIndex + optional presets that pre-fill/lock form fields.
type menuItem struct {
	label           string
	children        []menuItem          // non-empty ⇒ group
	wfIndex         int                 // leaf: workflowDefs index
	presets         map[string]string   // leaf: param label → preset value (locks field)
	optionOverrides map[string][]string // leaf: param label → dynamic option list
}

// buildMenuTree builds the top-level workflow menu with provider/currency
// drill-downs for every workflow that uses them.
func buildMenuTree(eng *engine.Engine) []menuItem {
	items := make([]menuItem, len(workflowDefs))
	for i, wf := range workflowDefs {
		if wf.flat {
			items[i] = menuItem{
				label:   wf.name,
				wfIndex: i,
			}
			continue
		}
		items[i] = menuItem{
			label:    wf.name,
			children: buildProviderMenu(eng, i),
		}
	}
	return items
}

func buildProviderMenu(eng *engine.Engine, wfIdx int) []menuItem {
	wf := workflowDefs[wfIdx]
	var provs []menuItem
	for _, p := range providerCurrencies {
		p := p
		var currencies []menuItem
		for _, c := range p.currencies {
			if wf.skipProviderCurrencies[p.id][c] {
				continue
			}
			cur, err := parseCurrency(c)
			var users []string
			if err == nil {
				users = listUsers(eng, p.id, cur)
			}
			overrides := map[string][]string{}
			for _, param := range wf.params {
				if param.dynamicUsers {
					overrides[param.label] = users
				}
			}
			currencies = append(currencies, menuItem{
				label:   c,
				wfIndex: wfIdx,
				presets: map[string]string{
					paramProvider: p.id,
					paramCurrency: c,
				},
				optionOverrides: overrides,
			})
		}
		if len(currencies) > 0 {
			provs = append(provs, menuItem{label: p.label, children: currencies})
		}
	}
	return provs
}

// listUsers returns sorted distinct user IDs having an account on the given
// provider+currency combination.
func listUsers(eng *engine.Engine, providerID string, cur engine.Currency) []string {
	if eng == nil {
		return nil
	}
	accounts, err := eng.ListAccounts()
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var users []string
	for _, a := range accounts {
		if a.Type != engine.AccountTypeUser {
			continue
		}
		if a.ProviderID != providerID || a.Currency.Code != cur.Code {
			continue
		}
		if seen[a.UserID] {
			continue
		}
		seen[a.UserID] = true
		users = append(users, a.UserID)
	}
	sort.Strings(users)
	return users
}

// ── form construction ─────────────────────────────────────────────────────────

func makeInputs(wf workflowDef, presets map[string]string, overrides map[string][]string) []formInput {
	inputs := make([]formInput, len(wf.params))
	for i, p := range wf.params {
		options := p.options
		if p.dynamicUsers {
			options = overrides[p.label]
		}

		if len(options) == 0 && p.dynamicUsers {
			sel := &selectInput{
				label:   p.label,
				options: []string{"(no users — onboard first)"},
				locked:  true,
			}
			inputs[i] = sel
			continue
		}

		if len(options) > 0 {
			sel := &selectInput{label: p.label, options: options}
			if preset, ok := presets[p.label]; ok {
				for j, opt := range options {
					if opt == preset {
						sel.idx = j
						break
					}
				}
				sel.locked = true
			}
			inputs[i] = sel
		} else {
			si := &simpleInput{label: p.label, placeholder: p.placeholder}
			if p.label == "Cutoff (RFC3339 or 'now')" {
				si.value = "now"
			}
			inputs[i] = si
		}
	}
	return inputs
}

// firstEditableInput returns the index of the first input that isn't locked,
// or len(inputs) if all are locked.
func firstEditableInput(inputs []formInput) int {
	for i, inp := range inputs {
		if !inp.isLocked() {
			return i
		}
	}
	return len(inputs)
}

// nextEditableInput returns the next non-locked index after start, wrapping.
// Returns start if nothing else is editable.
func nextEditableInput(inputs []formInput, start int) int {
	n := len(inputs)
	for i := 1; i <= n; i++ {
		idx := (start + i) % n
		if !inputs[idx].isLocked() {
			return idx
		}
	}
	return start
}

// lastEditableInput returns the index of the last non-locked input, or -1 if none.
func lastEditableInput(inputs []formInput) int {
	last := -1
	for i, inp := range inputs {
		if !inp.isLocked() {
			last = i
		}
	}
	return last
}

// ── parsers ───────────────────────────────────────────────────────────────────

func parseCurrency(s string) (engine.Currency, error) {
	switch strings.ToUpper(trim(s)) {
	case "EUR":
		return engine.EUR, nil
	case "ZAR":
		return engine.ZAR, nil
	default:
		return engine.Currency{}, fmt.Errorf("unknown currency %q — must be EUR or ZAR", trim(s))
	}
}

// parseAmount parses a decimal major-unit amount (e.g. "99", "99.50") and
// returns it in base units scaled by the currency's AssetScale.
// Example: scale=2, "99" → 9900; "99.5" → 9950; "0.01" → 1.
func parseAmount(s string, scale int) (int64, error) {
	s = trim(s)
	if s == "" {
		return 0, fmt.Errorf("amount required")
	}
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}

	whole, frac, hasDot := strings.Cut(s, ".")
	if whole == "" && frac == "" {
		return 0, fmt.Errorf("amount must be a number")
	}
	if whole == "" {
		whole = "0"
	}

	var wholeN int64
	if whole != "" {
		n, err := strconv.ParseInt(whole, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("amount must be a number")
		}
		wholeN = n
	}

	if !hasDot {
		frac = ""
	}
	if len(frac) > scale {
		return 0, fmt.Errorf("amount has more than %d decimal places", scale)
	}
	// Right-pad fractional part to `scale` digits.
	padded := frac + strings.Repeat("0", scale-len(frac))
	var fracN int64
	if padded != "" {
		n, err := strconv.ParseInt(padded, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("amount must be a number")
		}
		fracN = n
	}

	pow := int64(1)
	for i := 0; i < scale; i++ {
		pow *= 10
	}
	base := wholeN*pow + fracN
	if neg {
		base = -base
	}
	return base, nil
}

func trim(s string) string {
	return strings.TrimSpace(s)
}
