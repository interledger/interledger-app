package graph

import (
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/fundingsources"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/withdrawals"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
	"gitlab.com/fynbos/backend/user"
)

type GraphqlOpts struct {
	Db *sqlx.DB
	// TODO: refactor Identity -> Is etc.
	Identity                         identity.Service             `validate:"required"`
	User                             user.Service                 `validate:"required"`
	Account                          accounts.Service             `validate:"required"`
	AccountTransactions              account_transactions.Service `validate:"required"`
	Noop                             noop.Service                 `validate:"required"`
	Ds                               deposits.Service             `validate:"required"`
	Os                               onboarding.Service           `validate:"required"`
	Ws                               withdrawals.Service          `validate:"required"`
	Fs                               fundingsources.Service       `validate:"required"`
	Ps                               payments.Service             `validate:"required"`
	QueryCacheSize                   uint
	AutomaticPersistedQueryCacheSize uint
}

func NewService(opts GraphqlOpts) (*handler.Server, error) {
	validator := validator.New()
	if err := validator.Struct(opts); err != nil {
		return nil, err
	}

	var queryCacheSize uint = 1000
	if opts.QueryCacheSize != 0 {
		queryCacheSize = opts.QueryCacheSize
	}

	var automaticPersistedQuery uint = 100
	if opts.AutomaticPersistedQueryCacheSize != 0 {
		automaticPersistedQuery = opts.AutomaticPersistedQueryCacheSize
	}

	svc := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: &Resolver{
		Db:                  opts.Db,
		IdentityService:     opts.Identity,
		UserService:         opts.User,
		AccountService:      opts.Account,
		NoopService:         opts.Noop,
		AccountTransactions: opts.AccountTransactions,
		Ds:                  opts.Ds,
		Os:                  opts.Os,
		Ws:                  opts.Ws,
		Fs:                  opts.Fs,
		Ps:                  opts.Ps,
	}}))
	svc.SetQueryCache(lru.New(int(queryCacheSize)))
	svc.Use(extension.Introspection{})
	svc.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(int(automaticPersistedQuery)),
	})

	return svc, nil
}
