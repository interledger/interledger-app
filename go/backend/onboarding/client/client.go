package client

import (
	"context"

	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/onboarding/ops"
)

var _ onboarding.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) onboarding.Client {
	return &client{
		b: b,
	}
}

func (c client) GetOnboarding(ctx context.Context, args *onboarding.GetOnboardingArgs) (*onboarding.Onboarding, error) {
	return ops.GetOnboarding(ctx, c.b, args)
}

func (c client) UpdateOnboarding(ctx context.Context, args *onboarding.UpdateOnboardingArgs) (*onboarding.Onboarding, error) {
	return ops.UpdateOnboarding(ctx, c.b, args)
}
