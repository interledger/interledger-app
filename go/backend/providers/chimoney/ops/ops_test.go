package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	la_mock "github.com/interledger/interledger-app/go/backend/linkedaccounts/client/mock"
	"github.com/interledger/interledger-app/go/backend/providers/chimoney"
	"github.com/interledger/interledger-app/go/backend/providers/chimoney/ops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetInteracEmail(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	laMock := la_mock.NewMockClient(ctrl)
	b := backends{
		db:  db.MigrateTestDB(t, ctx),
		las: laMock,
	}

	laMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(
		&linkedaccounts.LinkedAccount{
			ProviderID: "test@test.com",
		}, nil,
	).Times(1)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return(
		[]linkedaccounts.LinkedAccount{}, nil,
	).Times(1)

	walletID := uuid.NewString()
	email := "test@test.com"

	ne, err := ops.SetInteracEmail(ctx, b, walletID, email)
	require.NoError(t, err)
	assert.Equal(t, email, ne.ProviderID)

	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return(
		[]linkedaccounts.LinkedAccount{{
			Provider:   chimoney.ProviderName,
			Type:       chimoney.AccTypeInterac,
			ProviderID: "test@test.com",
		}}, nil,
	).AnyTimes()
	e, err := ops.GetInteracEmail(ctx, b, walletID)
	require.NoError(t, err)
	assert.Equal(t, email, e)

	email = "test2@test2.com"

	ne, err = ops.SetInteracEmail(ctx, b, walletID, email)
	require.ErrorIs(t, err, chimoney.ErrInteracAlreadyLinked)
	require.Nil(t, ne)
}

func TestGetChiWallet(t *testing.T) {
	ctx := context.Background()

	b := backends{
		db: db.MigrateTestDB(t, ctx),
	}

	walletID := uuid.NewString()
	chiWallet := uuid.NewString()

	_, err := b.DB().ExecContext(ctx, "INSERT INTO chi_money_wallets (external_id, wallet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;", chiWallet, walletID)
	require.NoError(t, err)

	chiW, err := ops.GetChiWallet(ctx, b, walletID)
	require.NoError(t, err)
	require.Equal(t, chiWallet, chiW)
}
