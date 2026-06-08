package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/mock/mockxago/internal/models"
)

func newTestClient(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func seedXago(t *testing.T, client *redis.Client) {
	t.Helper()
	ctx := context.Background()

	sa := &models.SubAccount{
		ID:                  "id-1",
		WalletID:            "wallet-1",
		AccountID:           "account-1",
		Email:               "test@example.com",
		DepositReferenceZAR: "ZAR-REF-1",
		DepositReferenceUSD: "USD-REF-1",
	}
	require.NoError(t, client.Set(ctx, "subaccount:account-1", mustMarshal(t, sa), 0).Err())
	require.NoError(t, client.Set(ctx, "subaccount:wallet:wallet-1", "account-1", 0).Err())
	require.NoError(t, client.Set(ctx, "subaccount:depositref:ZAR-REF-1", "account-1", 0).Err())
	require.NoError(t, client.Set(ctx, "subaccount:depositref:USD-REF-1", "account-1", 0).Err())

	ben := &models.Beneficiary{
		ID:        "ben-1",
		AccountID: "account-1",
		WalletID:  "wallet-1",
		Name:      "Test Beneficiary",
	}
	require.NoError(t, client.Set(ctx, "beneficiary:ben-1", mustMarshal(t, ben), 0).Err())
	require.NoError(t, client.RPush(ctx, "beneficiaries:wallet:account-1", "ben-1").Err())
	require.NoError(t, client.RPush(ctx, "beneficiaries:account:account-1", "ben-1").Err())

	tx := &models.Transaction{
		ID:        "tx-1",
		AccountID: "account-1",
		WalletID:  "wallet-1",
		Amount:    100.0,
		Currency:  "ZAR",
		Status:    "completed",
	}
	require.NoError(t, client.Set(ctx, "transaction:tx-1", mustMarshal(t, tx), 0).Err())
	require.NoError(t, client.RPush(ctx, "transactions:account:account-1", "tx-1").Err())

	dep := &models.Deposit{
		ID:               "dep-1",
		AccountID:        "account-1",
		Amount:           50.0,
		Currency:         "ZAR",
		DepositReference: "DEP-REF-1",
		Status:           "completed",
	}
	require.NoError(t, client.Set(ctx, "deposit:dep-1", mustMarshal(t, dep), 0).Err())
	require.NoError(t, client.Set(ctx, "deposit:ref:DEP-REF-1", "dep-1", 0).Err())
	require.NoError(t, client.RPush(ctx, "deposits:all", "dep-1").Err())

	require.NoError(t, client.Set(ctx, "balance:wallet-1:ZAR:available", "150.00", 0).Err())
	require.NoError(t, client.Set(ctx, "balance:wallet-1:ZAR:reserved", "10.00", 0).Err())
}

func TestXagoRoundTrip(t *testing.T) {
	ctx := context.Background()

	src, _ := newTestClient(t)
	seedXago(t, src)

	snap, err := exportData(ctx, src)
	require.NoError(t, err)

	assert.Len(t, snap.SubAccounts, 1)
	assert.Equal(t, "account-1", snap.SubAccounts[0].AccountID)
	assert.Len(t, snap.Beneficiaries, 1)
	assert.Len(t, snap.Transactions, 1)
	assert.Len(t, snap.Deposits, 1)
	assert.NotEmpty(t, snap.Balances)

	dst, _ := newTestClient(t)
	require.NoError(t, importData(ctx, dst, snap))

	// Primary sub-account key
	val, err := dst.Get(ctx, "subaccount:account-1").Result()
	require.NoError(t, err)
	var gotSA models.SubAccount
	require.NoError(t, json.Unmarshal([]byte(val), &gotSA))
	assert.Equal(t, "account-1", gotSA.AccountID)

	// Wallet index key rebuilt
	ptr, err := dst.Get(ctx, "subaccount:wallet:wallet-1").Result()
	require.NoError(t, err)
	assert.Equal(t, "account-1", ptr)

	// Deposit-ref index keys rebuilt
	ptr, err = dst.Get(ctx, "subaccount:depositref:ZAR-REF-1").Result()
	require.NoError(t, err)
	assert.Equal(t, "account-1", ptr)

	ptr, err = dst.Get(ctx, "subaccount:depositref:USD-REF-1").Result()
	require.NoError(t, err)
	assert.Equal(t, "account-1", ptr)

	// Beneficiary
	_, err = dst.Get(ctx, "beneficiary:ben-1").Result()
	require.NoError(t, err)
	benIDs, err := dst.LRange(ctx, "beneficiaries:wallet:account-1", 0, -1).Result()
	require.NoError(t, err)
	assert.Contains(t, benIDs, "ben-1")

	// Transaction and account index
	_, err = dst.Get(ctx, "transaction:tx-1").Result()
	require.NoError(t, err)
	txIDs, err := dst.LRange(ctx, "transactions:account:account-1", 0, -1).Result()
	require.NoError(t, err)
	assert.Contains(t, txIDs, "tx-1")

	// Deposit list and ref lookup
	depIDs, err := dst.LRange(ctx, "deposits:all", 0, -1).Result()
	require.NoError(t, err)
	assert.Contains(t, depIDs, "dep-1")
	_, err = dst.Get(ctx, "deposit:ref:DEP-REF-1").Result()
	require.NoError(t, err)

	// Balance
	avail, err := dst.Get(ctx, "balance:wallet-1:ZAR:available").Result()
	require.NoError(t, err)
	assert.Equal(t, "150.00", avail)
}

func TestXagoIndexKeysExcluded(t *testing.T) {
	ctx := context.Background()
	src, _ := newTestClient(t)
	seedXago(t, src)

	snap, err := exportData(ctx, src)
	require.NoError(t, err)

	for _, sa := range snap.SubAccounts {
		assert.NotContains(t, sa.AccountID, "wallet", "wallet index key must not leak into SubAccounts")
		assert.NotContains(t, sa.AccountID, "depositref", "depositref index key must not leak into SubAccounts")
	}
}
