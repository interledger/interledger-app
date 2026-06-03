package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
)

func newTestClient(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func seedGatehub(t *testing.T, client *redis.Client) {
	t.Helper()
	ctx := context.Background()

	userID := "user-ghtest-1"
	user := &models.User{ID: userID, Email: "gh@test.com", KYCState: "accepted"}
	require.NoError(t, client.Set(ctx, fmt.Sprintf("user:%s", userID), mustJSON(user), 0).Err())
	require.NoError(t, client.Set(ctx, fmt.Sprintf("email:%s", user.Email), userID, 0).Err())

	orgID := "org-test-1"
	org := &models.Organization{ID: orgID, APIBaseURL: "https://example.com"}
	require.NoError(t, client.Set(ctx, fmt.Sprintf("organization:%s", orgID), mustJSON(org), 0).Err())

	walletAddr := "rTest123Addr"
	wallet := &models.Wallet{Address: walletAddr, UserID: userID, Name: "Test"}
	require.NoError(t, client.Set(ctx, fmt.Sprintf("wallet:%s", walletAddr), mustJSON(wallet), 0).Err())
	require.NoError(t, client.SAdd(ctx, fmt.Sprintf("user:%s:wallets", userID), walletAddr).Err())

	txID := "tx-ghtest-1"
	tx := &models.Transaction{ID: txID, UserID: userID, Amount: "100", Currency: "USD", Status: 1}
	require.NoError(t, client.Set(ctx, fmt.Sprintf("tx:%s", txID), mustJSON(tx), 0).Err())

	require.NoError(t, client.Set(ctx, fmt.Sprintf("balance:%s:%s", userID, "USD"), "250.5", 0).Err())

	custID := "cust-test-1"
	cust := &models.Customer{ID: &custID, SourceID: "src-test"}
	require.NoError(t, client.Set(ctx, fmt.Sprintf("customer:%s", custID), mustJSON(cust), 0).Err())
	require.NoError(t, client.Set(ctx, fmt.Sprintf("customer:source:%s", cust.SourceID), custID, 0).Err())

	addrID := "addr-test-1"
	addr := &models.CustomerDeliveryAddress{ID: addrID, CustomerID: custID, City: "TestCity", CountryCode: "US", ZipCode: "12345", Status: "active"}
	require.NoError(t, client.Set(ctx, fmt.Sprintf("customer:address:%s", addrID), mustJSON(addr), 0).Err())
	require.NoError(t, client.SAdd(ctx, fmt.Sprintf("customer:%s:addresses", custID), addrID).Err())

	accID := "acc-test-1"
	acc := &models.Account{ID: &accID, CustomerID: &custID}
	require.NoError(t, client.Set(ctx, fmt.Sprintf("account:%s", accID), mustJSON(acc), 0).Err())
	require.NoError(t, client.SAdd(ctx, fmt.Sprintf("customer:%s:accounts", custID), accID).Err())

	cardID := "card-test-1"
	card := &models.Card{ID: cardID, CustomerID: custID, AccountID: accID, Status: "active"}
	require.NoError(t, client.Set(ctx, fmt.Sprintf("card:%s", cardID), mustJSON(card), 0).Err())
	require.NoError(t, client.SAdd(ctx, fmt.Sprintf("customer:%s:cards", custID), cardID).Err())
	require.NoError(t, client.SAdd(ctx, fmt.Sprintf("account:%s:cards", accID), cardID).Err())

	limits := []models.CardLimit{{Type: "purchase", Limit: 500, Currency: "USD"}}
	require.NoError(t, client.Set(ctx, fmt.Sprintf("card:%s:limits", cardID), mustJSON(limits), 0).Err())

	ctxID := "cardtx-test-1"
	cardTx := &models.CardTransaction{TransactionID: ctxID, TerminalID: "TERM1", GHResponseCode: "00", GHResponseDescription: "OK", CardScheme: 1}
	require.NoError(t, client.Set(ctx, fmt.Sprintf("cardtx:%s", ctxID), mustJSON(cardTx), 0).Err())
	require.NoError(t, client.RPush(ctx, fmt.Sprintf("card:%s:transactions", cardID), ctxID).Err())

	challenge := &models.ThreeDSChallenge{
		TransactionID: "3ds-test-1", UserID: userID, CardID: cardID,
		Status: "pending", Timeout: time.Now().Add(10 * time.Minute),
	}
	ttl := time.Until(challenge.Timeout) + 10*time.Minute
	require.NoError(t, client.Set(ctx, fmt.Sprintf("3ds:challenge:%s", challenge.TransactionID), mustJSON(challenge), ttl).Err())
}

func TestGatehubRoundTrip(t *testing.T) {
	ctx := context.Background()

	// Seed source instance
	src, _ := newTestClient(t)
	seedGatehub(t, src)

	// Export
	snap, err := exportData(ctx, src)
	require.NoError(t, err)

	assert.Len(t, snap.Users, 1)
	assert.Equal(t, "gh@test.com", snap.Users[0].Email)
	assert.Len(t, snap.Organizations, 1)
	assert.Len(t, snap.Wallets, 1)
	assert.Len(t, snap.Transactions, 1)
	assert.NotEmpty(t, snap.Balances)
	assert.Len(t, snap.Customers, 1)
	assert.Len(t, snap.Accounts, 1)
	assert.Len(t, snap.CustomerAddresses, 1)
	assert.Len(t, snap.Cards, 1)
	assert.NotEmpty(t, snap.CardLimits)
	assert.Len(t, snap.CardTransactions, 1)
	assert.NotEmpty(t, snap.CardTransactionsByCard)
	assert.Len(t, snap.ThreeDSChallenges, 1)

	// Import into fresh instance
	dst, _ := newTestClient(t)
	require.NoError(t, importData(ctx, dst, snap))

	// Verify user and email pointer
	userVal, err := dst.Get(ctx, "user:user-ghtest-1").Result()
	require.NoError(t, err)
	var gotUser models.User
	require.NoError(t, json.Unmarshal([]byte(userVal), &gotUser))
	assert.Equal(t, "gh@test.com", gotUser.Email)

	emailPtr, err := dst.Get(ctx, "email:gh@test.com").Result()
	require.NoError(t, err)
	assert.Equal(t, "user-ghtest-1", emailPtr)

	// Verify wallet and user:wallets index
	walletMembers, err := dst.SMembers(ctx, "user:user-ghtest-1:wallets").Result()
	require.NoError(t, err)
	assert.Contains(t, walletMembers, "rTest123Addr")

	// Verify balance
	balVal, err := dst.Get(ctx, "balance:user-ghtest-1:USD").Result()
	require.NoError(t, err)
	assert.Equal(t, "250.5", balVal)

	// Verify customer source pointer
	srcPtr, err := dst.Get(ctx, "customer:source:src-test").Result()
	require.NoError(t, err)
	assert.Equal(t, "cust-test-1", srcPtr)

	// Verify customer:addresses set rebuilt
	addrMembers, err := dst.SMembers(ctx, "customer:cust-test-1:addresses").Result()
	require.NoError(t, err)
	assert.Contains(t, addrMembers, "addr-test-1")

	// Verify card limits
	limitsVal, err := dst.Get(ctx, "card:card-test-1:limits").Result()
	require.NoError(t, err)
	assert.Contains(t, limitsVal, "purchase")

	// Verify card-tx-by-card list
	txIDs, err := dst.LRange(ctx, "card:card-test-1:transactions", 0, -1).Result()
	require.NoError(t, err)
	assert.Contains(t, txIDs, "cardtx-test-1")

	// Verify 3DS challenge
	_, err = dst.Get(ctx, "3ds:challenge:3ds-test-1").Result()
	require.NoError(t, err)
}

func TestGatehubIndexKeysExcluded(t *testing.T) {
	ctx := context.Background()
	src, _ := newTestClient(t)
	seedGatehub(t, src)

	snap, err := exportData(ctx, src)
	require.NoError(t, err)

	// Index/pointer keys must NOT appear as primary entities
	for _, u := range snap.Users {
		assert.NotContains(t, u.ID, "wallets", "user index key leaked into Users")
	}
	for _, c := range snap.Customers {
		assert.NotContains(t, *c.ID, "source", "customer source key leaked into Customers")
		assert.NotContains(t, *c.ID, "accounts", "customer accounts key leaked into Customers")
		assert.NotContains(t, *c.ID, "cards", "customer cards key leaked into Customers")
		assert.NotContains(t, *c.ID, "addresses", "customer addresses key leaked into Customers")
	}
}

func TestGatehubExpired3DSExcluded(t *testing.T) {
	ctx := context.Background()
	src, _ := newTestClient(t)

	// Write an expired challenge (past Timeout — export should skip it)
	expired := &models.ThreeDSChallenge{
		TransactionID: "3ds-expired",
		UserID:        "u1",
		Status:        "pending",
		Timeout:       time.Now().Add(-5 * time.Minute),
	}
	require.NoError(t, src.Set(ctx, "3ds:challenge:3ds-expired", mustJSON(expired), 0).Err())

	snap, err := exportData(ctx, src)
	require.NoError(t, err)
	assert.Empty(t, snap.ThreeDSChallenges, "expired 3DS challenge must not appear in export")
}
