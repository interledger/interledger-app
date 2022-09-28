package external

import "context"

type Client interface {
	RegisterUser(ctx context.Context, user User) (*User, error)
	UpdateUser(ctx context.Context, id string, newValues User) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
}
