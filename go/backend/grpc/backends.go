package grpc

import (
	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/rafiki"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/supporttickets"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/waitlist"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	Accounts() accounts.Client
	AdminAuth() auth.Service
	Agreements() agreements.Client
	FundingSources() fundingsources.Client
	HealthCheck() healthcheck.Service
	Identity() identity.Client
	MX() mx.Client
	Onboarding() onboarding.Client
	Rafiki() rafiki.Service
	SupportTickets() supporttickets.Client
	Temporal() temporal.Client
	Twilio() twilio.Service
	Unit() unit.Client
	Users() user.Client
	Validator() *validator.Validate
	Waitlist() waitlist.Client
}
