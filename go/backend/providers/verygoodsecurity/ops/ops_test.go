package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/verygoodsecurity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCard(s *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("can create a vgs card", func(t *testing.T) {
		userId := uuid.NewString()
		wallet, err := c.Users().CreateNewWallet(ctx, userId, "")
		if err != nil {
			t.Fatal(err)
		}

		card, err := c.verygoodsecurity.CreateCard(ctx, verygoodsecurity.Card{
			WalletID: wallet.ID,
			Token:    "4111112436781111",
			Expiry:   "some_token_2937648273",
			CVV:      "some_token_7281687254",
			Last4:    "1111",
			Type:     "visa",
		})
		require.NoError(t, err)
		assert.NotNil(t, card)
		assert.NotNil(t, card.ID)
		assert.Equal(t, wallet.ID, card.WalletID)
		assert.Equal(t, "4111112436781111", card.Token)
		assert.Equal(t, "some_token_2937648273", card.Expiry)
		assert.Equal(t, "some_token_7281687254", card.CVV)
		assert.Equal(t, "1111", card.Last4)
		assert.Equal(t, "visa", card.Type)
	})

	s.Run("updates duplicates", func(t *testing.T) {
		userId := uuid.NewString()
		// Create Wallet
		wallet, err := c.Users().CreateNewWallet(ctx, userId, "")
		if err != nil {
			t.Fatal(err)
		}

		card, err := c.verygoodsecurity.CreateCard(ctx, verygoodsecurity.Card{
			WalletID: wallet.ID,
			Token:    "4111112436781111",
			Expiry:   "some_token_2937648273",
			CVV:      "some_token_7281687254",
			Last4:    "1111",
			Type:     "visa",
		})
		require.NoError(t, err)
		assert.NotNil(t, card)

		updatedCard, err := c.verygoodsecurity.CreateCard(ctx, verygoodsecurity.Card{
			WalletID: wallet.ID,
			Token:    "4111112436781111",
			Expiry:   "some_new_token_2937648273",
			CVV:      "some_new_token_7281687254",
			Last4:    "1111",
			Type:     "visa",
		})
		require.NoError(t, err)
		assert.Equal(t, "4111112436781111", updatedCard.Token)
		assert.Equal(t, card.ID, updatedCard.ID)
		assert.Equal(t, "some_new_token_2937648273", updatedCard.Expiry)
		assert.Equal(t, "some_new_token_7281687254", updatedCard.CVV)
		assert.Equal(t, "1111", updatedCard.Last4)
		assert.Equal(t, "visa", updatedCard.Type)
	})
}
