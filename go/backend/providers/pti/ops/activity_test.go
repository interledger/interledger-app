package ops_test

// func TestSavePtiUser(t *testing.T) {
// 	b := NewBackends(t)
// 	a := ops.NewActivity(b, nil)
// 	ctx := context.Background()

// 	walletID := uuid.NewString()
// 	u, err := a.SavePtiUser(ctx, "123", walletID)
// 	require.NoError(t, err)
// 	assert.Equal(t, "123", u.ExternalID)
// 	assert.Equal(t, walletID, u.WalletID)

// 	u, err = a.GetPtiUser(ctx, walletID)
// 	require.NoError(t, err)
// 	assert.Equal(t, "123", u.ExternalID)
// 	assert.Equal(t, walletID, u.WalletID)
// }

// func TestCreatePtiWallet(t *testing.T) {
// 	b := NewBackends(t)
// 	a := ops.NewActivity(b, nil)
// 	ctx := context.Background()

// 	laID := uuid.NewString()
// 	b.la.EXPECT().GetByProviderID(
// 		ctx, linkedaccounts.GetByProviderIDArgs{
// 			Provider:   pti.ProviderName,
// 			ProviderID: "123",
// 		}).Return(nil, linkedaccounts.ErrNotFound).Times(1)
// 	b.la.EXPECT().GetByProviderID(
// 		ctx, linkedaccounts.GetByProviderIDArgs{
// 			Provider:   pti.ProviderName,
// 			ProviderID: "123",
// 		}).Return(&linkedaccounts.LinkedAccount{ID: laID}, nil).Times(1)
// 	b.la.EXPECT().Create(ctx, gomock.Any()).Return(&linkedaccounts.LinkedAccount{ID: laID}, nil).Times(1)

// 	la, err := a.CreatePtiWalletLinkedAccount(ctx, linkedaccounts.CreateArgs{
// 		Provider:   pti.ProviderName,
// 		ProviderID: "123",
// 	})
// 	require.NoError(t, err)
// 	assert.Equal(t, laID, la.ID)

// 	la, err = a.CreatePtiWalletLinkedAccount(ctx, linkedaccounts.CreateArgs{
// 		Provider:   pti.ProviderName,
// 		ProviderID: "123",
// 	})
// 	require.NoError(t, err)
// 	assert.Equal(t, laID, la.ID)
// }
