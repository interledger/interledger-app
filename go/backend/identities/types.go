package identities

type AddArgs struct {
	WalletID string
	Platform Platform
	Handle   string
	Public   bool
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
	Handle            string // Can be either the website URL, Twiter handle, Instagram handle etc based on the Platform
	State             State
	VerificationCode  string `db:"code"`
	VerificationProof string `db:"proof"` // URL to the posted tweet or wellknown file etc.
	Public            bool   // Whether the Identity is visible to the public
}

type IdentityClaim struct {
	Wallet       string `json:"wallet"`
	Identifier   string `json:"identifier"`
	KeyID        string `json:"keyid"`
	Type         string `json:"type"`
	CreationTime string `json:"ctime"`
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
)
