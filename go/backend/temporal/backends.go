package temporal

import (
	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/backend/analytics"
	"github.com/interledger/interledger-app/go/backend/contacts"
	"github.com/interledger/interledger-app/go/backend/email"
	"github.com/interledger/interledger-app/go/backend/features"
	"github.com/interledger/interledger-app/go/backend/identities"
	"github.com/interledger/interledger-app/go/backend/images"
	"github.com/interledger/interledger-app/go/backend/keys"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/limits"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/notify"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/chimoney"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/pti"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/interledger/interledger-app/go/backend/rafiki"
	"github.com/interledger/interledger-app/go/backend/signup"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/backend/twilio"
	"github.com/interledger/interledger-app/go/backend/twitter"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/jmoiron/sqlx"
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
	Rafiki() rafiki.Client
	Xago() xago.Client
	Pacioli() pacioli.Client
	PTI() pti.Client
	Signup() signup.Client
	Gatehub() gatehub.Client
	Chimoney() chimoney.Client
}
