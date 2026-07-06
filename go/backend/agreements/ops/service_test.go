package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/agreements"
	"github.com/interledger/interledger-app/go/backend/agreements/migrations"
	"github.com/interledger/interledger-app/go/backend/agreements/ops"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignAgreements(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db, "../migrations/assets/testing"); err != nil {
		t.Fatal(err)
	}

	b := ops.NewTestBackends(t, db)

	err := ops.Sign(ctx, b, &agreements.SignArgs{
		AgreementIDs: []string{"privacy_policy-2.0.0"},
		UserID:       uuid.NewString(),
	})

	assert.NoError(t, err)
}

func TestAgreementSigns(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db, "../migrations/assets/testing"); err != nil {
		t.Fatal(err)
	}

	b := ops.NewTestBackends(t, db)

	userId := uuid.NewString()

	err := ops.Sign(ctx, b, &agreements.SignArgs{
		AgreementIDs: []string{"privacy_policy-2.0.0"},
		UserID:       userId,
	})
	if err != nil {
		t.Fatal(err)
	}

	signatures, err := ops.GetSignatures(ctx, b, userId)
	if err != nil {
		t.Fatal(err)
	}

	signature := signatures[0]

	assert.Equal(t, "privacy_policy-2.0.0", signature.AgreementID)
	assert.Equal(t, userId, signature.UserID)
}

func TestGetAgreement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db, "../migrations/assets/testing"); err != nil {
		t.Fatal(err)
	}

	b := ops.NewTestBackends(t, db)

	agreement, err := ops.Get(ctx, b, "privacy_policy-2.0.0")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "privacy_policy", agreement.Name)
	assert.Equal(t, "2.0.0", agreement.Version)
	assert.Equal(t, "privacy policy content v2", agreement.Content)
}

func TestListAffectedUserIDsPaginated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db, "../migrations/assets/testing"); err != nil {
		t.Fatal(err)
	}
	b := ops.NewTestBackends(t, db)

	userOld := uuid.NewString() // signed old version → should be affected
	userNew := uuid.NewString() // signed new version → should NOT be affected
	userOther := uuid.NewString() // signed unrelated agreement → should NOT be affected

	for _, tc := range []struct {
		userID      string
		agreementID string
	}{
		{userOld, "privacy_policy-1.0.0"},
		{userNew, "privacy_policy-2.0.0"},
		{userOther, "user_policy-1.0.0"},
	} {
		err := ops.Sign(ctx, b, &agreements.SignArgs{
			AgreementIDs: []string{tc.agreementID},
			UserID:       tc.userID,
		})
		require.NoError(t, err)
	}

	changes := []agreements.AgreementChange{{Name: "privacy_policy", ExceptID: "privacy_policy-2.0.0"}}

	t.Run("returns users who signed old versions", func(t *testing.T) {
		ids, err := ops.ListAffectedUserIDsPaginated(ctx, b, changes, 100, 0)
		require.NoError(t, err)
		assert.Contains(t, ids, userOld)
		assert.NotContains(t, ids, userNew)
		assert.NotContains(t, ids, userOther)
	})

	t.Run("pagination limit", func(t *testing.T) {
		ids, err := ops.ListAffectedUserIDsPaginated(ctx, b, changes, 0, 0)
		require.NoError(t, err)
		assert.Empty(t, ids)
	})

	t.Run("pagination offset beyond results", func(t *testing.T) {
		ids, err := ops.ListAffectedUserIDsPaginated(ctx, b, changes, 100, 1000)
		require.NoError(t, err)
		assert.Empty(t, ids)
	})

	t.Run("empty changes returns nil", func(t *testing.T) {
		ids, err := ops.ListAffectedUserIDsPaginated(ctx, b, nil, 100, 0)
		require.NoError(t, err)
		assert.Nil(t, ids)
	})
}

func TestMarkUsersNotified(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db, "../migrations/assets/testing"); err != nil {
		t.Fatal(err)
	}
	b := ops.NewTestBackends(t, db)

	userOld := uuid.NewString() // signed old version → should be marked
	userNew := uuid.NewString() // signed new version → should NOT be marked

	for _, tc := range []struct {
		userID      string
		agreementID string
	}{
		{userOld, "privacy_policy-1.0.0"},
		{userNew, "privacy_policy-2.0.0"},
	} {
		require.NoError(t, ops.Sign(ctx, b, &agreements.SignArgs{
			AgreementIDs: []string{tc.agreementID},
			UserID:       tc.userID,
		}))
	}

	changes := []agreements.AgreementChange{{Name: "privacy_policy", ExceptID: "privacy_policy-2.0.0"}}

	t.Run("marks old-version signer", func(t *testing.T) {
		require.NoError(t, ops.MarkUsersNotified(ctx, b, []string{userOld, userNew}, changes))

		sigs, err := ops.GetSignatures(ctx, b, userOld)
		require.NoError(t, err)
		require.Len(t, sigs, 1)
		require.NotNil(t, sigs[0].LastNotifiedAgreementID)
		assert.Equal(t, "privacy_policy-2.0.0", *sigs[0].LastNotifiedAgreementID)
	})

	t.Run("does not mark new-version signer", func(t *testing.T) {
		sigs, err := ops.GetSignatures(ctx, b, userNew)
		require.NoError(t, err)
		require.Len(t, sigs, 1)
		assert.Nil(t, sigs[0].LastNotifiedAgreementID)
	})

	t.Run("empty userIDs is a no-op", func(t *testing.T) {
		assert.NoError(t, ops.MarkUsersNotified(ctx, b, nil, changes))
	})

	t.Run("empty changes is a no-op", func(t *testing.T) {
		assert.NoError(t, ops.MarkUsersNotified(ctx, b, []string{userOld}, nil))
	})
}

func TestGetAgreementNamesSignedByUsersFromSet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	if err := migrations.MigrateFromMarkdowns(ctx, db, "../migrations/assets/testing"); err != nil {
		t.Fatal(err)
	}
	b := ops.NewTestBackends(t, db)

	userOld := uuid.NewString()   // signed privacy_policy-1.0.0 (old)
	userNew := uuid.NewString()   // signed privacy_policy-2.0.0 (new, the except)
	userBoth := uuid.NewString()  // signed both versions

	for _, tc := range []struct {
		userID      string
		agreementID string
	}{
		{userOld, "privacy_policy-1.0.0"},
		{userNew, "privacy_policy-2.0.0"},
		{userBoth, "privacy_policy-1.0.0"},
		{userBoth, "privacy_policy-2.0.0"},
	} {
		err := ops.Sign(ctx, b, &agreements.SignArgs{
			AgreementIDs: []string{tc.agreementID},
			UserID:       tc.userID,
		})
		require.NoError(t, err)
	}

	changes := []agreements.AgreementChange{{Name: "privacy_policy", ExceptID: "privacy_policy-2.0.0"}}

	t.Run("returns agreement names for users who signed old versions", func(t *testing.T) {
		result, err := ops.GetAgreementNamesSignedByUsersFromSet(ctx, b, []string{userOld, userNew, userBoth}, changes)
		require.NoError(t, err)
		assert.Equal(t, []string{"privacy_policy"}, result[userOld])
		assert.Empty(t, result[userNew])  // only signed the new version
		assert.Equal(t, []string{"privacy_policy"}, result[userBoth])
	})

	t.Run("empty userIDs returns empty map", func(t *testing.T) {
		result, err := ops.GetAgreementNamesSignedByUsersFromSet(ctx, b, nil, changes)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty changes returns empty map", func(t *testing.T) {
		result, err := ops.GetAgreementNamesSignedByUsersFromSet(ctx, b, []string{userOld}, nil)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
