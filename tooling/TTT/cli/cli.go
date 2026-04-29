// Package cli provides subcommand-based access to all TTT engine workflows.
// When os.Args contains a subcommand the binary runs in headless CLI mode;
// with no arguments it falls through to the interactive TUI.
package cli

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"ttt/engine"
	"ttt/engine/sqlite"
	"ttt/ods"
)

// Run dispatches args (os.Args[1:]) to the matching subcommand.
// Returns an exit code suitable for os.Exit.
func Run(store *sqlite.Store, eng *engine.Engine, args []string) int {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return 0
	}
	switch args[0] {
	case "init":
		return cmdInit(store, eng, args[1:])
	case "reset":
		return cmdReset(eng, args[1:])
	case "fund-liquidity":
		return cmdFundLiquidity(eng, args[1:])
	case "onboard":
		return cmdOnboard(eng, args[1:])
	case "offboard":
		return cmdOffboard(eng, args[1:])
	case "p2p":
		return cmdP2P(eng, args[1:])
	case "transfer":
		return cmdTransfer(eng, args[1:])
	case "settle":
		return cmdSettle(eng, args[1:])
	case "settlement-preview":
		return cmdSettlementPreview(eng, args[1:])
	case "set-charge":
		return cmdSetCharge(eng, args[1:])
	case "status":
		return cmdStatus(store, eng, args[1:])
	case "ledger":
		return cmdLedger(eng, args[1:])
	case "export-ods":
		return cmdExportODS(store, args[1:])
	case "help", "--help", "-h":
		printUsage(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command %q\n\n", args[0])
		printUsage(os.Stderr)
		return 1
	}
}

// ── commands ──────────────────────────────────────────────────────────────────

// ttt init --mode <mode>
// Modes: standard | legacy | self-exchange
func cmdInit(store *sqlite.Store, eng *engine.Engine, args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	mode := fs.String("mode", "", "paradigm: standard | legacy | self-exchange")
	if err := fs.Parse(args); err != nil {
		return cliErr(err)
	}
	if *mode == "" {
		return fail("init requires --mode <standard|legacy|self-exchange>")
	}
	p, err := parseMode(*mode)
	if err != nil {
		return cliErr(err)
	}
	if err := store.SetParadigm(p); err != nil {
		return cliErr(fmt.Errorf("setting paradigm: %w", err))
	}
	if err := engine.SeedParadigm(p, eng); err != nil {
		return cliErr(fmt.Errorf("seeding: %w", err))
	}
	fmt.Printf("Initialized: %s\n", p.Name())
	return 0
}

// ttt reset
// Wipes all ledger data (accounts, journal lines). Paradigm config is preserved.
func cmdReset(eng *engine.Engine, args []string) int {
	if err := eng.Reset(); err != nil {
		return cliErr(err)
	}
	fmt.Println("Ledger cleared.")
	return 0
}

// ttt fund-liquidity <provider> <currency> <amount>
func cmdFundLiquidity(eng *engine.Engine, args []string) int {
	if len(args) != 3 {
		return fail("usage: fund-liquidity <provider> <currency> <amount>")
	}
	cur, err := parseCurrency(args[1])
	if err != nil {
		return cliErr(err)
	}
	amount, err := parseAmount(args[2], cur.AssetScale)
	if err != nil {
		return cliErr(err)
	}
	if _, err := eng.FundProviderLiquidityLines(args[0], cur, amount); err != nil {
		return cliErr(err)
	}
	fmt.Printf("Funded %s %s liquidity +%s\n", args[0], cur.Code, fmtAmount(amount, cur.AssetScale))
	return 0
}

// ttt onboard <user> <provider> <currency> <amount>
func cmdOnboard(eng *engine.Engine, args []string) int {
	if len(args) != 4 {
		return fail("usage: onboard <user> <provider> <currency> <amount>")
	}
	cur, err := parseCurrency(args[2])
	if err != nil {
		return cliErr(err)
	}
	amount, err := parseAmount(args[3], cur.AssetScale)
	if err != nil {
		return cliErr(err)
	}
	if _, err := eng.UserOnboardLines(args[0], args[1], cur, amount); err != nil {
		return cliErr(err)
	}
	fmt.Printf("Onboarded %s @ %s/%s +%s\n", args[0], args[1], cur.Code, fmtAmount(amount, cur.AssetScale))
	return 0
}

// ttt offboard <provider> <currency> <user> <amount>
func cmdOffboard(eng *engine.Engine, args []string) int {
	if len(args) != 4 {
		return fail("usage: offboard <provider> <currency> <user> <amount>")
	}
	cur, err := parseCurrency(args[1])
	if err != nil {
		return cliErr(err)
	}
	amount, err := parseAmount(args[3], cur.AssetScale)
	if err != nil {
		return cliErr(err)
	}
	if _, err := eng.UserOffboardLines(args[2], args[0], cur, amount); err != nil {
		return cliErr(err)
	}
	fmt.Printf("Offboarded %s @ %s/%s -%s\n", args[2], args[0], cur.Code, fmtAmount(amount, cur.AssetScale))
	return 0
}

// ttt p2p <provider> <currency> <sender> <recipient> <amount>
func cmdP2P(eng *engine.Engine, args []string) int {
	if len(args) != 5 {
		return fail("usage: p2p <provider> <currency> <sender> <recipient> <amount>")
	}
	cur, err := parseCurrency(args[1])
	if err != nil {
		return cliErr(err)
	}
	amount, err := parseAmount(args[4], cur.AssetScale)
	if err != nil {
		return cliErr(err)
	}
	if _, err := eng.SameProviderP2PTransferLines(args[2], args[3], args[0], cur, amount); err != nil {
		return cliErr(err)
	}
	fmt.Printf("P2P %s → %s @ %s/%s %s\n", args[2], args[3], args[0], cur.Code, fmtAmount(amount, cur.AssetScale))
	return 0
}

// ttt transfer <sender-user> <sender-provider> <sender-currency>
//
//	<recipient-user> <recipient-provider> <recipient-currency>
//	<amount>
func cmdTransfer(eng *engine.Engine, args []string) int {
	if len(args) != 7 {
		return fail("usage: transfer <sender-user> <sender-provider> <sender-currency>" +
			" <recipient-user> <recipient-provider> <recipient-currency> <amount>")
	}
	srcCur, err := parseCurrency(args[2])
	if err != nil {
		return cliErr(err)
	}
	dstCur, err := parseCurrency(args[5])
	if err != nil {
		return cliErr(err)
	}
	amount, err := parseAmount(args[6], srcCur.AssetScale)
	if err != nil {
		return cliErr(err)
	}
	lines, rate, err := eng.CrossProviderTransferAutoLines(
		args[0], args[1], srcCur,
		args[3], args[4], dstCur,
		amount,
	)
	if err != nil {
		return cliErr(err)
	}
	destAmount := int64(0)
	for _, l := range lines {
		if l.CreditMetadata["step"] == "credit recipient user" {
			destAmount = l.Amount
			break
		}
	}
	fmt.Printf("Transfer %s/%s → %s/%s  %s %s @ %d/%d → %s %s\n",
		args[0], srcCur.Code,
		args[3], dstCur.Code,
		fmtAmount(amount, srcCur.AssetScale), srcCur.Code,
		rate.Num, rate.Den,
		fmtAmount(destAmount, dstCur.AssetScale), dstCur.Code,
	)
	return 0
}

// ttt settle <provider-a> <provider-b> <currency> [--cutoff <rfc3339|now>]
func cmdSettle(eng *engine.Engine, args []string) int {
	fs := flag.NewFlagSet("settle", flag.ContinueOnError)
	cutoffStr := fs.String("cutoff", "now", "settlement cutoff: RFC3339 timestamp or 'now'")
	if err := fs.Parse(args); err != nil {
		return cliErr(err)
	}
	pos := fs.Args()
	if len(pos) != 3 {
		return fail("usage: settle <provider-a> <provider-b> <currency> [--cutoff <time>]")
	}
	cur, err := parseCurrency(pos[2])
	if err != nil {
		return cliErr(err)
	}
	cutoff, err := parseCutoff(*cutoffStr)
	if err != nil {
		return cliErr(err)
	}
	if _, err := eng.SettleBilateralLines(pos[0], pos[1], cur, cutoff); err != nil {
		return cliErr(err)
	}
	fmt.Printf("Settled %s ↔ %s (%s) up to %s\n", pos[0], pos[1], cur.Code, cutoff.Format(time.RFC3339))
	return 0
}

// ttt settlement-preview <provider-a> <provider-b> <currency>
// Shows the net bilateral position without posting any entries.
func cmdSettlementPreview(eng *engine.Engine, args []string) int {
	if len(args) != 3 {
		return fail("usage: settlement-preview <provider-a> <provider-b> <currency>")
	}
	cur, err := parseCurrency(args[2])
	if err != nil {
		return cliErr(err)
	}
	pairs, err := eng.CheckBilateralPositions()
	if err != nil && len(pairs) == 0 {
		return cliErr(fmt.Errorf("reading positions: %w", err))
	}
	pA, pB := args[0], args[1]
	for _, p := range pairs {
		if p.Currency != cur.Code {
			continue
		}
		if !((p.ProviderA == pA && p.ProviderB == pB) || (p.ProviderA == pB && p.ProviderB == pA)) {
			continue
		}
		if p.BalanceA == 0 && p.BalanceB == 0 {
			fmt.Printf("Settlement preview: %s ↔ %s (%s)\nNothing to settle.\n", pA, pB, cur.Code)
			return 0
		}
		// Positive balance means creditor; the provider with positive balance is owed money.
		creditor, debtor := p.ProviderA, p.ProviderB
		amount := p.BalanceA
		if amount < 0 {
			creditor, debtor = p.ProviderB, p.ProviderA
			amount = p.BalanceB
		}
		fmt.Printf("Settlement preview: %s ↔ %s (%s)\n%s owes %s  %s %s\n",
			pA, pB, cur.Code,
			debtor, creditor,
			fmtAmount(amount, cur.AssetScale), cur.Code,
		)
		return 0
	}
	fmt.Printf("Settlement preview: %s ↔ %s (%s)\nNo position accounts found — run transfers first.\n", pA, pB, cur.Code)
	return 0
}

// ttt set-charge <from-provider> <to-provider> [<percent>]
// Omitting percent (or passing "") clears the charge.
func cmdSetCharge(eng *engine.Engine, args []string) int {
	if len(args) < 2 || len(args) > 3 {
		return fail("usage: set-charge <from-provider> <to-provider> [<percent>]")
	}
	pct := ""
	if len(args) == 3 {
		pct = args[2]
	}
	charge, err := parseChargePercent(pct)
	if err != nil {
		return cliErr(err)
	}
	if err := eng.SetCharge(args[0], args[1], charge); err != nil {
		return cliErr(err)
	}
	if charge == nil {
		fmt.Printf("Charge %s → %s cleared\n", args[0], args[1])
	} else {
		fmt.Printf("Charge %s → %s set to %s%%\n", args[0], args[1], fmtBasisPoints(charge.Num, charge.Den))
	}
	return 0
}

// ttt status
// Prints providers, account balances, FX rates, and integrity check results.
func cmdStatus(store *sqlite.Store, eng *engine.Engine, args []string) int {
	// Paradigm
	if ok, _ := store.IsParadigmSet(); ok {
		if p, err := store.GetParadigm(); err == nil {
			fmt.Printf("Paradigm: %s\n", p.Name())
		}
	} else {
		fmt.Println("Paradigm: (not set — run: ttt init --mode <mode>)")
	}

	// Accounts grouped by provider
	accounts, err := eng.ListAccounts()
	if err != nil {
		return cliErr(fmt.Errorf("listing accounts: %w", err))
	}
	providers, err := eng.ListProviders()
	if err != nil {
		return cliErr(fmt.Errorf("listing providers: %w", err))
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i].ID < providers[j].ID })

	if len(providers) > 0 {
		fmt.Println("\nAccounts:")
		byProvider := make(map[string][]engine.Account)
		for _, a := range accounts {
			byProvider[a.ProviderID] = append(byProvider[a.ProviderID], a)
		}
		for _, prov := range providers {
			fmt.Printf("  %s\n", prov.Name)
			accts := byProvider[prov.ID]
			sort.Slice(accts, func(i, j int) bool {
				if accts[i].Type != accts[j].Type {
					return accts[i].Type < accts[j].Type
				}
				return accts[i].Currency.Code < accts[j].Currency.Code
			})
			for _, a := range accts {
				bal, _ := eng.Balance(a.ID)
				fmt.Printf("    %-38s %12s %s\n",
					accountLabel(a),
					fmtAmount(bal, a.Currency.AssetScale),
					a.Currency.Code,
				)
			}
		}
	}

	// FX rates
	if fx := eng.FX(); fx != nil {
		if rate, err := fx.Rate("EUR", "ZAR"); err == nil {
			fmt.Printf("\nFX rates:\n  EUR/ZAR  %d/%d\n", rate.Num, rate.Den)
		}
	}

	// Charges
	providers2, _ := eng.ListProviders()
	var chargeLines []string
	for _, a := range providers2 {
		for _, b := range providers2 {
			if a.ID == b.ID {
				continue
			}
			if c, err := eng.GetCharge(a.ID, b.ID); err == nil && c != nil {
				chargeLines = append(chargeLines, fmt.Sprintf("  %s → %s: %s%%", a.ID, b.ID, fmtBasisPoints(c.Num, c.Den)))
			}
		}
	}
	if len(chargeLines) > 0 {
		fmt.Println("\nCharges:")
		for _, l := range chargeLines {
			fmt.Println(l)
		}
	}

	// Integrity checks
	fmt.Println("\nChecks:")
	if _, err := eng.CheckGlobalBalance(); err != nil {
		fmt.Printf("  Global balance:           FAIL — %v\n", err)
	} else {
		fmt.Println("  Global balance:           OK")
	}
	if _, err := eng.CheckBilateralPositions(); err != nil {
		fmt.Printf("  Bilateral positions:      FAIL — %v\n", err)
	} else {
		fmt.Println("  Bilateral positions:      OK")
	}
	if _, err := eng.CheckLiquidityDecomposition(); err != nil {
		fmt.Printf("  Liquidity decomposition:  FAIL — %v\n", err)
	} else {
		fmt.Println("  Liquidity decomposition:  OK")
	}

	return 0
}

// ttt ledger
// Dumps all journal lines in insertion order.
func cmdLedger(eng *engine.Engine, args []string) int {
	lines, err := eng.GetAllLines()
	if err != nil {
		return cliErr(err)
	}
	accounts, err := eng.ListAccounts()
	if err != nil {
		return cliErr(err)
	}
	byID := make(map[string]engine.Account, len(accounts))
	for _, a := range accounts {
		byID[a.ID] = a
	}

	if len(lines) == 0 {
		fmt.Println("(no journal entries)")
		return 0
	}

	fmt.Printf("%-20s %-10s %-28s %-38s %-38s %s\n",
		"Timestamp", "Event", "Workflow", "Debit", "Credit", "Amount")
	fmt.Println(strings.Repeat("-", 145))

	for _, l := range lines {
		ts := l.Timestamp.Format("2006-01-02 15:04:05")
		eventShort := l.EventID
		if len(eventShort) > 8 {
			eventShort = eventShort[:8]
		}
		workflow := l.Metadata[engine.MetaWorkflow]
		if len(workflow) > 26 {
			workflow = workflow[:26]
		}
		debit := accountLabel(byID[l.DebitAccountID])
		credit := accountLabel(byID[l.CreditAccountID])

		cur := byID[l.DebitAccountID].Currency
		amtStr := fmtAmount(l.Amount, cur.AssetScale) + " " + cur.Code

		fmt.Printf("%-20s %-10s %-28s %-38s %-38s %s\n",
			ts, eventShort, workflow, debit, credit, amtStr)
	}
	return 0
}

// ttt export-ods [path]
// Exports the current database as an OpenDocument spreadsheet.
func cmdExportODS(store *sqlite.Store, args []string) int {
	if len(args) > 1 {
		return fail("usage: export-ods [path]")
	}
	path := "output/ttt-export.ods"
	if len(args) == 1 && strings.TrimSpace(args[0]) != "" {
		path = args[0]
	}
	if err := ods.ExportStore(store, path); err != nil {
		return cliErr(err)
	}
	fmt.Printf("Exported ODS: %s\n", path)
	return 0
}

// ── helpers ───────────────────────────────────────────────────────────────────

func cliErr(err error) int {
	fmt.Fprintln(os.Stderr, "error:", err)
	return 1
}

func fail(msg string) int {
	fmt.Fprintln(os.Stderr, "error:", msg)
	return 1
}

func accountLabel(a engine.Account) string {
	switch a.Type {
	case engine.AccountTypeSystem:
		return fmt.Sprintf("%s/system(%s)", a.ProviderID, a.Currency.Code)
	case engine.AccountTypeLiquidity:
		return fmt.Sprintf("%s/liquidity(%s)", a.ProviderID, a.Currency.Code)
	case engine.AccountTypePosition:
		return fmt.Sprintf("%s/position(%s)@%s", a.ProviderID, a.Currency.Code, a.CounterpartyID)
	case engine.AccountTypeUser:
		return fmt.Sprintf("%s/%s(%s)", a.ProviderID, a.UserID, a.Currency.Code)
	case engine.AccountTypeFX:
		return fmt.Sprintf("%s/fx(%s)", a.ProviderID, a.Currency.Code)
	}
	return a.ID
}

func fmtAmount(amount int64, scale int) string {
	if scale <= 0 {
		return strconv.FormatInt(amount, 10)
	}
	neg := amount < 0
	if neg {
		amount = -amount
	}
	pow := int64(1)
	for i := 0; i < scale; i++ {
		pow *= 10
	}
	whole := amount / pow
	frac := amount % pow
	s := fmt.Sprintf("%d.%0*d", whole, scale, frac)
	if neg {
		return "-" + s
	}
	return s
}

func fmtBasisPoints(num, den int64) string {
	if den == 0 {
		return "0.00"
	}
	bp := num * 10000 / den
	return fmt.Sprintf("%d.%02d", bp/100, bp%100)
}

// ── parsers ───────────────────────────────────────────────────────────────────

func parseCurrency(s string) (engine.Currency, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "EUR":
		return engine.EUR, nil
	case "ZAR":
		return engine.ZAR, nil
	default:
		return engine.Currency{}, fmt.Errorf("unknown currency %q — supported: EUR, ZAR", s)
	}
}

func parseAmount(s string, scale int) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("amount required")
	}
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	whole, frac, hasDot := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	wholeN, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q", s)
	}
	if hasDot && len(frac) > scale {
		return 0, fmt.Errorf("amount %q has more than %d decimal places", s, scale)
	}
	padded := frac + strings.Repeat("0", scale-len(frac))
	fracN := int64(0)
	if padded != "" {
		fracN, err = strconv.ParseInt(padded, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid amount %q", s)
		}
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

func parseCutoff(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "now") {
		return time.Now().UTC(), nil
	}
	return time.Parse(time.RFC3339, s)
}

func parseChargePercent(s string) (*engine.ChargeRate, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	bp, err := parseAmount(s, 2)
	if err != nil {
		return nil, fmt.Errorf("invalid charge percentage %q: %w", s, err)
	}
	if bp < 0 {
		return nil, fmt.Errorf("charge percentage must be non-negative")
	}
	return &engine.ChargeRate{Num: bp, Den: 10000}, nil
}

func parseMode(s string) (engine.Paradigm, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "standard", "pos-two", "pos2":
		return engine.ParadigmPOSTwo, nil
	case "legacy", "single-gh-eur", "single":
		return engine.ParadigmSingleGHEUR, nil
	case "self-exchange", "selfexchange", "self_exchange":
		return engine.ParadigmSelfExchange, nil
	default:
		return 0, fmt.Errorf("unknown mode %q — supported: standard, legacy, self-exchange", s)
	}
}

// ── usage ─────────────────────────────────────────────────────────────────────

func printUsage(w *os.File) {
	fmt.Fprintln(w, `Toy Treasury Time (TTT)

Usage: ttt [command] [args...]

  No command         Launch interactive TUI

Commands:
  init --mode <mode>
    Initialize paradigm and seed default accounts.
    Modes: standard | legacy | self-exchange

  reset
    Wipe all ledger data (providers, accounts, journal lines).
    Paradigm config and charges are preserved.

  fund-liquidity <provider> <currency> <amount>
    Credit a provider's liquidity from its system account.
    Example: ttt fund-liquidity xago zar 15000

  onboard <user> <provider> <currency> <amount>
    Onboard a user with funds from the provider's system account.
    Example: ttt onboard alice gatehub eur 500

  offboard <provider> <currency> <user> <amount>
    Offboard a user (withdraw funds back to system account).
    Example: ttt offboard gatehub eur alice 100

  p2p <provider> <currency> <sender> <recipient> <amount>
    Transfer between two users on the same provider.
    Example: ttt p2p gatehub eur alice bob 50

  transfer <sender-user> <sender-provider> <sender-currency>
           <recipient-user> <recipient-provider> <recipient-currency>
           <amount>
    Cross-provider transfer using the live FX rate.
    Example: ttt transfer alice gatehub eur carlos xago zar 100

  settle <provider-a> <provider-b> <currency> [--cutoff <RFC3339|now>]
    Run bilateral settlement between two providers.
    Example: ttt settle gatehub xago eur
    Example: ttt settle gatehub xago eur --cutoff 2025-01-01T00:00:00Z

  settlement-preview <provider-a> <provider-b> <currency>
    Show the net bilateral position without posting any entries.
    Example: ttt settlement-preview gatehub xago eur

  set-charge <from-provider> <to-provider> [<percent>]
    Set a transfer charge percentage (omit percent to clear).
    Example: ttt set-charge gatehub xago 2.5
    Example: ttt set-charge gatehub xago        (clears charge)

  status
    Print account balances, FX rates, charges, and integrity checks.

  ledger
    Dump all journal lines in insertion order.

  export-ods [path]
    Export the current database as an OpenDocument spreadsheet.
    Default path: output/ttt-export.ods
    Example: ttt export-ods output/demo.ods`)
}
