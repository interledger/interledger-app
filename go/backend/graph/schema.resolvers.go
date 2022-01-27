package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"gitlab.com/fynbos/backend/graph/generated"
	_identity "gitlab.com/fynbos/backend/identity"
)

func (r *mutationResolver) CreateIdentity(ctx context.Context, input generated.IdentityInput) (*generated.CreateIdentityMutationResponse, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	identity, err := r.IdentityService.Create(_identity.CreateArgs{
		Country:   input.Country,
		LegalName: input.LegalName,
		User:      user,
	})
	if err != nil {
		switch err.(type) {
		case *_identity.ErrInvalidArgument:
			InvalidArgument(ctx, err.Error())
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	return &generated.CreateIdentityMutationResponse{
		Code:     "200",
		Success:  true,
		Message:  "Created account holder.",
		Identity: identity,
	}, nil
}

func (r *queryResolver) Identity(ctx context.Context) (*_identity.Identity, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	identity, err := r.IdentityService.Get(user.ID)
	if err != nil {
		switch err.(type) {
		case *_identity.ErrNotFound:
			NotFoundError(ctx)
			return nil, nil
		default:
			InternalServerError(ctx)
			return nil, nil
		}
	}

	return identity, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
