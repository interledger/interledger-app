package ops_test

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type backends struct {
	validator        *validator.Validate
	db               *sqlx.DB
	verygoodsecurity verygoodsecurity.Client
}

func (b backends) Validator() *validator.Validate {
	return b.validator
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) VGS() verygoodsecurity.Client {
	return b.verygoodsecurity
}

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

	s.Run("can't create duplicates on walletId and card token", func(t *testing.T) {
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

		_, err = c.verygoodsecurity.CreateCard(ctx, verygoodsecurity.Card{
			WalletID: wallet.ID,
			Token:    "4111112436781111",
			Expiry:   "some_token_2937648273",
			CVV:      "some_token_7281687254",
			Last4:    "1111",
			Type:     "visa",
		})
		assert.ErrorContains(t, err, "internal error pq: duplicate key value violates unique constraint \"card_token_wallet_id_uniq\"")
		assert.NotNil(t, err)

		_, err = c.verygoodsecurity.CreateCard(ctx, verygoodsecurity.Card{
			WalletID: wallet.ID,
			Token:    "4111112436781111",
			Expiry:   "changing",
			CVV:      "these",
			Last4:    "still",
			Type:     "errors",
		})
		assert.ErrorContains(t, err, "internal error pq: duplicate key value violates unique constraint \"card_token_wallet_id_uniq\"")
		assert.NotNil(t, err)

		_, err = c.verygoodsecurity.CreateCard(ctx, verygoodsecurity.Card{
			WalletID: wallet.ID,
			Token:    "41111123455641111",
			Expiry:   "will",
			CVV:      "yield",
			Last4:    "positive",
			Type:     "results",
		})
		require.NoError(t, err)
	})
}
