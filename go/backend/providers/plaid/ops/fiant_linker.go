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
// PTI/Fiant + linked_accounts machinery. The plaid handler does not import
// pti / linkedaccounts / wallets directly — main.go wires a real impl that
// does. Two methods so the handler can decide what to surface (dedupe vs new).
type FiantLinker interface {
	// ExistingLink returns the linked-account row already provisioned for this
	// (userID, plaidAccountID), if any. nil result + nil error means "no dupe".
	// Used by /plaid/link-to-fiant to short-circuit before minting a processor
	// token or posting to Fiant — see the partial unique index in schema.hcl.
	ExistingLink(ctx context.Context, userID, plaidAccountID string) (*LinkedIDs, error)

	// Register completes the cross-system write: it posts the processor token
	// to Fiant via /users/{externalId}/payment-information and persists a
	// linked_account row stamped with `plaid_account_id` for future dedupe.
	Register(ctx context.Context, args LinkPlaidArgs) (*LinkedIDs, error)
}
