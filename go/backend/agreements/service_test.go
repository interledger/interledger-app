package agreements

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestSignAgreements(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	if err := MigrateFromMarkdowns(ctx, &MigrateArgs{
		Db:            db,
		DirectoryPath: "../utils/agreements/test",
	}); err != nil {
		t.Fatal(err)
	}
	as, err := NewService(&ServiceArgs{
		Db:            db,
		AgreementsDir: "../utils/agreements/test",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = as.Sign(ctx, &SignArgs{
		AgreementIDs: []string{"privacy_policy-2.0.0"},
		IdentityID:   uuid.NewString(),
		IPAddress:    "123.123.123.123",
	})

	assert.NoError(t, err)
}

func TestAgreementSigns(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	if err := MigrateFromMarkdowns(ctx, &MigrateArgs{
		Db:            db,
		DirectoryPath: "../utils/agreements/test",
	}); err != nil {
		t.Fatal(err)
	}
	as, err := NewService(&ServiceArgs{
		Db:            db,
		AgreementsDir: "../utils/agreements/test",
	})
	if err != nil {
		t.Fatal(err)
	}

	userId := uuid.NewString()
	userIp := faker.IPv4()

	err = as.Sign(ctx, &SignArgs{
		AgreementIDs: []string{"privacy_policy-2.0.0"},
		IdentityID:   userId,
		IPAddress:    userIp,
	})
	if err != nil {
		t.Fatal(err)
	}

	signatures, err := as.GetSignatures(ctx, userId)
	if err != nil {
		t.Fatal(err)
	}

	signature := signatures[0]

	assert.Equal(t, "privacy_policy-2.0.0", signature.AgreementID)
	assert.Equal(t, userId, signature.IdentityID)
	assert.Equal(t, userIp, signature.IPAddress)
}

func TestGetAgreement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	if err := MigrateFromMarkdowns(ctx, &MigrateArgs{
		Db:            db,
		DirectoryPath: "../utils/agreements/test",
	}); err != nil {
		t.Fatal(err)
	}
	as, err := NewService(&ServiceArgs{
		Db:            db,
		AgreementsDir: "../utils/agreements/test",
	})
	if err != nil {
		t.Fatal(err)
	}

	agreement, err := as.Get(ctx, "privacy_policy-2.0.0")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "privacy_policy", agreement.Name)
	assert.Equal(t, "2.0.0", agreement.Version)
	assert.Equal(t, "privacy policy content v2", agreement.Content)
}

func TestLiveAgreements(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	if err := MigrateFromMarkdowns(ctx, &MigrateArgs{
		Db:            db,
		DirectoryPath: "../utils/agreements/test",
	}); err != nil {
		t.Fatal(err)
	}

	err := MigrateFromMarkdowns(ctx, &MigrateArgs{
		Db:            db,
		DirectoryPath: "../utils/agreements/live",
	})

	assert.NoError(t, err)
}

func TestMigrationIdempotency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	if err := MigrateFromMarkdowns(ctx, &MigrateArgs{
		Db:            db,
		DirectoryPath: "../utils/agreements/test",
	}); err != nil {
		t.Fatal(err)
	}
	err := MigrateFromMarkdowns(ctx, &MigrateArgs{
		Db:            db,
		DirectoryPath: "../utils/agreements/test",
	})

	assert.NoError(t, err)
}
