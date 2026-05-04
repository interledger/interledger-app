package ops_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/accountdeletion"
	"gitlab.com/fynbos/backend/accountdeletion/ops"
	"gitlab.com/fynbos/backend/db"
)

type backends struct {
	db *sqlx.DB
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func setupTest(t *testing.T) (context.Context, *backends) {
	ctx := context.Background()
	return ctx, &backends{db: db.MigrateTestDB(t, ctx)}
}

func TestRequest_FirstTimeInsertsRow(t *testing.T) {
	ctx, b := setupTest(t)
	userID := uuid.NewString()

	err := ops.Request(ctx, b, userID)
	require.NoError(t, err)

	got, err := ops.GetForUser(ctx, b, userID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, accountdeletion.StatusPending, got.Status)
	assert.False(t, got.CreatedAt.IsZero())
	assert.False(t, got.UpdatedAt.IsZero())
}

func TestRequest_DuplicateReturnsAlreadyRequested(t *testing.T) {
	ctx, b := setupTest(t)
	userID := uuid.NewString()

	require.NoError(t, ops.Request(ctx, b, userID))

	err := ops.Request(ctx, b, userID)
	require.Error(t, err)
	require.True(t, errors.Is(err, accountdeletion.ErrAlreadyRequested))

	var count int
	require.NoError(t, b.DB().GetContext(ctx, &count,
		"SELECT count(*) FROM account_deletion_requests WHERE user_id = $1", userID))
	assert.Equal(t, 1, count)
}

func TestRequest_DifferentUsersAreIndependent(t *testing.T) {
	ctx, b := setupTest(t)
	userA := uuid.NewString()
	userB := uuid.NewString()

	require.NoError(t, ops.Request(ctx, b, userA))
	require.NoError(t, ops.Request(ctx, b, userB))

	gotA, err := ops.GetForUser(ctx, b, userA)
	require.NoError(t, err)
	require.NotNil(t, gotA)

	gotB, err := ops.GetForUser(ctx, b, userB)
	require.NoError(t, err)
	require.NotNil(t, gotB)

	assert.NotEqual(t, gotA.ID, gotB.ID)
}

func TestGetForUser_NoRowReturnsNilNil(t *testing.T) {
	ctx, b := setupTest(t)

	got, err := ops.GetForUser(ctx, b, uuid.NewString())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestDelete_RemovesPendingRow(t *testing.T) {
	ctx, b := setupTest(t)
	userID := uuid.NewString()

	require.NoError(t, ops.Request(ctx, b, userID))
	require.NoError(t, ops.Delete(ctx, b, userID))

	got, err := ops.GetForUser(ctx, b, userID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestDelete_NoRowIsNoop(t *testing.T) {
	ctx, b := setupTest(t)

	require.NoError(t, ops.Delete(ctx, b, uuid.NewString()))
}

func TestDelete_AllowsReRequest(t *testing.T) {
	ctx, b := setupTest(t)
	userID := uuid.NewString()

	require.NoError(t, ops.Request(ctx, b, userID))
	require.NoError(t, ops.Delete(ctx, b, userID))
	require.NoError(t, ops.Request(ctx, b, userID))
}
