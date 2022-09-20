package ops

import (
	"context"
	"database/sql"
	"fmt"

	"gitlab.com/fynbos/backend/onboarding"
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
