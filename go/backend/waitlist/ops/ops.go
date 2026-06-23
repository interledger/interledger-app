package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/interledger/interledger-app/go/backend/waitlist"
)

func AddSignup(ctx context.Context, b Backends, email, country, fullName, mugID string, betaOptIn bool) error {
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

	var mug sql.NullString
	if mugID != "" && mugIDs[mugID] {
		available, err := IsMugAvailable(ctx, b, mugID)
		if err != nil {
			return err
		}

		mug = sql.NullString{
			String: mugID,
			Valid:  available,
		}
	}

	_, err = b.DB().ExecContext(ctx,
		"INSERT INTO waitlist_signups (email, country_code, full_name, beta_opt_in, mug_id) VALUES ($1, $2, $3, $4, $5) "+
			"ON CONFLICT (email, country_code) DO UPDATE SET mug_id = excluded.mug_id WHERE waitlist_signups.mug_id IS NULL AND waitlist_signups.email=$1 AND waitlist_signups.country_code=$2",
		email, country, fullName, betaOptIn, mug)
	if err != nil {
		return fmt.Errorf("%w %s", waitlist.ErrInternal, err.Error())
	}

	return nil
}

func IsMugAvailable(ctx context.Context, b Backends, mugID string) (bool, error) {
	// Does this even exists
	if !mugIDs[mugID] {
		return false, nil
	}

	var id string
	err := b.DB().GetContext(ctx, &id, "SELECT id FROM waitlist_signups WHERE mug_id=$1", mugID)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("%w %s", waitlist.ErrInternal, err)
	}

	return false, nil
}

type dbSignup struct {
	ID        string         `db:"id"`
	CanSignup bool           `db:"can_signup"`
	UserID    sql.NullString `db:"user_id"`
}

func GetIdByEmailAndCountryCode(ctx context.Context, b Backends, email, countryCode string) (string, error) {
	var record dbSignup

	err := b.DB().GetContext(ctx, &record,
		"SELECT id, can_signup, user_id from waitlist_signups where email = $1 and country_code = $2", email, countryCode)

	if err != nil {
		return "", fmt.Errorf("%w %s", waitlist.ErrInternal, err.Error())
	}

	return record.ID, err
}

func AllowSignupById(ctx context.Context, b Backends, id string) error {
	result, err := b.DB().ExecContext(ctx,
		"UPDATE waitlist_signups set can_signup = true where id = $1", id)
	if err != nil {
		return fmt.Errorf("%w %s", waitlist.ErrInternal, err.Error())
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w %s", waitlist.ErrInternal, err.Error())
	}
	if affected != 1 {
		return fmt.Errorf("%w incorrect number of rows updated (%d)", waitlist.ErrNotFound, affected)
	}

	return nil
}

func CanSignup(ctx context.Context, b Backends, id string) (bool, error) {
	var record dbSignup
	err := b.DB().GetContext(ctx, &record,
		"SELECT id, can_signup, user_id from waitlist_signups where id = $1 AND can_signup is true and user_id is null", id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("%w %s", waitlist.ErrInternal, err.Error())
	}

	return true, nil
}

func SetSignupComplete(ctx context.Context, b Backends, id, userID string) error {
	result, err := b.DB().ExecContext(ctx,
		"UPDATE waitlist_signups set user_id = $1 where id = $2 and (user_id is null OR user_id = $1)", userID, id)
	if err != nil {
		return fmt.Errorf("%w %s", waitlist.ErrInternal, err.Error())
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w %s", waitlist.ErrInternal, err.Error())
	}
	if affected != 1 {
		return fmt.Errorf("%w incorrect number of rows updated (%d)", waitlist.ErrNotFound, affected)
	}

	return nil
}

func ListSignups(ctx context.Context, b Backends) (signups []waitlist.Signup, err error) {
	type dbSignup struct {
		ID          string         `db:"id"`
		Name        string         `db:"full_name"`
		Email       string         `db:"email"`
		BetaOptIn   bool           `db:"beta_opt_in"`
		CanSignup   bool           `db:"can_signup"`
		MugID       sql.NullString `db:"mug_id"`
		CountryCode string         `db:"country_code"`
	}
	dbSignups := []dbSignup{}
	err = b.DB().SelectContext(ctx, &dbSignups,
		"SELECT id, full_name, email, beta_opt_in, can_signup, country_code, mug_id from waitlist_signups order by created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("%w %s", waitlist.ErrInternal, err.Error())
	}

	for _, signup := range dbSignups {
		signups = append(signups, waitlist.Signup{
			ID:          signup.ID,
			Name:        signup.Name,
			Email:       signup.Email,
			BetaOtpIn:   signup.BetaOptIn,
			CanSignup:   signup.CanSignup,
			MugID:       signup.MugID.String,
			CountryCode: signup.CountryCode,
		})
	}

	return signups, err
}
