package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"github.com/google/uuid"

	"gitlab.com/fynbos/backend/signup"
)

type dbSignup struct {
	signup.Signup
	MobileNumber sql.NullString `db:"mobile_number"`
	UserID       sql.NullString `db:"user_id"`
}

func Get(ctx context.Context, b Backends, id string) (*signup.Signup, error) {
	var s dbSignup
	err := b.DB().GetContext(ctx, &s, "SELECT * FROM  signups WHERE id = $1 LIMIT 1;", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, signup.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", signup.ErrInternal, err.Error())
	}

	resp := &signup.Signup{
		ID:           s.ID,
		UserID:       s.UserID.String,
		FirstName:    s.FirstName,
		LastName:     s.LastName,
		CountryCode:  s.CountryCode,
		Email:        s.Email,
		MobileNumber: s.MobileNumber.String,
		Completed:    s.UserID.Valid,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}

	return resp, nil
}

func GetForUser(ctx context.Context, b Backends, userID string) (*signup.Signup, error) {
	var s dbSignup
	err := b.DB().GetContext(ctx, &s, "SELECT * FROM  signups WHERE user_id = $1 LIMIT 1;", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, signup.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", signup.ErrInternal, err.Error())
	}

	resp := &signup.Signup{
		ID:           s.ID,
		UserID:       s.UserID.String,
		FirstName:    s.FirstName,
		LastName:     s.LastName,
		CountryCode:  s.CountryCode,
		Email:        s.Email,
		MobileNumber: s.MobileNumber.String,
		Completed:    s.UserID.Valid,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}

	return resp, nil
}

func SetUserData(ctx context.Context, b Backends, args signup.UserDataArgs) (string, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", signup.ErrInvalidArgument, err)
	}

	id := args.ID
	if id == "" {
		id = uuid.NewString()
	}

	existing, err := Get(ctx, b, id)
	if errors.Is(err, signup.ErrNotFound) {
	} else if err != nil {
		return "", err
	}

	// trim extra spaces from the name
	args.FirstName = strings.Join(strings.Fields(args.FirstName), " ")
	args.LastName = strings.Join(strings.Fields(args.LastName), " ")

	var r sql.Result
	if existing == nil {
		r, err = b.DB().ExecContext(ctx, "INSERT INTO signups (id, first_name, last_name, country_code, email) "+
			"VALUES ($1, $2, $3, $4, $5)", id, args.FirstName, args.LastName, args.CountryCode, args.Email)
	} else {
		// Update an existing signup
		r, err = b.DB().ExecContext(ctx, "UPDATE signups SET first_name=$1, last_name=$2, country_code=$3, email=$4, updated_at=now() WHERE id=$5 ",
			args.FirstName, args.LastName, args.CountryCode, args.Email, args.ID)
	}
	if err != nil {
		return "", fmt.Errorf("%w %s", signup.ErrInternal, err)
	}

	affected, err := r.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("%w %s", signup.ErrInternal, err)
	}
	if affected != 1 {
		return "", fmt.Errorf("%w incorrect number of rows updated (%d)", signup.ErrInternal, affected)
	}

	return id, nil
}

func SetMobileNumber(ctx context.Context, b Backends, args signup.MobileNumberArgs) error {
	err := b.Validator().Struct(args)
	if err != nil {
		return fmt.Errorf("%w %s", signup.ErrInvalidArgument, err)
	}

	v, err := b.Twilio().CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		PhoneNumber: args.MobileNumber,
		Code:        args.OTP,
	})
	if err != nil {
		return fmt.Errorf("%w %s", signup.ErrInvalidOTP, err)
	}
	if !v.IsValid() {
		return signup.ErrInvalidOTP
	}

	// Check if phone number already used. Note must be after OTP validation to prevent data leakage!
	var existsId string
	err = b.DB().GetContext(ctx, &existsId, "select id from signups where mobile_number=$1 and user_id is not null", args.MobileNumber)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w %s", signup.ErrInternal, err)
	}
	if existsId != "" {
		return fmt.Errorf("%w %s", signup.ErrDuplicatePhone, err)
	}

	r, err := b.DB().ExecContext(ctx, "UPDATE signups SET mobile_number=$1, updated_at=now() WHERE id=$2",
		args.MobileNumber, args.ID)
	if err != nil {
		return fmt.Errorf("%w %s", signup.ErrInternal, err)
	}

	affected, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w %s", signup.ErrInternal, err)
	}
	if affected != 1 {
		return fmt.Errorf("%w incorrect number of rows updated (%d)", signup.ErrInternal, affected)
	}

	return nil
}

func Complete(ctx context.Context, b Backends, id, userID string) error {
	err := b.Validator().Var(userID, "required,uuid")
	if err != nil {
		return fmt.Errorf("%w %s", signup.ErrInvalidArgument, err)
	}

	current, err := Get(ctx, b, id)
	if err != nil {
		return err
	}

	if current.UserID != "" && current.UserID != userID {
		return fmt.Errorf("%w tried to complete an already complete signup", signup.ErrInternal)
	}

	if current.MobileNumber == "" {
		log.Warn("signup did not have a mobile number, perhaps otp is disabled.", zap.String("userID", userID))
	}

	_, err = b.DB().ExecContext(ctx, "UPDATE signups SET user_id=$1, updated_at=now() WHERE id=$2 and user_id is null",
		userID, id)
	if err != nil {
		return fmt.Errorf("%w %s", signup.ErrInternal, err)
	}

	slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "Fynbot", fmt.Sprintf(":baby: New Sign Up\nID: %s\nUser ID: %s\nFull name: %s\nCountry: %s", current.ID, userID, current.FirstName+" "+current.LastName, current.CountryCode))

	return nil
}
