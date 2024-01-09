package pti

import (
	"context"
	"time"

	"gitlab.com/fynbos/backend/currency"
)

var (
	ProviderName   = "pti"
	AccTypeBalance = "balance"

	ScenarioDeposit    = "fynbos_deposit"
	ScenarioWithdrawal = "fynbos_withdrawal"
)

type User struct {
	ID               string `db:"id"`
	ExternalID       string `db:"external_id"`
	WalletID         string `db:"wallet_id"`
	Status           string `db:"status"`
	AssessmentStatus string `db:"assessment_status"`
	CreatedAt        string `db:"created_at"`
	UpdatedAt        string `db:"updated_at"`
}

type Wallet struct {
	ID        string            `db:"id"`
	UserID    string            `db:"external_user_id"`
	Reference string            `db:"reference"`
	Currency  currency.Currency `db:"currency"`
	Balance   currency.Amount
	CreatedAt time.Time `db:"created_at"`
}

type Await func(ctx context.Context, result interface{}) error

type CreateWalletArgs struct {
	WalletID string
	Currency currency.Currency
	Nickname string
	Title    string
}

type CreateExternalWalletArgs struct {
	ID       string
	UserID   string
	Currency currency.Currency
}
