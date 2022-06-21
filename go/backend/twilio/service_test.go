package twilio

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVerification(t *testing.T) {
	mockMxServer := NewMockServer()
	tws, err := NewService(&ServiceArgs{
		AccountSid:   "AC24566cec88f17f0607c23ba394a8de5c",
		AccountToken: "51ad2c0b2eaf88b10261ad26a54737c9",
		ServiceSid:   "VAfed340e6a933e63f95f3ab6058d7805b",
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
		AccountSid:   "AC24566cec88f17f0607c23ba394a8de5c",
		AccountToken: "51ad2c0b2eaf88b10261ad26a54737c9",
		ServiceSid:   "VAfed340e6a933e63f95f3ab6058d7805b",
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
