package signup

import (
	"context"
)

type Client interface {
	Get(ctx context.Context, id string) (*Signup, error)
	SetUserData(ctx context.Context, args UserDataArgs) (string, error)
	SetMobileNumber(ctx context.Context, args MobileNumberArgs) error
	Complete(ctx context.Context, id, userID string) error
	GetForUser(ctx context.Context, userID string) (*Signup, error)
}
