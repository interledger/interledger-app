package ops_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	la_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	ec_mock "gitlab.com/fynbos/backend/providers/gatehub/external/mock"
	"gitlab.com/fynbos/backend/providers/gatehub/ops"
	"gitlab.com/fynbos/backend/transactions"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"
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

func TestGetAccountStatement_LinkedAccountsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

	b := Backends{la: laMock}
	_, err := ops.GetAccountStatement(context.Background(), b, nil, "wallet-id", 2025, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrInternal)
}

func TestGetAccountStatement_NoGatehubLinkedAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{
		{Provider: "other-provider", Type: gatehub.AccTypeBalance},
	}, nil)

	b := Backends{la: laMock}
	_, err := ops.GetAccountStatement(context.Background(), b, nil, "wallet-id", 2025, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrNotFound)
}

func TestGetAccountStatement_DateBeforeAccountCreation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{
		{
			Provider:   gatehub.ProviderName,
			Type:       gatehub.AccTypeBalance,
			ProviderID: "rWalletAddress",
			CreatedAt:  time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}, nil)

	b := Backends{la: laMock}
	_, err := ops.GetAccountStatement(context.Background(), b, nil, "wallet-id", 2025, 5)

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrBadRequest)
}

func TestGetAccountStatement_DateInFuture(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now().UTC()
	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{
		{
			Provider:   gatehub.ProviderName,
			Type:       gatehub.AccTypeBalance,
			ProviderID: "rWalletAddress",
			CreatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}, nil)

	b := Backends{la: laMock}
	_, err := ops.GetAccountStatement(context.Background(), b, nil, "wallet-id", now.Year()+1, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrBadRequest)
}

func TestGetAccountStatement_ExternalIDsError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDB := db.MigrateTestDB(t, ctx)

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{
		{
			Provider:   gatehub.ProviderName,
			Type:       gatehub.AccTypeBalance,
			ProviderID: "rWalletAddress",
			CreatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}, nil)

	b := Backends{db: testDB, la: laMock}
	_, err := ops.GetAccountStatement(ctx, b, nil, "wallet-id-with-no-db-entry", 2025, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrInternal)
}

func TestGetAccountStatement_Success(t *testing.T) {
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
		{
			Provider:   gatehub.ProviderName,
			Type:       gatehub.AccTypeBalance,
			ProviderID: providerID,
			CreatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}, nil)

	ecMock := ec_mock.NewMockClient(ctrl)
	ecMock.EXPECT().GetAccountStatement(gomock.Any(), externalUserID, providerID, 2025, 1).
		Return(io.NopCloser(strings.NewReader("pdf-content")), nil)

	b := Backends{db: testDB, la: laMock}
	body, err := ops.GetAccountStatement(ctx, b, ecMock, walletID, 2025, 1)

	require.NoError(t, err)
	require.NotNil(t, body)
	defer body.Close()

	content, readErr := io.ReadAll(body)
	require.NoError(t, readErr)
	assert.Equal(t, "pdf-content", string(content))
}

func TestGetTransactionStatement_LinkedAccountsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

	b := Backends{la: laMock}
	_, err := ops.GetTransactionStatement(context.Background(), b, nil, "wallet-id", "tx-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrInternal)
}

func TestGetTransactionStatement_NoGatehubLinkedAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{
		{Provider: "other-provider", Type: gatehub.AccTypeBalance},
	}, nil)

	b := Backends{la: laMock}
	_, err := ops.GetTransactionStatement(context.Background(), b, nil, "wallet-id", "tx-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrNotFound)
}

func TestGetTransactionStatement_ExternalIDsError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDB := db.MigrateTestDB(t, ctx)

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), gomock.Any()).Return([]linkedaccounts.LinkedAccount{
		{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance, ProviderID: "rWalletAddress", WalletID: "wallet-id-with-no-db-entry"},
	}, nil)

	b := Backends{db: testDB, la: laMock}
	_, err := ops.GetTransactionStatement(ctx, b, nil, "wallet-id-with-no-db-entry", "tx-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrInternal)
}

func TestGetTransactionStatement_TransactionsError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	walletID := uuid.NewString()
	externalUserID := uuid.NewString()

	testDB := db.MigrateTestDB(t, ctx)
	_, err := testDB.ExecContext(ctx, `INSERT INTO wallets (id, name) VALUES ($1, 'test') ON CONFLICT DO NOTHING`, walletID)
	require.NoError(t, err)
	_, err = testDB.ExecContext(ctx, `INSERT INTO gatehub_users (external_id, wallet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, externalUserID, walletID)
	require.NoError(t, err)

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{
		{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance, ProviderID: "rWalletAddress", WalletID: walletID},
	}, nil)

	txMock := transactions_mock.NewMockClient(ctrl)
	txMock.EXPECT().GetTransaction(gomock.Any(), walletID, "tx-id").Return(nil, errors.New("tx db error"))

	b := Backends{db: testDB, la: laMock, tc: txMock}
	_, err = ops.GetTransactionStatement(ctx, b, nil, walletID, "tx-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrInternal)
}

func TestGetTransactionStatement_TransactionNotFound(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	walletID := uuid.NewString()
	externalUserID := uuid.NewString()

	testDB := db.MigrateTestDB(t, ctx)
	_, err := testDB.ExecContext(ctx, `INSERT INTO wallets (id, name) VALUES ($1, 'test') ON CONFLICT DO NOTHING`, walletID)
	require.NoError(t, err)
	_, err = testDB.ExecContext(ctx, `INSERT INTO gatehub_users (external_id, wallet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, externalUserID, walletID)
	require.NoError(t, err)

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{
		{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance, ProviderID: "rWalletAddress", WalletID: walletID},
	}, nil)

	txMock := transactions_mock.NewMockClient(ctrl)
	txMock.EXPECT().GetTransaction(gomock.Any(), walletID, "tx-id").Return(nil, nil)

	b := Backends{db: testDB, la: laMock, tc: txMock}
	_, err = ops.GetTransactionStatement(ctx, b, nil, walletID, "tx-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrNotFound)
}

func TestGetTransactionStatement_ExternalNotFound(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	walletID := uuid.NewString()
	externalUserID := uuid.NewString()
	foreignTxID := "external-transfer-id"

	testDB := db.MigrateTestDB(t, ctx)
	_, err := testDB.ExecContext(ctx, `INSERT INTO wallets (id, name) VALUES ($1, 'test') ON CONFLICT DO NOTHING`, walletID)
	require.NoError(t, err)
	_, err = testDB.ExecContext(ctx, `INSERT INTO gatehub_users (external_id, wallet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, externalUserID, walletID)
	require.NoError(t, err)

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{
		{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance, ProviderID: "rWalletAddress", WalletID: walletID},
	}, nil)

	txMock := transactions_mock.NewMockClient(ctrl)
	txMock.EXPECT().GetTransaction(gomock.Any(), walletID, "tx-id").Return(&transactions.Transaction{ForeignID: foreignTxID}, nil)

	ecMock := ec_mock.NewMockClient(ctrl)
	ecMock.EXPECT().GetTransferConfirmation(gomock.Any(), externalUserID, foreignTxID).Return(nil, external.ErrNotFound)

	b := Backends{db: testDB, la: laMock, tc: txMock}
	_, err = ops.GetTransactionStatement(ctx, b, ecMock, walletID, "tx-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrNotFound)
}

func TestGetTransactionStatement_ExternalError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	walletID := uuid.NewString()
	externalUserID := uuid.NewString()
	foreignTxID := "external-transfer-id"

	testDB := db.MigrateTestDB(t, ctx)
	_, err := testDB.ExecContext(ctx, `INSERT INTO wallets (id, name) VALUES ($1, 'test') ON CONFLICT DO NOTHING`, walletID)
	require.NoError(t, err)
	_, err = testDB.ExecContext(ctx, `INSERT INTO gatehub_users (external_id, wallet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, externalUserID, walletID)
	require.NoError(t, err)

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{
		{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance, ProviderID: "rWalletAddress", WalletID: walletID},
	}, nil)

	txMock := transactions_mock.NewMockClient(ctrl)
	txMock.EXPECT().GetTransaction(gomock.Any(), walletID, "tx-id").Return(&transactions.Transaction{ForeignID: foreignTxID}, nil)

	ecMock := ec_mock.NewMockClient(ctrl)
	ecMock.EXPECT().GetTransferConfirmation(gomock.Any(), externalUserID, foreignTxID).Return(nil, errors.New("gatehub error"))

	b := Backends{db: testDB, la: laMock, tc: txMock}
	_, err = ops.GetTransactionStatement(ctx, b, ecMock, walletID, "tx-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrInternal)
}

func TestGetTransactionStatement_Success(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	walletID := uuid.NewString()
	externalUserID := uuid.NewString()
	providerID := "rWalletAddress"
	foreignTxID := "external-transfer-id"

	testDB := db.MigrateTestDB(t, ctx)
	_, err := testDB.ExecContext(ctx, `INSERT INTO wallets (id, name) VALUES ($1, 'test') ON CONFLICT DO NOTHING`, walletID)
	require.NoError(t, err)
	_, err = testDB.ExecContext(ctx, `INSERT INTO gatehub_users (external_id, wallet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, externalUserID, walletID)
	require.NoError(t, err)

	laMock := la_mock.NewMockClient(ctrl)
	laMock.EXPECT().ListByWalletId(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{
		{Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance, ProviderID: providerID, WalletID: walletID},
	}, nil)

	txMock := transactions_mock.NewMockClient(ctrl)
	txMock.EXPECT().GetTransaction(gomock.Any(), walletID, "tx-id").Return(&transactions.Transaction{ForeignID: foreignTxID}, nil)

	ecMock := ec_mock.NewMockClient(ctrl)
	ecMock.EXPECT().GetTransferConfirmation(gomock.Any(), externalUserID, foreignTxID).
		Return(io.NopCloser(strings.NewReader("pdf-content")), nil)

	b := Backends{db: testDB, la: laMock, tc: txMock}
	body, err := ops.GetTransactionStatement(ctx, b, ecMock, walletID, "tx-id")

	require.NoError(t, err)
	require.NotNil(t, body)
	defer body.Close()

	content, readErr := io.ReadAll(body)
	require.NoError(t, readErr)
	assert.Equal(t, "pdf-content", string(content))
}
