package user

import "gitlab.com/fynbos/backend/country"

type User struct {
	ID          string
	Email       string
	PhoneNumber string
	Country     country.Country
}

type UserCtxKey string

var CtxKey = UserCtxKey("user")
