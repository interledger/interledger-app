package grpc

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/agreements"
	"gitlab.com/fynbos/backend/errcodes"
	"gitlab.com/fynbos/backend/signup"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func TestRpcService_SetSignupMobileNumber_InvalidOTPUsesCanonicalTwilioError(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	sID := uuid.NewString()
	mobile := faker.E164PhoneNumber()

	c.SignupService.EXPECT().SetMobileNumber(gomock.Any(), signup.MobileNumberArgs{
		ID:           sID,
		MobileNumber: mobile,
		OTP:          "123456",
	}).Return(signup.ErrInvalidOTP).Times(1)

	_, err := client.SetSignupMobileNumber(context.Background(), &pb.SetSignupMobileNumberRequest{
		Id:     sID,
		Mobile: mobile,
		Otp:    "123456",
	})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())

	br := statusFindDetail[*errdetails.BadRequest](st)
	require.NotNil(t, br)
	require.Len(t, br.FieldViolations, 1)
	assert.Equal(t, "otp", br.FieldViolations[0].Field)

	errInfo := statusFindDetail[*errdetails.ErrorInfo](st)
	require.NotNil(t, errInfo)
	assert.Equal(t, "TwilioError", errInfo.Reason)

	appErr := statusFindDetail[*pb.AppError](st)
	require.NotNil(t, appErr)
	assert.Equal(t, errcodes.ErrCodeTwilioInvalidOTP, appErr.ErrorCode)
	require.Len(t, appErr.Fields, 1)
	assert.Equal(t, "otp", appErr.Fields[0].Field)
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

func TestCompleteSignup_NoAgreementSigning(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	sID := uuid.NewString()
	userID := uuid.NewString()
	t.Setenv("SIGNUP_AGREEMENT_IDS", "")

	c.SignupService.EXPECT().Complete(gomock.Any(), sID, userID).Return(nil).Times(1)
	// Agreements().Sign must not be called when SIGNUP_AGREEMENT_IDS is unset

	_, err := client.CompleteSignup(context.Background(), &pb.CompleteSignupRequest{
		Id:     sID,
		UserId: userID,
	})
	require.NoError(t, err)
}

func TestCompleteSignup_WithAgreementSigning(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	sID := uuid.NewString()
	userID := uuid.NewString()
	agreementIDs := "privacy_policy-0.0.0,terms_of_service-0.0.0"
	t.Setenv("SIGNUP_AGREEMENT_IDS", agreementIDs)

	c.SignupService.EXPECT().Complete(gomock.Any(), sID, userID).Return(nil).Times(1)
	c.AgreementsService.EXPECT().Sign(gomock.Any(), &agreements.SignArgs{
		AgreementIDs: []string{"privacy_policy-0.0.0", "terms_of_service-0.0.0"},
		UserID:       userID,
	}).Return(nil).Times(1)

	_, err := client.CompleteSignup(context.Background(), &pb.CompleteSignupRequest{
		Id:     sID,
		UserId: userID,
	})
	require.NoError(t, err)
}

func TestCompleteSignup_SignFailsStillSucceeds(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	sID := uuid.NewString()
	userID := uuid.NewString()
	t.Setenv("SIGNUP_AGREEMENT_IDS", "privacy_policy-0.0.0")

	c.SignupService.EXPECT().Complete(gomock.Any(), sID, userID).Return(nil).Times(1)
	c.AgreementsService.EXPECT().Sign(gomock.Any(), gomock.Any()).Return(agreements.ErrNotFound).Times(1)

	_, err := client.CompleteSignup(context.Background(), &pb.CompleteSignupRequest{
		Id:     sID,
		UserId: userID,
	})
	require.NoError(t, err)
}
