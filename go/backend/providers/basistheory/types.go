package basistheory

import (
	"context"
)

const (
	ProviderName = "verygoodsecurity"
)

type Await func(context.Context, interface{}) error

type (
	Card struct {
		ID              string `db:"id"`
		WalletID        string `json:"walletId" db:"wallet_id"`
		TokenID         string `db:"token_id"`
		ExpirationMonth string `db:"expiration_month"`
		ExpirationYear  string `db:"expiration_year"`
		TokenizedNumber string `db:"tokenized_number"`
		Fingerprint     string `db:"fingerprint"`
		CreatedAt       string `db:"created_at"`
		UpdatedAt       string `db:"updated_at"`
	}
)
