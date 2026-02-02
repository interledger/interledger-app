package main

import (
	"log"
	"os"

	"gitlab.com/fynbos/backend/providers/chimoney"
	"gitlab.com/fynbos/backend/providers/gatehub"
	gh_client "gitlab.com/fynbos/backend/providers/gatehub/client"
	"gitlab.com/fynbos/backend/providers/pti"
	pti_client "gitlab.com/fynbos/backend/providers/pti/client"
	"gitlab.com/fynbos/pacioli"
	pacioli_client "gitlab.com/fynbos/pacioli/client"

	"gitlab.com/fynbos/backend/limits"

	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/rafiki"

	"gitlab.com/fynbos/backend/identities"

	payments_client "gitlab.com/fynbos/backend/payments/client"

	"gitlab.com/fynbos/backend/payments"
	wallets_client "gitlab.com/fynbos/backend/wallets/client"

	"gitlab.com/fynbos/backend/images"
	images_client "gitlab.com/fynbos/backend/images/client"
	"gitlab.com/fynbos/backend/keys"
	keys_client "gitlab.com/fynbos/backend/keys/client"
	"gitlab.com/fynbos/backend/vault"
	"gitlab.com/fynbos/backend/wallets"

	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"

	"gitlab.com/fynbos/backend/email"
	email_client "gitlab.com/fynbos/backend/email/client"
	"gitlab.com/fynbos/backend/notify"
	notify_client "gitlab.com/fynbos/backend/notify/client"

	transactions_client "gitlab.com/fynbos/backend/transactions/client"

	"gitlab.com/fynbos/backend/transactions"

	"github.com/go-playground/validator/v10"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	kratos "github.com/ory/kratos-client-go"
	"github.com/urfave/cli/v2"
	"gitlab.com/fynbos/backend/cli/actions"
	"gitlab.com/fynbos/backend/kyc"
	kyc_client "gitlab.com/fynbos/backend/kyc/client"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_client "gitlab.com/fynbos/backend/linkedaccounts/client"
	"gitlab.com/fynbos/backend/signup"
	signup_client "gitlab.com/fynbos/backend/signup/client"
	temporal_client "gitlab.com/fynbos/backend/temporal"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client"
	temporal "go.temporal.io/sdk/client"
)

func main() {
	envFile := os.Getenv("ENV_FILE")
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			log.Fatalln(err)
		}
	}

	b := &backends{}
	defer func() {
		if b.db != nil {
			if err := b.db.Close(); err != nil {
				log.Fatalln(err)
			}
		}
	}()

	app := &cli.App{
		Name:  "interledger",
		Usage: "Interact with Interledger application resources.",
		Commands: []*cli.Command{
			{
				Name:    "make",
				Aliases: []string{"m"},
				Usage:   "make new resources",
				Subcommands: []*cli.Command{
					{
						Name:   "wallet",
						Usage:  "create a new Interledger user and wallet",
						Flags:  actions.MakeWalletFlags,
						Action: actions.MakeWallet(b),
					},
					{
						Name:   "ed25519_key_pair",
						Usage:  "generate ed25519 key pair",
						Flags:  nil,
						Action: actions.MakeGenerateEd25519KeyPair(b),
					},
					{
						Name:   "sign_grant_request",
						Usage:  "sign test grant request",
						Flags:  actions.SignGrantRequestFlags,
						Action: actions.MakeSignGrantRequest(b),
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

type backends struct {
	db             *sqlx.DB
	kratos         *kratos.APIClient
	kyc            kyc.Client
	linkedaccounts linkedaccounts.Client
	signup         signup.Client
	temporal       temporal.Client
	twilio         twilio.Service
	user           user.Client
	notify         notify.Client
	val            *validator.Validate
	transactions   transactions.Client
	email          email.Client
	analytics      analytics.Client
	keys           keys.Client
	vault          vault.Client
	img            images.Client
	walletImpl     wallets.Client
	pay            payments.Client
	pc             pacioli.Client
	pcDB           *sqlx.DB
	pti            pti.Client
	gh             gatehub.Client
}

func (b *backends) Chimoney() chimoney.Client {
	return nil
}

func (b *backends) Gatehub() gatehub.Client {
	if b.gh == nil {
		// Create a minimal config for dev CLI
		cfg := gatehub.Config{
			AppID:                  os.Getenv("GATEHUB_APP_ID"),
			Secret:                 os.Getenv("GATEHUB_SECRET"),
			CardAppID:              os.Getenv("GATEHUB_CARD_APP_ID"),
			GatewayID:              os.Getenv("GATEHUB_GATEWAY_ID"),
			CardAccountProductCode: os.Getenv("GATEHUB_CARD_ACCOUNT_PRODUCT_CODE"),
			PaywiserEuroVaultID:    os.Getenv("GATEHUB_PAYWISER_EURO_VAULT_ID"),
			SendingUserID:          os.Getenv("GATEHUB_SENDING_USER_ID"),
			SendingUserAddress:     os.Getenv("GATEHUB_SENDING_USER_ADDRESS"),
			WebhookSecret:          os.Getenv("GATEHUB_WEBHOOK_SECRET"),
			FallbackWebhookURL:     os.Getenv("GATEHUB_FALLBACK_WEBHOOK_URL"),
		}
		b.gh = gh_client.New(b, cfg)
	}
	return b.gh
}

func (b *backends) PTI() pti.Client {
	if b.pti == nil {
		b.pti = pti_client.New(b)
	}
	return b.pti
}

func (b *backends) Limits() limits.Client {
	return nil
}

func (b *backends) PacioliDB() *sqlx.DB {
	if b.pcDB == nil {
		db, err := sqlx.Connect("postgres", "postgres://roach:roach@crdb.interledger.test:26256/pacioli")
		if err != nil {
			log.Fatalln(err)
		}

		b.pcDB = db
	}
	return b.pcDB
}

func (b *backends) Pacioli() pacioli.Client {
	if b.pc == nil {
		b.pc = pacioli_client.NewLocal(b.PacioliDB())
	}
	return b.pc
}

func (b *backends) Xago() xago.Client {
	return nil
}

func (b *backends) Rafiki() rafiki.Client {
	return nil
}

func (b *backends) Identities() identities.Client {
	return nil
}

func (b *backends) Payments() payments.Client {
	if b.pay == nil {
		b.pay = payments_client.New(b)
	}
	return b.pay
}

func (b *backends) Transactions() transactions.Client {
	if b.transactions == nil {
		b.transactions = transactions_client.New(b)
	}
	return b.transactions
}

func (b *backends) DB() *sqlx.DB {
	if b.db == nil {
		db, err := sqlx.Connect("postgres", "postgres://roach:roach@crdb.interledger.test:26256/backend")
		if err != nil {
			log.Fatalln(err)
		}

		b.db = db
	}
	return b.db
}

func (b *backends) Kratos() *kratos.APIClient {
	if b.kratos == nil {
		b.kratos = kratos.NewAPIClient(&kratos.Configuration{
			Servers: kratos.ServerConfigurations{
				{
					URL:         "https://kratos.interledger.test",
					Description: "Public Kratos",
				},
				{
					URL:         "https://kratos-admin.interledger.test",
					Description: "Admin Kratos",
				},
			},
		})
	}
	return b.kratos
}

func (b *backends) Validator() *validator.Validate {
	if b.val == nil {
		b.val = validator.New()
	}
	return b.val
}

func (b *backends) Users() user.Client {
	if b.user == nil {
		b.user = user_client.New(b, "https://kratos.interledger.test", "https://kratos-admin.interledger.test")
	}
	return b.user
}

func (b *backends) Wallets() wallets.Client {
	if b.walletImpl == nil {
		b.walletImpl = wallets_client.New(b)
	}
	return b.walletImpl
}

func (b *backends) Signup() signup.Client {
	if b.signup == nil {
		b.signup = signup_client.New(b)
	}
	return b.signup
}

func (b *backends) Notify() notify.Client {
	if b.notify == nil {
		b.notify = notify_client.New(b, "")
	}

	return b.notify
}

func (b *backends) Keys() keys.Client {
	if b.keys == nil {
		b.keys = keys_client.New(b)
	}

	return b.keys
}

func (b *backends) Vault() vault.Client {
	if b.vault == nil {
		vc, err := vault.NewClient()
		if err != nil {
			log.Fatalln(err)
		}
		b.vault = vc
	}

	return b.vault
}

func (b *backends) Twilio() twilio.Service {
	if b.twilio == nil {
		tw, err := twilio.NewService(&twilio.ServiceArgs{
			AccountSid:   "dev",
			AccountToken: "dev",
			ServiceSid:   "dev",
			ApiBaseUrl:   "http://localhost",
		})
		if err != nil {
			log.Fatalln(err)
		}
		b.twilio = tw
	}
	return b.twilio
}

func (b *backends) KYC() kyc.Client {
	if b.kyc == nil {
		b.kyc, _ = kyc_client.New(b, "ID", "TOKEN")
	}
	return b.kyc
}

func (b *backends) LinkedAccounts() linkedaccounts.Client {
	if b.linkedaccounts == nil {
		b.linkedaccounts = linkedaccounts_client.New(b)
	}
	return b.linkedaccounts
}

func (b *backends) Temporal() temporal.Client {
	if b.temporal == nil {
		tm, err := temporal_client.NewTemporalClient("temporal-frontend.interledger.test:80")
		if err != nil {
			log.Fatalln(err)
		}
		b.temporal = tm
	}
	return b.temporal
}

func (b *backends) Email() email.Client {
	if b.email == nil {
		return email_client.New(b, os.Getenv("SENDGRID_API_KEY"))
	}
	return b.email
}

func (b *backends) Analytics() analytics.Client {
	if b.analytics == nil {
		b.analytics = analytics_client.New(b, "")
	}
	return b.analytics
}

func (b *backends) Images() images.Client {
	if b.img == nil {
		b.img = images_client.New(b)
	}
	return b.img
}
