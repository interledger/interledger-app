package agreements

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestSignAgreements(t *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	as, err := NewService(&ServiceArgs{
		Db:            db,
		AgreementsDir: "markdowns",
	})
	if err != nil {
		t.Fatal(err)
	}

	signedAgreements, err := as.SignAgreement(ctx, &SignAgreementArgs{
		AgreementIDs: []string{"1", "2"},
		IdentityID:   uuid.NewString(),
		IPAddress:    "123.123.123.123",
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "1", signedAgreements.AgreementIDs[0])
}

func TestGetAgreement(t *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	as, err := NewService(&ServiceArgs{
		Db:            db,
		AgreementsDir: "markdowns",
	})
	if err != nil {
		t.Fatal(err)
	}

	agreement, err := as.GetAgreement(ctx, "privacy_policy-2.0.0")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "privacy_policy", agreement.Name)
	assert.Equal(t, "2.0.0", agreement.Version)
	assert.Equal(t, "privacy policy content v2", agreement.Content)
}

func TestGetAllAgreements(t *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)

	err := StoreAgreements(ctx, &StoreAgreementsArgs{
		db:  db,
		dir: "markdowns",
	})

	assert.NoError(t, err)
}
