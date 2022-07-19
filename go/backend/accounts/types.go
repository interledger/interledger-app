package accounts

type Account struct {
	ID                         string
	DebitsAccepted             uint64 // TODO change names to match pacioli names
	DebitsReserved             uint64
	CreditsAccepted            uint64
	CreditsReserved            uint64
	AvailableBalance           int64
	IdentityID                 string `db:"identity_id"`
	LedgerAccountID            string `db:"ledger_account_id"` // id returned by Pacioli.
	Provider                   string
	ProviderID                 string `db:"provider_id"`
	VerificationState          string `db:"verification_state"`
	DebitsMustNotExceedCredits bool
	CreditsMustNotExceedDebits bool
	CreatedAt                  string `db:"created_at"`
	UpdatedAt                  string `db:"updated_at"`
}

type CreateAccountArgs struct {
	IdentityID string `validate:"required,uuid"`
	Country    string `validate:"iso3166_1_alpha2"`

	// Points to the next account in array. Last one in array cannot have linked flag set.
	Linked                     bool
	DebitMustNotExceedCredits  bool
	CreditsMustNotExceedDebits bool
}

type VerifyArgs struct {
	AccountID  string `validate:"required,uuid"`
	Provider   string `validate:"oneof=noop"`
	ProviderID string `validate:"required"`
}

const (
	Verified string = "verified"
)

func (s Account) IsVerified() bool {
	return s.VerificationState == Verified
}
