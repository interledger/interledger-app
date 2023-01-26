package ilp

import "gitlab.com/fynbos/backend/currency"

type CreateStreamCredentialsArgs struct {
	PaymentTag string
	Currency   currency.Currency
}

type StreamCredentials struct {
	SharedSecret string
	IlpAddress   string
}

type IncomingPacket struct {
	PaymentTag string
	Amount     currency.Amount
	Peer       string
}
