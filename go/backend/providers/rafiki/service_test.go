package rafiki

import (
	"context"
	"flag"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/backend/utils"
)

var runIntegration = flag.Bool("integration", false, "Bool to run integration tests against Rafiki graphql server.")

func TestIntegration(t *testing.T) {
	t.Parallel()
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

	identifier, err := rafiki.CreateIdentifier(ctx, &CreateIdentifierArgs{
		AccountID:  uuid.NewString(),
		AssetCode:  "740",
		AssetScale: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = rafiki.CreateQuote(ctx, &CreateQuoteArgs{
		IdentifierID:   identifier.ID,
		Receiver:       "$ilp.test/123",
		SendAssetCode:  "740",
		SendAssetScale: 2,
		SendAmount:     100,
	})

	// just checking that passes validation for now.
	if err == nil {
		t.Fatal(err)
	}
	assert.True(t, strings.Contains(err.Error(), "invalid destination"))
}

func TestCreateIdentifier(t *testing.T) {
	t.Parallel()
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

func TestCreateQuote(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mockGraphqlServer := NewMockRafikiGraphqlServer(t)
	rafiki, err := NewService(&ServiceArgs{
		Db:  &sqlx.DB{},
		Url: mockGraphqlServer.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	identifierID := uuid.NewString()
	// values below are hard-coded in mock graphql server
	expectedAssetCode := "740"
	expectedAssetScale := uint8(2)
	expectedSendAmount := uint64(100)
	expectedReceiveAmount := uint64(99)
	expectedExchangeRate := float64(1.00)
	receiver := "$ilp.test/receiver"
	cases := []struct {
		Name          string
		Args          *CreateQuoteArgs
		ExpectedError error
	}{
		{
			Name:          "Send asset details required if SendAmount specified",
			ExpectedError: ErrInvalidArgument,
			Args: &CreateQuoteArgs{
				IdentifierID: identifierID,
				Receiver:     receiver,
				SendAmount:   100,
			},
		},
		{
			Name:          "Receive asset details required if SendAmount specified",
			ExpectedError: ErrInvalidArgument,
			Args: &CreateQuoteArgs{
				IdentifierID:  identifierID,
				Receiver:      receiver,
				ReceiveAmount: 100,
			},
		},
		{
			Name:          "Either SendAmount or ReceiveAmount must be specified",
			ExpectedError: ErrInvalidArgument,
			Args: &CreateQuoteArgs{
				IdentifierID: identifierID,
				Receiver:     receiver,
			},
		},
		{
			Name: "Can specify SendAmount only",
			Args: &CreateQuoteArgs{
				IdentifierID:   identifierID,
				Receiver:       receiver,
				SendAssetCode:  expectedAssetCode,
				SendAssetScale: expectedAssetScale,
				SendAmount:     expectedSendAmount,
			},
		},
		{
			Name: "Can specify ReceiveAmount only",
			Args: &CreateQuoteArgs{
				IdentifierID:      identifierID,
				Receiver:          receiver,
				ReceiveAssetCode:  expectedAssetCode,
				ReceiveAssetScale: expectedAssetScale,
				ReceiveAmount:     expectedReceiveAmount,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(st *testing.T) {
			quote, err := rafiki.CreateQuote(ctx, tc.Args)

			if tc.ExpectedError == nil {
				assert.NoError(st, err, tc.Name)
				assert.Equal(t, expectedSendAmount, quote.SendAmount)
				assert.Equal(t, expectedAssetScale, quote.SendAssetScale)
				assert.Equal(t, expectedAssetCode, quote.SendAssetCode)

				assert.Equal(t, expectedReceiveAmount, quote.ReceiveAmount)
				assert.Equal(t, expectedAssetScale, quote.ReceiveAssetScale)
				assert.Equal(t, expectedAssetCode, quote.ReceiveAssetCode)
				assert.Equal(t, expectedExchangeRate, quote.MinExchangeRate)
				assert.Equal(t, expectedExchangeRate, quote.LowEstimatedExchangeRate)
				assert.Equal(t, expectedExchangeRate, quote.HighEstimatedExchangeRate)
			} else {
				assert.ErrorIs(st, err, tc.ExpectedError, tc.Name)
				assert.Nil(st, quote, tc.Name)
			}
		})
	}
}
