package twilio

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVerification(t *testing.T) {
	mockMxServer := NewMockServer()
	tws, err := NewService(&ServiceArgs{
		AccountSid:   "testAccountSid",
		AccountToken: "testAccountToken",
		ServiceSid:   "testServiceSid",
		ApiBaseUrl:   mockMxServer.URL,
		Enabled:      true,
	})
	if err != nil {
		t.Fatal(err)
	}

	phoneNumber := "+90555555555"
	res, err := tws.SendVerificationCode(context.Background(), phoneNumber)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, phoneNumber, res.PhoneNumber)
	assert.Equal(t, "pending", res.Status)
}

func TestCheckVerification(t *testing.T) {
	mockMxServer := NewMockServer()
	tws, err := NewService(&ServiceArgs{
		AccountSid:   "testAccountSid",
		AccountToken: "testAccountToken",
		ServiceSid:   "testServiceSid",
		ApiBaseUrl:   mockMxServer.URL,
		Enabled:      true,
	})
	if err != nil {
		t.Fatal(err)
	}

	phoneNumber := "+90555555555"
	res, err := tws.CheckVerificationCode(context.Background(), &CheckVerificationCodeArgs{
		PhoneNumber: phoneNumber,
		Code:        "123456",
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, phoneNumber, res.PhoneNumber)
	assert.Equal(t, "approved", res.Status)
}

func TestServiceDisabled(t *testing.T) {
	tws, err := NewService(&ServiceArgs{
		Enabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}

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
