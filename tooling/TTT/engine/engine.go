// Package engine implements the core accounting domain logic for Toy Treasury Time (TTT).
// It has no dependency on any TUI package and can be used headlessly.
package engine

import (
	"crypto/rand"
	"fmt"
	"time"
)

// Engine is the core accounting service. Create one with New.
type Engine struct {
	store Store
	fx    *FXService
}

// New creates an Engine backed by the given store.
func New(store Store) *Engine {
	return &Engine{store: store}
}

// WithFX attaches a forex service used by CrossProviderTransferAuto. Returns
// the receiver for fluent construction.
func (e *Engine) WithFX(fx *FXService) *Engine {
	e.fx = fx
	return e
}

// FX returns the attached forex service, or nil if none was set.
func (e *Engine) FX() *FXService { return e.fx }

// Reset clears all providers, accounts, and ledger entries from the backing
// store. FX simulator state (if attached via WithFX) is untouched.
func (e *Engine) Reset() error {
	return e.store.Reset()
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand unavailable: %v", err))
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func now() time.Time {
	return time.Now().UTC()
}

// CreateProvider registers a new provider. Returns an error if id or name is empty,
// or if a provider with that ID already exists.
func (e *Engine) CreateProvider(id, name string) (Provider, error) {
	if id == "" {
		return Provider{}, fmt.Errorf("provider id must not be empty")
	}
	if name == "" {
		return Provider{}, fmt.Errorf("provider name must not be empty")
	}
	if _, err := e.store.GetProvider(id); err == nil {
		return Provider{}, fmt.Errorf("provider %q already exists", id)
	}
	p := Provider{ID: id, Name: name}
	return p, e.store.SaveProvider(p)
}

// GetProvider retrieves a provider by ID.
func (e *Engine) GetProvider(id string) (Provider, error) {
	return e.store.GetProvider(id)
}

// ListProviders returns all registered providers.
func (e *Engine) ListProviders() ([]Provider, error) {
	return e.store.ListProviders()
}

// CreateSystemAccount creates the printer account for a provider + currency pair.
// At most one system account may exist per provider + currency combination.
func (e *Engine) CreateSystemAccount(providerID string, currency Currency) (Account, error) {
	if err := e.validateProviderCurrency(providerID, currency); err != nil {
		return Account{}, err
	}
	if _, exists, err := e.store.FindSystemAccount(providerID, currency); err != nil {
		return Account{}, err
	} else if exists {
		return Account{}, fmt.Errorf("system account for provider %q currency %q already exists", providerID, currency.Code)
	}
	a := Account{
		ID:         newID(),
		Type:       AccountTypeSystem,
		ProviderID: providerID,
		Currency:   currency,
	}
	return a, e.store.SaveAccount(a)
}

// CreateLiquidityAccount creates the platform reserve pool account for a provider + currency pair.
// At most one liquidity account may exist per provider + currency combination.
func (e *Engine) CreateLiquidityAccount(providerID string, currency Currency) (Account, error) {
	if err := e.validateProviderCurrency(providerID, currency); err != nil {
		return Account{}, err
	}
	if _, exists, err := e.store.FindLiquidityAccount(providerID, currency); err != nil {
		return Account{}, err
	} else if exists {
		return Account{}, fmt.Errorf("liquidity account for provider %q currency %q already exists", providerID, currency.Code)
	}
	a := Account{
		ID:         newID(),
		Type:       AccountTypeLiquidity,
		ProviderID: providerID,
		Currency:   currency,
	}
	return a, e.store.SaveAccount(a)
}

// CreatePositionAccount creates an inter-provider settlement sub-account under a liquidity account.
// liquidityAccountID must refer to an existing AccountTypeLiquidity account.
// counterpartyID must refer to an existing Provider.
func (e *Engine) CreatePositionAccount(liquidityAccountID, counterpartyID string) (Account, error) {
	liq, err := e.store.GetAccount(liquidityAccountID)
	if err != nil {
		return Account{}, fmt.Errorf("liquidity account: %w", err)
	}
	if liq.Type != AccountTypeLiquidity {
		return Account{}, fmt.Errorf("account %q is not a liquidity account", liquidityAccountID)
	}
	if _, err := e.store.GetProvider(counterpartyID); err != nil {
		return Account{}, fmt.Errorf("counterparty provider: %w", err)
	}
	if _, exists, err := e.store.FindPositionAccount(liquidityAccountID, counterpartyID); err != nil {
		return Account{}, err
	} else if exists {
		return Account{}, fmt.Errorf("position account for liquidity %q counterparty %q already exists", liquidityAccountID, counterpartyID)
	}
	a := Account{
		ID:                 newID(),
		Type:               AccountTypePosition,
		ProviderID:         liq.ProviderID,
		Currency:           liq.Currency,
		LiquidityAccountID: liquidityAccountID,
		CounterpartyID:     counterpartyID,
	}
	return a, e.store.SaveAccount(a)
}

// CreateFXAccount creates the self-exchange pass-through account for a provider + currency pair.
// At most one FX account may exist per provider + currency combination.
func (e *Engine) CreateFXAccount(providerID string, currency Currency) (Account, error) {
	if err := e.validateProviderCurrency(providerID, currency); err != nil {
		return Account{}, err
	}
	if _, exists, err := e.store.FindFXAccount(providerID, currency); err != nil {
		return Account{}, err
	} else if exists {
		return Account{}, fmt.Errorf("FX account for provider %q currency %q already exists", providerID, currency.Code)
	}
	a := Account{
		ID:         newID(),
		Type:       AccountTypeFX,
		ProviderID: providerID,
		Currency:   currency,
	}
	return a, e.store.SaveAccount(a)
}

// FindFXAccount returns the self-exchange FX account for the given provider and currency.
func (e *Engine) FindFXAccount(providerID string, currency Currency) (Account, bool, error) {
	return e.store.FindFXAccount(providerID, currency)
}

// CreateUserAccount creates a user balance account for a given user + provider + currency triple.
// At most one such account may exist per combination.
func (e *Engine) CreateUserAccount(userID, providerID string, currency Currency) (Account, error) {
	if userID == "" {
		return Account{}, fmt.Errorf("userID must not be empty")
	}
	if err := e.validateProviderCurrency(providerID, currency); err != nil {
		return Account{}, err
	}
	if _, exists, err := e.store.FindUserAccount(userID, providerID, currency); err != nil {
		return Account{}, err
	} else if exists {
		return Account{}, fmt.Errorf("user account for user %q provider %q currency %q already exists", userID, providerID, currency.Code)
	}
	a := Account{
		ID:         newID(),
		Type:       AccountTypeUser,
		ProviderID: providerID,
		Currency:   currency,
		UserID:     userID,
	}
	return a, e.store.SaveAccount(a)
}

// GetAccount retrieves an account by ID.
func (e *Engine) GetAccount(id string) (Account, error) {
	return e.store.GetAccount(id)
}

// ListAccounts returns all accounts.
func (e *Engine) ListAccounts() ([]Account, error) {
	return e.store.ListAccounts()
}

// Balance returns the credit-normal balance of an account.
// A positive value means net credits exceed net debits; negative means the reverse.
func (e *Engine) Balance(accountID string) (int64, error) {
	lines, err := e.store.GetLinesByAccount(accountID)
	if err != nil {
		return 0, err
	}
	var bal int64
	for _, line := range lines {
		switch accountID {
		case line.DebitAccountID:
			bal -= line.Amount
		case line.CreditAccountID:
			bal += line.Amount
		}
	}
	return bal, nil
}

func (e *Engine) validateProviderCurrency(providerID string, currency Currency) error {
	if _, err := e.store.GetProvider(providerID); err != nil {
		return fmt.Errorf("provider: %w", err)
	}
	return validateCurrency(currency)
}

// SetCharge configures the per-direction transfer charge between two providers.
// Passing nil clears the charge (feature disabled for that direction).
// Num must be >= 0 and Den > 0 when charge is non-nil.
func (e *Engine) SetCharge(fromProviderID, toProviderID string, charge *ChargeRate) error {
	if charge != nil {
		if charge.Den <= 0 {
			return fmt.Errorf("charge rate denominator must be positive, got %d", charge.Den)
		}
		if charge.Num < 0 {
			return fmt.Errorf("charge rate numerator must be non-negative, got %d", charge.Num)
		}
	}
	return e.store.SetCharge(fromProviderID, toProviderID, charge)
}

// GetCharge returns the configured charge for a provider direction, or nil if
// no charge has been set.
func (e *Engine) GetCharge(fromProviderID, toProviderID string) (*ChargeRate, error) {
	return e.store.GetCharge(fromProviderID, toProviderID)
}

// GetAllEntries returns every ledger entry across all accounts.
func (e *Engine) GetAllEntries() ([]LedgerEntry, error) {
	return e.store.GetAllEntries()
}

// GetAllLines returns every journal line across all accounts.
func (e *Engine) GetAllLines() ([]JournalLine, error) {
	return e.store.GetAllLines()
}

// GetEntriesByAccount returns all ledger entries posted against a specific account.
func (e *Engine) GetEntriesByAccount(accountID string) ([]LedgerEntry, error) {
	return e.store.GetEntriesByAccount(accountID)
}

// ValidateJournalLines checks that a batch of journal lines is well-formed and
// self-consistent before posting.
func (e *Engine) ValidateJournalLines(lines []JournalLine) error {
	_, err := e.normalizeJournalLines(lines)
	return err
}

// PostJournalLines validates a batch of balanced journal lines and persists
// them atomically through the store's journal-line API. Returned lines include
// engine-assigned defaults for IDs, timestamps, and EventID when omitted.
func (e *Engine) PostJournalLines(lines []JournalLine) ([]JournalLine, error) {
	normalized, err := e.normalizeJournalLines(lines)
	if err != nil {
		return nil, err
	}
	if err := e.store.PostLines(normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

func (e *Engine) normalizeJournalLines(lines []JournalLine) ([]JournalLine, error) {
	if len(lines) == 0 {
		return nil, fmt.Errorf("event must contain at least one journal line")
	}
	var eventID string
	for _, line := range lines {
		if line.EventID == "" {
			continue
		}
		if eventID == "" {
			eventID = line.EventID
			continue
		}
		if line.EventID != eventID {
			return nil, fmt.Errorf("journal lines contain mixed event ids: %q and %q", eventID, line.EventID)
		}
	}
	if eventID == "" {
		eventID = newID()
	}
	normalized := make([]JournalLine, len(lines))
	for i, line := range lines {
		line.EventID = eventID
		if line.ID == "" {
			line.ID = newID()
		}
		if line.Timestamp.IsZero() {
			line.Timestamp = now()
		}
		if line.Amount <= 0 {
			return nil, fmt.Errorf("journal line %d amount must be positive, got %d", i, line.Amount)
		}
		if line.DebitAccountID == "" || line.CreditAccountID == "" {
			return nil, fmt.Errorf("journal line %d must have both debit and credit accounts", i)
		}
		if line.DebitAccountID == line.CreditAccountID {
			return nil, fmt.Errorf("journal line %d debit and credit accounts must differ", i)
		}
		debitAcct, err := e.store.GetAccount(line.DebitAccountID)
		if err != nil {
			return nil, fmt.Errorf("journal line %d debit account: %w", i, err)
		}
		creditAcct, err := e.store.GetAccount(line.CreditAccountID)
		if err != nil {
			return nil, fmt.Errorf("journal line %d credit account: %w", i, err)
		}
		if debitAcct.Currency.Code != creditAcct.Currency.Code || debitAcct.Currency.AssetScale != creditAcct.Currency.AssetScale {
			return nil, fmt.Errorf("journal line %d currency mismatch: debit=%s credit=%s", i, debitAcct.Currency.Code, creditAcct.Currency.Code)
		}
		line.Metadata = cloneMetadata(line.Metadata)
		line.DebitMetadata = firstNonEmptyMetadata(line.DebitMetadata, line.Metadata)
		line.CreditMetadata = firstNonEmptyMetadata(line.CreditMetadata, line.Metadata)
		normalized[i] = line
	}
	return normalized, nil
}

func cloneMetadata(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(m))
	for k, v := range m {
		cloned[k] = v
	}
	return cloned
}

func validateCurrency(c Currency) error {
	if len(c.Code) == 0 || len(c.Code) > 4 {
		return fmt.Errorf("currency code must be 1–4 characters, got %q", c.Code)
	}
	for _, r := range c.Code {
		if r < 'A' || r > 'Z' {
			return fmt.Errorf("currency code must be uppercase letters, got %q", c.Code)
		}
	}
	if c.AssetScale < 0 {
		return fmt.Errorf("asset scale must be non-negative")
	}
	return nil
}
