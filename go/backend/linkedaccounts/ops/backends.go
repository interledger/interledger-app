package ops

import (
	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/backend/email"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/notify"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/wallets"
	"github.com/jmoiron/sqlx"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Wallets() wallets.Client
	Notify() notify.Client
	Email() email.Client
	KYC() kyc.Client
	Payments() payments.Client
}
