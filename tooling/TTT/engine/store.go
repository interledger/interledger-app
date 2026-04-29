package engine

// Store is the data access abstraction for the ledger engine.
// It can be backed by an in-memory implementation or a persistent SQLite layer.
type Store interface {
	// Provider management
	SaveProvider(p Provider) error
	GetProvider(id string) (Provider, error)
	ListProviders() ([]Provider, error)

	// Account management
	SaveAccount(a Account) error
	GetAccount(id string) (Account, error)
	ListAccounts() ([]Account, error)
	FindSystemAccount(providerID string, currency Currency) (Account, bool, error)
	FindLiquidityAccount(providerID string, currency Currency) (Account, bool, error)
	FindPositionAccount(liquidityAccountID, counterpartyID string) (Account, bool, error)
	FindUserAccount(userID, providerID string, currency Currency) (Account, bool, error)
	FindFXAccount(providerID string, currency Currency) (Account, bool, error)

	// Journal-line ledger API.
	PostLines(lines []JournalLine) error
	GetLinesByAccount(accountID string) ([]JournalLine, error)
	GetLinesByEvent(eventID string) ([]JournalLine, error)
	GetAllLines() ([]JournalLine, error)

	// Transitional legacy entry API retained while the engine and UI migrate.
	PostEntries(entries []LedgerEntry) error
	GetEntriesByAccount(accountID string) ([]LedgerEntry, error)
	GetEntriesByEvent(eventID string) ([]LedgerEntry, error)
	GetAllEntries() ([]LedgerEntry, error)

	// Reset erases all providers, accounts, and ledger entries. Intended
	// for the "Clear Everything" UI action and for tests.
	Reset() error

	// Charge configuration — persists per-direction transfer charges.
	// Charges are not cleared by Reset; they survive "Clear Everything".
	GetCharge(fromProviderID, toProviderID string) (*ChargeRate, error)
	SetCharge(fromProviderID, toProviderID string, charge *ChargeRate) error
}
