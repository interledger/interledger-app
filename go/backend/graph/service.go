package graph

import (
	"errors"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/organisation"
	"gitlab.com/fynbos/backend/user"
)

type GraphqlOpts struct {
	Organisation                     organisation.Service
	User                             user.Service
	QueryCacheSize                   uint
	AutomaticPersistedQueryCacheSize uint
}

func NewService(opts GraphqlOpts) (*handler.Server, error) {
	if opts.Organisation == nil {
		return nil, errors.New("Organisation is required.")
	}

	if opts.User == nil {
		return nil, errors.New("User is required.")
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
		Organisations: opts.Organisation,
		User:          opts.User,
	}}))
	svc.SetQueryCache(lru.New(int(queryCacheSize)))
	svc.Use(extension.Introspection{})
	svc.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(int(automaticPersistedQuery)),
	})

	return svc, nil
}
