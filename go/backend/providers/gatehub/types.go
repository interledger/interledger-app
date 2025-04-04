package gatehub

import (
	"database/sql"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
)

const (
	ProviderName   = "gatehub"
	AccTypeBalance = "balance"

	LedgerIDEUR   uint32 = 4482387 // Spells ghubeur on a Nokia 3320 keyboard
	EUROpsAccount        = "1854f171-eafa-4e30-bf66-7dbfe167ccfa"
)

type Balance struct {
	Total     currency.Amount
	Available currency.Amount
}

type CreateTransferArgs struct {
	SendingLinkedAccountID   string
	ReceivingLinkedAccountID string
	Amount                   currency.Amount
}

type User = external.User

type Card = external.Card

type ExternalIDs struct {
	ID         string         `db:"external_id"`
	CustomerID sql.NullString `db:"external_customer_id"`
}
