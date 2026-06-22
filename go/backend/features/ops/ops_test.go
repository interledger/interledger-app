package ops_test

import (
	"context"
	"testing"

	"github.com/interledger/interledger-app/go/backend/providers/pti"
	"github.com/interledger/interledger-app/go/backend/wallets"
	wallet_mock "github.com/interledger/interledger-app/go/backend/wallets/client/mock"

	"github.com/interledger/interledger-app/go/backend/linkedaccounts"

	linked_accounts_mock "github.com/interledger/interledger-app/go/backend/linkedaccounts/client/mock"

	"github.com/interledger/interledger-app/go/backend/kyc"

	kyc_mock "github.com/interledger/interledger-app/go/backend/kyc/client/mock"

	"github.com/golang/mock/gomock"

	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/features"
	"github.com/interledger/interledger-app/go/backend/features/ops"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestSetFeatures(t *testing.T) {
	t.Skip("TODO - Broken test, needs to be fixed")
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, db, nil, nil, nil)

	cases := []struct {
		name  string
		feats *features.WalletFeatures
	}{
		{
			name: "all true",
			feats: &features.WalletFeatures{
				SendEnabled:       true,
				ReceiveEnabled:    true,
				LinkedAccEnabled:  true,
				CardsEnabled:      true,
				BanksEnabled:      true,
				IdentitiesEnabled: true,
				TwitterEnabled:    true,
				AddCardsEnabled:   true,
			},
		},
		{
			name:  "all false",
			feats: &features.WalletFeatures{},
		},
		{
			name: "mixed",
			feats: &features.WalletFeatures{
				SendEnabled:      true,
				ReceiveEnabled:   true,
				LinkedAccEnabled: true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wid := uuid.NewString()
			f, err := ops.SetFeatures(ctx, b, wid, *tc.feats)
			require.NoError(t, err)

			assert.DeepEqual(t, f, tc.feats)

			f, err = ops.Features(ctx, b, wid)
			require.NoError(t, err)

			assert.DeepEqual(t, f, tc.feats)
		})
	}
}

func TestFeatures(t *testing.T) {
	t.Skip("TODO - Broken test, needs to be fixed")

	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	kc := kyc_mock.NewMockClient(ctrl)
	fc := linked_accounts_mock.NewMockClient(ctrl)
	wc := wallet_mock.NewMockClient(ctrl)

	b := ops.NewTestBackends(t, db, kc, fc, wc)

	cases := []struct {
		name      string
		KycStatus kyc.Status
		wallet    *wallets.Wallet
		feats     *features.WalletFeatures
		numCards  int
	}{
		{
			name:      "KYC unapproved all false",
			KycStatus: kyc.StatusDenied,
			feats:     &features.WalletFeatures{},
		},
		{
			name:      "KYC NON US only identities",
			KycStatus: kyc.StatusLevel1,
			wallet:    &wallets.Wallet{Country: "UG"},
			feats: &features.WalletFeatures{
				IdentitiesEnabled: true,
				TwitterEnabled:    true,
			},
		},
		{
			name:      "KYC US send state",
			KycStatus: kyc.StatusLevel1,
			numCards:  2,
			wallet:    &wallets.Wallet{Country: "US"},
			feats: &features.WalletFeatures{
				IdentitiesEnabled: true,
				TwitterEnabled:    true,
				ReceiveEnabled:    true,
				LinkedAccEnabled:  true,
				CardsEnabled:      false,
				SendEnabled:       true,
				AddCardsEnabled:   false,
				BanksEnabled:      true,
			},
		},
		{
			name:      "KYC US send state, max cards added",
			KycStatus: kyc.StatusLevel1,
			numCards:  4,
			wallet:    &wallets.Wallet{Country: "US"},
			feats: &features.WalletFeatures{
				IdentitiesEnabled: true,
				TwitterEnabled:    true,
				ReceiveEnabled:    true,
				LinkedAccEnabled:  true,
				CardsEnabled:      false,
				SendEnabled:       true,
				AddCardsEnabled:   false,
				BanksEnabled:      true,
			},
		},
		{
			name:      "KYC ZA",
			KycStatus: kyc.StatusLevel1,
			numCards:  4,
			wallet:    &wallets.Wallet{Country: "ZA"},
			feats: &features.WalletFeatures{
				IdentitiesEnabled: true,
				TwitterEnabled:    true,
				ReceiveEnabled:    true,
				LinkedAccEnabled:  true,
				BanksEnabled:      true,
				CardsEnabled:      false,
				SendEnabled:       true,
				AddCardsEnabled:   false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wid := uuid.NewString()
			kc.EXPECT().GetKYCStatus(ctx, wid).Return(tc.KycStatus, nil)
			if tc.wallet != nil {
				wc.EXPECT().Get(ctx, wid).Return(tc.wallet, nil)
			}

			var lal []linkedaccounts.LinkedAccount
			for i := 0; i < tc.numCards; i++ {
				lal = append(lal, linkedaccounts.LinkedAccount{
					State:    linkedaccounts.Verified,
					Provider: pti.ProviderName,
					Type:     pti.TypeCard,
				})
			}

			fc.EXPECT().ListByWalletId(ctx, wid).Return(lal, nil).AnyTimes()

			f, err := ops.Features(ctx, b, wid)
			require.NoError(t, err)

			assert.DeepEqual(t, f, tc.feats)
		})
	}
}
