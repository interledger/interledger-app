package models

import "time"

// LinkSession is created when the frontend requests a link token. The mock Link
// UI later resolves the link token back to this session before minting a public
// token.
type LinkSession struct {
	LinkToken string    `json:"link_token"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Account is a single bank account exposed by a mock institution. The AccountID
// is fixed at selection time (see the bank catalog) and read back verbatim by
// every downstream Plaid call so determinism holds across the flow.
type Account struct {
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	Mask      string `json:"mask"`
	Type      string `json:"type"`
	Subtype   string `json:"subtype"`
}

// Item is the persisted result of a mock Link selection. It ties together the
// public token (used once at exchange), the access token (used by every read
// API), the institution, and the chosen account(s).
type Item struct {
	AccessToken     string    `json:"access_token"`
	ItemID          string    `json:"item_id"`
	InstitutionID   string    `json:"institution_id"`
	InstitutionName string    `json:"institution_name"`
	Accounts        []Account `json:"accounts"`
	PublicToken     string    `json:"public_token"`
}
