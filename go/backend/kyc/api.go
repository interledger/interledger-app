package kyc

import "context"

type Client interface {
	GetUserDetails(ctx context.Context, userID string) (*UserDetails, error)
	UpdateUserDetails(ctx context.Context, args UserDetails) (*UserDetails, error)
}
