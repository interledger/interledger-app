package linkedaccounts

type LinkedAccount struct {
	ID         string
	WalletId   string `db:"wallet_id"`
	Name       string
	Mask       string
	Provider   string
	ProviderID string `db:"provider_id"`
	Type       string
	CreatedAt  string `db:"created_at"`
	UpdatedAt  string `db:"updated_at"`
}

type CreateArgs struct {
	ID         string `validate:"omitempty,uuid4"`
	WalletID   string `validate:"required,uuid4"`
	Name       string `validate:"required"`
	Mask       string
	Provider   string `validate:"oneof=mx fakecash machnet"`
	ProviderID string
	Type       string `validate:"required"`
}

type GetByProviderIDArgs struct {
	Provider   string
	ProviderID string
	Type       string
	WalletID   string
}
