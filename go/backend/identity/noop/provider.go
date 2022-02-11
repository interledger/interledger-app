package noop

import (
	"time"

	"github.com/google/uuid"
)

const (
	Verified  string = "verified"
	Retry     string = "retry"
	Kba       string = "kba"
	Document  string = "document"
	Suspended string = "suspended"
)

type Provider interface {
	CreateCustomer(args *CreateCustomerArgs) (*Customer, error)
}

type provider struct{}

func NewProvider() Provider {
	return &provider{}
}

type CreateCustomerArgs struct {
	FirstName   string
	LastName    string
	Email       string
	Address1    string
	Address2    string
	State       string
	City        string
	PostalCode  string
	DateOfBirth string
	Ssn         string
}

type Customer struct {
	CreateCustomerArgs
	ID      string
	Status  string
	Created string
	Links   map[string]map[string]string
}

func (self *provider) CreateCustomer(args *CreateCustomerArgs) (*Customer, error) {
	return &Customer{
		ID:                 uuid.NewString(),
		CreateCustomerArgs: *args,
		Created:            time.Now().String(),
		Status:             Verified,
	}, nil
}
