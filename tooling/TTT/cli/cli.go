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
	case "create-account":
		return cmdCreateAccount(eng, args[1:])
	case "fund-liquidity":
		return cmdFundLiquidity(eng, args[1:])
	case "onboard":
		return cmdOnboard(eng, args[1:])
	case "offboard":
		return cmdOffboard(eng, args[1:])
	case "p2p":
		return cmdP2P(eng, args[1:])
	case "move":
		return cmdMove(eng, args[1:])
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

// ttt create-account --provider <provider> --user <user> --currency <currency> [--provider-name <name>]
func cmdCreateAccount(eng *engine.Engine, args []string) int {
	fs := flag.NewFlagSet("create-account", flag.ContinueOnError)
	providerID := fs.String("provider", "", "provider ID")
	providerName := fs.String("provider-name", "", "display name used when creating a new provider")
	userID := fs.String("user", "", "user/account ID")
	currency := fs.String("currency", "", "account currency: EUR | ZAR")
	accountType := fs.String("type", "user", "account type: user | system | liquidity | fx | position")
	counterpartyID := fs.String("counterparty", "", "counterparty provider ID for position accounts")
	if err := fs.Parse(args); err != nil {
		return cliErr(err)
	}
	if fs.NArg() != 0 {
		return fail("usage: create-account --provider <provider> --user <user> --currency <currency> [--provider-name <name>]")
	}
	*providerID = strings.TrimSpace(*providerID)
	*providerName = strings.TrimSpace(*providerName)
	*userID = strings.TrimSpace(*userID)
	if *providerID == "" {
		return fail("create-account requires --provider <provider>")
	}
	*accountType = strings.ToLower(strings.TrimSpace(*accountType))
	*counterpartyID = strings.TrimSpace(*counterpartyID)
	if *accountType == "" {
		*accountType = "user"
	}
	if *accountType == "user" && *userID == "" {
		return fail("create-account requires --user <user>")
	}
	if *currency == "" {
		return fail("create-account requires --currency <EUR|ZAR>")
	}
	cur, err := parseCurrency(*currency)
	if err != nil {
		return cliErr(err)
	}

	providerCreated, err := ensureProvider(eng, *providerID, *providerName)
	if err != nil {
		return cliErr(err)
	}
	account, systemCreated, liquidityCreated, err := createAccount(eng, *accountType, *providerID, *providerName, *userID, *counterpartyID, cur)
	if err != nil {
		return cliErr(err)
	}

	fmt.Printf("Created account %s\n", accountLabel(account))
	if providerCreated {
		fmt.Printf("Created provider %s\n", *providerID)
	}
	if systemCreated {
		fmt.Printf("Created system account %s/system(%s)\n", *providerID, cur.Code)
	}
	if liquidityCreated {
		fmt.Printf("Created liquidity account %s/liquidity(%s)\n", *providerID, cur.Code)
	}
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

// ttt move --from <provider/user> --to <provider/user> --currency <currency> --amount <amount> [--workflow <name>] [--step <name>]
func cmdMove(eng *engine.Engine, args []string) int {
	fs := flag.NewFlagSet("move", flag.ContinueOnError)
	from := fs.String("from", "", "sender account reference: provider/user")
	to := fs.String("to", "", "recipient account reference: provider/user")
	currency := fs.String("currency", "", "currency: EUR | ZAR")
	amountText := fs.String("amount", "", "amount to move")
	workflow := fs.String("workflow", engine.WorkflowDirectMove, "journal workflow metadata")
	step := fs.String("step", "direct move", "journal step metadata")
	if err := fs.Parse(args); err != nil {
		return cliErr(err)
	}
	if fs.NArg() != 0 {
		return fail("usage: move --from <provider/user> --to <provider/user> --currency <currency> --amount <amount> [--workflow <name>] [--step <name>]")
	}
	srcProvider, srcUser, err := parseAccountRef(*from)
	if err != nil {
		return cliErr(fmt.Errorf("--from: %w", err))
	}
	dstProvider, dstUser, err := parseAccountRef(*to)
	if err != nil {
		return cliErr(fmt.Errorf("--to: %w", err))
	}
	if strings.TrimSpace(*currency) == "" {
		return fail("move requires --currency <EUR|ZAR>")
	}
	cur, err := parseCurrency(*currency)
	if err != nil {
		return cliErr(err)
	}
	if strings.TrimSpace(*amountText) == "" {
		return fail("move requires --amount <amount>")
	}
	amount, err := parseAmount(*amountText, cur.AssetScale)
	if err != nil {
		return cliErr(err)
	}
	accounts, err := eng.ListAccounts()
	if err != nil {
		return cliErr(err)
	}
	srcAccount, err := resolveAccountRef(accounts, srcProvider, srcUser, cur)
	if err != nil {
		return cliErr(fmt.Errorf("--from: %w", err))
	}
	dstAccount, err := resolveAccountRef(accounts, dstProvider, dstUser, cur)
	if err != nil {
		return cliErr(fmt.Errorf("--to: %w", err))
	}
	if _, err := eng.MoveAccountsLinesWithMetadata(srcAccount, dstAccount, cur, amount, *workflow, *step); err != nil {
		return cliErr(err)
	}
	fmt.Printf("Moved %s %s %s/%s → %s/%s\n",
		fmtAmount(amount, cur.AssetScale),
		cur.Code,
		srcProvider, srcUser,
		dstProvider, dstUser,
	)
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
		// Fall back to single-POS settlement (legacy mode: only one provider has
		// the settlement currency). Try each provider as the hub in turn.
		err2 := trySettleSinglePOS(eng, pos[0], pos[1], cur, cutoff)
		if err2 != nil {
			err3 := trySettleSinglePOS(eng, pos[1], pos[0], cur, cutoff)
			if err3 != nil {
				return cliErr(fmt.Errorf("bilateral: %v", err))
			}
		}
	}
	fmt.Printf("Settled %s ↔ %s (%s) up to %s\n", pos[0], pos[1], cur.Code, cutoff.Format(time.RFC3339))
	return 0
}

func trySettleSinglePOS(eng *engine.Engine, hub, counterparty string, cur engine.Currency, cutoff time.Time) error {
	_, err := eng.SettleSinglePOSLines(hub, counterparty, cur, cutoff)
	return err
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
	pA, pB := args[0], args[1]

	// Try bilateral first (standard / self-exchange modes).
	pairs, _ := eng.CheckBilateralPositions()
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

	// Fall back to single-POS (legacy mode): look for a one-sided position under
	// whichever provider holds hub currency liquidity.
	for _, hub := range []string{pA, pB} {
		counterparty := pB
		if hub == pB {
			counterparty = pA
		}
		bal, err := eng.SinglePOSBalance(hub, counterparty, cur)
		if err != nil {
			continue
		}
		if bal == 0 {
			fmt.Printf("Settlement preview: %s ↔ %s (%s)\nNothing to settle.\n", pA, pB, cur.Code)
			return 0
		}
		amount := bal
		if amount < 0 {
			amount = -amount
		}
		// Positive position balance: hub is the creditor (counterparty owes hub).
		creditor, debtor := hub, counterparty
		if bal < 0 {
			creditor, debtor = counterparty, hub
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

	fmt.Printf("%-20s %-10s %-28s %-24s %-38s %-38s %s\n",
		"Timestamp", "Event", "Workflow", "Step", "Debit", "Credit", "Amount")
	fmt.Println(strings.Repeat("-", 170))

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
		step := l.Metadata[engine.MetaStep]
		if len(step) > 22 {
			step = step[:22]
		}
		debit := accountLabel(byID[l.DebitAccountID])
		credit := accountLabel(byID[l.CreditAccountID])

		cur := byID[l.DebitAccountID].Currency
		amtStr := fmtAmount(l.Amount, cur.AssetScale) + " " + cur.Code

		fmt.Printf("%-20s %-10s %-28s %-24s %-38s %-38s %s\n",
			ts, eventShort, workflow, step, debit, credit, amtStr)
	}
	return 0
}

// ttt export-ods [--sheet=name] [path]
// Exports the current database as an OpenDocument spreadsheet.
// With --sheet=name, adds or replaces only that sheet in an existing file.
func cmdExportODS(store *sqlite.Store, args []string) int {
	var sheetName string
	var pathArgs []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--sheet=") {
			sheetName = strings.TrimPrefix(arg, "--sheet=")
		} else if arg == "--sheet" {
			if i+1 >= len(args) {
				return fail("--sheet requires a value")
			}
			i++
			sheetName = args[i]
		} else {
			pathArgs = append(pathArgs, arg)
		}
	}
	if len(pathArgs) > 1 {
		return fail("usage: export-ods [--sheet=name] [path]")
	}
	path := "output/ttt-export.ods"
	if len(pathArgs) == 1 && strings.TrimSpace(pathArgs[0]) != "" {
		path = pathArgs[0]
	}
	if sheetName != "" {
		if err := ods.ExportSheetStore(store, path, sheetName); err != nil {
			return cliErr(err)
		}
	} else {
		if err := ods.ExportStore(store, path); err != nil {
			return cliErr(err)
		}
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

func ensureProvider(eng *engine.Engine, id, name string) (bool, error) {
	providers, err := eng.ListProviders()
	if err != nil {
		return false, err
	}
	for _, provider := range providers {
		if provider.ID == id {
			return false, nil
		}
	}
	if name == "" {
		name = id
	}
	if _, err := eng.CreateProvider(id, name); err != nil {
		return false, err
	}
	return true, nil
}

func createAccount(eng *engine.Engine, accountType, providerID, providerName, userID, counterpartyID string, currency engine.Currency) (engine.Account, bool, bool, error) {
	systemCreated := false
	liquidityCreated := false
	switch accountType {
	case "user":
		var err error
		systemCreated, liquidityCreated, err = ensureProviderCurrencyAccounts(eng, providerID, currency)
		if err != nil {
			return engine.Account{}, false, false, err
		}
		account, err := eng.CreateUserAccount(userID, providerID, currency)
		return account, systemCreated, liquidityCreated, err
	case "system":
		account, err := eng.CreateSystemAccount(providerID, currency)
		return account, true, false, err
	case "liquidity":
		account, err := eng.CreateLiquidityAccount(providerID, currency)
		return account, false, true, err
	case "fx":
		account, err := eng.CreateFXAccount(providerID, currency)
		return account, false, false, err
	case "position":
		if counterpartyID == "" {
			return engine.Account{}, false, false, fmt.Errorf("create-account --type position requires --counterparty <provider>")
		}
		if _, err := ensureProvider(eng, counterpartyID, counterpartyID); err != nil {
			return engine.Account{}, false, false, err
		}
		systemCreated, liquidityCreated, err := ensureProviderCurrencyAccounts(eng, providerID, currency)
		if err != nil {
			return engine.Account{}, false, false, err
		}
		accounts, err := eng.ListAccounts()
		if err != nil {
			return engine.Account{}, false, false, err
		}
		liquidity, err := resolveAccountRef(accounts, providerID, "liquidity", currency)
		if err != nil {
			return engine.Account{}, false, false, err
		}
		account, err := eng.CreatePositionAccount(liquidity.ID, counterpartyID)
		return account, systemCreated, liquidityCreated, err
	default:
		return engine.Account{}, false, false, fmt.Errorf("unknown account type %q — supported: user, system, liquidity, fx, position", accountType)
	}
}

func ensureProviderCurrencyAccounts(eng *engine.Engine, providerID string, currency engine.Currency) (bool, bool, error) {
	accounts, err := eng.ListAccounts()
	if err != nil {
		return false, false, err
	}
	systemCreated := false
	liquidityCreated := false

	if !hasProviderCurrencyAccount(accounts, engine.AccountTypeSystem, providerID, currency) {
		if _, err := eng.CreateSystemAccount(providerID, currency); err != nil {
			return false, false, err
		}
		systemCreated = true
	}
	if !hasProviderCurrencyAccount(accounts, engine.AccountTypeLiquidity, providerID, currency) {
		if _, err := eng.CreateLiquidityAccount(providerID, currency); err != nil {
			return false, false, err
		}
		liquidityCreated = true
	}
	return systemCreated, liquidityCreated, nil
}

func hasProviderCurrencyAccount(accounts []engine.Account, typ engine.AccountType, providerID string, currency engine.Currency) bool {
	for _, account := range accounts {
		if account.Type == typ && account.ProviderID == providerID && account.Currency.Code == currency.Code {
			return true
		}
	}
	return false
}

func parseAccountRef(ref string) (string, string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", "", fmt.Errorf("account reference is required")
	}
	provider, user, ok := strings.Cut(ref, "/")
	provider = strings.TrimSpace(provider)
	user = strings.TrimSpace(user)
	if !ok || provider == "" || user == "" || strings.Contains(user, "/") {
		return "", "", fmt.Errorf("expected <provider>/<user>, got %q", ref)
	}
	return provider, user, nil
}

func resolveAccountRef(accounts []engine.Account, providerID, name string, currency engine.Currency) (engine.Account, error) {
	for _, account := range accounts {
		if account.ProviderID != providerID || account.Currency.Code != currency.Code {
			continue
		}
		switch {
		case name == "system" && account.Type == engine.AccountTypeSystem:
			return account, nil
		case name == "liquidity" && account.Type == engine.AccountTypeLiquidity:
			return account, nil
		case name == "fx" && account.Type == engine.AccountTypeFX:
			return account, nil
		case strings.HasPrefix(name, "position:") && account.Type == engine.AccountTypePosition:
			if account.CounterpartyID == strings.TrimPrefix(name, "position:") {
				return account, nil
			}
		case account.Type == engine.AccountTypeUser && account.UserID == name:
			return account, nil
		}
	}
	return engine.Account{}, fmt.Errorf("account %s/%s(%s) not found", providerID, name, currency.Code)
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

  create-account --provider <provider> --user <user> --currency <currency>
                 [--provider-name <name>] [--type <type>] [--counterparty <provider>]
    Create an account. Default type is user; system, liquidity, fx, and position are also supported.
    Example: ttt create-account --provider gatehub --user alice --currency EUR
    Example: ttt create-account --provider blue --provider-name Blue --user bob --currency EUR
    Example: ttt create-account --provider gatehub --type position --currency EUR --counterparty xago

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

  move --from <provider/user> --to <provider/user> --currency <currency> --amount <amount>
       [--workflow <name>] [--step <name>]
    Directly move funds between two user accounts in the same currency.
    Built-in refs are supported: provider/system, provider/liquidity, provider/fx, provider/position:<counterparty>.
    Example: ttt move --from gatehub/alice --to xago/bob --currency EUR --amount 10 --workflow "doing-something" --step "part 0 out of 3"

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
