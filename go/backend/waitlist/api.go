package waitlist

import "context"

type Client interface {
	Add(ctx context.Context, email, countryCode string) error
}
