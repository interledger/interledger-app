package main

import (
	"log"
	"os"

	"gitlab.com/fynbos/backend/email"
	email_client "gitlab.com/fynbos/backend/email/client"
	"gitlab.com/fynbos/backend/notify"
	notify_client "gitlab.com/fynbos/backend/notify/client"

	"gitlab.com/fynbos/backend/statements"
	statements_client "gitlab.com/fynbos/backend/statements/client"
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
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_client "gitlab.com/fynbos/backend/providers/machnet/client"
	machnet_external "gitlab.com/fynbos/backend/providers/machnet/external"
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
		Name:  "fynbos",
		Usage: "Interact with Fynbos application resources.",
		Commands: []*cli.Command{
			{
				Name:    "make",
				Aliases: []string{"m"},
				Usage:   "make new resources",
				Subcommands: []*cli.Command{
					{
						Name:   "user",
						Usage:  "create a new Fynbos user",
						Flags:  actions.MakeUserFlags,
						Action: actions.MakeUser(b),
					},
					{
						Name:   "machnet_send_user",
						Usage:  "create a new Machnet send user",
						Flags:  actions.MakeMachnetSendUserFlags,
						Action: actions.MakeMachnetSendUser(b),
					},
					{
						Name:   "receive_bank_account",
						Usage:  "create a new receive bank account",
						Flags:  actions.MakeReceiveBankAccountFlags,
						Action: actions.MakeReceiveBankAccount(b),
					},
					{
						Name:   "machnet_transaction",
						Usage:  "create a transaction that uses machnet provider",
						Flags:  actions.MakeMachnetTransactionFlags,
						Action: actions.MakeMachnetTransaction(b),
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
	machnet        machnet.Client
	signup         signup.Client
	temporal       temporal.Client
	twilio         twilio.Service
	user           user.Client
	notify         notify.Client
	val            *validator.Validate
	transactions   transactions.Client
	statements     statements.Client
	email          email.Client
}

func (b *backends) Transactions() transactions.Client {
	if b.transactions == nil {
		b.transactions = transactions_client.New(b)
	}
	return b.transactions
}

func (b *backends) DB() *sqlx.DB {
	if b.db == nil {
		db, err := sqlx.Connect("postgres", "postgres://roach:roach@localhost:26257/backend")
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
					URL:         "http://localhost:4433",
					Description: "Public Kratos",
				},
				{
					URL:         "http://localhost:4434",
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
		b.user = user_client.New(b, "http://localhost:4433", "http://localhost:4434")
	}
	return b.user
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

func (b *backends) Machnet() machnet.Client {
	if b.machnet == nil {
		b.machnet = machnet_client.New(
			b,
			os.Getenv("MACHNET_CLIENT_ID"),
			os.Getenv("MACHNET_CLIENT_SECRET"),
		)
	}
	return b.machnet
}

func (b *backends) MachnetExternal() machnet_external.Client {
	return b.Machnet().External()
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
		tm, err := temporal_client.NewTemporalClient("localhost:7233")
		if err != nil {
			log.Fatalln(err)
		}
		b.temporal = tm
	}
	return b.temporal
}

func (b *backends) Statements() statements.Client {
	if b.statements == nil {
		return statements_client.New()
	}
	return b.statements
}

func (b *backends) Email() email.Client {
	if b.email == nil {
		return email_client.New(b, os.Getenv("SENDGRID_API_KEY"))
	}
	return b.email
}
