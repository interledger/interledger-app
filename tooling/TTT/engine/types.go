package engine

import "time"

// AccountType classifies the role of an account in the ledger.
type AccountType int

const (
	AccountTypeSystem    AccountType = iota // per-provider printer: origin of onboarded funds
	AccountTypeLiquidity                    // per-provider + currency reserve pool
	AccountTypePosition                     // sub-account of liquidity tracking inter-provider settlement
	AccountTypeUser                         // individual user balance
)

// EntryType is either Debit or Credit.
type EntryType int

const (
	EntryTypeDebit EntryType = iota
	EntryTypeCredit
)

// Currency represents a monetary unit with a decimal scale factor.
type Currency struct {
	Code       string // e.g. "EUR" — 1–4 uppercase letters
	AssetScale int    // decimal places; stored value / 10^AssetScale = display value
}

var (
	EUR = Currency{Code: "EUR", AssetScale: 2}
	ZAR = Currency{Code: "ZAR", AssetScale: 2}
)

// Provider is a licensed financial institution that holds funds on behalf of the platform.
type Provider struct {
	ID   string
	Name string
}

// Account is a record of financial transactions for a specific asset.
type Account struct {
	ID                 string
	Type               AccountType
	ProviderID         string
	Currency           Currency
	UserID             string // non-empty for AccountTypeUser
	CounterpartyID     string // non-empty for AccountTypePosition (counterparty provider ID)
	LiquidityAccountID string // non-empty for AccountTypePosition (parent liquidity account ID)
}

// LedgerEntry records a single debit or credit posted against an account.
type LedgerEntry struct {
	ID        string
	AccountID string
	Amount    int64 // always positive; in base units defined by currency's AssetScale
	Type      EntryType
	EventID   string // groups all entries belonging to the same workflow execution
	Timestamp time.Time
	Metadata  map[string]string
}

// JournalLine is the future balanced authoring model for accounting events.
// Each line explicitly names both sides of the posting.
//
// During migration, Metadata carries shared line metadata while the optional
// side-specific maps preserve older per-entry details such as step labels.
type JournalLine struct {
	ID              string
	EventID         string
	Timestamp       time.Time
	DebitAccountID  string
	CreditAccountID string
	Amount          int64
	Metadata        map[string]string
	DebitMetadata   map[string]string
	CreditMetadata  map[string]string
}

// SignedAmount returns +Amount for debits and -Amount for credits.
// In a balanced ledger the sum of all signed amounts is always zero.
func SignedAmount(e LedgerEntry) int64 {
	if e.Type == EntryTypeDebit {
		return e.Amount
	}
	return -e.Amount
}

// ChargeRate is the per-direction transfer charge expressed as a fraction of
// the dispatch amount: chargeAmount = dispatch * Num / Den (floor division).
// Nil means no charge is configured for that direction.
type ChargeRate struct {
	Num int64 // numerator; must be >= 0 when non-nil
	Den int64 // denominator; must be > 0 when non-nil
}

// ChargeAmount computes the integer charge for the given dispatch amount using
// floor division. Returns 0 when the receiver is nil or Num is zero.
func (r *ChargeRate) ChargeAmount(dispatch int64) int64 {
	if r == nil || r.Den == 0 || r.Num == 0 {
		return 0
	}
	return dispatch * r.Num / r.Den
}
