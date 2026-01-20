package temporal

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/aws"
	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/features"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/images"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/sdk/client"
)

type Backends interface {
	Twitter() twitter.Client
	Validator() *validator.Validate
	DB() *sqlx.DB
	Temporal() client.Client
	Users() user.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Email() email.Client
	Transactions() transactions.Client
	Notify() notify.Client
	Analytics() analytics.Client
	Contacts() contacts.Client
	Keys() keys.Client
	Limits() limits.Client
	Identities() identities.Client
	Images() images.Client
	Wallets() wallets.Client
	Features() features.Client
	Twilio() twilio.Service
	Payments() payments.Client
	AWS() aws.Client
	Rafiki() rafiki.Client
	Xago() xago.Client
	Pacioli() pacioli.Client
	PTI() pti.Client
	Signup() signup.Client
	Gatehub() gatehub.Client
	Chimoney() chimoney.Client
}
