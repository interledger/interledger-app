package admin

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/features"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/waitlist"
	"gitlab.com/fynbos/backend/wallets"
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
