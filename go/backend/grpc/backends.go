package grpc

import (
	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/rafiki"
	"gitlab.com/fynbos/backend/supporttickets"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/waitlist"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	AdminAuth() auth.Service
	Agreements() agreements.Client
	Countries() country.Client
	FundingSources() fundingsources.Client
	HealthCheck() healthcheck.Service
	Onboarding() onboarding.Client
	Rafiki() rafiki.Service
	SupportTickets() supporttickets.Client
	Temporal() temporal.Client
	Twilio() twilio.Service
	Users() user.Client
	Validator() *validator.Validate
	Waitlist() waitlist.Client
}
