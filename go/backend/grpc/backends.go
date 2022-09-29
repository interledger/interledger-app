package grpc

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/fakecash"
	"gitlab.com/fynbos/backend/providers/rafiki"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/supporttickets"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/waitlist"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	DB() *sqlx.DB
	AdminAuth() auth.Service
	Agreements() agreements.Client
	Countries() country.Client
	FakeCash() fakecash.Client
	LinkedAccounts() linkedaccounts.Client
	HealthCheck() healthcheck.Service
	Signup() signup.Client
	Rafiki() rafiki.Service
	SupportTickets() supporttickets.Client
	Temporal() temporal.Client
	Twilio() twilio.Service
	Users() user.Client
	Validator() *validator.Validate
	Waitlist() waitlist.Client
	KYC() kyc.Client
}
