package admin

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/onboarding"
	_user "gitlab.com/fynbos/backend/user"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestAccounts(s *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx)
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
		acc, err := c.Os.CreateAccount(ctx, &onboarding.CreateAccountArgs{
			IdentityID:   user.ID,
			Email:        user.Email,
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			MobileNumber: faker.E164PhoneNumber(),
			Country:      "US",
		})
		if err != nil {
			t.Fatal(err)
		}

		response, err := c.AdminClient.GetUserAccountByEmail(ctx, &backendv1.GetUserAccountByEmailRequest{
			Email: user.Email,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, acc.ID, response.GetId())
		assert.Equal(t, acc.AvailableBalance, response.GetBalance())
		assert.Equal(t, acc.DebitsAccepted, response.GetDebitsAccepted())
		assert.Equal(t, acc.DebitsReserved, response.GetDebitsReserved())
		assert.Equal(t, acc.CreditsAccepted, response.GetCreditsAccepted())
		assert.Equal(t, acc.CreditsReserved, response.GetCreditsReserved())
	})
}
