//go:build !codeanalysis
// +build !codeanalysis

// github.com/urfave/cli/v2 fails linting so disabling it for now

package main

import (
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/jmoiron/sqlx"
	kratos "github.com/ory/kratos-client-go"
	"github.com/urfave/cli/v2"
	"gitlab.com/fynbos/backend/cli/actions"
	"gitlab.com/fynbos/backend/signup"
	signup_client "gitlab.com/fynbos/backend/signup/client"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client"
)

func main() {
	os.Setenv("FYNBOS_ENV", "dev")
	db, err := sqlx.Connect("postgres", "postgres://roach:roach@localhost:26257/backend")
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalln(err)
		}
	}()
	b := &backends{
		db: db,
		kratos: kratos.NewAPIClient(&kratos.Configuration{
			Servers: kratos.ServerConfigurations{
				{
					URL:         "http://localhost:4434",
					Description: "Dev Kratos",
				},
			},
		}),
		val: validator.New(),
	}
	b.user = user_client.New(b, "http://localhost:4434")
	b.signup = signup_client.New(b)
	b.twilio, err = twilio.NewService(&twilio.ServiceArgs{
		AccountSid:   "dev",
		AccountToken: "dev",
		ServiceSid:   "dev",
		ApiBaseUrl:   "http://localhost",
	})
	if err != nil {
		log.Fatalln(err)
	}
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
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

type backends struct {
	db     *sqlx.DB
	kratos *kratos.APIClient
	signup signup.Client
	twilio twilio.Service
	user   user.Client
	val    *validator.Validate
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) Kratos() *kratos.APIClient {
	return b.kratos
}

func (b backends) Validator() *validator.Validate {
	return b.val
}

func (b backends) User() user.Client {
	return b.user
}

func (b backends) Signup() signup.Client {
	return b.signup
}

func (b backends) Twilio() twilio.Service {
	return b.twilio
}
