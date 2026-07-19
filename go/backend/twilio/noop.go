package twilio

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// NoOp is a stand-in Twilio Service used when Twilio is disabled
// (twilio.enabled=false). It never talks to Twilio: verification codes are
// always reported as sent, and any submitted code is treated as approved.
//
// This mirrors exactly how the real service behaved when its old Enabled flag
// was false. Disabling Twilio is only permitted outside environment.mode=prod
// (enforced in config.validateStart and in the Helm chart), so this NoOp only
// ever runs in non-production environments.
type NoOp struct {
	validator *validator.Validate
}

// NewNoOp returns a Twilio Service that performs no real verification.
func NewNoOp() Service {
	return &NoOp{validator: validator.New()}
}

func (n *NoOp) SendVerificationCode(_ context.Context, phoneNumber string) (*Verification, error) {
	if err := n.validator.Var(phoneNumber, "required,e164"); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidArgument, err)
	}

	return &Verification{
		Sid:         "1234",
		PhoneNumber: phoneNumber,
		Status:      "pending",
	}, nil
}

func (n *NoOp) CheckVerificationCode(_ context.Context, args *CheckVerificationCodeArgs) (*Verification, error) {
	return &Verification{
		Sid:         "1234",
		PhoneNumber: args.PhoneNumber,
		Status:      statusApproved,
	}, nil
}

func (n *NoOp) ListSuccessfulVerificationAttempts(_ context.Context, args ListSuccessfulVerificationAttemptsArgs) ([]Verification, error) {
	return []Verification{
		{
			Sid:         "1234",
			PhoneNumber: args.To,
			Status:      statusApproved,
			UpdatedAt:   time.Now(),
		},
	}, nil
}
