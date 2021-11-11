package graph

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"gitlab.com/fynbos/backend/graph/generated"
	org "gitlab.com/fynbos/backend/organisation"
	"gitlab.com/fynbos/backend/user"
)

func MakeHandler(org org.Service, user user.Service) http.Handler {
	return handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &Resolver{
		Organisations: org,
		User:          user,
	}}))
}
