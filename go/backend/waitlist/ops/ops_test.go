package ops_test

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/backend/waitlist"
	"gitlab.com/fynbos/backend/waitlist/ops"
)

func TestAddSignup(t *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	b := ops.NewBackends(t, db, validator.New())

	cases := []struct {
		name     string
		email    string
		country  string
		fullName string
		err      error
	}{
		{
			name:     "success",
			email:    "signup@fynbos.dev",
			country:  "ZA",
			fullName: "Bob",
		},
		{
			name:     "duplicate",
			email:    "signup@fynbos.dev",
			country:  "ZA",
			fullName: "Bob",
		},
		{
			name:     "duplicate new country",
			email:    "signup@fynbos.dev",
			country:  "GB",
			fullName: "Bob",
		},
		{
			name:     "invalid email",
			email:    "blahblah",
			country:  "GB",
			fullName: "Bob",
			err:      waitlist.ErrInvalidEmail,
		},
		{
			name:     "invalid country",
			email:    "nocountry@fynbos.dev",
			country:  "LALA",
			fullName: "Bob",
			err:      waitlist.ErrInvalidCountry,
		},
		{
			name:     "empty country",
			email:    "nocountry@fynbos.dev",
			fullName: "Bob",
			err:      waitlist.ErrInvalidCountry,
		},
		{
			name:    "empty name",
			email:   "signup@fynbos.dev",
			country: "ZA",
			err:     waitlist.ErrInvalidName,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ops.AddSignup(ctx, b, tc.email, tc.country, tc.fullName)
			if tc.err == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorIs(t, err, tc.err)
		})
	}
}
