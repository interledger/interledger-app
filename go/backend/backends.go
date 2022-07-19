package main

import (
	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	accounts_client "gitlab.com/fynbos/backend/accounts/client"
	transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/cli"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/temporal"
	"gitlab.com/fynbos/pacioli"
	pac_client "gitlab.com/fynbos/pacioli/client"
	"go.temporal.io/sdk/client"
	temporal_client "go.temporal.io/sdk/client"
)

type Backends interface {
	Validator() *validator.Validate
	DB() *sqlx.DB
	Identity() identity.Service
	Countries() country.Service
	Pacioli() pacioli.Client
	Accounts() accounts.Client
	FundingSources() fundingsources.Service
	Temporal() client.Client
	Transactions() transactions.Service
	Noop() noop.Service
}

func MakeBackends(args *cli.StartArgs, db *sqlx.DB, logger *zap.Logger) (Backends, error) {
	var b backends
	b.db = db
	b.val = validator.New()

	tp, err := temporal.NewTemporalClient()
	if err != nil {
		return nil, err
	}
	b.tc = tp
	b.countries = country.NewService(db)

	id, err := identity.NewService(identity.ServiceArgs{
		CountryService: b.countries,
		Db:             db,
	}) // TODO: Change to use Backends
	if err != nil {
		return nil, err
	}
	b.id = identity.NewLoggingService(id, logger)

	pClient, err := pac_client.Make(args.PacioliUrl)
	if err != nil {
		return nil, err
	}
	b.pac = pClient

	b.acc = accounts_client.Make(&b, logger) // Passing a pointer here will allow all other dependencies to be lazy loaded as backends gets built

	ts, err := transactions.NewService(&transactions.ServiceArgs{
		AccountService: b.acc,
		PacioliClient:  pClient,
		Db:             db,
	})
	if err != nil {
		return nil, err
	}
	b.tran = transactions.NewLoggingService(ts, logger)

	return &b, nil
}

type backends struct {
	val       *validator.Validate
	db        *sqlx.DB
	tc        temporal_client.Client
	countries country.Service
	id        identity.Service
	pac       pacioli.Client
	acc       accounts.Client
	tran      transactions.Service
	fs        fundingsources.Service
	noop      noop.Service
}

func (b backends) Validator() *validator.Validate {
	//TODO implement me
	panic("implement me")
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) Identity() identity.Service {
	return b.id
}

func (b backends) Countries() country.Service {
	return b.countries
}

func (b backends) Pacioli() pacioli.Client {
	return b.pac
}

func (b backends) Accounts() accounts.Client {
	return b.acc
}

func (b backends) FundingSources() fundingsources.Service {
	return b.fs
}

func (b backends) Temporal() temporal_client.Client {
	return b.tc
}

func (b backends) Transactions() transactions.Service {
	return b.tran
}

func (b backends) Noop() noop.Service {
	return b.noop
}
