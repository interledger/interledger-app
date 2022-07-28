package accounts

type Account struct {
	ID                         string
	DebitsAccepted             uint64 // TODO: Change naming to Pending to conform with tigerbeetle
	DebitsReserved             uint64
	CreditsAccepted            uint64
	CreditsReserved            uint64
	AvailableBalance           int64
	IdentityID                 string `db:"identity_id"`
	LedgerAccountID            string `db:"ledger_account_id"` // id returned by Pacioli.
	Provider                   string
	ProviderID                 string `db:"provider_id"`
	DebitsMustNotExceedCredits bool
	CreditsMustNotExceedDebits bool
	CreatedAt                  string `db:"created_at"`
	UpdatedAt                  string `db:"updated_at"`
}

type CreateAccountArgs struct {
	IdentityID string `validate:"required,uuid"`
	Provider   string `validate:"oneof=unit"`
	ProviderID string `validate:"required"`

	// Points to the next account in array. Last one in array cannot have linked flag set.
	Linked                     bool
	DebitMustNotExceedCredits  bool
	CreditsMustNotExceedDebits bool
}

const (
	Verified string = "verified"
)
