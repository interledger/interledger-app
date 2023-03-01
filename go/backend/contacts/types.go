package contacts

type Contact struct {
	ID             string
	Name           string
	PaymentPointer string `db:"payment_pointer"`
	WalletID       string `db:"wallet_id"`
}

type CreateContactArgs struct {
	Name           string
	PaymentPointer string `validate:"required"`
	WalletID       string `validate:"required,uuid"`
}
