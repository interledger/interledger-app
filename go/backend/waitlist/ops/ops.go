package ops

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/waitlist"
)

func AddSignup(ctx context.Context, b Backends, email, country, fullName string) error {
	err := b.Validator().Var(email, "required,email")
	if err != nil {
		return fmt.Errorf("%w %s", waitlist.ErrInvalidEmail, err.Error())
	}

	err = b.Validator().Var(country, "required,iso3166_1_alpha2")
	if err != nil {
		return fmt.Errorf("%w %s", waitlist.ErrInvalidCountry, err.Error())
	}

	err = b.Validator().Var(fullName, "required")
	if err != nil {
		return fmt.Errorf("%w %s", waitlist.ErrInvalidName, err.Error())
	}

	_, err = b.DB().ExecContext(ctx,
		"INSERT INTO waitlist_signups (email, country_code, full_name) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING ",
		email, country, fullName)
	if err != nil {
		return fmt.Errorf("%w %s", waitlist.ErrInternal, err.Error())
	}

	return nil
}
