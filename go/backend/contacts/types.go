package contacts

import (
	"database/sql"
	"gitlab.com/fynbos/backend/paymentpointers"
)

type Contact struct {
	ID             string
	Name           string
	PaymentPointer paymentpointers.PaymentPointer `db:"payment_pointer"`
	WalletID       string                         `db:"wallet_id"`
	LastPaidAt     sql.NullTime                   `db:"last_paid_at"`
}

type CreateContactArgs struct {
	Name           string
	PaymentPointer paymentpointers.PaymentPointer `validate:"required"`
	WalletID       string                         `validate:"required,uuid"`
}
