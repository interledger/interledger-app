package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	pb "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestUpdateUserKYC(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	wallet := wallets.Wallet{
		ID:   uuid.NewString(),
		Name: "testing",
	}
	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()

	ud := kyc.IndividualDetails{
		WalletID:    wallet.ID,
		FirstName:   "FirstName",
		LastName:    "LastName",
		CountryCode: "ZA",
		Gender:      kyc.GenderMale,
		IPAddress:   "41.71.7.130",
	}
	genderMale := int32(kyc.GenderMale)

	c.KYCClient.EXPECT().UpdateIndividualDetails(gomock.Any(), ud)

	// Create KYC data
	_, err := client.UpdateIndividualKYC(user_mock.ActingAsContext(t, context.Background(), u), &pb.UpdateIndividualKYCRequest{
		FirstName:   &ud.FirstName,
		LastName:    &ud.LastName,
		CountryCode: &ud.CountryCode,
		Gender:      &genderMale,
		IpAddress:   ud.IPAddress,
	})
	require.NoError(t, err)

	// Update KYC
	ud = kyc.IndividualDetails{
		WalletID:    wallet.ID,
		FirstName:   "FirstName1",
		LastName:    "LastName2",
		CountryCode: "US",
		Gender:      kyc.GenderFemale,
		IPAddress:   "41.71.7.130",
		DateOfBirth: time.Date(2000, time.April, 4, 0, 0, 0, 0, time.UTC),
		Address: &kyc.Address{
			Line1:       "Line1",
			Line2:       "Line2",
			Building:    "Building",
			Apartment:   "Apartment",
			City:        "City",
			State:       "US-CA",
			ZipCode:     "ZipCode",
			CountryCode: "US",
		},
	}
	genderFemale := int32(kyc.GenderFemale)

	c.KYCClient.EXPECT().UpdateIndividualDetails(gomock.Any(), ud)

	_, err = client.UpdateIndividualKYC(user_mock.ActingAsContext(t, context.Background(), u), &pb.UpdateIndividualKYCRequest{
		FirstName:   &ud.FirstName,
		LastName:    &ud.LastName,
		CountryCode: &ud.CountryCode,
		Gender:      &genderFemale,
		IpAddress:   "41.71.7.130",
		DateOfBirth: timestamppb.New(ud.DateOfBirth),
		Address: &pb.Address{
			Line1:       &ud.Address.Line1,
			Line2:       &ud.Address.Line2,
			Building:    &ud.Address.Building,
			Apartment:   &ud.Address.Apartment,
			City:        &ud.Address.City,
			State:       &ud.Address.State,
			ZipCode:     &ud.Address.ZipCode,
			CountryCode: &ud.Address.CountryCode,
		},
	})
	require.NoError(t, err)

	// validation errors
	badState := "CA"
	badCountryCode := "FF"
	_, err = client.UpdateIndividualKYC(user_mock.ActingAsContext(t, context.Background(), u), &pb.UpdateIndividualKYCRequest{
		FirstName: &ud.FirstName,
		LastName:  &ud.LastName,
		// CountryCode: &badCountryCode,
		Gender:      &genderFemale,
		DateOfBirth: timestamppb.New(ud.DateOfBirth),
		Address: &pb.Address{
			Line1:       &ud.Address.Line1,
			Line2:       &ud.Address.Line2,
			Building:    &ud.Address.Building,
			Apartment:   &ud.Address.Apartment,
			City:        &ud.Address.City,
			State:       &badState,
			ZipCode:     &ud.Address.ZipCode,
			CountryCode: &badCountryCode,
		},
	})
	require.Error(t, err)
	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	errorFields := []string{}

	badRequest := statusFindDetail[*errdetails.BadRequest](grpcStatus)
	require.NotNil(t, badRequest)

	for _, violation := range badRequest.FieldViolations {
		errorFields = append(errorFields, violation.Field)
	}

	assert.EqualValues(t, errorFields, []string{"IpAddress", "AddressCountryCode", "AddressState"})
}

func TestGetIndividualKYC(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	wallet := wallets.Wallet{
		ID:   uuid.NewString(),
		Name: "testing",
	}
	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()

	ud := kyc.IndividualDetails{
		WalletID:    wallet.ID,
		FirstName:   "FirstName1",
		LastName:    "LastName2",
		CountryCode: "US",
		Gender:      kyc.GenderFemale,
		IPAddress:   "41.71.7.130",
		DateOfBirth: time.Date(2000, time.April, 4, 0, 0, 0, 0, time.UTC),
		Address: &kyc.Address{
			Line1:       "Line1",
			Line2:       "Line2",
			Building:    "Building",
			Apartment:   "Apartment",
			City:        "City",
			State:       "US-CA",
			ZipCode:     "ZipCode",
			CountryCode: "US",
		},
	}

	c.KYCClient.EXPECT().GetIndividualDetails(gomock.Any(), wallet.ID).Return(&ud, nil)

	details, err := client.GetIndividualKYC(user_mock.ActingAsContext(t, context.Background(), u), &pb.Empty{})
	require.NoError(t, err)

	require.Equal(t, ud.FirstName, details.FirstName)
	require.Equal(t, ud.LastName, details.LastName)
	require.Equal(t, ud.CountryCode, details.CountryCode)
	require.Equal(t, int32(ud.Gender), details.Gender)
	require.True(t, details.DateOfBirth.AsTime().Equal(ud.DateOfBirth))
}

func TestKYCStatus(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	wallet := wallets.Wallet{
		ID:   uuid.NewString(),
		Name: "testing",
	}
	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()

	c.KYCClient.EXPECT().GetKYCStatus(gomock.Any(), wallet.ID).Return(kyc.StatusApproved, nil)

	status, err := client.KYCStatus(user_mock.ActingAsContext(t, context.Background(), u), &pb.Empty{})
	require.NoError(t, err)

	require.Equal(t, kyc.StatusApproved.ToInt32(), status.KycStatus)
}

func TestSetKYCStatusPending_AllowsDocumentsRequired(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	wallet := wallets.Wallet{ID: uuid.NewString(), Name: "testing"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()
	c.KYCClient.EXPECT().GetKYCStatus(gomock.Any(), wallet.ID).Return(kyc.StatusDocumentsRequired, nil)
	c.KYCClient.EXPECT().SetKYCStatus(gomock.Any(), wallet.ID, kyc.StatusPending)

	_, err := client.SetKYCStatusPending(user_mock.ActingAsContext(t, context.Background(), u), &pb.Empty{})
	require.NoError(t, err)
}

func TestSetKYCStatusPending_DoesNothingForApproved(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	wallet := wallets.Wallet{ID: uuid.NewString(), Name: "testing"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()
	c.KYCClient.EXPECT().GetKYCStatus(gomock.Any(), wallet.ID).Return(kyc.StatusApproved, nil)

	_, err := client.SetKYCStatusPending(user_mock.ActingAsContext(t, context.Background(), u), &pb.Empty{})
	require.NoError(t, err)
}

func TestSetKYCStatusPending_DoesNothingForPending(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	wallet := wallets.Wallet{ID: uuid.NewString(), Name: "testing"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()
	c.KYCClient.EXPECT().GetKYCStatus(gomock.Any(), wallet.ID).Return(kyc.StatusPending, nil)

	_, err := client.SetKYCStatusPending(user_mock.ActingAsContext(t, context.Background(), u), &pb.Empty{})
	require.NoError(t, err)
}

func TestGetKYCProviderWidget_BlockedForApprovedStatus(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	wallet := wallets.Wallet{ID: uuid.NewString(), Name: "testing", Country: "ZA"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()
	c.KYCClient.EXPECT().GetKYCStatus(gomock.Any(), wallet.ID).Return(kyc.StatusApproved, nil)

	_, err := client.GetKYCProviderWidget(user_mock.ActingAsContext(t, context.Background(), u), &pb.GetKYCProviderWidgetRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGetKYCProviderWidget_AllowsDocumentsRequired(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	u := &user.User{ID: uuid.NewString()}
	wallet := wallets.Wallet{ID: uuid.NewString(), Name: "testing", Country: "ZA"}

	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()
	c.KYCClient.EXPECT().GetKYCStatus(gomock.Any(), wallet.ID).Return(kyc.StatusDocumentsRequired, nil)
	c.KYCClient.EXPECT().GetPersonaInquiry(gomock.Any(), wallet.ID, "").Return(&kyc.PersonaInquiry{ID: "inq-1"}, nil)

	resp, err := client.GetKYCProviderWidget(user_mock.ActingAsContext(t, context.Background(), u), &pb.GetKYCProviderWidgetRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.PersonaInquiry)
	require.Equal(t, "inq-1", resp.PersonaInquiry.Id)
}
