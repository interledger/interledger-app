package ops_test

import (
	"context"
	"testing"

	"github.com/Basis-Theory/basistheory-go/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	fynbos_basistheory "gitlab.com/fynbos/backend/providers/basistheory"
	"gitlab.com/fynbos/backend/providers/basistheory/external/client/mock"
	"gitlab.com/fynbos/backend/providers/basistheory/ops"
)

func TestCreateCard(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	b := &ops.TestBackends{
		Db:             db.MigrateTestDB(t, ctx),
		ExternalClient: mock.NewMockClient(ctrl),
	}

	tokenID, walletID, fingerprint := uuid.NewString(), uuid.NewString(), uuid.NewString()
	b.ExternalClient.EXPECT().GetToken(ctx, tokenID).Return(
		&basistheory.Token{
			Id: &tokenID,
			Data: map[string]interface{}{
				"number":           "1234",
				"expiration_month": float64(2),
				"expiration_year":  float64(2024),
			},
			Type:        basistheory.PtrString("card"),
			Fingerprint: *basistheory.NewNullableString(&fingerprint),
		},
		nil,
	).AnyTimes()

	card, err := ops.CreateCard(ctx, b, fynbos_basistheory.CreateCardArgs{
		WalletID:         walletID,
		TokenID:          tokenID,
		Bin:              "5678",
		PullNetwork:      "VISA",
		PullEnabled:      true,
		PullType:         "Debit",
		PullCountry:      "US",
		PushNetwork:      "Mastercard",
		PushEnabled:      true,
		PushType:         "Credit",
		PushAvailability: "Immediate",
		PushCountry:      "US",
	})
	require.NoError(t, err)
	assert.Equal(t, tokenID, card.TokenID)
	assert.Equal(t, "1234", card.TokenizedNumber)
	assert.Equal(t, "02", card.ExpirationMonth)
	assert.Equal(t, "2024", card.ExpirationYear)
	assert.Equal(t, fingerprint, card.Fingerprint)
	assert.Equal(t, "5678", card.Bin)
	assert.Equal(t, "VISA", card.PullNetwork)
	assert.True(t, card.PullEnabled)
	assert.Equal(t, "Debit", card.PullType)
	assert.Equal(t, "US", card.PullCountry)
	assert.Equal(t, "Mastercard", card.PushNetwork)
	assert.True(t, card.PushEnabled)
	assert.Equal(t, "Credit", card.PushType)
	assert.Equal(t, "Immediate", card.PushAvailability)
	assert.Equal(t, "US", card.PushCountry)
	assert.Equal(t, fingerprint, card.Fingerprint)

	idempotentCard, err := ops.CreateCard(ctx, b, fynbos_basistheory.CreateCardArgs{
		WalletID: walletID,
		TokenID:  tokenID,
	})
	require.NoError(t, err)
	assert.Equal(t, tokenID, idempotentCard.TokenID)
	assert.Equal(t, card.ID, idempotentCard.ID)
	assert.Equal(t, fingerprint, idempotentCard.Fingerprint)
}
