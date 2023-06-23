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

var CtxKey = UserCtxKey("user")
