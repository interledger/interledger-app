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
