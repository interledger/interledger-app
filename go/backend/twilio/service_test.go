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
