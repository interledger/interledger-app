package external

import "context"

type Client interface {
	RegisterUser(ctx context.Context, user User) (*User, error)
}
