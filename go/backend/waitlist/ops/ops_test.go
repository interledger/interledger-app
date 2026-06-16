package ops_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"github.com/go-playground/validator/v10"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/waitlist"
	"github.com/interledger/interledger-app/go/backend/waitlist/ops"
	"github.com/stretchr/testify/assert"
)

func TestAddSignup(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewBackends(t, db, validator.New())

	cases := []struct {
		name      string
		email     string
		country   string
		fullName  string
		betaOptIn bool
		err       error
	}{
		{
			name:      "success",
			email:     "signup@interledger.test",
			country:   "ZA",
			fullName:  "Bob",
			betaOptIn: false,
		},
		{
			name:      "duplicate",
			email:     "signup@interledger.test",
			country:   "ZA",
			fullName:  "Bob",
			betaOptIn: false,
		},
		{
			name:      "beta opt in",
			email:     "signup@interledger.test",
			country:   "ZA",
			fullName:  "Bob",
			betaOptIn: true,
		},
		{
			name:      "duplicate new country",
			email:     "signup@interledger.test",
			country:   "GB",
			fullName:  "Bob",
			betaOptIn: false,
		},
		{
			name:      "invalid email",
			email:     "blahblah",
			country:   "GB",
			fullName:  "Bob",
			betaOptIn: false,
			err:       waitlist.ErrInvalidEmail,
		},
		{
			name:      "invalid country",
			email:     "nocountry@interledger.test",
			country:   "LALA",
			fullName:  "Bob",
			betaOptIn: false,
			err:       waitlist.ErrInvalidCountry,
		},
		{
			name:      "empty country",
			email:     "nocountry@interledger.test",
			fullName:  "Bob",
			betaOptIn: false,
			err:       waitlist.ErrInvalidCountry,
		},
		{
			name:      "empty name",
			email:     "signup@interledger.test",
			country:   "ZA",
			betaOptIn: false,
			err:       waitlist.ErrInvalidName,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ops.AddSignup(ctx, b, tc.email, tc.country, tc.fullName, "", tc.betaOptIn)
			if tc.err == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestAddSignupWithMug(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewBackends(t, db, validator.New())

	cases := []struct {
		name     string
		email    string
		country  string
		fullName string
		mugID    string
		hasMug   string
	}{
		{
			name:     "success no mug",
			email:    "signup@interledger.test",
			country:  "ZA",
			fullName: "Bob",
		},
		{
			name:     "override with mug",
			email:    "signup@interledger.test",
			country:  "ZA",
			fullName: "Bob",
			mugID:    "1e25f533",
			hasMug:   "1e25f533",
		},
		{
			name:     "signup with duplicate mug",
			email:    "taken@interledger.test",
			country:  "ZA",
			fullName: "Bob",
			mugID:    "1e25f533",
			hasMug:   "",
		},
		{
			name:     "not override mug with null",
			email:    "signup@interledger.test",
			country:  "ZA",
			fullName: "Bob",
			mugID:    "",
			hasMug:   "1e25f533",
		},
		{
			name:     "not override mug with new mug",
			email:    "signup@interledger.test",
			country:  "ZA",
			fullName: "Bob",
			mugID:    "16ba8774",
			hasMug:   "1e25f533",
		},
		{
			name:     "success with mug",
			email:    "new_signup@interledger.test",
			country:  "ZA",
			fullName: "Bob",
			mugID:    "625ee641",
			hasMug:   "625ee641",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ops.AddSignup(ctx, b, tc.email, tc.country, tc.fullName, tc.mugID, false)
			require.NoError(t, err)

			var mug sql.NullString
			err = b.DB().GetContext(ctx, &mug, "SELECT mug_id FROM waitlist_signups WHERE country_code=$1 AND email=$2", tc.country, tc.email)
			require.NoError(t, err)

			assert.Equal(t, tc.hasMug, mug.String)
		})
	}
}

func TestCanSignup(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewBackends(t, db, validator.New())

	cases := []struct {
		name      string
		email     string
		country   string
		fullName  string
		canSignup bool
		betaOptIn bool
		err       error
	}{
		{
			name:      "allowed",
			email:     "allowed@interledger.test",
			country:   "ZA",
			fullName:  "Bob",
			betaOptIn: false,
			canSignup: true,
		},
		{
			name:      "not allowed",
			email:     "nowallowed@interledger.test",
			country:   "ZA",
			fullName:  "Robert",
			betaOptIn: false,
			canSignup: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ops.AddSignup(ctx, b, tc.email, tc.country, tc.fullName, "", tc.betaOptIn)
			if tc.err == nil {
				assert.NoError(t, err)
			}
			signupId, err := ops.GetIdByEmailAndCountryCode(ctx, b, tc.email, tc.country)
			assert.NoError(t, err)

			if tc.canSignup == true {
				err = ops.AllowSignupById(ctx, b, signupId)
				assert.NoError(t, err)
			}

			canSignup, err := ops.CanSignup(ctx, b, signupId)
			assert.NoError(t, err)
			assert.Equal(t, tc.canSignup, canSignup)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestSetSignupComplete(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewBackends(t, db, validator.New())

	cases := []struct {
		name        string
		email       string
		country     string
		fullName    string
		canSignup   bool
		setComplete bool
		err         error
	}{
		{
			name:        "complete signup",
			email:       "allowed@interledger.test",
			country:     "ZA",
			fullName:    "Bob",
			canSignup:   true,
			setComplete: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ops.AddSignup(ctx, b, tc.email, tc.country, tc.fullName, "", false)
			if tc.err == nil {
				assert.NoError(t, err)
			}
			signupId, err := ops.GetIdByEmailAndCountryCode(ctx, b, tc.email, tc.country)
			assert.NoError(t, err)

			if tc.canSignup == true {
				err = ops.AllowSignupById(ctx, b, signupId)
				assert.NoError(t, err)
			}

			canSignup, err := ops.CanSignup(ctx, b, signupId)
			assert.NoError(t, err)
			assert.Equal(t, tc.canSignup, canSignup)

			userId := uuid.NewString()
			err = ops.SetSignupComplete(ctx, b, signupId, userId)
			assert.NoError(t, err)

			canSignup, err = ops.CanSignup(ctx, b, signupId)
			assert.NoError(t, err)
			assert.Equal(t, false, canSignup)
		})
	}
}

func TestListSignups(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewBackends(t, db, validator.New())

	sups := []struct {
		email     string
		country   string
		fullName  string
		betaOptIn bool
	}{
		{
			email:     "bob@interledger.test",
			country:   "ZA",
			fullName:  "Bob",
			betaOptIn: false,
		},
		{
			email:     "alice@interledger.test",
			country:   "ZA",
			fullName:  "Alice",
			betaOptIn: false,
		},
		{
			email:     "beta@interledger.test",
			country:   "US",
			fullName:  "Beta Max",
			betaOptIn: true,
		},
	}

	for _, signup := range sups {
		err := ops.AddSignup(ctx, b, signup.email, signup.country, signup.fullName, "", false)
		assert.NoError(t, err)
	}

	signups, err := ops.ListSignups(ctx, b)
	assert.NoError(t, err)

	assert.Len(t, signups, 3)
}
