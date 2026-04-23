package engine

import "fmt"

// Metadata keys used by the engine's built-in workflows.
const (
	MetaWorkflow = "workflow"
	MetaStep     = "step"

	// Cross-provider FX metadata keys — present on every entry produced by
	// CrossProviderTransfer. The rate is stored as an integer ratio to keep
	// the ledger integer-clean.
	MetaFXRateNum = "fx.rate_num"
	MetaFXRateDen = "fx.rate_den"
	MetaFXBase    = "fx.base"
	MetaFXQuote   = "fx.quote"

	// Settlement metadata keys.
	MetaSettlementCounterparty = "settlement.counterparty"
	MetaSettlementCutoff       = "settlement.cutoff"

	// Charge metadata keys — present on cross-provider events when a charge is applied.
	MetaChargeRateNum = "charge.rate_num"
	MetaChargeRateDen = "charge.rate_den"
	MetaChargeAmount  = "charge.amount"
)

// Workflow name constants.
const (
	WorkflowFundProviderLiquidity = "Fund Provider Liquidity"
	WorkflowUserOnboard           = "User Onboard"
	WorkflowP2PSameProvider       = "P2P Transfer (Same Provider)"
	WorkflowUserOffboard          = "User Offboard"
	WorkflowCrossProviderXfer     = "Cross-Provider Transfer"
	WorkflowBilateralSettlement   = "Bilateral Settlement"
)

// FundProviderLiquidity seeds a provider's liquidity account from its system (printer) account.
// Debit: system account — Credit: liquidity account.
// Compatibility wrapper: returns expanded legacy entry view.
func (e *Engine) FundProviderLiquidity(providerID string, currency Currency, amount int64) ([]LedgerEntry, error) {
	lines, err := e.FundProviderLiquidityLines(providerID, currency, amount)
	if err != nil {
		return nil, err
	}
	return ExpandJournalLines(lines, ""), nil
}

// FundProviderLiquidityLines is the line-native variant of FundProviderLiquidity.
func (e *Engine) FundProviderLiquidityLines(providerID string, currency Currency, amount int64) ([]JournalLine, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive, got %d", amount)
	}
	sys, ok, err := e.store.FindSystemAccount(providerID, currency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("system account for provider %q currency %q not found", providerID, currency.Code)
	}
	liq, ok, err := e.store.FindLiquidityAccount(providerID, currency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("liquidity account for provider %q currency %q not found", providerID, currency.Code)
	}
	meta := workflowMeta(WorkflowFundProviderLiquidity, "fund liquidity")
	line := JournalLine{
		DebitAccountID:  sys.ID,
		CreditAccountID: liq.ID,
		Amount:          amount,
		Metadata:        meta,
	}
	posted, err := e.PostJournalLines([]JournalLine{line})
	if err != nil {
		return nil, err
	}
	return posted, nil
}

// UserOnboard records a user depositing money into the system via their provider.
// Debit: system account — Credit: user account.
// Compatibility wrapper: returns expanded legacy entry view.
func (e *Engine) UserOnboard(userID, providerID string, currency Currency, amount int64) ([]LedgerEntry, error) {
	lines, err := e.UserOnboardLines(userID, providerID, currency, amount)
	if err != nil {
		return nil, err
	}
	return ExpandJournalLines(lines, ""), nil
}

// UserOnboardLines is the line-native variant of UserOnboard.
func (e *Engine) UserOnboardLines(userID, providerID string, currency Currency, amount int64) ([]JournalLine, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive, got %d", amount)
	}
	sys, ok, err := e.store.FindSystemAccount(providerID, currency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("system account for provider %q currency %q not found", providerID, currency.Code)
	}
	user, ok, err := e.store.FindUserAccount(userID, providerID, currency)
	if err != nil {
		return nil, err
	}
	if !ok {
		user, err = e.CreateUserAccount(userID, providerID, currency)
		if err != nil {
			return nil, fmt.Errorf("creating user account: %w", err)
		}
	}
	meta := workflowMeta(WorkflowUserOnboard, "onboard user")
	line := JournalLine{
		DebitAccountID:  sys.ID,
		CreditAccountID: user.ID,
		Amount:          amount,
		Metadata:        meta,
	}
	posted, err := e.PostJournalLines([]JournalLine{line})
	if err != nil {
		return nil, err
	}
	return posted, nil
}

// SameProviderP2PTransfer moves funds between two users on the same provider and currency.
// Debit: sender account — Credit: recipient account.
// Returns an error if the sender has insufficient balance.
// Compatibility wrapper: returns expanded legacy entry view.
func (e *Engine) SameProviderP2PTransfer(senderUserID, recipientUserID, providerID string, currency Currency, amount int64) ([]LedgerEntry, error) {
	lines, err := e.SameProviderP2PTransferLines(senderUserID, recipientUserID, providerID, currency, amount)
	if err != nil {
		return nil, err
	}
	return ExpandJournalLines(lines, ""), nil
}

// SameProviderP2PTransferLines is the line-native variant of SameProviderP2PTransfer.
func (e *Engine) SameProviderP2PTransferLines(senderUserID, recipientUserID, providerID string, currency Currency, amount int64) ([]JournalLine, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive, got %d", amount)
	}
	if senderUserID == recipientUserID {
		return nil, fmt.Errorf("sender and recipient must be different users")
	}
	sender, ok, err := e.store.FindUserAccount(senderUserID, providerID, currency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("sender account for user %q provider %q currency %q not found", senderUserID, providerID, currency.Code)
	}
	recipient, ok, err := e.store.FindUserAccount(recipientUserID, providerID, currency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("recipient account for user %q provider %q currency %q not found", recipientUserID, providerID, currency.Code)
	}
	senderBalance, err := e.Balance(sender.ID)
	if err != nil {
		return nil, err
	}
	if senderBalance < amount {
		return nil, fmt.Errorf("insufficient balance: have %s, need %s %s",
			formatScaledAmount(senderBalance, currency.AssetScale),
			formatScaledAmount(amount, currency.AssetScale),
			currency.Code)
	}
	meta := workflowMeta(WorkflowP2PSameProvider, "p2p transfer")
	line := JournalLine{
		DebitAccountID:  sender.ID,
		CreditAccountID: recipient.ID,
		Amount:          amount,
		Metadata:        meta,
	}
	posted, err := e.PostJournalLines([]JournalLine{line})
	if err != nil {
		return nil, err
	}
	return posted, nil
}

// UserOffboard records a user withdrawing money from the system back to their provider.
// Debit: user account — Credit: system account.
// Returns an error if the user has insufficient balance.
// Compatibility wrapper: returns expanded legacy entry view.
func (e *Engine) UserOffboard(userID, providerID string, currency Currency, amount int64) ([]LedgerEntry, error) {
	lines, err := e.UserOffboardLines(userID, providerID, currency, amount)
	if err != nil {
		return nil, err
	}
	return ExpandJournalLines(lines, ""), nil
}

// UserOffboardLines is the line-native variant of UserOffboard.
func (e *Engine) UserOffboardLines(userID, providerID string, currency Currency, amount int64) ([]JournalLine, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive, got %d", amount)
	}
	user, ok, err := e.store.FindUserAccount(userID, providerID, currency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("user account for user %q provider %q currency %q not found", userID, providerID, currency.Code)
	}
	sys, ok, err := e.store.FindSystemAccount(providerID, currency)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("system account for provider %q currency %q not found", providerID, currency.Code)
	}
	userBalance, err := e.Balance(user.ID)
	if err != nil {
		return nil, err
	}
	if userBalance < amount {
		return nil, fmt.Errorf("insufficient balance: have %s, need %s %s",
			formatScaledAmount(userBalance, currency.AssetScale),
			formatScaledAmount(amount, currency.AssetScale),
			currency.Code)
	}
	meta := workflowMeta(WorkflowUserOffboard, "offboard user")
	line := JournalLine{
		DebitAccountID:  user.ID,
		CreditAccountID: sys.ID,
		Amount:          amount,
		Metadata:        meta,
	}
	posted, err := e.PostJournalLines([]JournalLine{line})
	if err != nil {
		return nil, err
	}
	return posted, nil
}

func workflowMeta(workflow, step string) map[string]string {
	return map[string]string{
		MetaWorkflow: workflow,
		MetaStep:     step,
	}
}

// formatScaledAmount renders a base-unit integer as a decimal string using the
// given scale (e.g. amount=9900, scale=2 → "99.00").
func formatScaledAmount(amount int64, scale int) string {
	if scale <= 0 {
		return fmt.Sprintf("%d", amount)
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
	sign := ""
	if neg {
		sign = "-"
	}
	return fmt.Sprintf("%s%d.%0*d", sign, whole, scale, frac)
}
