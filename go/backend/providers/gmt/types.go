package gmt

import (
	"context"
)

const (
	ProviderName    = "gmt"
	TypeBankAccount = "bankAccount"
	TypeSendCard    = "sendCard"
)

type Await func(context.Context, interface{}) error
