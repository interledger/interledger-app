package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/bxcodec/faker/v3"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/signup/ops"
	"gitlab.com/fynbos/backend/twilio"
)

type backends struct {
	validator *validator.Validate
	db        *sqlx.DB
	twilio    twilio.Service
}

func (b backends) Validator() *validator.Validate {
	return b.validator
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) Twilio() twilio.Service {
	return b.twilio
}

func TestSetUserData(t *testing.T) {
	ctx := context.Background()

	db := db.MigrateTestDB(t, ctx)

	b := &backends{
		validator: validator.New(),
		db:        db,
	}

	id, err := ops.SetUserData(ctx, b, signup.UserDataArgs{
		FirstName:   "FirstName",
		LastName:    "LastName",
		Email:       "test@interledger.test",
		CountryCode: "ZA",
	})
	require.NoError(t, err)

	// Lookup and verify
	su, err := ops.Get(ctx, b, id)
	require.NoError(t, err)

	assert.Equal(t, "FirstName", su.FirstName)
	assert.Equal(t, "LastName", su.LastName)
	assert.Equal(t, "test@interledger.test", su.Email)
	assert.Equal(t, "ZA", su.CountryCode)

	// Update
	id, err = ops.SetUserData(ctx, b, signup.UserDataArgs{
		FirstName:   "Jason",
		LastName:    "Real Person",
		Email:       "random@interledger.test",
		CountryCode: "GB",
	})
	require.NoError(t, err)

	// Lookup and verify
	su, err = ops.Get(ctx, b, id)
	require.NoError(t, err)

	assert.Equal(t, "Jason", su.FirstName)
	assert.Equal(t, "Real Person", su.LastName)
	assert.Equal(t, "random@interledger.test", su.Email)
	assert.Equal(t, "GB", su.CountryCode)
}

func TestSetMobileNumber(t *testing.T) {
	ctx := context.Background()

	db := db.MigrateTestDB(t, ctx)

	tw := twilio.NewMockService(gomock.NewController(t))
	tw.EXPECT().CheckVerificationCode(ctx, gomock.Any()).Return(&twilio.Verification{Status: "approved"}, nil).AnyTimes()
	b := &backends{
		validator: validator.New(),
		db:        db,
		twilio:    tw,
	}

	id, err := ops.SetUserData(ctx, b, signup.UserDataArgs{
		FirstName:   "FirstName",
		LastName:    "LastName",
		Email:       "test@interledger.test",
		CountryCode: "ZA",
	})
	require.NoError(t, err)

	mobile := faker.E164PhoneNumber()
	err = ops.SetMobileNumber(ctx, b, signup.MobileNumberArgs{
		ID:           id,
		MobileNumber: mobile,
		OTP:          "123456",
	})
	require.NoError(t, err)

	su, err := ops.Get(ctx, b, id)
	require.NoError(t, err)

	assert.Equal(t, "FirstName", su.FirstName)
	assert.Equal(t, "LastName", su.LastName)
	assert.Equal(t, "test@interledger.test", su.Email)
	assert.Equal(t, "ZA", su.CountryCode)
	assert.Equal(t, mobile, su.MobileNumber)
	assert.False(t, su.Completed)
}

func TestFailsDuplicateCompleteMobileNumber(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	tw := twilio.NewMockService(gomock.NewController(t))
	tw.EXPECT().CheckVerificationCode(ctx, gomock.Any()).Return(&twilio.Verification{Status: "approved"}, nil).AnyTimes()
	b := &backends{
		validator: validator.New(),
		db:        db,
		twilio:    tw,
	}
	id, err := ops.SetUserData(ctx, b, signup.UserDataArgs{
		FirstName:   "FirstName",
		LastName:    "LastName",
		Email:       "test@interledger.test",
		CountryCode: "ZA",
	})
	require.NoError(t, err)
	mobile := faker.E164PhoneNumber()
	err = ops.SetMobileNumber(ctx, b, signup.MobileNumberArgs{
		ID:           id,
		MobileNumber: mobile,
		OTP:          "123456",
	})
	require.NoError(t, err)
	userID := uuid.NewString()
	err = ops.Complete(ctx, b, id, userID)
	require.NoError(t, err)

	id1, err := ops.SetUserData(ctx, b, signup.UserDataArgs{
		FirstName:   "FirstName1",
		LastName:    "LastName1",
		Email:       "test1@interledger.test",
		CountryCode: "ZA",
	})
	require.NoError(t, err)
	err = ops.SetMobileNumber(ctx, b, signup.MobileNumberArgs{
		ID:           id1,
		MobileNumber: mobile,
		OTP:          "123456",
	})
	require.ErrorIs(t, err, signup.ErrDuplicatePhone)
}

func TestComplete(t *testing.T) {
	ctx := context.Background()

	db := db.MigrateTestDB(t, ctx)

	tw := twilio.NewMockService(gomock.NewController(t))
	tw.EXPECT().CheckVerificationCode(ctx, gomock.Any()).Return(&twilio.Verification{Status: "approved"}, nil).AnyTimes()
	b := &backends{
		validator: validator.New(),
		db:        db,
		twilio:    tw,
	}

	id, err := ops.SetUserData(ctx, b, signup.UserDataArgs{
		FirstName:   "FirstName",
		LastName:    "LastName",
		Email:       "test@interledger.test",
		CountryCode: "ZA",
	})
	require.NoError(t, err)

	mobile := faker.E164PhoneNumber()
	err = ops.SetMobileNumber(ctx, b, signup.MobileNumberArgs{
		ID:           id,
		MobileNumber: mobile,
		OTP:          "123456",
	})
	require.NoError(t, err)

	userID := uuid.NewString()
	err = ops.Complete(ctx, b, id, userID)
	require.NoError(t, err)

	su, err := ops.Get(ctx, b, id)
	require.NoError(t, err)

	assert.Equal(t, "FirstName", su.FirstName)
	assert.Equal(t, "LastName", su.LastName)
	assert.Equal(t, "test@interledger.test", su.Email)
	assert.Equal(t, "ZA", su.CountryCode)
	assert.Equal(t, mobile, su.MobileNumber)
	assert.Equal(t, userID, su.UserID)
	assert.True(t, su.Completed)

	su, err = ops.GetForUser(ctx, b, userID)
	require.NoError(t, err)

	assert.Equal(t, "FirstName", su.FirstName)
	assert.Equal(t, "LastName", su.LastName)
	assert.Equal(t, "test@interledger.test", su.Email)
	assert.Equal(t, "ZA", su.CountryCode)
	assert.Equal(t, mobile, su.MobileNumber)
	assert.Equal(t, userID, su.UserID)
	assert.True(t, su.Completed)
}

func TestCompleteIdempotent(t *testing.T) {
	ctx := context.Background()

	db := db.MigrateTestDB(t, ctx)

	tw := twilio.NewMockService(gomock.NewController(t))
	tw.EXPECT().CheckVerificationCode(ctx, gomock.Any()).Return(&twilio.Verification{Status: "approved"}, nil).AnyTimes()
	b := &backends{
		validator: validator.New(),
		db:        db,
		twilio:    tw,
	}

	id, err := ops.SetUserData(ctx, b, signup.UserDataArgs{
		FirstName:   "FirstName",
		LastName:    "LastName",
		Email:       "test@interledger.test",
		CountryCode: "ZA",
	})
	require.NoError(t, err)

	mobile := faker.E164PhoneNumber()
	err = ops.SetMobileNumber(ctx, b, signup.MobileNumberArgs{
		ID:           id,
		MobileNumber: mobile,
		OTP:          "123456",
	})
	require.NoError(t, err)

	userID := uuid.NewString()
	err = ops.Complete(ctx, b, id, userID)
	require.NoError(t, err)

	err = ops.Complete(ctx, b, id, userID)
	require.NoError(t, err)
}

func TestCompleteFailsAnotherUser(t *testing.T) {
	ctx := context.Background()

	db := db.MigrateTestDB(t, ctx)

	tw := twilio.NewMockService(gomock.NewController(t))
	tw.EXPECT().CheckVerificationCode(ctx, gomock.Any()).Return(&twilio.Verification{Status: "approved"}, nil).AnyTimes()
	b := &backends{
		validator: validator.New(),
		db:        db,
		twilio:    tw,
	}

	id, err := ops.SetUserData(ctx, b, signup.UserDataArgs{
		FirstName:   "FirstName",
		LastName:    "LastName",
		Email:       "test@interledger.test",
		CountryCode: "ZA",
	})
	require.NoError(t, err)

	mobile := faker.E164PhoneNumber()
	err = ops.SetMobileNumber(ctx, b, signup.MobileNumberArgs{
		ID:           id,
		MobileNumber: mobile,
		OTP:          "123456",
	})
	require.NoError(t, err)

	userID := uuid.NewString()
	err = ops.Complete(ctx, b, id, userID)
	require.NoError(t, err)

	anotherUserID := uuid.NewString()
	err = ops.Complete(ctx, b, id, anotherUserID)
	require.Error(t, err)

	su, err := ops.GetForUser(ctx, b, userID)
	require.NoError(t, err)
	assert.Equal(t, userID, su.UserID)
	assert.True(t, su.Completed)
}
