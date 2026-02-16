package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/backend/providers/xago/ops"
)

var testXagoConfig = xago.Config{
	APIBaseURL:      "https://test-api.xago.io:8085/v1",
	IdentityBaseURL: "https://test-api.xago.io:9000/v1",
	APIPublicKey:    "test-public-key",
	APISecret:       "test-secret",
	PolicyID:        "5e2585a474b0e90012ce8ff1",
	USDOpsAccount:   "868196c3-f6b4-4920-bbfb-d1c7f6a98183",
	ZAROpsAccount:   "b0944908-16e6-4ef4-8677-192165e33c59",
	LedgerIDZAR:     9246927,
	LedgerIDUSD:     9246873,
}

func TestActivity_SaveBeneficiary(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(t, func(tb *ops.TestBackends) {
		tb.DBC = db.MigrateTestDB(t, ctx)
	})

	a := ops.NewActivity(b, testXagoConfig)

	walletID := uuid.NewString()
	beneficiaryID := uuid.NewString()

	err := a.SaveBeneficiary(ctx, walletID, external.AccountBeneficiaries{
		BranchCode:         "branchy face",
		Reference:          "Ref me",
		BeneficiaryAddress: "Dark side of the moon",
		BankName:           "FNB",
		AccountNumber:      "acc_1234",
		Status:             "open",
		CurrencyCode:       "ZAR",
		ID:                 beneficiaryID,
		Scope:              "read",
		Name:               "nothing",
		Wallet:             nil,
	})
	require.NoError(t, err)

	var entry xago.Beneficiary
	err = b.DB().GetContext(ctx, &entry, "SELECT id, wallet_id, address, reference, status, currency, scope, name FROM xago_beneficiaries WHERE id =$1", beneficiaryID)
	require.NoError(t, err)

	assert.Equal(t, "Ref me", entry.Reference)
	assert.Equal(t, "Dark side of the moon", entry.Address)
	assert.Equal(t, "ZAR", entry.Currency)
	assert.Equal(t, "open", entry.Status)
	assert.Equal(t, beneficiaryID, entry.ID)
	assert.Equal(t, "read", entry.Scope)
	assert.Equal(t, "nothing", entry.Name)
}

func TestActivity_SaveSubAccount(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(t, func(tb *ops.TestBackends) {
		tb.DBC = db.MigrateTestDB(t, ctx)
	})

	a := ops.NewActivity(b, testXagoConfig)

	walletID := uuid.NewString()
	accountID := uuid.NewString()

	err := a.SaveSubAccount(ctx, walletID, accountID, external.SubAccount{
		AccountID:      accountID,
		DepositAddress: "Fluffy",
		DepositTag:     1945,
		Beneficiaries: []external.Beneficiaries{
			{
				DepositReference: "fluffels",
				BeneficiaryType:  "rollup",
			},
		},
	})
	require.NoError(t, err)

	entry, err := ops.LookupByAccountID(ctx, b, accountID)
	require.NoError(t, err)

	assert.Equal(t, walletID, entry.WalletID)
	assert.Equal(t, accountID, entry.ID)
	assert.Equal(t, accountID, entry.AccountID)
	assert.Equal(t, "Fluffy", entry.DepositAddress)
	assert.Equal(t, 1945, entry.DepositTag)
	assert.Equal(t, "fluffels", entry.DepositReference)

	entry, err = ops.LookupSubAccount(ctx, b, walletID)
	require.NoError(t, err)

	assert.Equal(t, walletID, entry.WalletID)
	assert.Equal(t, accountID, entry.ID)
	assert.Equal(t, accountID, entry.AccountID)
	assert.Equal(t, "Fluffy", entry.DepositAddress)
	assert.Equal(t, 1945, entry.DepositTag)
	assert.Equal(t, "fluffels", entry.DepositReference)

}
