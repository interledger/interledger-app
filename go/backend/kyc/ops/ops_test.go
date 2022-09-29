package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/ops"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestUpdateUserDetails(t *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := ops.NewTestBackends(t, db)
	userID := uuid.NewString()

	ud, err := ops.UpdateUserDetails(ctx, b, kyc.UserDetails{
		UserID:      userID,
		FirstName:   "InitFirst",
		CountryCode: "ZA",
		Gender:      kyc.GenderMale,
	})
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "InitFirst", ud.FirstName)
	assert.Empty(t, ud.LastName)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	var revisionCnt int
	err = db.GetContext(ctx, &revisionCnt, "select count(*) from user_kyc_details where user_id=$1", userID)
	require.NoError(t, err)
	assert.Equal(t, 1, revisionCnt)

	// Now look it up
	ud, err = ops.GetUserDetails(ctx, b, userID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "InitFirst", ud.FirstName)
	assert.Empty(t, ud.LastName)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	// Update
	ud, err = ops.UpdateUserDetails(ctx, b, kyc.UserDetails{
		UserID:    userID,
		FirstName: "Updated",
		LastName:  "New",
	})
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	err = db.GetContext(ctx, &revisionCnt, "select count(*) from user_kyc_details where user_id=$1", userID)
	require.NoError(t, err)
	assert.Equal(t, 2, revisionCnt)

	// Now look it up
	ud, err = ops.GetUserDetails(ctx, b, userID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.True(t, ud.DateOfBirth.IsZero())
	assert.Nil(t, ud.Address)

	// Add an address
	ud, err = ops.UpdateUserDetails(ctx, b, kyc.UserDetails{
		UserID: userID,
		Address: &kyc.Address{
			Line1:       "Line1",
			Line2:       "Line2",
			Building:    "Building",
			Apartment:   "Apartment",
			City:        "Cape Town",
			State:       "Western Cape",
			ZipCode:     "8001",
			CountryCode: "ZA",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.True(t, ud.DateOfBirth.IsZero())
	require.NotNil(t, ud.Address)
	assert.Equal(t, "Line1", ud.Address.Line1)
	assert.Equal(t, "Line2", ud.Address.Line2)
	assert.Equal(t, "Building", ud.Address.Building)
	assert.Equal(t, "Apartment", ud.Address.Apartment)
	assert.Equal(t, "Cape Town", ud.Address.City)
	assert.Equal(t, "Western Cape", ud.Address.State)
	assert.Equal(t, "8001", ud.Address.ZipCode)
	assert.Equal(t, "ZA", ud.Address.CountryCode)

	err = db.GetContext(ctx, &revisionCnt, "select count(*) from user_kyc_details where user_id=$1", userID)
	require.NoError(t, err)
	assert.Equal(t, 3, revisionCnt)

	// Now look it up
	ud, err = ops.GetUserDetails(ctx, b, userID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, ud.Gender)
	assert.Equal(t, "ZA", ud.CountryCode)
	assert.Equal(t, "Updated", ud.FirstName)
	assert.Equal(t, "New", ud.LastName)
	assert.True(t, ud.DateOfBirth.IsZero())
	require.NotNil(t, ud.Address)
	assert.Equal(t, "Line1", ud.Address.Line1)
	assert.Equal(t, "Line2", ud.Address.Line2)
	assert.Equal(t, "Building", ud.Address.Building)
	assert.Equal(t, "Apartment", ud.Address.Apartment)
	assert.Equal(t, "Cape Town", ud.Address.City)
	assert.Equal(t, "Western Cape", ud.Address.State)
	assert.Equal(t, "8001", ud.Address.ZipCode)
	assert.Equal(t, "ZA", ud.Address.CountryCode)

	// Noop Update
	_, err = ops.UpdateUserDetails(ctx, b, kyc.UserDetails{
		UserID:      userID,
		FirstName:   "Updated",
		LastName:    "New",
		CountryCode: "ZA",
		Gender:      kyc.GenderMale,
	})
	require.NoError(t, err)

	err = db.GetContext(ctx, &revisionCnt, "select count(*) from user_kyc_details where user_id=$1", userID)
	require.NoError(t, err)
	assert.Equal(t, 3, revisionCnt)
}
