package mx

import (
	"time"

	"gitlab.com/fynbos/backend/currency"
)

var (
	ProviderName    = "mx"
	TypeBankAccount = "bankAccount"
	TypeSavings     = "SAVINGS"
	TypeChecking    = "CHECKING"
)

type CreateBankAccountsArgs struct {
	WalletID    string
	SessionGuid string
	MemberGuid  string
	UserGuid    string
}

type Account struct {
	Guid             string
	MemberGuid       string
	UserGuid         string
	Name             string
	AccountNumber    string
	Balance          currency.Amount
	AvailableBalance currency.Amount
	InstitutionCode  string
	IsHidden         bool
	Nickname         string
	RoutingNumber    string
	Type             string
	UpdatedAt        time.Time
}

type User struct {
	WalletID string
	GUID     string
	Enabled  bool
}
