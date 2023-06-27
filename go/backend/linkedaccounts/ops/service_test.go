package ops_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/mx"

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
		walletID := uuid.NewString()

		linkedAccount, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
			WalletID:   walletID,
			Name:       "Test",
			Mask:       "1234",
			Provider:   "mx",
			Type:       "bank",
			CanSend:    true,
			CanReceive: true,
			State:      linkedaccounts.OwnershipReviewRequired,
		})
		require.NoError(t, err)

		assert.NotNil(t, linkedAccount)
		assert.Equal(t, linkedAccount.Provider, "mx")
		assert.Equal(t, linkedAccount.ProviderID, "")
		assert.Equal(t, linkedAccount.Type, "bank")
		assert.Equal(t, linkedAccount.WalletID, walletID)
		assert.True(t, linkedAccount.CanSend)
		assert.True(t, linkedAccount.CanReceive)
		assert.Equal(t, linkedAccount.State, linkedaccounts.OwnershipReviewRequired)
	})

	s.Run("can get a linked account", func(t *testing.T) {
		walletID := uuid.NewString()
		linkedAccount, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
			WalletID:   walletID,
			Name:       "Test",
			Mask:       "1234",
			Provider:   "mx",
			ProviderID: "123",
			Type:       "bank",
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
			Provider:   "mx",
			ProviderID: "123",
			WalletID:   wallet.ID,
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
			Provider:   "mx",
			Type:       "bank",
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

	walletID := uuid.NewString()
	la, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID: walletID,
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

	walletID := uuid.NewString()

	linkedAccounts, err := c.LinkedAccounts.CreateBatch(ctx, []linkedaccounts.CreateArgs{
		{
			WalletID:   walletID,
			Name:       "Test",
			Nickname:   "TestNickname",
			Mask:       "1234",
			Provider:   mx.ProviderName,
			Type:       mx.TypeBankAccount,
			ProviderID: "2345",
			CanSend:    true,
			CanReceive: true,
		},
		{
			WalletID:   walletID,
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
			assert.True(t, la.CanReceive)
			assert.True(t, la.CanSend)
		} else {
			assert.Equal(t, "Test2", la.Name)
			assert.Equal(t, "Test2Nickname", la.Nickname)
			assert.Equal(t, "4321", la.Mask)
			assert.False(t, la.CanReceive)
			assert.False(t, la.CanSend)
		}
	}
}

func TestSetNickname(s *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, s)
	require.NoError(s, err)

	walletID := uuid.NewString()

	la, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID: walletID,
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

func TestReviews(t *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, t)
	require.NoError(t, err)

	walletID := uuid.NewString()
	la, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID: walletID,
		Name:     "Test",
		Mask:     "1234",
		Provider: "mx",
		Type:     "bank",
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
		ReviewedBy: "test@fynbos.dev",
		NewState:   linkedaccounts.Verified,
		Reason:     "Manual check passed.",
	})
	require.NoError(t, err)
	assert.Equal(t, "Manual check passed.", review.Reason)
	assert.Equal(t, "test@fynbos.dev", review.ReviewedBy)
	assert.Equal(t, linkedaccounts.Verified, review.NewState)
	assert.NotEmpty(t, review.CompletedAt)

	review, err = c.LinkedAccounts.GetReview(ctx, reviews[0].ID)
	require.NoError(t, err)
	assert.Equal(t, "Manual check passed.", review.Reason)
	assert.Equal(t, "test@fynbos.dev", review.ReviewedBy)
	assert.Equal(t, linkedaccounts.OwnershipReviewRequired, review.State)
	assert.Equal(t, linkedaccounts.Verified, review.NewState)
	assert.NotEmpty(t, review.CompletedAt)
}

func TestListReviews(t *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, t)
	require.NoError(t, err)

	walletID := uuid.NewString()
	la, err := c.LinkedAccounts.Create(ctx, &linkedaccounts.CreateArgs{
		WalletID: walletID,
		Name:     "Test",
		Mask:     "1234",
		Provider: "mx",
		Type:     "bank",
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
		ReviewedBy: "test@fynbos.dev",
		NewState:   linkedaccounts.Verified,
	})
	require.NoError(t, err)

	reviews, err = c.LinkedAccounts.ListIncompleteReviews(ctx, db.Pagination{})
	require.NoError(t, err)
	assert.Len(t, reviews, 3)
}
