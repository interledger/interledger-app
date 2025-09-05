package ops_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/user/ops"
)

func TestUserForContext(t *testing.T) {
	ctx := context.Background()

	_, err := ops.UserForContext(ctx)
	require.ErrorIs(t, err, user.ErrNoUserFound)

	ctx = context.WithValue(ctx, user.UserCtxKey("user"), &user.User{
		ID:    "1235",
		Email: "test@interledger.test",
	})

	u, err := ops.UserForContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, u.ID, "1235")
	assert.Equal(t, u.Email, "test@interledger.test")
}
