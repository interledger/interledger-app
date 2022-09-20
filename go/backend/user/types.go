package user

type User struct {
	ID    string
	Email string
}

type Wallet struct {
	ID   string
	Name string
}

type UserCtxKey string
type WalletCtxKey string
