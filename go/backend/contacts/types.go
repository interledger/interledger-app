package contacts

import "gitlab.com/fynbos/backend/paymentpointers"

type Contact struct {
	ID             string
	Name           string
	PaymentPointer paymentpointers.PaymentPointer `db:"payment_pointer"`
	WalletID       string                         `db:"wallet_id"`
}

type CreateContactArgs struct {
	Name           string
	PaymentPointer paymentpointers.PaymentPointer `validate:"required"`
	WalletID       string                         `validate:"required,uuid"`
}
