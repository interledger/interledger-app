package mx

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
