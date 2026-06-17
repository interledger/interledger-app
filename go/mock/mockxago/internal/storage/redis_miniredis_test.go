package storage

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/interledger/interledger-app/go/mock/mockxago/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRedisStoreForTest(t *testing.T) (*RedisStorage, func()) {
	t.Helper()

	mini := miniredis.RunT(t)
	storeIface, err := NewRedisStorage("redis://"+mini.Addr(), 0)
	require.NoError(t, err)

	store, ok := storeIface.(*RedisStorage)
	require.True(t, ok)

	cleanup := func() {
		_ = store.Reset(context.Background())
		mini.Close()
	}

	return store, cleanup
}

func TestRedisStorage_TokenAndSubAccountFlow(t *testing.T) {
	store, cleanup := newRedisStoreForTest(t)
	defer cleanup()
	ctx := context.Background()

	token := &models.AccessToken{ID: "t1", Token: "token-1", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, store.SaveAccessToken(ctx, token))

	gotToken, err := store.GetAccessToken(ctx, token.Token)
	require.NoError(t, err)
	assert.Equal(t, token.Token, gotToken.Token)

	require.NoError(t, store.SaveTokenAccount(ctx, token.Token, "acc-1"))
	accountID, err := store.GetAccountIDByToken(ctx, token.Token)
	require.NoError(t, err)
	assert.Equal(t, "acc-1", accountID)

	require.NoError(t, store.InvalidateAccessToken(ctx, token.Token))
	_, err = store.GetAccessToken(ctx, token.Token)
	assert.ErrorIs(t, err, ErrTokenNotFound)

	sa := &models.SubAccount{ID: "sa-1", WalletID: "wallet-1", AccountID: "account-1", FirstName: "A"}
	require.NoError(t, store.SaveSubAccount(ctx, sa))

	gotByWallet, err := store.GetSubAccountByWalletID(ctx, "wallet-1")
	require.NoError(t, err)
	assert.Equal(t, "account-1", gotByWallet.AccountID)

	sa.FirstName = "Updated"
	require.NoError(t, store.UpdateSubAccount(ctx, sa))
	gotByID, err := store.GetSubAccount(ctx, "account-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated", gotByID.FirstName)
}

func TestRedisStorage_BeneficiaryTransactionAndBalanceFlow(t *testing.T) {
	store, cleanup := newRedisStoreForTest(t)
	defer cleanup()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		ben := &models.Beneficiary{
			ID:        "ben-" + string(rune('a'+i)),
			AccountID: "account-1",
			WalletID:  "account-1",
			Name:      "B",
			Currency:  "ZAR",
			Status:    "pending",
		}
		require.NoError(t, store.SaveBeneficiary(ctx, ben))
	}

	benList, benTotal, err := store.ListBeneficiariesByAccountID(ctx, "account-1", 2, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, benTotal)
	assert.Len(t, benList, 2)

	require.NoError(t, store.UpdateBeneficiaryStatus(ctx, "ben-a", "approved"))
	updatedBen, err := store.GetBeneficiary(ctx, "ben-a")
	require.NoError(t, err)
	assert.Equal(t, "approved", updatedBen.Status)

	tx := &models.Transaction{ID: "tx-1", AccountID: "account-1", WalletID: "wallet-1", Status: "pending"}
	require.NoError(t, store.SaveTransaction(ctx, tx))
	require.NoError(t, store.SaveIdempotencyKey(ctx, "idem-1", "tx-1"))

	gotByIdem, err := store.GetTransactionByIdempotencyKey(ctx, "idem-1")
	require.NoError(t, err)
	require.NotNil(t, gotByIdem)
	assert.Equal(t, "tx-1", gotByIdem.ID)

	require.NoError(t, store.UpdateTransactionStatus(ctx, "tx-1", "completed"))
	gotTx, err := store.GetTransaction(ctx, "tx-1")
	require.NoError(t, err)
	assert.NotNil(t, gotTx.SettledAt)

	txList, txTotal, err := store.ListTransactionsByAccount(ctx, "account-1", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, txTotal, 1)
	assert.GreaterOrEqual(t, len(txList), 1)

	require.NoError(t, store.SetBalance(ctx, "wallet-1", "USD", 100, 5))
	require.NoError(t, store.AddBalance(ctx, "wallet-1", "USD", 20))
	require.NoError(t, store.SubtractBalance(ctx, "wallet-1", "USD", 30))
	available, reserved, err := store.GetBalance(ctx, "wallet-1", "USD")
	require.NoError(t, err)
	assert.Equal(t, 90.0, available)
	assert.Equal(t, 5.0, reserved)

	err = store.SubtractBalance(ctx, "wallet-1", "USD", 1000)
	assert.ErrorIs(t, err, ErrInsufficientBalance)
}

func TestRedisStorage_DepositAndJobFlow(t *testing.T) {
	store, cleanup := newRedisStoreForTest(t)
	defer cleanup()
	ctx := context.Background()

	dep := &models.Deposit{ID: "dep-1", AccountID: "account-1", DepositReference: "ref-1", Status: "pending"}
	require.NoError(t, store.SaveDeposit(ctx, dep))

	gotRef, err := store.GetDepositByReference(ctx, "ref-1")
	require.NoError(t, err)
	require.NotNil(t, gotRef)
	assert.Equal(t, "dep-1", gotRef.ID)

	require.NoError(t, store.UpdateDepositStatus(ctx, "dep-1", "completed"))
	gotDep, err := store.GetDeposit(ctx, "dep-1")
	require.NoError(t, err)
	assert.NotNil(t, gotDep.SettledAt)

	deps, total, err := store.ListDeposits(ctx, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, deps, 1)

	job := &models.Job{ID: "job-1", JobType: "test", Status: "pending", CreatedAt: time.Now(), NotBefore: time.Now().Add(-time.Second)}
	require.NoError(t, store.SaveJob(ctx, job))

	ready, err := store.ListReadyJobs(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, ready, 1)

	require.NoError(t, store.IncrementJobAttempts(ctx, "job-1"))
	now := time.Now()
	require.NoError(t, store.UpdateJobStatus(ctx, "job-1", "failed", &now, "boom"))
	gotJob, err := store.GetJob(ctx, "job-1")
	require.NoError(t, err)
	assert.Equal(t, "failed", gotJob.Status)
	assert.Equal(t, "boom", gotJob.LastError)

	require.NoError(t, store.ClearJobs(ctx))
	require.NoError(t, store.ClearDeposits(ctx))
	require.NoError(t, store.ClearTransactions(ctx))
	require.NoError(t, store.ClearBalances(ctx))

	require.NoError(t, store.Reset(ctx))
}
