package ops

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/wallets"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error
	Wallets() wallets.Client
	Notify() notify.Client
	Email() email.Client
	KYC() kyc.Client
	Payments() payments.Client
}
