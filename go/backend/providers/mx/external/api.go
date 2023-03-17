package external

import "context"

type Client interface {
	CreateUser(ctx context.Context, id string) (*User, error)
	ListUsersByID(ctx context.Context, id string) (*ListUsersResponse, error)
	GetWidgetURL(ctx context.Context, args GetWidgetURLArgs) (*WidgetURL, error)
}
