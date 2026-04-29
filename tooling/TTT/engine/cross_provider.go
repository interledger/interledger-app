package engine

import (
	"fmt"
	"strconv"
	"time"
)

// CrossProviderTransfer moves funds from a user on one provider + currency to
// a user on another provider + currency, using the supplied FX ratio.
//
// destAmount = srcAmount * rateNum / rateDen using floor division so the
// ledger stays integer-clean.
//
// Produces a single event with entries on both sides.
//
// Conversion routing is dynamic:
//   - If sender provider has system+liquidity in destination currency,
//     conversion is posted on sender side and bilateral positions are in
//     destination currency.
//   - Otherwise, if recipient provider has system+liquidity in source
//     currency, conversion is posted on recipient side and bilateral
//     positions are in source currency.
//
// Position accounts are created on demand in the chosen settlement currency.
// Compatibility wrapper: returns expanded legacy entry view.
func (e *Engine) CrossProviderTransfer(
	senderUserID, senderProviderID string, senderCurrency Currency,
	recipientUserID, recipientProviderID string, recipientCurrency Currency,
	srcAmount int64, rateNum, rateDen int64,
) ([]LedgerEntry, error) {
	lines, err := e.CrossProviderTransferLines(
		senderUserID, senderProviderID, senderCurrency,
		recipientUserID, recipientProviderID, recipientCurrency,
		srcAmount, rateNum, rateDen,
	)
	if err != nil {
		return nil, err
	}
	return ExpandJournalLines(lines, ""), nil
}

// CrossProviderTransferLines is the line-native variant of CrossProviderTransfer.
// It does not apply any configured charge; use CrossProviderTransferAutoLines for
// the charge-aware path driven by the FX service and stored charge config.
func (e *Engine) CrossProviderTransferLines(
	senderUserID, senderProviderID string, senderCurrency Currency,
	recipientUserID, recipientProviderID string, recipientCurrency Currency,
	srcAmount int64, rateNum, rateDen int64,
) ([]JournalLine, error) {
	return e.crossProviderTransferCoreLines(
		senderUserID, senderProviderID, senderCurrency,
		recipientUserID, recipientProviderID, recipientCurrency,
		srcAmount, rateNum, rateDen, nil,
	)
}

// crossProviderTransferCoreLines is the shared implementation used by both the
// explicit-rate public API and the FX-service / charge-aware Auto variant.
// charge may be nil (no charge applied).
func (e *Engine) crossProviderTransferCoreLines(
	senderUserID, senderProviderID string, senderCurrency Currency,
	recipientUserID, recipientProviderID string, recipientCurrency Currency,
	srcAmount int64, rateNum, rateDen int64,
	charge *ChargeRate,
) ([]JournalLine, error) {
	if srcAmount <= 0 {
		return nil, fmt.Errorf("amount must be positive, got %d", srcAmount)
	}
	if senderProviderID == recipientProviderID {
		return nil, fmt.Errorf("cross-provider transfer requires distinct providers")
	}
	if rateNum <= 0 || rateDen <= 0 {
		return nil, fmt.Errorf("fx rate numerator and denominator must be positive, got %d/%d", rateNum, rateDen)
	}

	// Convert src → dest using floor division. Source and destination legs
	// are balanced independently per currency, so fractional remainders in
	// the FX calculation do not break double-entry. The resulting
	// destination amount must be positive to be meaningful.
	scaled := srcAmount * rateNum
	destAmount := scaled / rateDen
	if destAmount <= 0 {
		return nil, fmt.Errorf("fx rate %d/%d produces non-positive destination amount for %d %s",
			rateNum, rateDen, srcAmount, senderCurrency.Code)
	}

	// Resolve sender side.
	senderUser, ok, err := e.store.FindUserAccount(senderUserID, senderProviderID, senderCurrency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("sender account for user %q provider %q currency %q not found",
			senderUserID, senderProviderID, senderCurrency.Code)
	}
	senderBalance, err := e.Balance(senderUser.ID)
	if err != nil {
		return nil, err
	}

	// Compute charge (may be zero when nil or 0% rate). Balance check covers
	// dispatch + charge so the sender cannot be over-debited.
	chargeAmount := charge.ChargeAmount(srcAmount)
	totalRequired := srcAmount + chargeAmount
	if senderBalance < totalRequired {
		return nil, fmt.Errorf("insufficient balance: have %s, need %s %s",
			formatScaledAmount(senderBalance, senderCurrency.AssetScale),
			formatScaledAmount(totalRequired, senderCurrency.AssetScale),
			senderCurrency.Code)
	}

	// Sender side — source currency plumbing.
	senderLiqSrc, ok, err := e.store.FindLiquidityAccount(senderProviderID, senderCurrency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("liquidity account for provider %q currency %q not found",
			senderProviderID, senderCurrency.Code)
	}
	senderSysSrc, ok, err := e.store.FindSystemAccount(senderProviderID, senderCurrency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("system account for provider %q currency %q not found",
			senderProviderID, senderCurrency.Code)
	}

	// Resolve recipient user destination account (auto-create if needed).
	recipientUser, ok, err := e.store.FindUserAccount(recipientUserID, recipientProviderID, recipientCurrency)
	if err != nil {
		return nil, err
	}
	if !ok {
		recipientUser, err = e.CreateUserAccount(recipientUserID, recipientProviderID, recipientCurrency)
		if err != nil {
			return nil, fmt.Errorf("creating recipient user account: %w", err)
		}
	}

	// Destination-currency capabilities on sender (preferred route).
	senderSysDst, senderHasSysDst, err := e.store.FindSystemAccount(senderProviderID, recipientCurrency)
	if err != nil {
		return nil, err
	}
	senderLiqDst, senderHasLiqDst, err := e.store.FindLiquidityAccount(senderProviderID, recipientCurrency)
	if err != nil {
		return nil, err
	}
	senderCanConvert := senderHasSysDst && senderHasLiqDst

	// Source-currency capabilities on recipient (Mode B fallback).
	recipientSysSrc, recipientHasSysSrc, err := e.store.FindSystemAccount(recipientProviderID, senderCurrency)
	if err != nil {
		return nil, err
	}
	recipientLiqSrc, recipientHasLiqSrc, err := e.store.FindLiquidityAccount(recipientProviderID, senderCurrency)
	if err != nil {
		return nil, err
	}
	recipientCanConvert := recipientHasSysSrc && recipientHasLiqSrc

	// Self-exchange capability: recipient has a pre-funded FX account in the
	// destination currency. Preferred over Mode B when present.
	recipientFXDst, recipientHasFXDst, err := e.store.FindFXAccount(recipientProviderID, recipientCurrency)
	if err != nil {
		return nil, err
	}

	if !senderCanConvert && !recipientHasFXDst && !recipientCanConvert {
		return nil, fmt.Errorf("cross-provider transfer requires either sender conversion (%s/%s on %s), recipient self-exchange FX account (%s on %s), or recipient conversion (%s/%s on %s)",
			recipientCurrency.Code, recipientCurrency.Code, senderProviderID,
			recipientCurrency.Code, recipientProviderID,
			senderCurrency.Code, senderCurrency.Code, recipientProviderID)
	}

	recipientSysDst, recipientHasSysDst, err := e.store.FindSystemAccount(recipientProviderID, recipientCurrency)
	if err != nil {
		return nil, err
	}
	recipientLiqDst, recipientHasLiqDst, err := e.store.FindLiquidityAccount(recipientProviderID, recipientCurrency)
	if err != nil {
		return nil, err
	}

	eventID := newID()
	ts := now()
	baseMeta := map[string]string{
		MetaWorkflow:  WorkflowCrossProviderXfer,
		MetaFXRateNum: strconv.FormatInt(rateNum, 10),
		MetaFXRateDen: strconv.FormatInt(rateDen, 10),
		MetaFXBase:    senderCurrency.Code,
		MetaFXQuote:   recipientCurrency.Code,
	}
	// Attach charge metadata to all lines in this event when a charge is applied.
	if charge != nil && chargeAmount > 0 {
		baseMeta[MetaChargeRateNum] = strconv.FormatInt(charge.Num, 10)
		baseMeta[MetaChargeRateDen] = strconv.FormatInt(charge.Den, 10)
		baseMeta[MetaChargeAmount] = strconv.FormatInt(chargeAmount, 10)
	}

	// The first line debits the sender for the dispatch amount plus any charge.
	// When chargeAmount > 0 the extra funds stay in senderLiqSrc (net of the
	// second line which only moves srcAmount to system/position).
	firstLineAmount := srcAmount + chargeAmount

	var lines []JournalLine
	if recipientHasFXDst {
		baseMeta[MetaSelfExchange] = "true"
	}
	if senderCanConvert {
		if !recipientHasLiqDst {
			return nil, fmt.Errorf("recipient liquidity account for provider %q currency %q not found",
				recipientProviderID, recipientCurrency.Code)
		}
		senderPos, err := e.ensurePositionAccount(senderLiqDst.ID, recipientProviderID)
		if err != nil {
			return nil, fmt.Errorf("sender position: %w", err)
		}
		recipientPos, err := e.ensurePositionAccount(recipientLiqDst.ID, senderProviderID)
		if err != nil {
			return nil, fmt.Errorf("recipient position: %w", err)
		}

		lines = []JournalLine{
			{EventID: eventID, Timestamp: ts, DebitAccountID: senderUser.ID, CreditAccountID: senderLiqSrc.ID, Amount: firstLineAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit sender user"},
				CreditMetadata: map[string]string{MetaStep: "credit sender liquidity"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: senderLiqSrc.ID, CreditAccountID: senderSysSrc.ID, Amount: srcAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit sender liquidity"},
				CreditMetadata: map[string]string{MetaStep: "credit sender system"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: senderSysDst.ID, CreditAccountID: senderLiqDst.ID, Amount: destAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit sender system (converted)"},
				CreditMetadata: map[string]string{MetaStep: "credit sender liquidity (converted)"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: senderLiqDst.ID, CreditAccountID: senderPos.ID, Amount: destAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit sender liquidity"},
				CreditMetadata: map[string]string{MetaStep: "credit sender position"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: recipientPos.ID, CreditAccountID: recipientLiqDst.ID, Amount: destAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit recipient position"},
				CreditMetadata: map[string]string{MetaStep: "credit recipient liquidity"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: recipientLiqDst.ID, CreditAccountID: recipientUser.ID, Amount: destAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit recipient liquidity"},
				CreditMetadata: map[string]string{MetaStep: "credit recipient user"}},
		}
	} else if recipientHasFXDst {
		// Self-exchange routing: EUR flows through bilateral position accounts and
		// accumulates in recipientLiqSrc; ZAR is drawn from the pre-funded ZAR
		// liquidity pool and routed via the FX pass-through account to the recipient.
		if !recipientHasLiqSrc {
			return nil, fmt.Errorf("recipient liquidity account for provider %q currency %q not found",
				recipientProviderID, senderCurrency.Code)
		}
		if !recipientHasLiqDst {
			return nil, fmt.Errorf("recipient liquidity account for provider %q currency %q not found",
				recipientProviderID, recipientCurrency.Code)
		}
		liqDstBal, err := e.Balance(recipientLiqDst.ID)
		if err != nil {
			return nil, err
		}
		if liqDstBal < destAmount {
			return nil, fmt.Errorf("insufficient %s liquidity on %q for self-exchange: have %s, need %s",
				recipientCurrency.Code, recipientProviderID,
				formatScaledAmount(liqDstBal, recipientCurrency.AssetScale),
				formatScaledAmount(destAmount, recipientCurrency.AssetScale))
		}
		senderPos, err := e.ensurePositionAccount(senderLiqSrc.ID, recipientProviderID)
		if err != nil {
			return nil, fmt.Errorf("sender position: %w", err)
		}
		recipientPos, err := e.ensurePositionAccount(recipientLiqSrc.ID, senderProviderID)
		if err != nil {
			return nil, fmt.Errorf("recipient position: %w", err)
		}

		lines = []JournalLine{
			{EventID: eventID, Timestamp: ts, DebitAccountID: senderUser.ID, CreditAccountID: senderLiqSrc.ID, Amount: firstLineAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit sender user"},
				CreditMetadata: map[string]string{MetaStep: "credit sender liquidity"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: senderLiqSrc.ID, CreditAccountID: senderPos.ID, Amount: srcAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit sender liquidity"},
				CreditMetadata: map[string]string{MetaStep: "credit sender position"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: recipientPos.ID, CreditAccountID: recipientLiqSrc.ID, Amount: srcAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit recipient position"},
				CreditMetadata: map[string]string{MetaStep: "credit recipient liquidity (EUR)"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: recipientLiqDst.ID, CreditAccountID: recipientFXDst.ID, Amount: destAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit recipient liquidity (ZAR)"},
				CreditMetadata: map[string]string{MetaStep: "credit recipient FX account"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: recipientFXDst.ID, CreditAccountID: recipientUser.ID, Amount: destAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit recipient FX account"},
				CreditMetadata: map[string]string{MetaStep: "credit recipient user"}},
		}
	} else {
		if !recipientHasSysSrc {
			return nil, fmt.Errorf("recipient system account for provider %q currency %q not found",
				recipientProviderID, senderCurrency.Code)
		}
		if !recipientHasLiqSrc {
			return nil, fmt.Errorf("recipient liquidity account for provider %q currency %q not found",
				recipientProviderID, senderCurrency.Code)
		}
		if !recipientHasSysDst {
			return nil, fmt.Errorf("recipient system account for provider %q currency %q not found",
				recipientProviderID, recipientCurrency.Code)
		}
		if !recipientHasLiqDst {
			return nil, fmt.Errorf("recipient liquidity account for provider %q currency %q not found",
				recipientProviderID, recipientCurrency.Code)
		}

		senderPos, err := e.ensurePositionAccount(senderLiqSrc.ID, recipientProviderID)
		if err != nil {
			return nil, fmt.Errorf("sender position: %w", err)
		}
		recipientPos, err := e.ensurePositionAccount(recipientLiqSrc.ID, senderProviderID)
		if err != nil {
			return nil, fmt.Errorf("recipient position: %w", err)
		}

		lines = []JournalLine{
			{EventID: eventID, Timestamp: ts, DebitAccountID: senderUser.ID, CreditAccountID: senderLiqSrc.ID, Amount: firstLineAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit sender user"},
				CreditMetadata: map[string]string{MetaStep: "credit sender liquidity"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: senderLiqSrc.ID, CreditAccountID: senderPos.ID, Amount: srcAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit sender liquidity"},
				CreditMetadata: map[string]string{MetaStep: "credit sender position"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: recipientPos.ID, CreditAccountID: recipientLiqSrc.ID, Amount: srcAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit recipient position"},
				CreditMetadata: map[string]string{MetaStep: "credit recipient liquidity"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: recipientLiqSrc.ID, CreditAccountID: recipientSysSrc.ID, Amount: srcAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit recipient liquidity"},
				CreditMetadata: map[string]string{MetaStep: "credit recipient system"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: recipientSysDst.ID, CreditAccountID: recipientLiqDst.ID, Amount: destAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit recipient system (converted)"},
				CreditMetadata: map[string]string{MetaStep: "credit recipient liquidity (converted)"}},
			{EventID: eventID, Timestamp: ts, DebitAccountID: recipientLiqDst.ID, CreditAccountID: recipientUser.ID, Amount: destAmount, Metadata: baseMeta,
				DebitMetadata:  map[string]string{MetaStep: "debit recipient liquidity"},
				CreditMetadata: map[string]string{MetaStep: "credit recipient user"}},
		}
	}
	posted, err := e.PostJournalLines(lines)
	if err != nil {
		return nil, err
	}
	return posted, nil
}

// SettleBilateral closes out the open obligation between two providers for a
// given currency up to the supplied cutoff timestamp (inclusive).
//
// Both providers must already have a liquidity account for the currency and
// a position account for each other; these are normally created by
// CrossProviderTransfer. The workflow verifies the bilateral mirror invariant
// before posting any entries.
// Compatibility wrapper: returns expanded legacy entry view.
func (e *Engine) SettleBilateral(
	providerAID, providerBID string, currency Currency, cutoff time.Time,
) ([]LedgerEntry, error) {
	lines, err := e.SettleBilateralLines(providerAID, providerBID, currency, cutoff)
	if err != nil {
		return nil, err
	}
	return ExpandJournalLines(lines, ""), nil
}

// SettleBilateralLines is the line-native variant of SettleBilateral.
func (e *Engine) SettleBilateralLines(
	providerAID, providerBID string, currency Currency, cutoff time.Time,
) ([]JournalLine, error) {
	if providerAID == providerBID {
		return nil, fmt.Errorf("settlement requires distinct providers")
	}
	liqA, ok, err := e.store.FindLiquidityAccount(providerAID, currency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("liquidity account for provider %q currency %q not found", providerAID, currency.Code)
	}
	liqB, ok, err := e.store.FindLiquidityAccount(providerBID, currency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("liquidity account for provider %q currency %q not found", providerBID, currency.Code)
	}
	posA, ok, err := e.store.FindPositionAccount(liqA.ID, providerBID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("no position account in %q liquidity for counterparty %q", providerAID, providerBID)
	}
	posB, ok, err := e.store.FindPositionAccount(liqB.ID, providerAID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("no position account in %q liquidity for counterparty %q", providerBID, providerAID)
	}

	balA, err := e.balanceAsOf(posA.ID, cutoff)
	if err != nil {
		return nil, err
	}
	balB, err := e.balanceAsOf(posB.ID, cutoff)
	if err != nil {
		return nil, err
	}
	if balA+balB != 0 {
		return nil, fmt.Errorf("bilateral mirror broken: pos(%s→%s)=%d, pos(%s→%s)=%d",
			providerAID, providerBID, balA, providerBID, providerAID, balB)
	}
	if balA == 0 {
		return nil, fmt.Errorf("nothing to settle between %q and %q for %s up to %s",
			providerAID, providerBID, currency.Code, cutoff.Format(time.RFC3339))
	}

	// Whichever side has a positive credit-normal balance is the creditor.
	// The creditor's position is debited (clears its credit); its liquidity
	// is credited (reserves grow). The debtor's liquidity is debited (reserves
	// shrink); its position is credited (clears its debit balance).
	creditorLiq, creditorPos := liqA, posA
	debtorLiq, debtorPos := liqB, posB
	creditorID, debtorID := providerAID, providerBID
	amount := balA
	if amount < 0 {
		creditorLiq, creditorPos = liqB, posB
		debtorLiq, debtorPos = liqA, posA
		creditorID, debtorID = providerBID, providerAID
		amount = -amount
	}

	eventID := newID()
	ts := now()
	baseMeta := map[string]string{
		MetaWorkflow:               WorkflowBilateralSettlement,
		MetaSettlementCounterparty: debtorID,
		MetaSettlementCutoff:       cutoff.UTC().Format(time.RFC3339),
	}

	lines := []JournalLine{
		// Creditor's position → creditor's liquidity
		{
			EventID:         eventID,
			Timestamp:       ts,
			DebitAccountID:  creditorPos.ID,
			CreditAccountID: creditorLiq.ID,
			Amount:          amount,
			Metadata:        baseMeta,
			DebitMetadata: map[string]string{
				MetaStep: "debit creditor position",
			},
			CreditMetadata: map[string]string{
				MetaStep: "credit creditor liquidity",
			},
		},
		// Debtor's liquidity → debtor's position
		{
			EventID:         eventID,
			Timestamp:       ts,
			DebitAccountID:  debtorLiq.ID,
			CreditAccountID: debtorPos.ID,
			Amount:          amount,
			Metadata: map[string]string{
				MetaWorkflow:               WorkflowBilateralSettlement,
				MetaSettlementCounterparty: creditorID,
				MetaSettlementCutoff:       cutoff.UTC().Format(time.RFC3339),
			},
			DebitMetadata: map[string]string{
				MetaStep: "debit debtor liquidity",
			},
			CreditMetadata: map[string]string{
				MetaStep: "credit debtor position",
			},
		},
	}
	posted, err := e.PostJournalLines(lines)
	if err != nil {
		return nil, err
	}
	return posted, nil
}

// ensurePositionAccount returns the position account for a given liquidity
// account + counterparty, creating it if necessary.
func (e *Engine) ensurePositionAccount(liquidityAccountID, counterpartyID string) (Account, error) {
	if existing, ok, err := e.store.FindPositionAccount(liquidityAccountID, counterpartyID); err != nil {
		return Account{}, err
	} else if ok {
		return existing, nil
	}
	return e.CreatePositionAccount(liquidityAccountID, counterpartyID)
}

// balanceAsOf returns the credit-normal balance of an account including only
// journal lines with timestamp <= cutoff.
func (e *Engine) balanceAsOf(accountID string, cutoff time.Time) (int64, error) {
	lines, err := e.store.GetLinesByAccount(accountID)
	if err != nil {
		return 0, err
	}
	var bal int64
	for _, line := range lines {
		if line.Timestamp.After(cutoff) {
			continue
		}
		switch accountID {
		case line.DebitAccountID:
			bal -= line.Amount
		case line.CreditAccountID:
			bal += line.Amount
		}
	}
	return bal, nil
}

// CrossProviderTransferAuto is the FX-aware entry point used by the UI: it
// looks up the current base→quote rate from the attached FXService, performs
// the transfer using floor-division rounding on the destination amount, then
// — on success only — mutates the FX rate by +5 % or −5 % for the next
// conversion. Phase-4 behavior.
//
// Returns the posted ledger entries and the rate that was actually applied
// to this event (not the mutated next-event rate).
// Compatibility wrapper: returns expanded legacy entry view.
func (e *Engine) CrossProviderTransferAuto(
	senderUserID, senderProviderID string, senderCurrency Currency,
	recipientUserID, recipientProviderID string, recipientCurrency Currency,
	srcAmount int64,
) ([]LedgerEntry, Rate, error) {
	lines, rate, err := e.CrossProviderTransferAutoLines(
		senderUserID, senderProviderID, senderCurrency,
		recipientUserID, recipientProviderID, recipientCurrency,
		srcAmount,
	)
	if err != nil {
		return nil, Rate{}, err
	}
	return ExpandJournalLines(lines, ""), rate, nil
}

// CrossProviderTransferAutoLines is the line-native variant of CrossProviderTransferAuto.
func (e *Engine) CrossProviderTransferAutoLines(
	senderUserID, senderProviderID string, senderCurrency Currency,
	recipientUserID, recipientProviderID string, recipientCurrency Currency,
	srcAmount int64,
) ([]JournalLine, Rate, error) {
	if e.fx == nil {
		return nil, Rate{}, fmt.Errorf("no FX service attached; use CrossProviderTransferLines with an explicit rate")
	}
	rate, err := e.fx.Rate(senderCurrency.Code, recipientCurrency.Code)
	if err != nil {
		return nil, Rate{}, err
	}
	// Look up the configured charge for this provider direction (nil = no charge).
	charge, err := e.store.GetCharge(senderProviderID, recipientProviderID)
	if err != nil {
		return nil, Rate{}, fmt.Errorf("looking up charge config: %w", err)
	}
	lines, err := e.crossProviderTransferCoreLines(
		senderUserID, senderProviderID, senderCurrency,
		recipientUserID, recipientProviderID, recipientCurrency,
		srcAmount, rate.Num, rate.Den, charge,
	)
	if err != nil {
		// Per spec: do NOT mutate on failure.
		return nil, Rate{}, err
	}
	if _, mErr := e.fx.Mutate(senderCurrency.Code, recipientCurrency.Code); mErr != nil {
		// Extremely unlikely (rate existed on lookup); don't unwind the posted
		// entries — surface the error for visibility.
		return lines, rate, fmt.Errorf("transfer posted but rate mutation failed: %w", mErr)
	}
	return lines, rate, nil
}
