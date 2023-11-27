package contacts

import (
	"database/sql"

	"gitlab.com/fynbos/backend/wallets"
)

type Contact struct {
	ID            string
	Name          string
	WalletAddress wallets.Address `db:"wallet_address"`
	WalletID      string          `db:"wallet_id"`
	LastPaidAt    sql.NullTime    `db:"last_paid_at"`
}

type CreateContactArgs struct {
	Name          string
	WalletAddress wallets.Address `validate:"required"`
	WalletID      string          `validate:"required,uuid"`
}
