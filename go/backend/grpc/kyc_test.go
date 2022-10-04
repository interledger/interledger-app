package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	pb "gitlab.com/fynbos/proto/backend/v1"
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
	user := &user.User{
		ID: uuid.NewString(),
	}
	wallet, err := c.Users().CreateNewWallet(context.Background(), user.ID, "default")
	require.NoError(t, err)

	ud := kyc.IndividualDetails{
		WalletID:    wallet.ID,
		FirstName:   "FirstName",
		LastName:    "LastName",
		CountryCode: "ZA",
		Gender:      kyc.GenderMale,
	}
	genderMale := int32(kyc.GenderMale)

	c.KYCClient.EXPECT().UpdateIndividualDetails(gomock.Any(), ud)

	// Create KYC data
	_, err = client.UpdateIndividualKYC(user_mock.ActingAsContext(t, context.Background(), user), &pb.UpdateIndividualKYCRequest{
		FirstName:   &ud.FirstName,
		LastName:    &ud.LastName,
		CountryCode: &ud.CountryCode,
		Gender:      &genderMale,
	})
	require.NoError(t, err)

	// Update KYC
	ud = kyc.IndividualDetails{
		WalletID:    wallet.ID,
		FirstName:   "FirstName1",
		LastName:    "LastName2",
		CountryCode: "US",
		Gender:      kyc.GenderFemale,
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

	_, err = client.UpdateIndividualKYC(user_mock.ActingAsContext(t, context.Background(), user), &pb.UpdateIndividualKYCRequest{
		FirstName:   &ud.FirstName,
		LastName:    &ud.LastName,
		CountryCode: &ud.CountryCode,
		Gender:      &genderFemale,
		DateOfBirth: timestamppb.New(ud.DateOfBirth),
		Address: &pb.UpdateIndividualKYCRequest_Address{
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
}
