package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	osoErrors "github.com/osohq/go-oso/errors"
	"gitlab.com/fynbos/backend/graph/generated"
	"gitlab.com/fynbos/backend/organisation"
)

func (r *mutationResolver) CreateOrganisation(ctx context.Context, name string) (*generated.OrganisationMutationResponse, error) {
	user, err := r.User.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	org, err := r.Org.Create(name, *user)
	if err != nil {
		switch err.(type) {
		case *osoErrors.NotFoundError:
			NotFoundError(ctx)
			return nil, nil
		case *osoErrors.ForbiddenError:
			ForbiddenError(ctx)
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	return &generated.OrganisationMutationResponse{
		Code:         "200",
		Success:      true,
		Message:      "Created organisation.",
		Organisation: org,
	}, nil
}

func (r *queryResolver) Organisation(ctx context.Context, id string) (*organisation.Organisation, error) {
	user, err := r.User.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	org, err := r.Org.Get(id, *user)
	if err != nil {
		switch err.(type) {
		case *osoErrors.NotFoundError:
			NotFoundError(ctx)
			return nil, nil
		case *osoErrors.ForbiddenError:
			ForbiddenError(ctx)
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	return org, nil
}

func (r *queryResolver) Organisations(ctx context.Context) ([]*organisation.Organisation, error) {
	user, err := r.User.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	orgs, err := r.Org.GetAllOwnedBy(*user)
	if err != nil {
		InternalServerError(ctx)
		return nil, nil
	}

	return orgs, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
