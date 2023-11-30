package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/xago"
	external_mock "gitlab.com/fynbos/backend/providers/xago/external/mock"
	"gitlab.com/fynbos/backend/providers/xago/ops"
)

func TestCreateTransaction(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)

	ex := external_mock.NewMockClient(ctrl)
	la := linkedaccounts_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(t, func(tb *ops.TestBackends) {
		tb.DBC = db.MigrateTestDB(t, ctx)
		tb.Extr = ex
		tb.La = la
	})

	args := xago.CreateTransactionArgs{
		WalletID:        uuid.NewString(),
		LinkedAccountID: uuid.NewString(),
		TransactionID:   uuid.NewString(),
		Amount:          currency.FromFloat64(10.2, currency.ZAR),
		Reference:       "MagicRef",
	}

	la.EXPECT().Get(ctx, args.LinkedAccountID).Return(&linkedaccounts.LinkedAccount{WalletID: args.WalletID, ProviderID: uuid.NewString(), ID: args.LinkedAccountID}, nil)

	externalID := uuid.NewString()

	ex.EXPECT().CreateTransaction(ctx, gomock.Any(), args.TransactionID, gomock.Any(), "MagicRef").Return(externalID, nil)

	tx, err := ops.CreateTransaction(ctx, b, args)
	require.NoError(t, err)

	assert.Equal(t, args.WalletID, tx.WalletID)
	assert.Equal(t, args.LinkedAccountID, tx.LinkedAccountID)
	assert.Equal(t, args.Amount.Float64(), tx.Amount.Float64())
	assert.Equal(t, args.TransactionID, tx.TransactionID)
	assert.Equal(t, externalID, tx.ID)
}
