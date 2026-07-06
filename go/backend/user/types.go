package user

import "github.com/interledger/interledger-app/go/backend/country"

type User struct {
	ID          string
	Email       string
	PhoneNumber string
	Country     country.Country
	FirstName   string
	LastName    string
}

type UserCtxKey string

var CtxKey = UserCtxKey("user")
