package linkedaccounts

import (
	"database/sql"
	"time"

	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/currency"
)

type LinkedAccount struct {
	ID                  string
	WalletID            string `db:"wallet_id"`
	Name                string
	Nickname            string `db:"nickname"`
	Mask                string
	Provider            string
	ProviderID          string `db:"provider_id"`
	Type                string
	CanSend             bool `db:"can_send"`
	CanReceive          bool `db:"can_receive"`
	DefaultReceive      bool `db:"default_receive"`
	DefaultSend         bool `db:"default_send"`
	State               State
	SendCountry         country.Country   `db:"send_country"`
	SendCurrency        currency.Currency `db:"send_currency"`
	SendAvailability    string            `db:"send_availability"`
	SendNetwork         string            `db:"send_network"`
	ReceiveCountry      country.Country   `db:"receive_country"`
	ReceiveCurrency     currency.Currency `db:"receive_currency"`
	ReceiveAvailability string            `db:"receive_availability"`
	ReceiveNetwork      string            `db:"receive_network"`
	CreatedAt           time.Time         `db:"created_at"`
	UpdatedAt           time.Time         `db:"updated_at"`
	DeletedAt           sql.NullTime      `db:"deleted_at"`
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

// Check that the receiving linked account:
// 1) is verified
// 2) can receive
// 3) has the same currency as the sending account OR sending account is sending in USD
func (la *LinkedAccount) CanPay(recvAcc LinkedAccount) bool {
	if la == nil {
		return false
	}

	return recvAcc.State == Verified &&
		recvAcc.CanReceive &&
		recvAcc.ReceiveCurrency == la.SendCurrency
}

type CreateArgs struct {
	ID                  string `validate:"omitempty,uuid4"`
	WalletID            string `validate:"required,uuid4"`
	Name                string `validate:"required"`
	Nickname            string
	Mask                string
	Provider            string `validate:"oneof=xago pti gatehub chimoney"`
	ProviderID          string
	Type                string `validate:"required"`
	CanSend             bool
	CanReceive          bool
	State               State
	SendCountry         country.Country
	SendCurrency        currency.Currency
	SendAvailability    FundsAvailability
	SendNetwork         string
	ReceiveCountry      country.Country
	ReceiveCurrency     currency.Currency
	ReceiveAvailability FundsAvailability
	ReceiveNetwork      string
}

type GetByProviderIDArgs struct {
	Provider   string
	ProviderID string
	WalletID   string
}

type State string

var (
	Verified                State = "Verified"
	OwnershipReviewRequired State = "OwnershipReviewRequired"
	Rejected                State = "Rejected"
)

type Review struct {
	ID                string
	LinkedAccountID   string
	State             State
	NewState          State
	Reason            string
	LinkedAccountMask string
	WalletID          string
	WalletName        string
	ReviewedBy        string
	CreatedAt         time.Time
	CompletedAt       time.Time
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

var validStates = map[State]bool{
	Verified:                true,
	OwnershipReviewRequired: true,
	Rejected:                true,
}

func IsValidState(s State) bool {
	_, ok := validStates[s]
	return ok
}

type FundsAvailability string

const (
	Immediate FundsAvailability = "Immediate"
	Next      FundsAvailability = "Next"
	Few       FundsAvailability = "Few"
)
