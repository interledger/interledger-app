package admin

import (
	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/backend/admin/auth"
	"github.com/interledger/interledger-app/go/backend/email"
	"github.com/interledger/interledger-app/go/backend/features"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/pti"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/interledger/interledger-app/go/backend/transactions"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/waitlist"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	DB() *sqlx.DB
	AdminAuth() auth.Service
	Validator() *validator.Validate
	Waitlist() waitlist.Client
	Users() user.Client
	KYC() kyc.Client
	Email() email.Client
	LinkedAccounts() linkedaccounts.Client
	Transactions() transactions.Client
	Features() features.Client
	Wallets() wallets.Client
	Payments() payments.Client
	Xago() xago.Client
	PTI() pti.Client
	Gatehub() gatehub.Client
}
