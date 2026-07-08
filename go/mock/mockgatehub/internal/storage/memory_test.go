package storage

import (
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_CreateUser(t *testing.T) {
	store := NewMemoryStorage()

	user := &models.User{
		Email: "test@example.com",
	}

	err := store.CreateUser(user)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
}

func TestMemoryStorage_CreateUser_DuplicateEmail(t *testing.T) {
	store := NewMemoryStorage()

	user1 := &models.User{Email: "test@example.com"}
	err := store.CreateUser(user1)
	require.NoError(t, err)

	user2 := &models.User{Email: "test@example.com"}
	err = store.CreateUser(user2)
	assert.Error(t, err)
}

func TestMemoryStorage_GetUser(t *testing.T) {
	store := NewMemoryStorage()

	user := &models.User{Email: "test@example.com"}
	err := store.CreateUser(user)
	require.NoError(t, err)

	retrieved, err := store.GetUser(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestMemoryStorage_GetUserByEmail(t *testing.T) {
	store := NewMemoryStorage()

	user := &models.User{Email: "test@example.com"}
	err := store.CreateUser(user)
	require.NoError(t, err)

	retrieved, err := store.GetUserByEmail("test@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
}

func TestMemoryStorage_UpdateUser(t *testing.T) {
	store := NewMemoryStorage()

	user := &models.User{Email: "test@example.com"}
	err := store.CreateUser(user)
	require.NoError(t, err)

	user.KYCState = "accepted"
	err = store.UpdateUser(user)
	require.NoError(t, err)

	retrieved, err := store.GetUser(user.ID)
	require.NoError(t, err)
	assert.Equal(t, "accepted", retrieved.KYCState)
}

func TestMemoryStorage_CreateWallet(t *testing.T) {
	store := NewMemoryStorage()

	wallet := &models.Wallet{
		Address: "rTestAddress123",
		UserID:  "user-123",
		Name:    "My Wallet",
	}

	err := store.CreateWallet(wallet)
	require.NoError(t, err)
	assert.NotZero(t, wallet.CreatedAt)
}

func TestMemoryStorage_GetWallet(t *testing.T) {
	store := NewMemoryStorage()

	wallet := &models.Wallet{
		Address: "rTestAddress123",
		UserID:  "user-123",
	}
	err := store.CreateWallet(wallet)
	require.NoError(t, err)

	retrieved, err := store.GetWallet("rTestAddress123")
	require.NoError(t, err)
	assert.Equal(t, wallet.UserID, retrieved.UserID)
}

func TestMemoryStorage_GetWalletsByUser(t *testing.T) {
	store := NewMemoryStorage()

	wallet1 := &models.Wallet{Address: "rAddr1", UserID: "user-123"}
	wallet2 := &models.Wallet{Address: "rAddr2", UserID: "user-123"}
	wallet3 := &models.Wallet{Address: "rAddr3", UserID: "user-456"}

	store.CreateWallet(wallet1)
	store.CreateWallet(wallet2)
	store.CreateWallet(wallet3)

	wallets, err := store.GetWalletsByUser("user-123")
	require.NoError(t, err)
	assert.Len(t, wallets, 2)
}

func TestMemoryStorage_CreateTransaction(t *testing.T) {
	store := NewMemoryStorage()

	tx := &models.Transaction{
		UserID:      "user-123",
		Amount:      "100.50",
		TotalAmount: "100.50",
		Fee:         "0.00",
		Currency:    "USD",
		Status:      1,
	}

	err := store.CreateTransaction(tx)
	require.NoError(t, err)
	assert.NotEmpty(t, tx.ID)
	assert.NotZero(t, tx.CreatedAt)
}

func TestMemoryStorage_GetTransaction(t *testing.T) {
	store := NewMemoryStorage()

	tx := &models.Transaction{
		UserID:      "user-123",
		Amount:      "100.50",
		TotalAmount: "100.50",
		Fee:         "0.00",
		Currency:    "USD",
		Status:      1,
	}
	err := store.CreateTransaction(tx)
	require.NoError(t, err)

	retrieved, err := store.GetTransaction(tx.ID)
	require.NoError(t, err)
	assert.Equal(t, tx.Amount, retrieved.Amount)
	assert.Equal(t, "100.50", retrieved.Amount)
}

func TestMemoryStorage_Balance(t *testing.T) {
	store := NewMemoryStorage()

	// Initial balance should be 0
	balance, err := store.GetBalance("user-123", "USD")
	require.NoError(t, err)
	assert.Equal(t, 0.0, balance)

	// Add balance
	err = store.AddBalance("user-123", "USD", 100.50)
	require.NoError(t, err)

	balance, err = store.GetBalance("user-123", "USD")
	require.NoError(t, err)
	assert.Equal(t, 100.50, balance)

	// Add more
	err = store.AddBalance("user-123", "USD", 50.25)
	require.NoError(t, err)

	balance, err = store.GetBalance("user-123", "USD")
	require.NoError(t, err)
	assert.Equal(t, 150.75, balance)

	// Deduct
	err = store.DeductBalance("user-123", "USD", 50.00)
	require.NoError(t, err)

	balance, err = store.GetBalance("user-123", "USD")
	require.NoError(t, err)
	assert.Equal(t, 100.75, balance)
}

func TestMemoryStorage_DeductBalance_Insufficient(t *testing.T) {
	store := NewMemoryStorage()

	err := store.AddBalance("user-123", "USD", 50.00)
	require.NoError(t, err)

	err = store.DeductBalance("user-123", "USD", 100.00)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")
}

func TestMemoryStorage_Concurrent(t *testing.T) {
	store := NewMemoryStorage()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(i int) {
			user := &models.User{
				Email:     "test" + string(rune(i)) + "@example.com",
				CreatedAt: time.Now(),
			}
			store.CreateUser(user)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all users were created
	assert.Len(t, store.users, 10)
}

func TestMemoryStorage_TransactionWithAllFields(t *testing.T) {
	store := NewMemoryStorage()

	tx := &models.Transaction{
		UserID:           "user-456",
		UID:              "ext-ref-123",
		Amount:           "250.75",
		TotalAmount:      "252.50",
		Fee:              "1.75",
		Currency:         "EUR",
		VaultUUID:        "a09a0a2c-1a3a-44c5-a1b9-603a6eea9341",
		ReceivingAddress: "rTestAddr123",
		Type:             1,
		DepositType:      "external",
		Status:           1,
	}

	err := store.CreateTransaction(tx)
	require.NoError(t, err)
	assert.NotEmpty(t, tx.ID)

	retrieved, err := store.GetTransaction(tx.ID)
	require.NoError(t, err)
	assert.Equal(t, "250.75", retrieved.Amount)
	assert.Equal(t, "252.50", retrieved.TotalAmount)
	assert.Equal(t, "1.75", retrieved.Fee)
	assert.Equal(t, 1, retrieved.Status)
	assert.Equal(t, "ext-ref-123", retrieved.UID)
}

func TestMemoryStorage_TransactionStatusTypes(t *testing.T) {
	store := NewMemoryStorage()

	testCases := []struct {
		name   string
		status int
	}{
		{"pending", 0},
		{"completed", 1},
		{"failed", 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx := &models.Transaction{
				UserID:      "user-789",
				Amount:      "100.00",
				TotalAmount: "100.00",
				Fee:         "0.00",
				Currency:    "USD",
				Status:      tc.status,
			}

			err := store.CreateTransaction(tx)
			require.NoError(t, err)

			retrieved, err := store.GetTransaction(tx.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.status, retrieved.Status)
		})
	}
}

// Organization tests

func TestMemoryStorage_OrganizationCRUD(t *testing.T) {
	store := NewMemoryStorage()

	org := &models.Organization{
		ID:         "test-org",
		APIBaseURL: "https://api.example.com",
		TwoFAType:  "sms",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Test Create
	err := store.CreateOrganization(org)
	require.NoError(t, err)

	// Test duplicate create
	err = store.CreateOrganization(org)
	require.Error(t, err)

	// Test Get
	retrieved, err := store.GetOrganization("test-org")
	require.NoError(t, err)
	assert.Equal(t, org.ID, retrieved.ID)
	assert.Equal(t, org.APIBaseURL, retrieved.APIBaseURL)
	assert.Equal(t, org.TwoFAType, retrieved.TwoFAType)

	// Test Get returns a copy (mutation safety)
	retrieved.APIBaseURL = "https://mutated.example.com"
	original, err := store.GetOrganization("test-org")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", original.APIBaseURL)

	// Test Update
	org.TwoFAType = "totp"
	org.APIBaseURL = "https://api.newurl.com"
	err = store.UpdateOrganization(org)
	require.NoError(t, err)

	retrieved, err = store.GetOrganization("test-org")
	require.NoError(t, err)
	assert.Equal(t, "totp", retrieved.TwoFAType)
	assert.Equal(t, "https://api.newurl.com", retrieved.APIBaseURL)

	// Test Get non-existent
	_, err = store.GetOrganization("nonexistent")
	require.Error(t, err)

	// Test Update non-existent
	nonExistent := &models.Organization{ID: "nonexistent"}
	err = store.UpdateOrganization(nonExistent)
	require.Error(t, err)
}

// Customer tests

func TestMemoryStorage_CustomerCRUD(t *testing.T) {
	store := NewMemoryStorage()

	id := "cust-001"
	customer := &models.Customer{
		ID:       &id,
		SourceID: "src-001",
		Type:     "Citizen",
		Code:     "CUST-001",
	}

	err := store.CreateCustomer(customer)
	require.NoError(t, err)
	assert.NotZero(t, customer.CreatedAt)

	// Duplicate
	err = store.CreateCustomer(customer)
	assert.Error(t, err)

	// Get
	got, err := store.GetCustomer(id)
	require.NoError(t, err)
	assert.Equal(t, "Citizen", got.Type)

	// Get by source ID
	got, err = store.GetCustomerBySourceID("src-001")
	require.NoError(t, err)
	assert.Equal(t, id, *got.ID)

	// Not found
	_, err = store.GetCustomer("nonexistent")
	assert.Error(t, err)

	_, err = store.GetCustomerBySourceID("nonexistent")
	assert.Error(t, err)

	// Update
	customer.Code = "CUST-UPDATED"
	err = store.UpdateCustomer(customer)
	require.NoError(t, err)
	got, _ = store.GetCustomer(id)
	assert.Equal(t, "CUST-UPDATED", got.Code)

	// Update non-existent
	missingID := "missing"
	err = store.UpdateCustomer(&models.Customer{ID: &missingID, SourceID: "x"})
	assert.Error(t, err)

	// Update with nil ID
	err = store.UpdateCustomer(&models.Customer{SourceID: "x"})
	assert.Error(t, err)
}

func TestMemoryStorage_CreateCustomer_MissingSourceID(t *testing.T) {
	store := NewMemoryStorage()
	id := "cust-002"
	err := store.CreateCustomer(&models.Customer{ID: &id})
	assert.Error(t, err)
}

// Account tests

func TestMemoryStorage_AccountCRUD(t *testing.T) {
	store := NewMemoryStorage()

	id := "acc-001"
	account := &models.Account{
		ID:       &id,
		Currency: "EUR",
		Type:     "DEBIT",
	}

	err := store.CreateAccount(account)
	require.NoError(t, err)
	assert.NotZero(t, account.CreatedAt)

	// Duplicate
	err = store.CreateAccount(account)
	assert.Error(t, err)

	// Get
	got, err := store.GetAccount(id)
	require.NoError(t, err)
	assert.Equal(t, "EUR", got.Currency)

	// Not found
	_, err = store.GetAccount("nonexistent")
	assert.Error(t, err)

	// Update
	account.Status = "CLOSED"
	err = store.UpdateAccount(account)
	require.NoError(t, err)
	got, _ = store.GetAccount(id)
	assert.Equal(t, "CLOSED", got.Status)

	// Update non-existent
	missingID := "missing"
	err = store.UpdateAccount(&models.Account{ID: &missingID})
	assert.Error(t, err)

	// Update with nil ID
	err = store.UpdateAccount(&models.Account{})
	assert.Error(t, err)
}

// Card tests

func TestMemoryStorage_CardCRUD(t *testing.T) {
	store := NewMemoryStorage()

	card := &models.Card{
		ID:         "card-001",
		CustomerID: "cust-001",
		AccountID:  "acc-001",
		Status:     "Active",
	}

	err := store.CreateCard(card)
	require.NoError(t, err)
	assert.NotZero(t, card.CreatedAt)

	// Duplicate
	err = store.CreateCard(card)
	assert.Error(t, err)

	// Get
	got, err := store.GetCard("card-001")
	require.NoError(t, err)
	assert.Equal(t, "Active", got.Status)

	// Not found
	_, err = store.GetCard("nonexistent")
	assert.Error(t, err)

	// Update
	card.Status = "Blocked"
	err = store.UpdateCard(card)
	require.NoError(t, err)
	got, _ = store.GetCard("card-001")
	assert.Equal(t, "Blocked", got.Status)

	// Update non-existent
	err = store.UpdateCard(&models.Card{ID: "missing"})
	assert.Error(t, err)

	// Update with empty ID
	err = store.UpdateCard(&models.Card{})
	assert.Error(t, err)

	// Get by customer
	cards, err := store.GetCardsByCustomer("cust-001")
	require.NoError(t, err)
	assert.Len(t, cards, 1)

	cards, _ = store.GetCardsByCustomer("nonexistent")
	assert.Empty(t, cards)

	// Get by account
	cards, err = store.GetCardsByAccount("acc-001")
	require.NoError(t, err)
	assert.Len(t, cards, 1)

	cards, _ = store.GetCardsByAccount("nonexistent")
	assert.Empty(t, cards)
}

// Card limits tests

func TestMemoryStorage_CardLimits(t *testing.T) {
	store := NewMemoryStorage()

	// Empty limits
	limits, err := store.GetCardLimits("card-001")
	require.NoError(t, err)
	assert.Empty(t, limits)

	// Set limits
	newLimits := []models.CardLimit{
		{Type: "dailyOverall", Limit: 1000.00, Currency: "EUR"},
		{Type: "perTransaction", Limit: 500.00, Currency: "EUR"},
	}
	err = store.SetCardLimits("card-001", newLimits)
	require.NoError(t, err)

	limits, err = store.GetCardLimits("card-001")
	require.NoError(t, err)
	assert.Len(t, limits, 2)
	assert.Equal(t, 1000.0, limits[0].Limit)
}

// Card transaction tests

func TestMemoryStorage_CardTransactionCRUD(t *testing.T) {
	store := NewMemoryStorage()

	amount := "100.00"
	currency := "EUR"
	tx := &models.CardTransaction{
		TransactionID:       "ctx-001",
		TransactionAmount:   &amount,
		TransactionCurrency: &currency,
		Type:                1,
	}

	err := store.CreateCardTransaction(tx)
	require.NoError(t, err)

	// Duplicate
	err = store.CreateCardTransaction(tx)
	assert.Error(t, err)

	// Missing ID
	err = store.CreateCardTransaction(&models.CardTransaction{})
	assert.Error(t, err)

	// Get
	got, err := store.GetCardTransaction("ctx-001")
	require.NoError(t, err)
	assert.Equal(t, "100.00", *got.TransactionAmount)

	// Not found
	_, err = store.GetCardTransaction("nonexistent")
	assert.Error(t, err)

	// Index by card
	err = store.AddCardTransactionIndex("card-001", "ctx-001")
	require.NoError(t, err)

	ids, err := store.GetCardTransactionIDs("card-001")
	require.NoError(t, err)
	assert.Equal(t, []string{"ctx-001"}, ids)

	// Empty index
	ids, err = store.GetCardTransactionIDs("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, ids)

	// Missing args
	err = store.AddCardTransactionIndex("", "ctx-001")
	assert.Error(t, err)
	err = store.AddCardTransactionIndex("card-001", "")
	assert.Error(t, err)
}

// Customer address tests

func TestMemoryStorage_CustomerAddress(t *testing.T) {
	store := NewMemoryStorage()

	addr := &models.CustomerDeliveryAddress{
		Type:        "HOME",
		Line1:       "123 Main St",
		City:        "Berlin",
		CountryCode: "DE",
	}

	err := store.CreateCustomerAddress("cust-001", addr)
	require.NoError(t, err)
	assert.NotEmpty(t, addr.ID)

	addresses, err := store.GetCustomerAddresses("cust-001")
	require.NoError(t, err)
	assert.Len(t, addresses, 1)
	assert.Equal(t, "Berlin", addresses[0].City)

	// Empty customer
	addresses, err = store.GetCustomerAddresses("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, addresses)

	// Missing customer ID
	err = store.CreateCustomerAddress("", addr)
	assert.Error(t, err)

	// Nil address
	err = store.CreateCustomerAddress("cust-001", nil)
	assert.Error(t, err)
}

// UpdateTransactionStatus tests

func TestMemoryStorage_UpdateTransactionStatus(t *testing.T) {
	store := NewMemoryStorage()

	tx := &models.Transaction{
		UserID:      "user-123",
		Amount:      "100.00",
		TotalAmount: "100.00",
		Fee:         "0.00",
		Currency:    "USD",
		Status:      1, // pending
	}
	require.NoError(t, store.CreateTransaction(tx))

	err := store.UpdateTransactionStatus(tx.ID, 100) // completed
	require.NoError(t, err)

	got, _ := store.GetTransaction(tx.ID)
	assert.Equal(t, 100, got.Status)

	// Non-existent
	err = store.UpdateTransactionStatus("nonexistent", 100)
	assert.Error(t, err)
}

// 3DS Challenge tests

func TestMemoryStorage_ThreeDSChallenge(t *testing.T) {
	store := NewMemoryStorage()

	challenge := &models.ThreeDSChallenge{
		TransactionID:    "3ds-001",
		CardID:           "card-001",
		UserID:           "user-001",
		MerchantName:     "Test Shop",
		PurchaseAmount:   "50.00",
		PurchaseCurrency: "EUR",
		PurchaseDate:     time.Now().UTC().Format(time.RFC3339),
		Timeout:          time.Now().Add(5 * time.Minute),
		Status:           "pending",
		CreatedAt:        time.Now(),
	}

	err := store.CreateThreeDSChallenge(challenge)
	require.NoError(t, err)

	// Duplicate overwrites (no error)
	err = store.CreateThreeDSChallenge(challenge)
	assert.NoError(t, err)

	// Get
	got, err := store.GetThreeDSChallenge("3ds-001")
	require.NoError(t, err)
	assert.Equal(t, "Test Shop", got.MerchantName)

	// Not found
	_, err = store.GetThreeDSChallenge("nonexistent")
	assert.Error(t, err)

	// Get pending
	pending, err := store.GetPendingThreeDSChallenges("user-001")
	require.NoError(t, err)
	assert.Len(t, pending, 1)

	// No pending for other user
	pending, err = store.GetPendingThreeDSChallenges("user-other")
	require.NoError(t, err)
	assert.Empty(t, pending)

	// Update to approved
	challenge.Status = "approved"
	err = store.UpdateThreeDSChallenge(challenge)
	require.NoError(t, err)

	// Should no longer be pending
	pending, _ = store.GetPendingThreeDSChallenges("user-001")
	assert.Empty(t, pending)

	// Update non-existent
	err = store.UpdateThreeDSChallenge(&models.ThreeDSChallenge{TransactionID: "nonexistent"})
	assert.Error(t, err)
}

// ── ListUsers ──────────────────────────────────────────────────────────────

func TestListUsers_Empty(t *testing.T) {
	store := NewMemoryStorage()
	users, err := store.ListUsers()
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestListUsers_WithSeededUsers(t *testing.T) {
	store := NewMemoryStorage()
	SeedTestUsers(store)

	users, err := store.ListUsers()
	require.NoError(t, err)
	assert.Len(t, users, 2)

	emails := make([]string, 0, len(users))
	for _, u := range users {
		emails = append(emails, u.Email)
	}
	assert.Contains(t, emails, "testuser1@mockgatehub.local")
	assert.Contains(t, emails, "testuser2@mockgatehub.local")
}

func TestListUsers_AfterCreate(t *testing.T) {
	store := NewMemoryStorage()
	require.NoError(t, store.CreateUser(&models.User{Email: "a@example.com"}))
	require.NoError(t, store.CreateUser(&models.User{Email: "b@example.com"}))

	users, err := store.ListUsers()
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

// ── ListTransactionsByUser ─────────────────────────────────────────────────

func TestListTransactionsByUser_Empty(t *testing.T) {
	store := NewMemoryStorage()
	SeedTestUsers(store)

	txns, err := store.ListTransactionsByUser("00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)
	assert.Empty(t, txns)
}

func TestListTransactionsByUser(t *testing.T) {
	store := NewMemoryStorage()
	SeedTestUsers(store)

	user1 := "00000000-0000-0000-0000-000000000001"
	user2 := "00000000-0000-0000-0000-000000000002"

	require.NoError(t, store.CreateTransaction(&models.Transaction{
		ID: "tx-u1-a", UserID: user1, Amount: "10.00", Currency: "USD",
	}))
	require.NoError(t, store.CreateTransaction(&models.Transaction{
		ID: "tx-u1-b", UserID: user1, Amount: "20.00", Currency: "EUR",
	}))
	require.NoError(t, store.CreateTransaction(&models.Transaction{
		ID: "tx-u2-a", UserID: user2, Amount: "5.00", Currency: "GBP",
	}))

	u1Txns, err := store.ListTransactionsByUser(user1)
	require.NoError(t, err)
	assert.Len(t, u1Txns, 2)

	u2Txns, err := store.ListTransactionsByUser(user2)
	require.NoError(t, err)
	assert.Len(t, u2Txns, 1)
	assert.Equal(t, "tx-u2-a", u2Txns[0].ID)
}

// ── GetAllBalances ─────────────────────────────────────────────────────────

func TestGetAllBalances_Empty(t *testing.T) {
	store := NewMemoryStorage()
	balances, err := store.GetAllBalances("nonexistent-user")
	require.NoError(t, err)
	assert.NotNil(t, balances)
	assert.Empty(t, balances)
}

func TestGetAllBalances(t *testing.T) {
	store := NewMemoryStorage()
	userID := "00000000-0000-0000-0000-000000000001"
	require.NoError(t, store.CreateUser(&models.User{ID: userID, Email: "u@test.com"}))

	require.NoError(t, store.AddBalance(userID, "USD", 100.0))
	require.NoError(t, store.AddBalance(userID, "EUR", 200.0))

	balances, err := store.GetAllBalances(userID)
	require.NoError(t, err)
	assert.Equal(t, 100.0, balances["USD"])
	assert.Equal(t, 200.0, balances["EUR"])
	assert.Len(t, balances, 2)
}
