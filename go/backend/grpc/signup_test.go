package grpc

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/signup"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func TestSetSignupUserData(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	sID := uuid.NewString()
	c.SignupService.EXPECT().SetUserData(gomock.Any(), signup.UserDataArgs{
		FirstName:   "FirstName",
		LastName:    "LastName",
		Email:       "test@interledger.test",
		CountryCode: "ZA",
	}).Return(sID, nil).Times(1)

	resp, err := client.SetSignupUserData(context.Background(), &pb.SetSignupUserDataRequest{
		FirstName:   "FirstName",
		LastName:    "LastName",
		Email:       "test@interledger.test",
		CountryCode: "ZA",
	})

	require.NoError(t, err)
	assert.Equal(t, sID, resp.Id)
}

func TestRpcService_SetSignupMobileNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	sID := uuid.NewString()
	mobile := faker.E164PhoneNumber()

	c.SignupService.EXPECT().SetMobileNumber(gomock.Any(), signup.MobileNumberArgs{
		ID:           sID,
		MobileNumber: mobile,
		OTP:          "123456",
	}).Return(nil).Times(1)

	_, err := client.SetSignupMobileNumber(context.Background(), &pb.SetSignupMobileNumberRequest{
		Id:     sID,
		Mobile: mobile,
		Otp:    "123456",
	})
	require.NoError(t, err)
}

func TestGetSignup(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	sID := uuid.NewString()
	mobile := faker.E164PhoneNumber()
	c.SignupService.EXPECT().Get(gomock.Any(), sID).Return(&signup.Signup{
		ID:           sID,
		UserID:       "123",
		FirstName:    "FirstName",
		LastName:     "LastName",
		CountryCode:  "ZA",
		Email:        "test@interledger.test",
		MobileNumber: mobile,
		Completed:    true,
	}, nil)

	resp, err := client.GetSignup(context.Background(), &pb.GetSignupRequest{
		Id: sID,
	})

	require.NoError(t, err)
	assert.Equal(t, sID, resp.Id)
	assert.Equal(t, "123", resp.UserId)
	assert.Equal(t, "FirstName", resp.FirstName)
	assert.Equal(t, "LastName", resp.LastName)
	assert.Equal(t, "ZA", resp.CountryCode)
	assert.Equal(t, "test@interledger.test", resp.Email)
	assert.Equal(t, mobile, resp.MobileNumber)
	assert.True(t, resp.Completed)
}
