package fundingsources

type FundingSource struct {
	ID        string
	WalletId  string `db:"wallet_id"`
	Name      string
	Mask      string
	Provider  string
	Type      string
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type CreateArgs struct {
	ID       string `validate:"omitempty,uuid4"`
	WalletID string `validate:"required,uuid4"`
	Name     string `validate:"required"`
	Mask     string
	Provider string `validate:"oneof= mx"`
	Type     string `validate:"required"`
}
