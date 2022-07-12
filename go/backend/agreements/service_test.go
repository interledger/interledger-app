package agreements

import (
	"context"
	"testing"

	"github.com/google/uuid"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestSignAgreements(t *testing.T) {
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	as, err := NewService(&ServiceArgs{
		Db: db,
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

	t.Log(signedAgreements)
}
