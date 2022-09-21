package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/fundingsources"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFundingSources(s *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("can create a funding source", func(t *testing.T) {
		userId := uuid.NewString()
		wallet, err := c.Users().CreateNewWallet(ctx, userId, "")
		if err != nil {
			t.Fatal(err)
		}

		fs, err := c.Fs.Create(ctx, &fundingsources.CreateArgs{
			WalletID: wallet.ID,
			Name:     "Test",
			Mask:     "1234",
			Provider: "mx",
			Type:     "bank",
		})
		if err != nil {
			t.Fatal("Should be able to create a funding source")
		}

		assert.NotNil(t, fs)
		assert.Equal(t, fs.Provider, "mx")
		assert.Equal(t, fs.Type, "bank")
		assert.Equal(t, fs.WalletId, wallet.ID)
	})

	s.Run("can get a funding source", func(t *testing.T) {
		userId := uuid.NewString()
		wallet, err := c.Users().CreateNewWallet(ctx, userId, "")
		if err != nil {
			t.Fatal(err)
		}
		fsCreated, err := c.Fs.Create(ctx, &fundingsources.CreateArgs{
			WalletID: wallet.ID,
			Name:     "Test",
			Mask:     "1234",
			Provider: "mx",
			Type:     "bank",
		})
		if err != nil {
			t.Fatal("Should be able to create a funding source")
		}

		fs, err := c.Fs.Get(ctx, fsCreated.ID)

		assert.NotNil(t, fs)
		assert.Equal(t, fs.ID, fsCreated.ID)
		assert.Equal(t, fs.WalletId, wallet.ID)
	})

	s.Run("can get a list of wallet funding sources", func(t *testing.T) {
		userId := uuid.NewString()
		wallet, err := c.Users().CreateNewWallet(ctx, userId, "")
		if err != nil {
			t.Fatal(err)
		}
		fsCreated, err := c.Fs.Create(ctx, &fundingsources.CreateArgs{
			WalletID: wallet.ID,
			Name:     "Test",
			Mask:     "1234",
			Provider: "mx",
			Type:     "bank",
		})
		if err != nil {
			t.Fatal("Should be able to create a funding source")
		}

		fses, err := c.Fs.ListByWalletId(ctx, wallet.ID)

		assert.NotNil(t, fses)
		assert.Len(t, fses, 1)
		fs := fses[0]
		assert.NotNil(t, fs)
		assert.Equal(t, fs.ID, fsCreated.ID)
		assert.Equal(t, fs.WalletId, wallet.ID)
	})
}
