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

	id, err := ops.UpdateUserDetails(ctx, b, kyc.UserDetails{
		UserID:      userID,
		FirstName:   "InitFirst",
		CountryCode: "ZA",
		Gender:      kyc.GenderMale,
	})
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, id.Gender)
	assert.Equal(t, "ZA", id.CountryCode)
	assert.Equal(t, "InitFirst", id.FirstName)
	assert.Empty(t, id.LastName)
	assert.True(t, id.DateOfBirth.IsZero())
	assert.Nil(t, id.Address)

	var revisionCnt int
	err = db.GetContext(ctx, &revisionCnt, "select count(*) from user_kyc_details where user_id=$1", userID)
	require.NoError(t, err)
	assert.Equal(t, 1, revisionCnt)

	// Now look it up
	id, err = ops.GetUserDetails(ctx, b, userID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, id.Gender)
	assert.Equal(t, "ZA", id.CountryCode)
	assert.Equal(t, "InitFirst", id.FirstName)
	assert.Empty(t, id.LastName)
	assert.True(t, id.DateOfBirth.IsZero())
	assert.Nil(t, id.Address)

	// Update
	id, err = ops.UpdateUserDetails(ctx, b, kyc.UserDetails{
		UserID:    userID,
		FirstName: "Updated",
		LastName:  "New",
	})
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, id.Gender)
	assert.Equal(t, "ZA", id.CountryCode)
	assert.Equal(t, "Updated", id.FirstName)
	assert.Equal(t, "New", id.LastName)
	assert.True(t, id.DateOfBirth.IsZero())
	assert.Nil(t, id.Address)

	err = db.GetContext(ctx, &revisionCnt, "select count(*) from user_kyc_details where user_id=$1", userID)
	require.NoError(t, err)
	assert.Equal(t, 2, revisionCnt)

	// Now look it up
	id, err = ops.GetUserDetails(ctx, b, userID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, id.Gender)
	assert.Equal(t, "ZA", id.CountryCode)
	assert.Equal(t, "Updated", id.FirstName)
	assert.Equal(t, "New", id.LastName)
	assert.True(t, id.DateOfBirth.IsZero())
	assert.Nil(t, id.Address)

	// Add an address
	id, err = ops.UpdateUserDetails(ctx, b, kyc.UserDetails{
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

	assert.Equal(t, kyc.GenderMale, id.Gender)
	assert.Equal(t, "ZA", id.CountryCode)
	assert.Equal(t, "Updated", id.FirstName)
	assert.Equal(t, "New", id.LastName)
	assert.True(t, id.DateOfBirth.IsZero())
	require.NotNil(t, id.Address)
	assert.Equal(t, "Line1", id.Address.Line1)
	assert.Equal(t, "Line2", id.Address.Line2)
	assert.Equal(t, "Building", id.Address.Building)
	assert.Equal(t, "Apartment", id.Address.Apartment)
	assert.Equal(t, "Cape Town", id.Address.City)
	assert.Equal(t, "Western Cape", id.Address.State)
	assert.Equal(t, "8001", id.Address.ZipCode)
	assert.Equal(t, "ZA", id.Address.CountryCode)

	err = db.GetContext(ctx, &revisionCnt, "select count(*) from user_kyc_details where user_id=$1", userID)
	require.NoError(t, err)
	assert.Equal(t, 3, revisionCnt)

	// Now look it up
	id, err = ops.GetUserDetails(ctx, b, userID)
	require.NoError(t, err)

	assert.Equal(t, kyc.GenderMale, id.Gender)
	assert.Equal(t, "ZA", id.CountryCode)
	assert.Equal(t, "Updated", id.FirstName)
	assert.Equal(t, "New", id.LastName)
	assert.True(t, id.DateOfBirth.IsZero())
	require.NotNil(t, id.Address)
	assert.Equal(t, "Line1", id.Address.Line1)
	assert.Equal(t, "Line2", id.Address.Line2)
	assert.Equal(t, "Building", id.Address.Building)
	assert.Equal(t, "Apartment", id.Address.Apartment)
	assert.Equal(t, "Cape Town", id.Address.City)
	assert.Equal(t, "Western Cape", id.Address.State)
	assert.Equal(t, "8001", id.Address.ZipCode)
	assert.Equal(t, "ZA", id.Address.CountryCode)

	// Noop Update
	id, err = ops.UpdateUserDetails(ctx, b, kyc.UserDetails{
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
