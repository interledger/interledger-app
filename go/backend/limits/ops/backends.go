package ops

import (
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/kyc"

	"gitlab.com/fynbos/backend/user"

	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/notify"

	"gitlab.com/fynbos/backend/openpayments"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	OpenPayments() openpayments.Client
	Notify() notify.Client
	Analytics() analytics.Client
	Users() user.Client
	Keys() keys.Client
	KYC() kyc.Client
}
