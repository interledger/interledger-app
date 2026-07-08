package contacts

import (
	"database/sql"

	"github.com/interledger/interledger-app/go/backend/wallets"
)

type Contact struct {
	ID             string
	Name           string
	PaymentPointer wallets.Address `db:"payment_pointer"`
	WalletID       string          `db:"wallet_id"`
	LastPaidAt     sql.NullTime    `db:"last_paid_at"`
}

type CreateContactArgs struct {
	Name           string
	PaymentPointer wallets.Address `validate:"required"`
	WalletID       string          `validate:"required,uuid"`
}
