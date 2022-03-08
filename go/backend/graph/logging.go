package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"go.uber.org/zap"
)

func NewLoggingService(gs *handler.Server, logger *zap.Logger) *handler.Server {
	childLogger := logger.With(zap.String("service", "graphql"))
	gs.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		oc := graphql.GetOperationContext(ctx)

		// NB: Do not add variables to this logging so that
		// sensitive variables are not leaked in the logs.
		childLogger.Debug(
			"Graphql query",
			zap.String("operation", oc.OperationName),
			zap.String("query", oc.RawQuery),
		)

		return next(ctx)
	})

	return gs
}
