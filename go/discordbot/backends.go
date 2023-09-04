package main

import (
	limits_client "gitlab.com/fynbos/backend/limits/client"

	payments_client "gitlab.com/fynbos/backend/payments/client"

	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/mx"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/identities"
	id_client "gitlab.com/fynbos/backend/identities/client"
	"gitlab.com/fynbos/backend/images"
	"gitlab.com/fynbos/backend/keys"
	keys_client "gitlab.com/fynbos/backend/keys/client"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccount_client "gitlab.com/fynbos/backend/linkedaccounts/client"
	"gitlab.com/fynbos/backend/notify"
	notify_client "gitlab.com/fynbos/backend/notify/client"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/transactions"
	transaction_client "gitlab.com/fynbos/backend/transactions/client"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/vault"
	"gitlab.com/fynbos/backend/wallets"
	wallet_client "gitlab.com/fynbos/backend/wallets/client"
	temporal_client "go.temporal.io/sdk/client"
)

func NewBackends(env *Environment) *Backends {
	return &Backends{}
}

type Backends struct {
	db   *sqlx.DB
	kyc  kyc.Client
	user user.Client
}

func (b *Backends) Twilio() twilio.Service {
	return nil
}

func (b *Backends) Signup() signup.Client {
	return nil
}

func (b *Backends) Analytics() analytics.Client {
	return analytics_client.New(b, "")
}

func (b *Backends) KYC() kyc.Client {
	return b.kyc
}

func (b *Backends) Transactions() transactions.Client {
	return transaction_client.New(b)
}

func (b *Backends) Tabapay() tabapay.Client {
	return nil
}

func (b *Backends) LinkedAccounts() linkedaccounts.Client {
	return linkedaccount_client.New(b)
}

func (b *Backends) Identities() identities.Client {
	return id_client.New(b)
}

func (b *Backends) Twitter() twitter.Client {
	return nil
}

func (b *Backends) Images() images.Client {
	return nil
}

func (b *Backends) Keys() keys.Client {
	return keys_client.New(b)
}

func (b *Backends) Vault() vault.Client {
	return nil
}

func (b *Backends) Users() user.Client {
	return b.user
}

func (b *Backends) Validator() *validator.Validate {
	return validator.New()
}

func (b *Backends) DB() *sqlx.DB {
	return b.db
}

func (b *Backends) Temporal() temporal_client.Client {
	return nil
}

func (b *Backends) Notify() notify.Client {
	return notify_client.New(b, "")
}

func (b *Backends) Email() email.Client {
	return nil
}

func (b *Backends) Wallets() wallets.Client {
	return wallet_client.New(b)
}

func (b *Backends) MX() mx.Client {
	return nil
}

func (b *Backends) Limits() limits.Client {
	return limits_client.New(b)
}

func (b *Backends) Payments() payments.Client {
	return payments_client.New(b)
}
