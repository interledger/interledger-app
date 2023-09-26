package main

import (
	limits_client "gitlab.com/fynbos/backend/limits/client"
	"gitlab.com/fynbos/discordbot/ops"
	"gitlab.com/fynbos/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"

	payments_client "gitlab.com/fynbos/backend/payments/client"

	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/mx"

	"github.com/bwmarrin/discordgo"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
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
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/vault"
	"gitlab.com/fynbos/backend/wallets"
	wallet_client "gitlab.com/fynbos/backend/wallets/client"
	temporal_client "go.temporal.io/sdk/client"
)

func NewBackends(env *Environment, discordSession *discordgo.Session) *Backends {
	db, err := otelsqlx.Connect("postgres", env.DbURL, otelsql.WithAttributes(semconv.DBSystemCockroachdb), otelsql.WithDBName("cockroachdb"))
	if err != nil {
		log.Fatal("Failed to initialize db client.", zap.Error(err))
	}

	return &Backends{
		db:      db,
		user:    user_mock.NewMock(),
		discord: discordSession,
	}
}

func CloseBackends(b *Backends) {
	if b == nil {
		return
	}

	if b.db != nil {
		_ = b.db.Close()
	}
}

type Backends struct {
	db      *sqlx.DB
	kyc     kyc.Client
	user    user.Client
	discord *discordgo.Session
}

func (b *Backends) Discord() ops.Discord {
	return b.discord
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
