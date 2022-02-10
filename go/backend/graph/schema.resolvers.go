package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/graph/generated"
	_identity "gitlab.com/fynbos/backend/identity"
)

func (r *mutationResolver) CreateIdentity(ctx context.Context, input generated.CreateIdentityInput) (*generated.CreateIdentityMutationResponse, error) {
	user, err := r.UserService.ForContext(ctx)
	if err != nil {
		ForbiddenError(ctx)
		return nil, nil
	}

	var identity *_identity.Identity
	err = crdbsqlx.ExecuteTx(ctx, r.Db, nil, func(tx *sqlx.Tx) error {
		_identity, err := r.IdentityService.Create(ctx, tx, _identity.CreateArgs{
			ID:           user.ID,
			FirstName:    input.FirstName,
			LastName:     input.LastName,
			MobileNumber: input.MobileNumber,
			Country:      input.Country,
			Email:        user.Email,
		})
		if err != nil {
			return err
		}

		_, err = r.AccountService.Create(ctx, tx, &accounts.CreateAccountArgs{
			IdentityID: _identity.ID,
			Country:    input.Country,
		})
		if err != nil {
			return err
		}

		identity = _identity
		return nil
	})
	if err != nil {
		switch err.(type) {
		case *_identity.ErrInvalidArgument:
		case *accounts.ErrInvalidArgument:
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

	var identity *_identity.Identity
	err = crdbsqlx.ExecuteTx(ctx, r.Db, nil, func(tx *sqlx.Tx) error {
		_identity, err := r.IdentityService.Get(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		identity = _identity
		return nil
	})
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
