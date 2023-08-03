package ops_test

import (
	"context"
	"testing"

	email_client "gitlab.com/fynbos/backend/email/client/mock"
	notify_client "gitlab.com/fynbos/backend/notify/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	wallet_client "gitlab.com/fynbos/backend/wallets/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/ops"
)

func TestUpdateUserDetails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, db, nil, nil, nil, nil, nil, nil)

	walletID := uuid.NewString()

	ud, err := ops.UpdateIndividualDetails(ctx, b, kyc.IndividualDetails{
		WalletID:    walletID,
		FirstName:   "InitFirst",
		CountryCode: "ZA",
		Gender:      kyc.GenderMale,
		IPAddress:   "198.16.0.1",
	})
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "InitFirst", ud.FirstName)
	assert.Equal(t, "198.16.0.1", ud.IPAddress)
	assert.Empty(t, ud.LastName)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	var revisionCnt int
	err = db.GetContext(ctx, &revisionCnt, "select count(*) from individual_kyc_details where wallet_id=$1", walletID)
	require.NoError(t, err)
	assert.Equal(t, 1, revisionCnt)

	// Now look it up
	ud, err = ops.GetIndividualDetails(ctx, b, walletID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "InitFirst", ud.FirstName)
	assert.Equal(t, "198.16.0.1", ud.IPAddress)
	assert.Empty(t, ud.LastName)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	// Update
	ud, err = ops.UpdateIndividualDetails(ctx, b, kyc.IndividualDetails{
		WalletID:  walletID,
		FirstName: "Updated",
		LastName:  "New",
		IPAddress: "198.16.0.2",
	})
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.Equal(t, "198.16.0.2", ud.IPAddress)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	err = db.GetContext(ctx, &revisionCnt, "select count(*) from individual_kyc_details where wallet_id=$1", walletID)
	require.NoError(t, err)
	assert.Equal(t, 2, revisionCnt)

	// Now look it up
	ud, err = ops.GetIndividualDetails(ctx, b, walletID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.Equal(t, "198.16.0.2", ud.IPAddress)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	// Add an address
	ud, err = ops.UpdateIndividualDetails(ctx, b, kyc.IndividualDetails{
		WalletID:  walletID,
		IPAddress: "198.16.0.3",
		Address: &kyc.Address{
			Line1:       "Line1",
			Line2:       "Line2",
			Building:    "Building",
			Apartment:   "Apartment",
			City:        "Cape Town",
			State:       "ZA-WC",
			ZipCode:     "8001",
			CountryCode: "ZA",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.Equal(t, "198.16.0.3", ud.IPAddress)
	assert.True(t, ud.DateOfBirth.IsZero())
	require.NotNil(t, ud.Address)
	assert.Equal(t, "Line1", ud.Address.Line1)
	assert.Equal(t, "Line2", ud.Address.Line2)
	assert.Equal(t, "Building", ud.Address.Building)
	assert.Equal(t, "Apartment", ud.Address.Apartment)
	assert.Equal(t, "Cape Town", ud.Address.City)
	assert.Equal(t, "ZA-WC", ud.Address.State)
	assert.Equal(t, "8001", ud.Address.ZipCode)
	assert.Equal(t, "ZA", ud.Address.CountryCode)

	err = db.GetContext(ctx, &revisionCnt, "select count(*) from individual_kyc_details where wallet_id=$1", walletID)
	require.NoError(t, err)
	assert.Equal(t, 3, revisionCnt)

	// Now look it up
	ud, err = ops.GetIndividualDetails(ctx, b, walletID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.Equal(t, "198.16.0.3", ud.IPAddress)
	assert.True(t, ud.DateOfBirth.IsZero())
	require.NotNil(t, ud.Address)
	assert.Equal(t, "Line1", ud.Address.Line1)
	assert.Equal(t, "Line2", ud.Address.Line2)
	assert.Equal(t, "Building", ud.Address.Building)
	assert.Equal(t, "Apartment", ud.Address.Apartment)
	assert.Equal(t, "Cape Town", ud.Address.City)
	assert.Equal(t, "ZA-WC", ud.Address.State)
	assert.Equal(t, "8001", ud.Address.ZipCode)
	assert.Equal(t, "ZA", ud.Address.CountryCode)

	// Noop Update
	_, err = ops.UpdateIndividualDetails(ctx, b, kyc.IndividualDetails{
		WalletID:    walletID,
		FirstName:   "Updated",
		LastName:    "New",
		CountryCode: "ZA",
		IPAddress:   "198.16.0.3",
		Gender:      kyc.GenderMale,
	})
	require.NoError(t, err)

	err = db.GetContext(ctx, &revisionCnt, "select count(*) from individual_kyc_details where wallet_id=$1", walletID)
	require.NoError(t, err)
	assert.Equal(t, 3, revisionCnt)
}

func TestKYCStatus(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	db := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	nc := notify_client.NewMockClient(ctrl)
	nc.EXPECT().NotifyWallet(ctx, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	em := email_client.NewMockClient(ctrl)
	wc := wallet_client.NewMockClient(ctrl)
	b := ops.NewTestBackends(t, db, nil, nil, nil, nc, em, wc)

	walletID := uuid.NewString()
	wc.EXPECT().Get(ctx, gomock.Any()).Return(&wallets.Wallet{}, nil).AnyTimes()

	// Defaults to unknown if not set
	s, err := ops.GetKYCStatus(ctx, b, walletID)
	require.NoError(t, err)
	assert.Equal(t, kyc.StatusUnknown, s)

	// Can set status and sends out pending email
	err = ops.SetKYCStatus(ctx, b, walletID, kyc.StatusPending)
	require.NoError(t, err)
	s, err = ops.GetKYCStatus(ctx, b, walletID)
	require.NoError(t, err)
	assert.Equal(t, kyc.StatusPending, s)

	// send out email when going into review
	em.EXPECT().SendApplicationPendingEmail(ctx, walletID).Times(1)
	err = ops.SetKYCStatus(ctx, b, walletID, kyc.StatusInReview)
	require.NoError(t, err)
	s, err = ops.GetKYCStatus(ctx, b, walletID)
	assert.Equal(t, kyc.StatusInReview, s)
	require.NoError(t, err)

	// Setting status to kyc level 1 or kyc level 2 sends out approved email
	em.EXPECT().SendApplicationApprovedEmail(ctx, walletID).Times(1) // only send 1 email
	err = ops.SetKYCStatus(ctx, b, walletID, kyc.StatusLevel1)
	require.NoError(t, err)
	s, err = ops.GetKYCStatus(ctx, b, walletID)
	require.NoError(t, err)
	assert.Equal(t, kyc.StatusLevel1, s)

	err = ops.SetKYCStatus(ctx, b, walletID, kyc.StatusLevel2)
	require.NoError(t, err)
	s, err = ops.GetKYCStatus(ctx, b, walletID)
	require.NoError(t, err)
	assert.Equal(t, kyc.StatusLevel2, s)

	// Setting status to denied also sends out email
	em.EXPECT().SendApplicationDeniedEmail(ctx, walletID).Times(1)
	err = ops.SetKYCStatus(ctx, b, walletID, kyc.StatusDenied)
	require.NoError(t, err)

	s, err = ops.GetKYCStatus(ctx, b, walletID)
	require.NoError(t, err)
	assert.Equal(t, kyc.StatusDenied, s)

	// don't send out email if kyc is already denied
	err = ops.SetKYCStatus(ctx, b, walletID, kyc.StatusDenied)
	require.NoError(t, err)
}
