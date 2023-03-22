package mx

var (
	ProviderName    = "mx"
	TypeBankAccount = "bankAccount"
)

type CreateBankAccountsArgs struct {
	WalletID    string
	SessionGuid string
	MemberGuid  string
	UserGuid    string
}
