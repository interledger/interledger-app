package linkedaccounts

import (
	"time"

	"gitlab.com/fynbos/backend/providers/tabapay"
)

type LinkedAccount struct {
	ID         string
	WalletID   string `db:"wallet_id"`
	Name       string
	Nickname   string `db:"nickname"`
	Mask       string
	Provider   string
	ProviderID string `db:"provider_id"`
	Type       string
	CanSend    bool `db:"can_send"`
	CanReceive bool `db:"can_receive"`
	State      State
	CreatedAt  string `db:"created_at"`
	UpdatedAt  string `db:"updated_at"`
}

func (la *LinkedAccount) Title() string {
	if la == nil {
		return ""
	}
	if la.Nickname != "" {
		return la.Nickname
	}
	return la.Mask
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
	State      State
}

type GetByProviderIDArgs struct {
	Provider   string
	ProviderID string
	WalletID   string
}

func Requires3DS(la *LinkedAccount) bool {
	if la == nil {
		return false
	}

	return la.Provider == tabapay.ProviderName
}

type State string

var (
	Verified                State = "Verified"
	OwnershipReviewRequired State = "OwnershipReviewRequired"
	Rejected                State = "Rejected"
)

type Review struct {
	ID              string
	LinkedAccountID string
	State           State
	NewState        State
	Reason          string
	ReviewedBy      string
	CreatedAt       time.Time
	CompletedAt     time.Time
}

type CreateReviewArgs struct {
	LinkedAccountID string
	State           State
}

type CompleteReviewArgs struct {
	ID         string
	Reason     string
	NewState   State
	ReviewedBy string
}
