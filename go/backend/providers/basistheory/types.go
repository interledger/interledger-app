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
		ID               string `db:"id"`
		WalletID         string `json:"walletId" db:"wallet_id"`
		TokenID          string `db:"token_id"`
		ExpirationMonth  string `db:"expiration_month"`
		ExpirationYear   string `db:"expiration_year"`
		TokenizedNumber  string `db:"tokenized_number"`
		Bin              string `db:"bin"`
		Fingerprint      string `db:"fingerprint"`
		PullNetwork      string `db:"pull_network"`
		PullEnabled      bool   `db:"pull_enabled"`
		PullType         string `db:"pull_type"`
		PullCountry      string `db:"pull_country"`
		PushNetwork      string `db:"push_network"`
		PushEnabled      bool   `db:"push_enabled"`
		PushType         string `db:"push_type"`
		PushAvailability string `db:"push_availability"`
		PushCountry      string `db:"push_country"`
		CreatedAt        string `db:"created_at"`
		UpdatedAt        string `db:"updated_at"`
	}

	CreateCardArgs struct {
		WalletID         string
		TokenID          string
		Bin              string
		PullNetwork      string
		PullEnabled      bool
		PullType         string
		PullCountry      string
		PushNetwork      string
		PushEnabled      bool
		PushType         string
		PushAvailability string
		PushCountry      string
		CreatedAt        string
		UpdatedAt        string
	}

	CreateCardTokenArgs struct {
		WalletID        string
		Number          string `json:"number"`
		ExpirationMonth int    `json:"expiration_month"`
		ExpirationYear  int    `json:"expiration_year"`
		CVC             string `json:"cvc"`
	}
)
