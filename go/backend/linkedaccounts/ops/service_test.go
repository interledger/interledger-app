package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/user"

	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLinkedAccounts(s *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("can create a linked account", func(t *testing.T) {
		userId := uuid.NewString()
		// Create Wallet
		wallet, err := c.Users().CreateNewWallet(ctx, user.CreateWalletArgs{
			UserID: userId,
			Name:   "",
		})
		if err != nil {
			t.Fatal(err)
		}

		linkedAccount, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
			WalletID: wallet.ID,
			Name:     "Test",
			Mask:     "1234",
			Provider: "mx",
			Type:     "bank",
		})
		require.NoError(t, err)

		assert.NotNil(t, linkedAccount)
		assert.Equal(t, linkedAccount.Provider, "mx")
		assert.Equal(t, linkedAccount.ProviderID, "")
		assert.Equal(t, linkedAccount.Type, "bank")
		assert.Equal(t, linkedAccount.WalletID, wallet.ID)
	})

	s.Run("can get a linked account", func(t *testing.T) {
		userId := uuid.NewString()
		// Create Wallet
		wallet, err := c.Users().CreateNewWallet(ctx, user.CreateWalletArgs{
			UserID: userId,
			Name:   "",
		})
		if err != nil {
			t.Fatal(err)
		}
		linkedAccount, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
			WalletID:   wallet.ID,
			Name:       "Test",
			Mask:       "1234",
			Provider:   "mx",
			ProviderID: "123",
			Type:       "bank",
		})
		require.NoError(t, err)

		fs, err := c.LinkedAccounts.Get(ctx, linkedAccount.ID)

		require.NoError(t, err)
		assert.NotNil(t, fs)
		assert.Equal(t, fs.ID, linkedAccount.ID)
		assert.Equal(t, fs.WalletID, wallet.ID)
		assert.Equal(t, "123", fs.ProviderID)

		laByProviderID, err := c.LinkedAccounts.GetByProviderID(ctx, linkedaccounts.GetByProviderIDArgs{
			Provider:   "mx",
			ProviderID: "123",
		})
		require.NoError(t, err)
		assert.Equal(t, linkedAccount.ID, laByProviderID.ID)
		assert.Equal(t, wallet.ID, laByProviderID.WalletID)
	})

	s.Run("can get a list of wallet linked accounts", func(t *testing.T) {
		userId := uuid.NewString()
		// Create Wallet
		wallet, err := c.Users().CreateNewWallet(ctx, user.CreateWalletArgs{
			UserID: userId,
			Name:   "",
		})
		if err != nil {
			t.Fatal(err)
		}
		linkedAccount, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
			WalletID: wallet.ID,
			Name:     "Test",
			Mask:     "1234",
			Provider: "mx",
			Type:     "bank",
		})
		require.NoError(t, err)

		linkedAccounts, err := c.LinkedAccounts.ListByWalletId(ctx, wallet.ID)
		require.NoError(t, err)

		assert.NotNil(t, linkedAccounts)
		assert.Len(t, linkedAccounts, 1)
		la := linkedAccounts[0]
		assert.NotNil(t, la)
		assert.Equal(t, la.ID, linkedAccount.ID)
		assert.Equal(t, la.WalletID, wallet.ID)
	})
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, t)
	require.NoError(t, err)

	userId := uuid.NewString()
	// Create Wallet
	wallet, err := c.Users().CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: userId,
		Name:   "",
	})
	if err != nil {
		t.Fatal(err)
	}
	la, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID: wallet.ID,
		Name:     "Test",
		Mask:     "1234",
		Provider: "mx",
		Type:     "bank",
	})
	require.NoError(t, err)

	_, err = c.LinkedAccounts.Get(ctx, la.ID)
	require.NoError(t, err)

	err = c.LinkedAccounts.Delete(ctx, la.ID)
	require.NoError(t, err)

	_, err = c.LinkedAccounts.Get(ctx, la.ID)
	assert.ErrorIs(t, err, linkedaccounts.ErrNotFound)

	la, err = c.LinkedAccounts.MarkNotDeleted(ctx, la.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test", la.Name)

	_, err = c.LinkedAccounts.Get(ctx, la.ID)
	require.NoError(t, err)

	_, err = c.LinkedAccounts.MarkNotDeleted(ctx, uuid.NewString())
	assert.ErrorIs(t, err, linkedaccounts.ErrNotFound)
}

func TestListMXBankAccounts(t *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, t)
	require.NoError(t, err)

	userId := uuid.NewString()
	// Create Wallet
	wallet, err := c.Users().CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: userId,
		Name:   "",
	})
	if err != nil {
		t.Fatal(err)
	}

	linkedAccounts, err := c.LinkedAccounts.CreateBatch(ctx, []linkedaccounts.CreateArgs{
		{
			WalletID:   wallet.ID,
			Name:       "Test",
			Nickname:   "TestNickname",
			Mask:       "1234",
			Provider:   mx.ProviderName,
			Type:       mx.TypeBankAccount,
			ProviderID: "2345",
		},
		{
			WalletID:   wallet.ID,
			Name:       "Test2",
			Nickname:   "Test2Nickname",
			Mask:       "4321",
			Provider:   mx.ProviderName,
			Type:       mx.TypeBankAccount,
			ProviderID: "5432",
		},
	})
	require.NoError(t, err)
	require.Len(t, linkedAccounts, 2)

	mxBankAccounts, err := c.LinkedAccounts.ListMXBankAccounts(ctx)
	require.NoError(t, err)
	require.Len(t, mxBankAccounts, 2)
	for _, la := range mxBankAccounts {
		if la.ProviderID == "2345" {
			assert.Equal(t, "Test", la.Name)
			assert.Equal(t, "TestNickname", la.Nickname)
			assert.Equal(t, "1234", la.Mask)
		} else {
			assert.Equal(t, "Test2", la.Name)
			assert.Equal(t, "Test2Nickname", la.Nickname)
			assert.Equal(t, "4321", la.Mask)
		}
	}
}

func TestSetNickname(s *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, s)
	require.NoError(s, err)

	userId := uuid.NewString()
	// Create Wallet
	wallet, err := c.Users().CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: userId,
		Name:   "",
	})
	if err != nil {
		s.Fatal(err)
	}
	la, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID: wallet.ID,
		Name:     "Test",
		Mask:     "1234",
		Provider: "mx",
		Type:     "bank",
	})
	require.NoError(s, err)
	assert.Equal(s, "", la.Nickname)

	s.Run("can set nickname", func(t *testing.T) {
		la, err = c.LinkedAccounts.SetNickname(ctx, la.ID, "New name")
		require.NoError(t, err)

		assert.Equal(t, "New name", la.Nickname)
	})

	s.Run("returns error if no account found", func(t *testing.T) {
		rid := uuid.NewString()
		la, err = c.LinkedAccounts.SetNickname(ctx, rid, "New name")
		require.ErrorIs(t, err, linkedaccounts.ErrNotFound)
	})
}
