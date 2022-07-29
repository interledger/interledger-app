package ops

//go:generate mockgen -destination=./mock.go -package=onboarding -source=./service.go

import (
	"context"
	"database/sql"
	"fmt"

	"gitlab.com/fynbos/backend/onboarding"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/providers/unit"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

// Fetch a users onboarding data
func GetOnboarding(
	ctx context.Context,
	b Backends,
	args *onboarding.GetOnboardingArgs,
) (*onboarding.Onboarding, error) {
	if err := b.Validator().Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", onboarding.ErrInvalidArgument, err)
	}

	var ob onboarding.Onboarding
	err := b.DB().GetContext(ctx, &ob, "SELECT * FROM  onboarding WHERE id = $1 LIMIT 1;", args.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, onboarding.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", onboarding.ErrInternal, err.Error())
	}

	return &ob, err
}

func UpdateOnboarding(
	ctx context.Context,
	b Backends,
	args *onboarding.UpdateOnboardingArgs,
) (*onboarding.Onboarding, error) {
	if err := b.Validator().Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", onboarding.ErrInvalidArgument, err)
	}

	var ob onboarding.Onboarding
	if args.Id == "" {
		err := b.DB().GetContext(ctx, &ob,
			`INSERT INTO onboarding (first_name,last_name,country_of_residence,email,phone,phone_verified,service_agreement) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING *;`,
			args.FirstName, args.LastName, args.Country, args.Email, args.Phone, args.PhoneVerified, args.ServiceAgreement,
		)
		if err != nil {
			return nil, err
		}
	} else {
		currentOnboarding, err := GetOnboarding(ctx, b, &onboarding.GetOnboardingArgs{
			Id: args.Id,
		})
		if err != nil {
			return nil, err
		}
		// Manually replace currentOnboarding vals with args if arg not empty
		if args.Country == "" {
			args.Country = currentOnboarding.Country
		}
		if args.FirstName == "" {
			args.FirstName = currentOnboarding.FirstName
		}
		if args.LastName == "" {
			args.LastName = currentOnboarding.LastName
		}
		if args.Email == "" {
			args.Email = currentOnboarding.Email
		}
		if args.Phone == "" {
			args.Phone = currentOnboarding.Phone
		}
		isPhoneVerified := currentOnboarding.PhoneVerified
		if !currentOnboarding.PhoneVerified && args.PhoneVerified != isPhoneVerified {
			isPhoneVerified = args.PhoneVerified
		}
		if args.ServiceAgreement != currentOnboarding.ServiceAgreement {
			args.ServiceAgreement = currentOnboarding.ServiceAgreement
		}
		err = b.DB().GetContext(ctx, &ob,
			`UPDATE onboarding SET (first_name,last_name,country_of_residence,email,phone,phone_verified,service_agreement) = ($2,$3,$4,$5,$6,$7,$8) WHERE id = $1 RETURNING *;`,
			args.Id, args.FirstName, args.LastName, args.Country, args.Email, args.Phone, isPhoneVerified, args.ServiceAgreement,
		)
		if err != nil {
			return nil, err
		}
	}

	return &ob, nil
}

// Creating an account with Fynbos means that your identity is stored in our system and you get
// a Fynbos account. The account is not yet backed by any provider and as such the user won't be
// able to do anything with it.
func CreateAccount(
	ctx context.Context,
	b Backends,
	args *onboarding.CreateAccountArgs,
) (*accounts.Account, error) {
	if err := b.Validator().Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", onboarding.ErrInvalidArgument, err)
	}

	id, err := b.Identity().Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", onboarding.ErrInternal, err)
	}
	account, err := b.Accounts().Create(ctx, &accounts.CreateAccountArgs{
		IdentityID:                 id.ID,
		CreditsMustNotExceedDebits: true,
		Provider:                   "unit",
		ProviderID:                 uuid.NewString(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", onboarding.ErrInternal, err)
	}
	if err != nil {
		return nil, err
	}

	return account, err
}

func InitiateUnitCustomerOnboarding(ctx context.Context, b Backends, args *onboarding.InitiateUnitCustomerOnboardingArgs) error {
	if err := b.Validator().Struct(args); err != nil {
		return fmt.Errorf("%w %s", onboarding.ErrInvalidArgument, err)
	}

	// TODO: store these args in vault and just pass the key to the workflow.

	_, err := b.Temporal().ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:                    "unit_onboarding_" + args.IdentityID,
			TaskQueue:             "backend",
			WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		},
		unit.UnitOnboardCustomerWorkflow, unit.UnitOnboardCustomerState{
			CustomerID: "",
			Type:       "",
			IdentityID: args.IdentityID,
			AccountID:  "",
			ApplicationArgs: unit.CreateApplicationArgs{
				Ssn:                args.Ssn,
				DateOfBirth:        args.DateOfBirth,
				Street:             args.Street,
				Street2:            args.Street2,
				City:               args.City,
				State:              args.State,
				PostalCode:         args.PostalCode,
				IpAddress:          args.IpAddress,
				UserID:             args.IdentityID,
				DeviceFingerprints: args.DeviceFingerprints,
			},
		})
	if err != nil {
		return fmt.Errorf("%w %s", onboarding.ErrInternal, err)
	}

	return nil
}
