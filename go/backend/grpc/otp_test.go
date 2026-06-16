package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/errcodes"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/user"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRpcService_ConfirmUserPhone_InvalidOTPUsesCanonicalTwilioError(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	svc := &rpcService{b: c}

	usr := &user.User{ID: "user-1", PhoneNumber: "+27830000000"}
	ctx := context.WithValue(context.Background(), user.CtxKey, usr)

	c.TwilioService.EXPECT().CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		PhoneNumber: usr.PhoneNumber,
		Code:        "123456",
	}).Return(nil, twilio.ErrInvalidOTP).Times(1)

	_, err := svc.ConfirmUserPhone(ctx, &pb.ConfirmUserPhoneRequest{Otp: "123456"})
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

func TestRpcService_SendPhoneVerification_InvalidPhoneUsesCanonicalPhoneField(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	svc := &rpcService{b: c}

	_, err := svc.SendPhoneVerification(context.Background(), &pb.SendPhoneVerificationRequest{To: "invalid"})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())

	br := statusFindDetail[*errdetails.BadRequest](st)
	require.NotNil(t, br)
	require.Len(t, br.FieldViolations, 1)
	assert.Equal(t, "phone", br.FieldViolations[0].Field)

	appErr := statusFindDetail[*pb.AppError](st)
	require.NotNil(t, appErr)
	require.Len(t, appErr.Fields, 1)
	assert.Equal(t, "phone", appErr.Fields[0].Field)
}

func TestRpcService_CheckPhoneVerification_InvalidOTPUsesCanonicalTwilioError(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	svc := &rpcService{b: c}

	ctx := context.Background()
	phoneNumber := "+27830000000"

	c.TwilioService.EXPECT().CheckVerificationCode(ctx, &twilio.CheckVerificationCodeArgs{
		PhoneNumber: phoneNumber,
		Code:        "123456",
	}).Return(nil, twilio.ErrInvalidOTP).Times(1)

	_, err := svc.CheckPhoneVerification(ctx, &pb.CheckPhoneVerificationRequest{To: phoneNumber, Otp: "123456"})
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

func TestRpcService_CheckPhoneVerification_InvalidPhoneUsesCanonicalPhoneField(t *testing.T) {
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	svc := &rpcService{b: c}

	_, err := svc.CheckPhoneVerification(context.Background(), &pb.CheckPhoneVerificationRequest{To: "invalid", Otp: "123456"})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())

	br := statusFindDetail[*errdetails.BadRequest](st)
	require.NotNil(t, br)
	require.Len(t, br.FieldViolations, 1)
	assert.Equal(t, "phone", br.FieldViolations[0].Field)

	appErr := statusFindDetail[*pb.AppError](st)
	require.NotNil(t, appErr)
	require.Len(t, appErr.Fields, 1)
	assert.Equal(t, "phone", appErr.Fields[0].Field)
}
