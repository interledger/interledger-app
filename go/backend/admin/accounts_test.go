package admin

import (
	"context"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestAccounts(s *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		if err = c.Cleanup(ctx); err != nil {
			s.Fatal(err)
		}
	})

	s.Run("can get user account by email", func(t *testing.T) {
		user := _user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		id, err := c.Is.Create(ctx, &identity.CreateArgs{
			ID:           user.ID,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Email:        user.Email,
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		acc, err := c.Os.CreateAccount(ctx, &onboarding.CreateAccountArgs{
			IdentityID: id.ID,
			Country:    id.Country,
		})
		if err != nil {
			t.Fatal(err)
		}

		scenarios := []struct {
			Name          string
			AdminUser     string
			ExpectedError string
		}{
			{
				Name:          "Non approved admin user is not allowed to read.",
				AdminUser:     "test@fynbos.test",
				ExpectedError: "Forbidden.",
			},
			{
				Name:          "Unauthenticated request is denied",
				AdminUser:     "",
				ExpectedError: "Unauthenticated.",
			},
			{
				Name:          "Approved admin user is allowed to read",
				AdminUser:     "don@fynbos.dev",
				ExpectedError: "",
			},
			{
				Name:          "Approved admin user is allowed to read",
				AdminUser:     "matt@fynbos.dev",
				ExpectedError: "",
			},
			{
				Name:          "Approved admin user is allowed to read",
				AdminUser:     "cairin@fynbos.dev",
				ExpectedError: "",
			},
		}

		for _, scenario := range scenarios {
			response, err := c.AdminClient.GetUserAccountByEmail(
				auth.ActingAs(ctx, scenario.AdminUser),
				&backendv1.GetUserAccountByEmailRequest{
					Email: user.Email,
				})

			if scenario.ExpectedError == "" {
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, acc.ID, response.GetId())
				assert.Equal(t, acc.AvailableBalance, response.GetBalance())
				assert.Equal(t, acc.DebitsAccepted, response.GetDebitsAccepted())
				assert.Equal(t, acc.DebitsReserved, response.GetDebitsReserved())
				assert.Equal(t, acc.CreditsAccepted, response.GetCreditsAccepted())
				assert.Equal(t, acc.CreditsReserved, response.GetCreditsReserved())
			} else {
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), scenario.ExpectedError))
			}
		}
	})
}
