package grpc

import (
	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/backend/accountdeletion"
	"github.com/interledger/interledger-app/go/backend/admin/auth"
	"github.com/interledger/interledger-app/go/backend/agreements"
	"github.com/interledger/interledger-app/go/backend/analytics"
	"github.com/interledger/interledger-app/go/backend/contacts"
	"github.com/interledger/interledger-app/go/backend/email"
	"github.com/interledger/interledger-app/go/backend/features"
	"github.com/interledger/interledger-app/go/backend/healthcheck"
	"github.com/interledger/interledger-app/go/backend/identities"
	"github.com/interledger/interledger-app/go/backend/keys"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/limits"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/chimoney"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/pti"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/interledger/interledger-app/go/backend/rafiki"
	"github.com/interledger/interledger-app/go/backend/signup"
	"github.com/interledger/interledger-app/go/backend/slack"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/backend/twilio"
	"github.com/interledger/interledger-app/go/backend/twitter"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/waitlist"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/jmoiron/sqlx"
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
	Slack() slack.Client
	Rafiki() rafiki.Client
	Xago() xago.Client
	PTI() pti.Client
	Gatehub() gatehub.Client
	Chimoney() chimoney.Client
	AccountDeletion() accountdeletion.Client
}
