package user

type User struct {
	ID          string
	Email       string
	PhoneNumber string
}

type Wallet struct {
	ID   string
	Name string
}

type UserCtxKey string
type WalletCtxKey string

type CreateWalletArgs struct {
	ID     string
	UserID string
	Name   string
}
