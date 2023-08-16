package client_test

import (
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/identities"
	id_client "gitlab.com/fynbos/backend/identities/client"
	"gitlab.com/fynbos/backend/images"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/kyc"
	kyc_client "gitlab.com/fynbos/backend/kyc/client"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccount_client "gitlab.com/fynbos/backend/linkedaccounts/client"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/transactions"
	transaction_client "gitlab.com/fynbos/backend/transactions/client"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client"
	"gitlab.com/fynbos/backend/wallets"
	wallet_client "gitlab.com/fynbos/backend/wallets/client"
	temporal_client "go.temporal.io/sdk/client"
)

type TestBackends struct {
	db       *sqlx.DB
	email    email.Client
	tabapay  tabapay.Client
	user     user.Client
	temporal temporal_client.Client
}

func (b *TestBackends) Twilio() twilio.Service {
	return nil
}

func (b *TestBackends) Signup() signup.Client {
	return nil
}

func (b *TestBackends) Analytics() analytics.Client {
	return nil
}

func (b *TestBackends) KYC() kyc.Client {
	c, err := kyc_client.New(b, "", "")
	if err != nil {
		panic("Can't initialise kyc client")
	}
	return c
}

func (b *TestBackends) Transactions() transactions.Client {
	return transaction_client.New(b)
}

func (b *TestBackends) Tabapay() tabapay.Client {
	return b.tabapay
}

func (b *TestBackends) LinkedAccounts() linkedaccounts.Client {
	return linkedaccount_client.New(b)
}

func (b *TestBackends) Identities() identities.Client {
	return id_client.New(b)
}

func (b *TestBackends) Twitter() twitter.Client {
	return nil
}

func (b *TestBackends) Images() images.Client {
	return nil
}

func (b *TestBackends) Keys() keys.Client {
	return nil
}

func (b *TestBackends) Users() user.Client {
	return user_client.New(b, "", "")
}

func (b *TestBackends) Validator() *validator.Validate {
	return nil
}

func (b *TestBackends) DB() *sqlx.DB {
	return b.db
}

func (b *TestBackends) Temporal() temporal_client.Client {
	return b.temporal
}

func (b *TestBackends) Notify() notify.Client {
	return nil
}

func (b *TestBackends) Email() email.Client {
	return b.email
}

func (b *TestBackends) Wallets() wallets.Client {
	return wallet_client.New(b)
}
