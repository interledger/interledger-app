package waitlist

import "context"

type Client interface {
	Add(ctx context.Context, email, countryCode, fullName string) error
	CanSignup(ctx context.Context, id string) (bool, error)
	SetSignupComplete(ctx context.Context, id, userSignupId string) error
}
