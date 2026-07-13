package twilio

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNoOp verifies the no-op service behaves exactly as the real service used
// to when its Enabled flag was false: codes are always "sent" and any submitted
// code is treated as approved, without ever calling Twilio.
func TestNoOp(t *testing.T) {
	tws := NewNoOp()

	phoneNumber := "+90555555555"

	sendRes, err := tws.SendVerificationCode(context.Background(), phoneNumber)
	assert.NoError(t, err)
	assert.Equal(t, phoneNumber, sendRes.PhoneNumber)
	assert.Equal(t, "pending", sendRes.Status)

	checkRes, err := tws.CheckVerificationCode(context.Background(), &CheckVerificationCodeArgs{
		PhoneNumber: phoneNumber,
		Code:        "123456",
	})
	assert.NoError(t, err)
	assert.Equal(t, phoneNumber, checkRes.PhoneNumber)
	assert.True(t, checkRes.IsValid())

	listRes, err := tws.ListSuccessfulVerificationAttempts(context.Background(), ListSuccessfulVerificationAttemptsArgs{
		To: phoneNumber,
	})
	assert.NoError(t, err)
	assert.Len(t, listRes, 1)
	assert.Equal(t, phoneNumber, listRes[0].PhoneNumber)
	assert.True(t, listRes[0].IsValid())
}

func TestNoOpRejectsInvalidPhoneNumber(t *testing.T) {
	tws := NewNoOp()

	_, err := tws.SendVerificationCode(context.Background(), "not-a-phone")
	assert.ErrorIs(t, err, ErrInvalidArgument)
}
