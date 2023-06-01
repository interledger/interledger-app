package linkedaccounts

import "gitlab.com/fynbos/backend/providers/tabapay"

type LinkedAccount struct {
	ID         string
	WalletID   string `db:"wallet_id"`
	Name       string
	Nickname   string `db:"nickname"`
	Mask       string
	Provider   string
	ProviderID string `db:"provider_id"`
	Type       string
	CanSend    bool   `db:"can_send"`
	CanReceive bool   `db:"can_receive"`
	CreatedAt  string `db:"created_at"`
	UpdatedAt  string `db:"updated_at"`
}

type CreateArgs struct {
	ID         string `validate:"omitempty,uuid4"`
	WalletID   string `validate:"required,uuid4"`
	Name       string `validate:"required"`
	Nickname   string
	Mask       string
	Provider   string `validate:"oneof=mx gmt tabapay"`
	ProviderID string
	Type       string `validate:"required"`
	CanSend    bool
	CanReceive bool
}

type GetByProviderIDArgs struct {
	Provider   string
	ProviderID string
}

func Requires3DS(la *LinkedAccount) bool {
	if la == nil {
		return false
	}

	return la.Provider == tabapay.ProviderName
}
