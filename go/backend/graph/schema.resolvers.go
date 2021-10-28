package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/graph/model"
)

func (r *queryResolver) Organisation(ctx context.Context) (*model.Organisation, error) {
	org := &model.Organisation{
		ID: "1",
		Name: "My first organisation",
	}

	return org, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
