package ops_test

import (
	"context"
	"testing"

	"github.com/Basis-Theory/basistheory-go/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/basistheory/external"
	"gitlab.com/fynbos/backend/providers/basistheory/external/client/mock"
	"gitlab.com/fynbos/backend/providers/basistheory/ops"
	"gotest.tools/assert"
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

	tokenID, walletID := uuid.NewString(), uuid.NewString()
	b.ExternalClient.EXPECT().GetToken(ctx, tokenID).Return(
		&basistheory.Token{
			Id: &tokenID,
			Data: &external.CardData{
				TokenizedNumber: "1234",
				ExpirationMonth: "02",
				ExpirationYear:  "2024",
			},
			Type: basistheory.PtrString("card"),
		},
		nil,
	).Times(1)

	card, err := ops.CreateCard(ctx, b, tokenID, walletID)
	require.NoError(t, err)
	assert.Equal(t, tokenID, card.TokenID)
	assert.Equal(t, "1234", card.TokenizedNumber)
	assert.Equal(t, "02", card.ExpirationMonth)
	assert.Equal(t, "2024", card.ExpirationYear)

	idempotentCard, err := ops.CreateCard(ctx, b, tokenID, walletID)
	require.NoError(t, err)
	assert.Equal(t, tokenID, idempotentCard.TokenID)
	assert.Equal(t, card.ID, idempotentCard.ID)
}
