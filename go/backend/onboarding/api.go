package onboarding

import (
	"context"

	"gitlab.com/fynbos/backend/accounts"
)

type Client interface {
	GetOnboarding(ctx context.Context, args *GetOnboardingArgs) (*Onboarding, error)
	UpdateOnboarding(ctx context.Context, args *UpdateOnboardingArgs) (*Onboarding, error)
	CreateAccount(ctx context.Context, args *CreateAccountArgs) (*accounts.Account, error)
	InitiateUnitCustomerOnboarding(ctx context.Context, args *InitiateUnitCustomerOnboardingArgs) error
}
