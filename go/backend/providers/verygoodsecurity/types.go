package verygoodsecurity

import (
	"context"
)

const (
	ProviderName = "verygoodsecurity"
)

type Await func(context.Context, interface{}) error

type (
	Card struct {
		ID        string `db:"id"`
		Token     string `json:"card-number" db:"card_token"`
		Expiry    string `json:"exp-date" db:"expiry"`
		CVV       string `json:"card-security-code" db:"card_security_code"`
		WalletID  string `json:"walletId" db:"wallet_id"`
		Last4     string `json:"last4" db:"last4"`
		Type      string `json:"cardType" db:"card_type"`
		CreatedAt string `db:"created_at"`
		UpdatedAt string `db:"updated_at"`
	}
)
