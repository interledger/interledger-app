package onboarding

import (
	"context"
)

type Client interface {
	GetOnboarding(ctx context.Context, args *GetOnboardingArgs) (*Onboarding, error)
	UpdateOnboarding(ctx context.Context, args *UpdateOnboardingArgs) (*Onboarding, error)
}
