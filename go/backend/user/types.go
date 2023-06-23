package user

type User struct {
	ID          string
	Email       string
	PhoneNumber string
}

type UserCtxKey string

var CtxKey = UserCtxKey("user")
