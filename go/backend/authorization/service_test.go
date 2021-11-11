package authorization

import (
	"testing"

	"github.com/osohq/go-oso/errors"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/organisation"
	"gitlab.com/fynbos/backend/user"
)

func TestAuthorizationService(s *testing.T) {

	authz, err := NewService()
	if err != nil {
		s.Fatal(err)
	}

	s.Run("only allows owner to read organisation", func(t *testing.T) {
		owner := user.User{
			ID: "123",
		}
		notOwner := user.User{
			ID: "124",
		}
		org := organisation.Organisation{
			ID:        "321",
			OwnerID:   owner.ID,
			Name:      "My organisation.",
			CreatedAt: "",
			UpdatedAt: "",
		}

		err := authz.Authorize(owner, "read", org)
		assert.NoError(t, err)

		err = authz.Authorize(notOwner, "read", org)
		assert.Error(t, err, errors.NotFoundError{})
	})

}
