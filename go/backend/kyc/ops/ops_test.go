package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/kyc/ops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
