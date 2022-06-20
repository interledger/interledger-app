package rafiki

import (
	"context"
	"flag"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/backend/utils"
)

var runIntegration = flag.Bool("integration", false, "Bool to run integration tests against Rafiki graphql server.")

func TestIntegration(t *testing.T) {
	if !*runIntegration {
		t.Skip()
	}
	ctx := context.Background()
	db := test_utils.MigrateCockroachDB(t, ctx)
	rafiki, err := NewService(&ServiceArgs{
		Db:  db,
		Url: "http://localhost:3001/graphql",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = rafiki.CreateIdentifier(ctx, &CreateIdentifierArgs{
		AccountID:  uuid.NewString(),
		AssetCode:  "740",
		AssetScale: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateIdentifier(t *testing.T) {
	ctx := context.Background()
	mockGraphqlServer := NewMockRafikiGraphqlServer(t)
	db := test_utils.MigrateCockroachDB(t, ctx)
	rafiki, err := NewService(&ServiceArgs{
		Db:  db,
		Url: mockGraphqlServer.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	accountID := uuid.NewString()

	identifier, err := rafiki.CreateIdentifier(ctx, &CreateIdentifierArgs{
		AccountID:  accountID,
		AssetCode:  "740",
		AssetScale: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, accountID, identifier.AccountID)
	_, err = uuid.Parse(identifier.ID)
	assert.NoError(t, err)

	freshIdentifer, err := rafiki.GetIdentifier(ctx, identifier.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, identifier.ID, freshIdentifer.ID)
	assert.Equal(t, accountID, freshIdentifer.AccountID)
	assert.Equal(t, "740", freshIdentifer.AssetCode)
	assert.Equal(t, uint8(2), freshIdentifer.AssetScale)
}
