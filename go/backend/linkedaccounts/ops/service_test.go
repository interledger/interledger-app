package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/providers/pti"

	"gitlab.com/fynbos/backend/linkedaccounts/ops"
	"gitlab.com/fynbos/backend/wallets"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
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
	c.Ec.EXPECT().SendConnectedAccountEmail(ctx, gomock.Any()).AnyTimes()
	s.Run("can create a linked account under review", func(t *testing.T) {
		walletID := uuid.NewString()
		c.Ec.EXPECT().SendConnectedAccountDocumentsNeededEmail(ctx, gomock.Any()).Times(1)
		linkedAccount, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
			WalletID:   walletID,
			Name:       "Test",
			Mask:       "1234",
			Provider:   pti.ProviderName,
			Type:       pti.TypeCard,
			CanSend:    true,
			CanReceive: true,
			State:      linkedaccounts.OwnershipReviewRequired,
		})
		require.NoError(t, err)

		assert.NotNil(t, linkedAccount)
		assert.Equal(t, linkedAccount.Provider, pti.ProviderName)
		assert.Equal(t, linkedAccount.ProviderID, "")
		assert.Equal(t, linkedAccount.Type, pti.TypeCard)
		assert.Equal(t, linkedAccount.WalletID, walletID)
		assert.True(t, linkedAccount.CanSend)
		assert.True(t, linkedAccount.CanReceive)
		assert.Equal(t, linkedAccount.State, linkedaccounts.OwnershipReviewRequired)
	})

	s.Run("can create a linked account", func(t *testing.T) {
		walletID := uuid.NewString()
		linkedAccount, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
			WalletID:   walletID,
			Name:       "Test",
			Mask:       "1234",
			Provider:   pti.ProviderName,
			Type:       pti.AccTypeBalance,
			CanSend:    true,
			CanReceive: true,
			State:      linkedaccounts.Verified,
		})
		require.NoError(t, err)

		assert.NotNil(t, linkedAccount)
		assert.Equal(t, linkedAccount.Provider, pti.ProviderName)
		assert.Equal(t, linkedAccount.ProviderID, "")
		assert.Equal(t, linkedAccount.Type, pti.AccTypeBalance)
		assert.Equal(t, linkedAccount.WalletID, walletID)
		assert.True(t, linkedAccount.CanSend)
		assert.True(t, linkedAccount.CanReceive)
		assert.Equal(t, linkedAccount.State, linkedaccounts.Verified)
	})

	s.Run("can get a linked account", func(t *testing.T) {
		walletID := uuid.NewString()
		linkedAccount, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
			WalletID:   walletID,
			Name:       "Test",
			Mask:       "1234",
			Provider:   pti.ProviderName,
			ProviderID: "123",
			Type:       pti.AccTypeBalance,
			CanSend:    true,
			CanReceive: true,
		})
		require.NoError(t, err)

		fs, err := c.LinkedAccounts.Get(ctx, linkedAccount.ID)

		require.NoError(t, err)
		assert.NotNil(t, fs)
		assert.Equal(t, fs.ID, linkedAccount.ID)
		assert.Equal(t, fs.WalletID, walletID)
		assert.Equal(t, "123", fs.ProviderID)

		laByProviderID, err := c.LinkedAccounts.GetByProviderID(ctx, linkedaccounts.GetByProviderIDArgs{
			Provider:   pti.ProviderName,
			ProviderID: "123",
			WalletID:   walletID,
		})
		require.NoError(t, err)
		assert.Equal(t, linkedAccount.ID, laByProviderID.ID)
		assert.Equal(t, walletID, laByProviderID.WalletID)
		assert.True(t, laByProviderID.CanSend)
		assert.True(t, laByProviderID.CanReceive)
		assert.Equal(t, linkedAccount.State, linkedaccounts.Verified)
	})

	s.Run("can get a list of wallet linked accounts", func(t *testing.T) {
		walletID := uuid.NewString()

		linkedAccount, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
			WalletID:   walletID,
			Name:       "Test",
			Mask:       "1234",
			Provider:   pti.ProviderName,
			Type:       pti.AccTypeBalance,
			CanSend:    true,
			CanReceive: true,
		})
		require.NoError(t, err)

		linkedAccounts, err := c.LinkedAccounts.ListByWalletId(ctx, walletID)
		require.NoError(t, err)

		assert.NotNil(t, linkedAccounts)
		assert.Len(t, linkedAccounts, 1)
		la := linkedAccounts[0]
		assert.NotNil(t, la)
		assert.Equal(t, la.ID, linkedAccount.ID)
		assert.Equal(t, la.WalletID, walletID)
		assert.True(t, la.CanSend)
		assert.True(t, la.CanReceive)
	})
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, t)
	require.NoError(t, err)
	c.Ec.EXPECT().SendConnectedAccountEmail(ctx, gomock.Any()).AnyTimes()

	walletID := uuid.NewString()
	la, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID: walletID,
		Name:     "Test",
		Mask:     "1234",
		Provider: pti.ProviderName,
		Type:     pti.AccTypeBalance,
	})
	require.NoError(t, err)

	la, err = c.LinkedAccounts.Get(ctx, la.ID)
	require.NoError(t, err)
	assert.False(t, la.DeletedAt.Valid)

	err = c.LinkedAccounts.Delete(ctx, la.ID)
	require.NoError(t, err)

	la, err = c.LinkedAccounts.Get(ctx, la.ID)
	require.NoError(t, err)
	assert.True(t, la.DeletedAt.Valid)

	la, err = c.LinkedAccounts.MarkNotDeleted(ctx, la.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test", la.Name)
	assert.False(t, la.DeletedAt.Valid)

	la, err = c.LinkedAccounts.Get(ctx, la.ID)
	require.NoError(t, err)
	assert.False(t, la.DeletedAt.Valid)

	_, err = c.LinkedAccounts.MarkNotDeleted(ctx, uuid.NewString())
	assert.ErrorIs(t, err, linkedaccounts.ErrNotFound)
}

func TestSetNickname(s *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, s)
	require.NoError(s, err)
	c.Ec.EXPECT().SendConnectedAccountEmail(ctx, gomock.Any()).AnyTimes()
	walletID := uuid.NewString()

	la, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID: walletID,
		Name:     "Test",
		Mask:     "1234",
		Provider: pti.ProviderName,
		Type:     pti.AccTypeBalance,
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

func TestReviews(t *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, t)
	require.NoError(t, err)
	c.Ec.EXPECT().SendConnectedAccountEmail(ctx, gomock.Any()).AnyTimes()
	walletID := uuid.NewString()
	c.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{
		ID:   walletID,
		Name: "Test Wallet",
	}, nil).AnyTimes()

	la, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID: walletID,
		Name:     "Test",
		Mask:     "1234",
		Provider: pti.ProviderName,
		Type:     pti.AccTypeBalance,
	})
	require.NoError(t, err)

	reviews, err := c.LinkedAccounts.CreateReviews(ctx, []linkedaccounts.CreateReviewArgs{
		{LinkedAccountID: la.ID, State: linkedaccounts.OwnershipReviewRequired},
	})
	require.NoError(t, err)
	require.Len(t, reviews, 1)
	assert.Equal(t, reviews[0].LinkedAccountID, la.ID)
	assert.Equal(t, reviews[0].State, linkedaccounts.OwnershipReviewRequired)
	assert.Empty(t, reviews[0].NewState)
	assert.Empty(t, reviews[0].ReviewedBy)

	review, err := c.LinkedAccounts.CompleteReview(ctx, linkedaccounts.CompleteReviewArgs{
		ID:         reviews[0].ID,
		ReviewedBy: "test@interledger.test",
		NewState:   linkedaccounts.Verified,
		Reason:     "Manual check passed.",
	})
	require.NoError(t, err)
	assert.Equal(t, "Manual check passed.", review.Reason)
	assert.Equal(t, "test@interledger.test", review.ReviewedBy)
	assert.Equal(t, linkedaccounts.Verified, review.NewState)
	assert.NotEmpty(t, review.CompletedAt)

	review, err = c.LinkedAccounts.GetReview(ctx, reviews[0].ID)
	require.NoError(t, err)
	assert.Equal(t, "Manual check passed.", review.Reason)
	assert.Equal(t, "test@interledger.test", review.ReviewedBy)
	assert.Equal(t, linkedaccounts.OwnershipReviewRequired, review.State)
	assert.Equal(t, linkedaccounts.Verified, review.NewState)
	assert.Equal(t, la.Mask, review.LinkedAccountMask)
	assert.NotEmpty(t, review.CompletedAt)
}

func TestListReviews(t *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, t)
	require.NoError(t, err)
	c.Ec.EXPECT().SendConnectedAccountEmail(ctx, gomock.Any()).AnyTimes()
	walletID := uuid.NewString()
	c.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{
		ID:   walletID,
		Name: "Test Wallet",
	}, nil).AnyTimes()

	la, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID: walletID,
		Name:     "Test",
		Mask:     "1234",
		Provider: pti.ProviderName,
		Type:     pti.AccTypeBalance,
	})
	require.NoError(t, err)

	args := []linkedaccounts.CreateReviewArgs{
		{LinkedAccountID: la.ID, State: linkedaccounts.OwnershipReviewRequired},
		{LinkedAccountID: la.ID, State: linkedaccounts.OwnershipReviewRequired},
		{LinkedAccountID: la.ID, State: linkedaccounts.OwnershipReviewRequired},
		{LinkedAccountID: la.ID, State: linkedaccounts.Rejected},
	}
	for _, r := range args {
		_, err = c.LinkedAccounts.CreateReviews(ctx, []linkedaccounts.CreateReviewArgs{r})
		require.NoError(t, err)
	}

	reviews, err := c.LinkedAccounts.ListReviews(ctx, db.Pagination{})
	require.NoError(t, err)
	assert.Len(t, reviews, 4)

	reviews, err = c.LinkedAccounts.ListReviews(ctx, db.Pagination{PageToken: reviews[0].ID})
	require.NoError(t, err)
	assert.Len(t, reviews, 3)

	_, err = c.LinkedAccounts.CompleteReview(ctx, linkedaccounts.CompleteReviewArgs{
		ID:         reviews[0].ID,
		ReviewedBy: "test@interledger.test",
		NewState:   linkedaccounts.Verified,
	})
	require.NoError(t, err)

	reviews, err = c.LinkedAccounts.ListIncompleteReviews(ctx, db.Pagination{})
	require.NoError(t, err)
	assert.Len(t, reviews, 3)
}

func TestDefaultSendReceive(t *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, t)
	require.NoError(t, err)
	c.Ec.EXPECT().SendConnectedAccountEmail(ctx, gomock.Any()).AnyTimes()
	walletID := uuid.NewString()
	c.Wc.EXPECT().Get(ctx, walletID).Return(&wallets.Wallet{
		ID:   walletID,
		Name: "Test Wallet",
	}, nil).AnyTimes()

	la1, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:   walletID,
		Name:       "Test",
		Mask:       "1234",
		Provider:   pti.ProviderName,
		ProviderID: "1234",
		Type:       pti.AccTypeBalance,
		State:      linkedaccounts.Verified,
		CanSend:    true,
		CanReceive: true,
	})
	require.NoError(t, err)
	assert.False(t, la1.DefaultReceive)
	assert.False(t, la1.DefaultSend)

	la2, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:   walletID,
		Name:       "Test2",
		Mask:       "4321",
		Provider:   pti.ProviderName,
		ProviderID: "4321",
		Type:       pti.AccTypeBalance,
		State:      linkedaccounts.Verified,
		CanSend:    true,
		CanReceive: true,
	})
	require.NoError(t, err)
	assert.False(t, la2.DefaultReceive)
	assert.False(t, la2.DefaultSend)

	defaultSend, err := ops.SetDefaultSend(ctx, c, la1.ID)
	require.NoError(t, err)
	assert.True(t, defaultSend.DefaultSend)

	defaultReceive, err := ops.SetDefaultReceive(ctx, c, la1.ID)
	require.NoError(t, err)
	assert.True(t, defaultReceive.DefaultReceive)

	las, err := ops.ListByWalletId(ctx, c, walletID)
	require.NoError(t, err)
	for _, la := range las {
		if la.ID == la1.ID {
			assert.True(t, la.DefaultSend)
			assert.True(t, la.DefaultReceive)
		} else {
			assert.False(t, la.DefaultSend)
			assert.False(t, la.DefaultReceive)
		}
	}

	defaultSend, err = ops.SetDefaultSend(ctx, c, la2.ID)
	require.NoError(t, err)
	assert.True(t, defaultSend.DefaultSend)

	defaultReceive, err = ops.SetDefaultReceive(ctx, c, la2.ID)
	require.NoError(t, err)
	assert.True(t, defaultReceive.DefaultReceive)

	las, err = ops.ListByWalletId(ctx, c, walletID)
	require.NoError(t, err)
	for _, la := range las {
		if la.ID == la2.ID {
			assert.True(t, la.DefaultSend)
			assert.True(t, la.DefaultReceive)
		} else {
			assert.False(t, la.DefaultSend)
			assert.False(t, la.DefaultReceive)
		}
	}
}
