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

func TestAgreementSigns(t *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	as, err := NewService(&ServiceArgs{
		Db:            db,
		AgreementsDir: "markdowns",
	})
	if err != nil {
		t.Fatal(err)
	}

	userId := uuid.NewString()
	userIp := faker.IPv4()

	_, err = as.SignAgreement(ctx, &SignAgreementArgs{
		AgreementIDs: []string{"privacy_policy-2.0.0"},
		IdentityID:   userId,
		IPAddress:    userIp,
	})
	if err != nil {
		t.Fatal(err)
	}

	agreementSigns, err := as.GetAgreementSigns(ctx, userId)
	if err != nil {
		t.Fatal(err)
	}

	agreementSign := agreementSigns[0]

	assert.Equal(t, "privacy_policy-2.0.0", agreementSign.AgreementIDs[0])
	assert.Equal(t, userId, agreementSign.IdentityID)
	assert.Equal(t, userIp, agreementSign.IPAddress)
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
