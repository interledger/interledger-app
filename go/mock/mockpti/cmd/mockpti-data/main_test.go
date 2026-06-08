package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
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

func seedPTI(t *testing.T, client *redis.Client) {
	t.Helper()
	ctx := context.Background()

	userID := "user-pti-1"
	u := &models.User{
		ID:     userID,
		Type:   "individual",
		Status: "active",
	}
	require.NoError(t, client.Set(ctx, "pti:user:"+userID, mustMarshal(t, u), 0).Err())

	assessment := &models.Assessment{
		ResourceType: "assessment",
		ClientID:     "client-1",
		RequestID:    "req-1",
		UserID:       userID,
		Date:         "2024-01-01",
		Assessment:   "accepted",
		Tier:         1,
	}
	require.NoError(t, client.RPush(ctx, "pti:assessments:"+userID, mustMarshal(t, assessment)).Err())

	walletID := "wallet-pti-1"
	w := &models.Wallet{
		WalletID: walletID,
		Currency: "USD",
		Balance:  100.0,
	}
	require.NoError(t, client.Set(ctx, "pti:wallet:"+userID+":"+walletID, mustMarshal(t, w), 0).Err())
	require.NoError(t, client.SAdd(ctx, "pti:wallets:"+userID, walletID).Err())

	piID := "pi-1"
	pi := &models.PaymentInformation{
		ID:                "pi-1",
		Type:              "bank",
		BankAccountNumber: "123456789",
	}
	require.NoError(t, client.Set(ctx, "pti:paymentinfo:"+userID+":"+piID, mustMarshal(t, pi), 0).Err())

	txID := "req-tx-1"
	tx := &models.Transaction{
		RequestID:       txID,
		Status:          "completed",
		TransactionType: "deposit",
		Amount:          50.0,
		Currency:        "USD",
		ResourceType:    "transaction",
	}
	require.NoError(t, client.Set(ctx, "pti:transaction:"+txID, mustMarshal(t, tx), 0).Err())

	upd := &models.TransactionUpdate{
		ID:            "upd-1",
		TransactionID: "tx-abc",
		Feedback:      "ok",
		Date:          time.Now().UTC().Truncate(time.Second),
		ProviderName:  "pti",
		Payload:       "{}",
	}
	require.NoError(t, client.RPush(ctx, "pti:txupdates:"+txID, mustMarshal(t, upd)).Err())
}

func TestPTIRoundTrip(t *testing.T) {
	ctx := context.Background()

	src, _ := newTestClient(t)
	seedPTI(t, src)

	snap, err := exportData(ctx, src)
	require.NoError(t, err)

	assert.Len(t, snap.Users, 1)
	assert.Equal(t, "user-pti-1", snap.Users[0].ID)
	assert.Len(t, snap.Assessments["user-pti-1"], 1)
	assert.Len(t, snap.Wallets, 1)
	assert.Equal(t, "USD", snap.Wallets[0].Currency)
	assert.Len(t, snap.PaymentInformation, 1)
	assert.Len(t, snap.Transactions, 1)
	assert.Len(t, snap.TransactionUpdates["req-tx-1"], 1)

	dst, _ := newTestClient(t)
	require.NoError(t, importData(ctx, dst, snap, false))

	// User key
	userVal, err := dst.Get(ctx, "pti:user:user-pti-1").Result()
	require.NoError(t, err)
	var gotUser models.User
	require.NoError(t, json.Unmarshal([]byte(userVal), &gotUser))
	assert.Equal(t, "user-pti-1", gotUser.ID)

	// Assessment list
	items, err := dst.LRange(ctx, "pti:assessments:user-pti-1", 0, -1).Result()
	require.NoError(t, err)
	assert.Len(t, items, 1)

	// Wallet key and index set
	_, err = dst.Get(ctx, "pti:wallet:user-pti-1:wallet-pti-1").Result()
	require.NoError(t, err)
	members, err := dst.SMembers(ctx, "pti:wallets:user-pti-1").Result()
	require.NoError(t, err)
	assert.Contains(t, members, "wallet-pti-1")

	// Payment information
	_, err = dst.Get(ctx, "pti:paymentinfo:user-pti-1:pi-1").Result()
	require.NoError(t, err)

	// Transaction
	_, err = dst.Get(ctx, "pti:transaction:req-tx-1").Result()
	require.NoError(t, err)

	// Transaction update list
	updItems, err := dst.LRange(ctx, "pti:txupdates:req-tx-1", 0, -1).Result()
	require.NoError(t, err)
	assert.Len(t, updItems, 1)
}

func TestPTIImportNoFlush(t *testing.T) {
	ctx := context.Background()

	src, _ := newTestClient(t)
	seedPTI(t, src)

	snap, err := exportData(ctx, src)
	require.NoError(t, err)

	dst, _ := newTestClient(t)
	// Pre-seed a key that should survive the no-flush import
	require.NoError(t, dst.Set(ctx, "pti:user:existing", `{"id":"existing"}`, 0).Err())

	require.NoError(t, importData(ctx, dst, snap, true))

	// Both the pre-seeded key and the imported key should exist
	_, err = dst.Get(ctx, "pti:user:existing").Result()
	require.NoError(t, err)
	_, err = dst.Get(ctx, "pti:user:user-pti-1").Result()
	require.NoError(t, err)
}
