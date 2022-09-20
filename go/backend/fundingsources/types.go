package fundingsources

type VerificationState string

const (
	REQUIRED   = VerificationState("required")
	PROCESSING = VerificationState("processing")
	VERIFIED   = VerificationState("verified")
)

type FundingSource struct {
	ID                string
	AccountID         string `db:"account_id"`
	Name              string
	VerificationState string `db:"verification_state"`
	Mask              string
	Type              string
	SubType           string `db:"subtype"`
	CreatedAt         string `db:"created_at"`
	UpdatedAt         string `db:"updated_at"`
}

type UnitCounterParty struct {
	ID                 string
	UnitCounterpartyID string `db:"unit_counterparty_id"`
	CreatedAt          string `db:"created_at"`
	UpdatedAt          string `db:"updated_at"`
}

type CreateArgs struct {
	ID                string `validate:"omitempty,uuid4"`
	WalletID          string `validate:"required,uuid4"`
	Name              string `validate:"required"`
	Mask              string
	VerificationState string `validate:"required"`
	Type              string `validate:"oneof= mx"`
	SubType           string `validate:"required"`
}

type VerifyArgs struct {
	IdentityID      string `validate:"required,uuid4"`
	FundingSourceID string `validate:"required,uuid4"`
}

type CreateBankAccountArgs struct {
	IdentityID    string `validate:"required,uuid4"`
	AccountID     string `validate:"required,uuid4"`
	Name          string `validate:"required"`
	AccountNumber string `validate:"required"`
	RoutingNumber string `validate:"required"`
	Institution   string `validate:"required"`
	Type          string `validate:"required"`
}
