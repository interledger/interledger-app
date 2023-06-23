package grpc

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/features"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/linkedin"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/providers/basistheory"
	"gitlab.com/fynbos/backend/providers/gmt"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/supporttickets"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/waitlist"
	temporal "go.temporal.io/sdk/client"
)

type Backends interface {
	Twitter() twitter.Client
	DB() *sqlx.DB
	AdminAuth() auth.Service
	Agreements() agreements.Client
	LinkedAccounts() linkedaccounts.Client
	HealthCheck() healthcheck.Service
	Signup() signup.Client
	SupportTickets() supporttickets.Client
	Temporal() temporal.Client
	Twilio() twilio.Service
	Users() user.Client
	Validator() *validator.Validate
	Waitlist() waitlist.Client
	KYC() kyc.Client
	Email() email.Client
	Transactions() transactions.Client
	Authorisation() authorisation.InternalClient
	Analytics() analytics.Client
	OpenPayments() openpayments.Client
	Limits() limits.Client
	Contacts() contacts.Client
	Identities() identities.Client
	MX() mx.Client
	GMT() gmt.Client
	Tabapay() tabapay.Client
	Keys() keys.Client
	BasisTheory() basistheory.Client
	Features() features.Client
	Linkedin() linkedin.Client
}
