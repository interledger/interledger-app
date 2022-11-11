package waitlist

import "context"

type Client interface {
	Add(ctx context.Context, email, countryCode, fullName, mugID string, betaOptIn bool) error
	CanSignup(ctx context.Context, id string) (bool, error)
	SetSignupComplete(ctx context.Context, id, userSignupId string) error
	ListSignups(ctx context.Context) ([]Signup, error)
	IsMugAvailable(ctx context.Context, mugID string) (bool, error)
	AllowSignupById(ctx context.Context, id string) error
}
