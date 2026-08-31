package user

import (
	"time"

	"github.com/interledger/interledger-app/go/backend/country"
)

type User struct {
	ID          string
	Email       string
	PhoneNumber string
	Country     country.Country
	FirstName   string
	LastName    string
	CreatedAt   time.Time
}

type Stats struct {
	Total     int
	ThisYear  int
	ByQuarter [4]int
	Year      int
}

type UserCtxKey string

var CtxKey = UserCtxKey("user")
