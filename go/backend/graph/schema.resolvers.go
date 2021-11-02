package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/models"
)

func (r *mutationResolver) CreateOrganisation(ctx context.Context, name string) (*models.Organisation, error) {
	org, err := r.Organisations.Create(name)
	if err != nil {
		return nil, err
	}

	return org, nil
}

func (r *queryResolver) Organisation(ctx context.Context, id string) (*models.Organisation, error) {
	org, err := r.Organisations.Get(id)
	// TODO: error handling
	if err != nil {
		return nil, err
	}

	return org, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
