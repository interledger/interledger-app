package main

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
)

func newTestClient(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

func seedChimoney(t *testing.T, client *redis.Client) {
	t.Helper()
	ctx := context.Background()

	sa := models.SubAccount{
		ID:        "sub-1",
		ParentID:  "parent-1",
		UID:       "uid-1",
		Name:      "Test Account",
		Email:     "test@example.com",
		KYCStatus: "approved",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	require.NoError(t, client.Set(ctx, "chimoney:subaccount:sub-1", mustMarshal(t, sa), 0).Err())
	require.NoError(t, client.SAdd(ctx, subAccountsIdx, "sub-1").Err())

	payment := models.Payment{
		ID:         "pay-1",
		IssueID:    "issue-pay-1",
		SubAccount: "sub-1",
		Amount:     25.0,
		Currency:   "USD",
		Status:     "pending",
		CreatedAt:  time.Now().UTC().Truncate(time.Second),
	}
	require.NoError(t, client.Set(ctx, "chimoney:payment:issue-pay-1", mustMarshal(t, payment), 0).Err())

	payout := models.Payout{
		ID:        "pout-1",
		IssueID:   "issue-pout-1",
		SubAccount: "sub-1",
		Amount:    10.0,
		Currency:  "USD",
		Status:    "completed",
		ChiRef:    "chiref-abc",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	require.NoError(t, client.Set(ctx, "chimoney:payout:issue-pout-1", mustMarshal(t, payout), 0).Err())
	require.NoError(t, client.Set(ctx, "chimoney:payout:chiref:chiref-abc", "issue-pout-1", 0).Err())
}

func TestChimoneyRoundTrip(t *testing.T) {
	ctx := context.Background()

	src, _ := newTestClient(t)
	seedChimoney(t, src)

	// Export into a buffer
	var buf bytes.Buffer
	require.NoError(t, runExport(ctx, src, &buf))

	// Verify snapshot contents
	var snap ChimoneySnapshot
	require.NoError(t, json.Unmarshal(buf.Bytes(), &snap))
	assert.Len(t, snap.SubAccounts, 1)
	assert.Equal(t, "sub-1", snap.SubAccounts[0].ID)
	assert.Len(t, snap.Payments, 1)
	assert.Len(t, snap.Payouts, 1)

	// Import into a fresh instance
	dst, _ := newTestClient(t)
	require.NoError(t, runImport(ctx, dst, bytes.NewReader(buf.Bytes())))

	// Verify sub-account primary key and index set
	raw, err := dst.Get(ctx, "chimoney:subaccount:sub-1").Result()
	require.NoError(t, err)
	var gotSA models.SubAccount
	require.NoError(t, json.Unmarshal([]byte(raw), &gotSA))
	assert.Equal(t, "sub-1", gotSA.ID)

	members, err := dst.SMembers(ctx, subAccountsIdx).Result()
	require.NoError(t, err)
	assert.Contains(t, members, "sub-1")

	// Verify payment
	_, err = dst.Get(ctx, "chimoney:payment:issue-pay-1").Result()
	require.NoError(t, err)

	// Verify payout and chiref lookup key rebuilt
	_, err = dst.Get(ctx, "chimoney:payout:issue-pout-1").Result()
	require.NoError(t, err)

	chiRef, err := dst.Get(ctx, "chimoney:payout:chiref:chiref-abc").Result()
	require.NoError(t, err)
	assert.Equal(t, "issue-pout-1", chiRef)
}

func TestChimoneyPayoutChirefExcluded(t *testing.T) {
	ctx := context.Background()
	src, _ := newTestClient(t)
	seedChimoney(t, src)

	var buf bytes.Buffer
	require.NoError(t, runExport(ctx, src, &buf))

	var snap ChimoneySnapshot
	require.NoError(t, json.Unmarshal(buf.Bytes(), &snap))

	// chiref lookup keys must not appear as standalone Payout entries
	for _, p := range snap.Payouts {
		assert.NotContains(t, p.IssueID, "chiref", "chiref lookup key must not leak into Payouts")
	}
}
