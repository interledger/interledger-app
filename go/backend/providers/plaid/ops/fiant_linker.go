package ops

import "context"

type LinkedIDs struct {
	LinkedAccountID      string `json:"linked_account_id"`
	PaymentInformationID string `json:"payment_information_id"`
}

type LinkPlaidArgs struct {
	UserID         string
	PlaidAccountID string
	AccountName    string
	AccountMask    string
	ProcessorToken string
}

// FiantLinker is the cross-package seam between the Plaid HTTP handler and the
// PTI/Fiant.
type FiantLinker interface {
	// WithAccountLock runs fn while holding a per-(userID, plaidAccountID)
	// advisory lock, serializing the dedupe-check → mint → Register critical
	// section so concurrent requests for the same account can't double-register
	// with Fiant. The lock is released when fn returns.
	WithAccountLock(ctx context.Context, userID, plaidAccountID string, fn func(context.Context) error) error

	// ExistingLink returns the linked-account row already provisioned for this
	// (userID, plaidAccountID), if any. nil result + nil error means "no dupe".
	ExistingLink(ctx context.Context, userID, plaidAccountID string) (*LinkedIDs, error)

	// Register completes the cross-system write: it posts the processor token
	// to Fiant, persists a linked_account row, and records a plaid_links row
	// (plaid_account_id) for future dedupe.
	Register(ctx context.Context, args LinkPlaidArgs) (*LinkedIDs, error)

	// ListLinkedPlaidAccountIDs returns the Plaid account_ids that the user
	// has already registered with Fiant via this flow
	ListLinkedPlaidAccountIDs(ctx context.Context, userID string) ([]string, error)

	// IsActivated reports whether the user's wallet is provisioned enough to
	// link a bank: it holds a US balance (the PTI balance linked account that is
	// created asynchronously after KYC). Used to reject link attempts upfront —
	// before any Plaid exchange / processor-token mint — for a not-yet-activated
	// user. Mirrors the "has US balance" signal the frontend gate uses.
	IsActivated(ctx context.Context, userID string) (bool, error)
}
