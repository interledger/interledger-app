package grpc

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/discord"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/features"
	"gitlab.com/fynbos/backend/healthcheck"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/waitlist"
	"gitlab.com/fynbos/backend/wallets"
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
	Temporal() temporal.Client
	Twilio() twilio.Service
	Users() user.Client
	Validator() *validator.Validate
	Waitlist() waitlist.Client
	KYC() kyc.Client
	Email() email.Client
	Transactions() transactions.Client
	Analytics() analytics.Client
	Limits() limits.Client
	Contacts() contacts.Client
	Identities() identities.Client
	Keys() keys.Client
	Features() features.Client
	Wallets() wallets.Client
	Payments() payments.Client
	Discord() discord.Client
	Slack() slack.Client
	Rafiki() rafiki.Client
	Xago() xago.Client
	PTI() pti.Client
	Gatehub() gatehub.Client
	Chimoney() chimoney.Client
}
