package ops_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/kyc"

	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"

	"github.com/golang/mock/gomock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/features"
	"gitlab.com/fynbos/backend/features/ops"
	"gotest.tools/assert"
)

func TestSetFeatures(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, db, nil)

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
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	ctrl := gomock.NewController(t)
	kc := kyc_mock.NewMockClient(ctrl)

	b := ops.NewTestBackends(t, db, kc)

	cases := []struct {
		name      string
		KycStatus kyc.Status
		id        *kyc.IndividualDetails
		feats     *features.WalletFeatures
	}{
		{
			name:      "KYC unapproved all false",
			KycStatus: kyc.StatusDenied,
			feats:     &features.WalletFeatures{},
		},
		{
			name:      "KYC NON US only identities",
			KycStatus: kyc.StatusApproved,
			id:        &kyc.IndividualDetails{CountryCode: "UG"},
			feats: &features.WalletFeatures{
				IdentitiesEnabled: true,
				TwitterEnabled:    true,
			},
		},
		{
			name:      "KYC US non send state",
			KycStatus: kyc.StatusApproved,
			id:        &kyc.IndividualDetails{CountryCode: "US", Address: &kyc.Address{State: "US-XX"}},
			feats: &features.WalletFeatures{
				IdentitiesEnabled: true,
				TwitterEnabled:    true,
				ReceiveEnabled:    true,
				LinkedAccEnabled:  true,
				CardsEnabled:      true,
			},
		},
		{
			name:      "KYC US send state",
			KycStatus: kyc.StatusApproved,
			id:        &kyc.IndividualDetails{CountryCode: "US", Address: &kyc.Address{State: "US-SD"}},
			feats: &features.WalletFeatures{
				IdentitiesEnabled: true,
				TwitterEnabled:    true,
				ReceiveEnabled:    true,
				LinkedAccEnabled:  true,
				CardsEnabled:      true,
				SendEnabled:       true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wid := uuid.NewString()
			kc.EXPECT().GetKYCStatus(ctx, wid).Return(tc.KycStatus, nil)
			if tc.id != nil {
				kc.EXPECT().GetIndividualDetails(ctx, wid).Return(tc.id, nil)
			}

			f, err := ops.Features(ctx, b, wid)
			require.NoError(t, err)

			assert.DeepEqual(t, f, tc.feats)
		})
	}
}
