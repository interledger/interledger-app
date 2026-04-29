package ops

import (
	"context"
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"gitlab.com/fynbos/backend/user"
)

func ValidateTotpAgainstURL(totpURL, code string, now time.Time) error {
	otpKey, err := otp.NewKeyFromURL(totpURL)
	if err != nil {
		return user.ErrInvalidTotpConfig
	}

	valid, err := totp.ValidateCustom(code, otpKey.Secret(), now, totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otpKey.Digits(),
		Algorithm: otpKey.Algorithm(),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", user.ErrInternal, err)
	}
	if !valid {
		return user.ErrInvalidTotpCode
	}
	return nil
}

func ValidateTotpCode(ctx context.Context, b Backends, userID, code string, now time.Time) error {
	totpURL, err := GetTotpURL(ctx, b, userID)
	if err != nil {
		return err
	}
	if totpURL == "" {
		return user.ErrTotpNotConfigured
	}
	return ValidateTotpAgainstURL(totpURL, code, now)
}
