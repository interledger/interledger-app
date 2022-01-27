package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// This returns a forbidden error according to
// https://www.apollographql.com/docs/apollo-server/data/errors/#error-codes
func ForbiddenError(ctx context.Context) {
	graphql.AddError(ctx, &gqlerror.Error{
		Path:    graphql.GetPath(ctx),
		Message: "Forbidden.",
		Extensions: map[string]interface{}{
			"code": "FORBIDDEN",
		},
	})
}

// This corresponds to the Oso not found error.
func NotFoundError(ctx context.Context) {
	graphql.AddError(ctx, &gqlerror.Error{
		Path:    graphql.GetPath(ctx),
		Message: "Not found.",
		Extensions: map[string]interface{}{
			"code": "NOT_FOUND",
		},
	})
}

// Generic error when the server fails to processa request.
func InternalServerError(ctx context.Context) {
	graphql.AddError(ctx, &gqlerror.Error{
		Path:    graphql.GetPath(ctx),
		Message: "Unable to process request.",
		Extensions: map[string]interface{}{
			"code": "INTERNAL_SERVER_ERROR",
		},
	})
}

// Argument validation has failed. Argument may have passed
// graphql validation but has now failed at the service level.
func InvalidArgument(ctx context.Context, message string) {
	graphql.AddError(ctx, &gqlerror.Error{
		Path:    graphql.GetPath(ctx),
		Message: "Bad input: " + message,
		Extensions: map[string]interface{}{
			"code": "BAD_USER_INPUT",
		},
	})
}
