package ops_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	la_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/gatehub"
	ec_mock "gitlab.com/fynbos/backend/providers/gatehub/external/client/mock"
	"gitlab.com/fynbos/backend/providers/gatehub/ops"
)

func TestGetAccountConfirmation_LinkedAccountsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

	b := Backends{la: laMock}
	_, err := ops.GetAccountConfirmation(context.Background(), b, nil, "wallet-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrInternal)
}

func TestGetAccountConfirmation_NoGatehubLinkedAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{
		{Provider: "other-provider", Type: gatehub.AccTypeBalance},
	}, nil)

	b := Backends{la: laMock}
	_, err := ops.GetAccountConfirmation(context.Background(), b, nil, "wallet-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrNotFound)
}

func TestGetAccountConfirmation_ExternalIDsError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDB := db.MigrateTestDB(t, ctx)

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{
		{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance, ProviderID: "rWalletAddress"},
	}, nil)

	b := Backends{db: testDB, la: laMock}
	_, err := ops.GetAccountConfirmation(ctx, b, nil, "wallet-id-with-no-db-entry")

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrInternal)
}

func TestGetAccountConfirmation_Success(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	walletID := uuid.NewString()
	externalUserID := uuid.NewString()
	providerID := "rWalletAddress"

	testDB := db.MigrateTestDB(t, ctx)
	_, err := testDB.ExecContext(ctx, `INSERT INTO wallets (id, name) VALUES ($1, 'test') ON CONFLICT DO NOTHING`, walletID)
	require.NoError(t, err)
	_, err = testDB.ExecContext(ctx, `INSERT INTO gatehub_users (external_id, wallet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, externalUserID, walletID)
	require.NoError(t, err)

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{
		{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance, ProviderID: providerID},
	}, nil)

	ecMock := ec_mock.NewMockClient(ctrl)
	ecMock.EXPECT().GetAccountConfirmation(gomock.Any(), externalUserID, providerID).
		Return(io.NopCloser(strings.NewReader("pdf-content")), nil)

	b := Backends{db: testDB, la: laMock}
	body, err := ops.GetAccountConfirmation(ctx, b, ecMock, walletID)

	require.NoError(t, err)
	require.NotNil(t, body)
	defer body.Close()

	content, readErr := io.ReadAll(body)
	require.NoError(t, readErr)
	assert.Equal(t, "pdf-content", string(content))
}
