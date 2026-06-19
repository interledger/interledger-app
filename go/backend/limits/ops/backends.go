package ops

import (
	"github.com/interledger/interledger-app/go/backend/keys"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/wallets"

	"github.com/interledger/interledger-app/go/backend/user"

	"github.com/interledger/interledger-app/go/backend/analytics"
	"github.com/interledger/interledger-app/go/backend/notify"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Notify() notify.Client
	Analytics() analytics.Client
	Users() user.Client
	Wallets() wallets.Client
	Keys() keys.Client
	KYC() kyc.Client
}
