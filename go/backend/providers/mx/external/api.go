package external

import "context"

type Client interface {
	CreateUser(ctx context.Context, id string) (*User, error)
	ListUsersByID(ctx context.Context, id string) (*ListUsersResponse, error)
	GetWidgetURL(ctx context.Context, args GetWidgetURLArgs) (*WidgetURL, error)
	ListAccountOwnersByMember(ctx context.Context, userGuid, memberGuid string) (*ListAccountOwnersResponse, error)
	ListAccountNumbersByMember(ctx context.Context, userGuid, memberGuid string) (*ListAccountNumbersResponse, error)
	ListAccountsByMember(ctx context.Context, userGuid, memberGuid string) (*ListAccountsResponse, error)
	ReadUsersAccount(ctx context.Context, userGuid, accountGuid string) (*Account, error)
	ListUsers(ctx context.Context) ([]User, error)
	DeleteUser(ctx context.Context, guid string) error
}
