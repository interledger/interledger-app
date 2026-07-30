package cppairs

import (
	"context"
	"errors"
	"testing"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/features"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeFeaturesClient is a minimal hand-written fake for features.Client,
// keyed by wallet ID so sender/receiver can be given different flag states.
type fakeFeaturesClient struct {
	byWallet map[string]*features.WalletFeatures
	errs     map[string]error
}

func (f *fakeFeaturesClient) Features(ctx context.Context, walletID string) (*features.WalletFeatures, error) {
	if err, ok := f.errs[walletID]; ok {
		return nil, err
	}
	return f.byWallet[walletID], nil
}

func (f *fakeFeaturesClient) SetFeatures(ctx context.Context, walletID string, wf features.WalletFeatures) (*features.WalletFeatures, error) {
	return &wf, nil
}

type fakeBackends struct {
	fc features.Client
}

func (b *fakeBackends) Features() features.Client {
	return b.fc
}

func validXagoAcc() linkedaccounts.LinkedAccount {
	return linkedaccounts.LinkedAccount{
		WalletID:        "wallet-xago",
		Provider:        xago.ProviderName,
		Type:            xago.AccTypeBalance,
		State:           linkedaccounts.Verified,
		CanSend:         true,
		CanReceive:      true,
		SendCurrency:    currency.ZAR,
		ReceiveCurrency: currency.ZAR,
	}
}

func validGatehubAcc() linkedaccounts.LinkedAccount {
	return linkedaccounts.LinkedAccount{
		WalletID:        "wallet-gatehub",
		Provider:        gatehub.ProviderName,
		Type:            gatehub.AccTypeBalance,
		State:           linkedaccounts.Verified,
		CanSend:         true,
		CanReceive:      true,
		SendCurrency:    currency.EUR,
		ReceiveCurrency: currency.EUR,
	}
}

func TestLookupCrossProviderPair(t *testing.T) {
	xagoAcc := validXagoAcc()
	gatehubAcc := validGatehubAcc()
	unsupportedAcc := linkedaccounts.LinkedAccount{Provider: "other"}

	t.Run("sender-receiver order", func(t *testing.T) {
		pair := LookupCrossProviderPair(&xagoAcc, &gatehubAcc)
		assert.NotNil(t, pair)
		assert.IsType(t, &XagoGatehubPair{}, pair)
	})

	t.Run("receiver-sender order", func(t *testing.T) {
		pair := LookupCrossProviderPair(&gatehubAcc, &xagoAcc)
		assert.NotNil(t, pair)
		assert.IsType(t, &XagoGatehubPair{}, pair)
	})

	t.Run("unsupported pair", func(t *testing.T) {
		pair := LookupCrossProviderPair(&xagoAcc, &unsupportedAcc)
		assert.Nil(t, pair)
	})
}

func TestIsCrossProviderPair(t *testing.T) {
	xagoAcc := validXagoAcc()
	gatehubAcc := validGatehubAcc()
	sameProviderAcc := validXagoAcc()

	assert.True(t, IsCrossProviderPair(&xagoAcc, &gatehubAcc))
	assert.False(t, IsCrossProviderPair(&xagoAcc, &sameProviderAcc))
}

func TestIsSupportedCrossProviderPair(t *testing.T) {
	xagoAcc := validXagoAcc()
	gatehubAcc := validGatehubAcc()
	unsupportedAcc := linkedaccounts.LinkedAccount{Provider: "other"}

	assert.True(t, IsSupportedCrossProviderPair(&xagoAcc, &gatehubAcc))
	assert.False(t, IsSupportedCrossProviderPair(&xagoAcc, &unsupportedAcc))
}

func TestValidateAccountCompatibility(t *testing.T) {
	t.Run("valid accounts", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := validGatehubAcc()
		err := ValidateAccountCompatibility(&sender, &receiver)
		assert.NoError(t, err)
	})

	t.Run("unsupported pair", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := linkedaccounts.LinkedAccount{Provider: "other"}
		err := ValidateAccountCompatibility(&sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})

	t.Run("currency not supported", func(t *testing.T) {
		sender := validXagoAcc()
		sender.SendCurrency = currency.EUR
		receiver := validGatehubAcc()
		err := ValidateAccountCompatibility(&sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})

	t.Run("invalid state", func(t *testing.T) {
		sender := validXagoAcc()
		sender.State = linkedaccounts.State("Rejected")
		receiver := validGatehubAcc()
		err := ValidateAccountCompatibility(&sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})

	t.Run("sender can't send", func(t *testing.T) {
		sender := validXagoAcc()
		sender.CanSend = false
		receiver := validGatehubAcc()
		err := ValidateAccountCompatibility(&sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})

	t.Run("receiver can't receive", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := validGatehubAcc()
		receiver.CanReceive = false
		err := ValidateAccountCompatibility(&sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})

	t.Run("account type not supported", func(t *testing.T) {
		sender := validXagoAcc()
		sender.Type = xago.AccTypeBank
		receiver := validGatehubAcc()
		err := ValidateAccountCompatibility(&sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})
}

func TestValidateWalletFlags(t *testing.T) {
	ctx := context.Background()

	t.Run("both enabled", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := validGatehubAcc()
		b := &fakeBackends{fc: &fakeFeaturesClient{byWallet: map[string]*features.WalletFeatures{
			sender.WalletID:   {XagoGatehubPaymentsEnabled: true},
			receiver.WalletID: {XagoGatehubPaymentsEnabled: true},
		}}}

		err := ValidateWalletFlags(ctx, b, &sender, &receiver)
		assert.NoError(t, err)
	})

	t.Run("flag disabled for one wallet", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := validGatehubAcc()
		b := &fakeBackends{fc: &fakeFeaturesClient{byWallet: map[string]*features.WalletFeatures{
			sender.WalletID:   {XagoGatehubPaymentsEnabled: true},
			receiver.WalletID: {XagoGatehubPaymentsEnabled: false},
		}}}

		err := ValidateWalletFlags(ctx, b, &sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})

	t.Run("unsupported pair", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := linkedaccounts.LinkedAccount{Provider: "other"}
		b := &fakeBackends{fc: &fakeFeaturesClient{}}

		err := ValidateWalletFlags(ctx, b, &sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})

	t.Run("sender features lookup error", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := validGatehubAcc()
		wantErr := errors.New("boom")
		b := &fakeBackends{fc: &fakeFeaturesClient{errs: map[string]error{
			sender.WalletID: wantErr,
		}}}

		err := ValidateWalletFlags(ctx, b, &sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("receiver features lookup error", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := validGatehubAcc()
		wantErr := errors.New("boom")
		b := &fakeBackends{fc: &fakeFeaturesClient{
			byWallet: map[string]*features.WalletFeatures{
				sender.WalletID: {XagoGatehubPaymentsEnabled: true},
			},
			errs: map[string]error{
				receiver.WalletID: wantErr,
			},
		}}

		err := ValidateWalletFlags(ctx, b, &sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, wantErr)
	})
}

func TestValidateAccounts(t *testing.T) {
	ctx := context.Background()

	enabledFeaturesFor := func(sender, receiver linkedaccounts.LinkedAccount) *fakeBackends {
		return &fakeBackends{fc: &fakeFeaturesClient{byWallet: map[string]*features.WalletFeatures{
			sender.WalletID:   {XagoGatehubPaymentsEnabled: true},
			receiver.WalletID: {XagoGatehubPaymentsEnabled: true},
		}}}
	}

	t.Run("valid p2p payment", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := validGatehubAcc()
		b := enabledFeaturesFor(sender, receiver)

		err := ValidateAccounts(ctx, b, payments.TypePeer2Peer, &sender, &receiver)
		assert.NoError(t, err)
	})

	t.Run("same provider accounts", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := validXagoAcc()
		b := enabledFeaturesFor(sender, receiver)

		err := ValidateAccounts(ctx, b, payments.TypePeer2Peer, &sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})

	t.Run("unsupported provider pair", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := linkedaccounts.LinkedAccount{Provider: "other"}
		b := &fakeBackends{fc: &fakeFeaturesClient{}}

		err := ValidateAccounts(ctx, b, payments.TypePeer2Peer, &sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})

	t.Run("unsupported payment type", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := validGatehubAcc()
		b := enabledFeaturesFor(sender, receiver)

		err := ValidateAccounts(ctx, b, payments.TypeWithdrawal, &sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})

	t.Run("account compatibility failure", func(t *testing.T) {
		sender := validXagoAcc()
		sender.CanSend = false
		receiver := validGatehubAcc()
		b := enabledFeaturesFor(sender, receiver)

		err := ValidateAccounts(ctx, b, payments.TypePeer2Peer, &sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})

	t.Run("feature flag failure", func(t *testing.T) {
		sender := validXagoAcc()
		receiver := validGatehubAcc()
		b := &fakeBackends{fc: &fakeFeaturesClient{byWallet: map[string]*features.WalletFeatures{
			sender.WalletID:   {XagoGatehubPaymentsEnabled: false},
			receiver.WalletID: {XagoGatehubPaymentsEnabled: true},
		}}}

		err := ValidateAccounts(ctx, b, payments.TypePeer2Peer, &sender, &receiver)
		require.Error(t, err)
		assert.ErrorIs(t, err, payments.ErrIncompatibleAccounts)
	})
}

func TestIsStateValid(t *testing.T) {
	verified := validXagoAcc()
	unverified := validXagoAcc()
	unverified.State = linkedaccounts.State("Rejected")

	assert.True(t, isStateValid(&verified, &verified))
	assert.False(t, isStateValid(&unverified, &verified))
	assert.False(t, isStateValid(&verified, &unverified))
}

func TestCanSendAndReceive(t *testing.T) {
	canSend := validXagoAcc()
	canReceive := validGatehubAcc()
	cantSend := validXagoAcc()
	cantSend.CanSend = false
	cantReceive := validGatehubAcc()
	cantReceive.CanReceive = false

	assert.True(t, canSendAndReceive(&canSend, &canReceive))
	assert.False(t, canSendAndReceive(&cantSend, &canReceive))
	assert.False(t, canSendAndReceive(&canSend, &cantReceive))
}
