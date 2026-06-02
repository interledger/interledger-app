package ops

import "context"

// LinkedIDs identifies the two persisted artefacts created (or rediscovered)
// when a Plaid account is registered as a Fiant deposit source.
type LinkedIDs struct {
	LinkedAccountID      string `json:"linked_account_id"`
	PaymentInformationID string `json:"payment_information_id"`
}

// LinkPlaidArgs is what the handler hands to FiantLinker.Register once a fresh
// processor token has been minted. `AccountName` / `AccountMask` come from the
// Plaid metadata sent by the frontend.
type LinkPlaidArgs struct {
	UserID         string
	PlaidAccountID string
	AccountName    string
	AccountMask    string
	ProcessorToken string
}

// FiantLinker is the cross-package seam between the Plaid HTTP handler and the
// PTI/Fiant + linked_accounts machinery.
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
	// to Fiant and persists a linked_account row stamped with `plaid_account_id` for future dedupe.
	Register(ctx context.Context, args LinkPlaidArgs) (*LinkedIDs, error)

	// ListLinkedPlaidAccountIDs returns the Plaid account_ids that the user
	// has already registered with Fiant via this flow
	ListLinkedPlaidAccountIDs(ctx context.Context, userID string) ([]string, error)
}
