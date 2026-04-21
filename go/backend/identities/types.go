package identities

import (
	"database/sql"
	"time"
)

type AddArgs struct {
	WalletID   string
	Platform   Platform
	Identifier string
}

type VerifyInstructions struct {
	IdentityID   string
	Code         string // Code to publish somewhere public so that we can verify it.
	Instructions string // Long format explanation of how to use to code for us to verify it
}

type Identity struct {
	ID                string
	WalletID          string `db:"wallet_id"`
	Platform          Platform
	State             State
	Public            bool         // Whether the Identity is visible to the public
	KeyID             string       `db:"key_id"`
	Identifier        string       // Can be either the website URL, Twiter handle, Instagram handle etc based on the Platform
	VerificationProof string       `db:"proof"` // URL to the posted tweet or wellknown file etc.
	Signature         []byte       `db:"signature"`
	SignatureHash     []byte       `db:"signature_hash"`
	CreatedAt         time.Time    `db:"created_at"`
	VerifiedAt        sql.NullTime `db:"verified_at"`
}

type State string

const (
	StateUnverified State = "unverified"
	StatePending    State = "pending"
	StateVerified   State = "verified"
	StateFailed     State = "failed"
)

type Platform string

const (
	PlatformTwitter Platform = "twitter"
	PlatformDomain  Platform = "domain"
	PlatformSlack   Platform = "slack"
)

/*
	Claim

Example structure

	{
		"wallet": "https://ilp.link/adrian",
		"type": "twitter",
		"identifier": "@adrian",
		"kid": "external_1",
		"ctime": 837457834876
	}
*/
type Claim struct {
	Wallet     string `json:"wallet"`
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
	Kid        string `json:"kid"`
	Ctime      int64  `json:"ctime"`
}

type SearchResult struct {
	WalletID       string  `db:"wallet_id"`
	WalletUrl      string  `db:"url"`
	WalletName     string  `db:"name"`
	Identifier     string  `db:"identifier"`
	IdentifierType string  `db:"identifier_type"`
	Rank           float64 `db:"rank"`
	SubResults     []SearchResult
}
