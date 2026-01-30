package ops

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/payments"
	payments_mock "gitlab.com/fynbos/backend/payments/client/mock"
	"gitlab.com/fynbos/backend/providers/gatehub"
	gatehub_external "gitlab.com/fynbos/backend/providers/gatehub/external"
	rafiki_external "gitlab.com/fynbos/backend/rafiki/external"
	rafiki_external_mock "gitlab.com/fynbos/backend/rafiki/external/mock"
	"gitlab.com/fynbos/backend/wallets"
	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"
)

// startPostgresContainer starts a Postgres container and returns a connection string template
// suitable for db.MigrateTestDB (format: postgres://user:pass@host:port/%s?sslmode=disable).
func startPostgresContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("ready to accept connections").WithStartupTimeout(30*time.Second),
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(10*time.Second),
		),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}
	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", err
	}
	if host == "localhost" {
		host = "127.0.0.1"
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", err
	}
	// Connection string with %s for database name (MigrateTestDB creates a temp DB per run).
	connStr := fmt.Sprintf("postgres://postgres:password@%s:%s/%%s?sslmode=disable", host, port.Port())
	return container, connStr, nil
}

func TestMain(m *testing.M) {
	flag.Parse()
	if os.Getenv("DB_URL") == "" {
		runFlag := flag.Lookup("test.run").Value.String()
		runIntegration := runFlag == "" || strings.Contains(runFlag, "Integration")
		if runIntegration {
			ctx := context.Background()
			container, connStr, err := startPostgresContainer(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "rafiki/ops: Postgres container failed to start (Docker required): %v\n", err)
				os.Exit(1)
			}
			_ = os.Setenv("DB_URL", connStr)
			defer func() {
				_ = container.Terminate(ctx)
			}()
		}
	}
	os.Exit(m.Run())
}

// requireTestDB skips the test if DB_URL is not set or connection fails (integration tests require a running Postgres with testmigrations).
func requireTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	baseString := os.Getenv("DB_URL")
	if baseString == "" {
		t.Skip("DB_URL not set; skipping integration test (run with: make generatetestmigrations && DB_URL='postgres://...' go test ./rafiki/ops/... -run Integration)")
	}
	connString := baseString
	if strings.Contains(baseString, "%s") {
		connString = fmt.Sprintf(baseString, "postgres")
	}
	tryConn, err := sqlx.Connect("postgres", connString)
	if err != nil {
		t.Skipf("DB_URL set but connection failed; skipping integration test: %v (is Postgres running and DB_URL correct?)", err)
	}
	_ = tryConn.Close()
	conn := db.MigrateTestDB(t, context.Background())
	return conn
}

// fakeGatehub implements gatehub.Client for tests; only GetBalance, ReserveBalance, FinaliseReserve, AssignBalance are used by webhooks.
type fakeGatehub struct {
	getBalance      func(ctx context.Context, linkedAccountID string) (*gatehub.Balance, error)
	reserveBalance  func(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*gatehub.Balance, error)
	finaliseReserve func(ctx context.Context, txID string) error
	assignBalance   func(ctx context.Context, linkedAccountID, trxID string, amount currency.Amount) (*gatehub.Balance, error)
}

func (f *fakeGatehub) GetBalance(ctx context.Context, linkedAccountID string) (*gatehub.Balance, error) {
	if f.getBalance != nil {
		return f.getBalance(ctx, linkedAccountID)
	}
	return &gatehub.Balance{Available: currency.Amount{Value: 10000, Currency: currency.EUR, Scale: 2}}, nil
}

func (f *fakeGatehub) ReserveBalance(ctx context.Context, linkedAccountID, txID string, amt currency.Amount, timeout time.Duration) (*gatehub.Balance, error) {
	if f.reserveBalance != nil {
		return f.reserveBalance(ctx, linkedAccountID, txID, amt, timeout)
	}
	return &gatehub.Balance{}, nil
}

func (f *fakeGatehub) FinaliseReserve(ctx context.Context, txID string) error {
	if f.finaliseReserve != nil {
		return f.finaliseReserve(ctx, txID)
	}
	return nil
}

func (f *fakeGatehub) AssignBalance(ctx context.Context, linkedAccountID, trxID string, amount currency.Amount) (*gatehub.Balance, error) {
	if f.assignBalance != nil {
		return f.assignBalance(ctx, linkedAccountID, trxID, amount)
	}
	return &gatehub.Balance{}, nil
}

func (f *fakeGatehub) CreateUser(ctx context.Context, walletID string) (gatehub.Await, error) { return nil, nil }
func (f *fakeGatehub) GetUser(ctx context.Context, walletID string) (*gatehub.User, error)   { return nil, nil }
func (f *fakeGatehub) GetOnboardingWidget(ctx context.Context, walletID string) (string, error) {
	return "", nil
}
func (f *fakeGatehub) GetOnOffRampWidget(ctx context.Context, walletID string, isDeposit bool) (string, error) {
	return "", nil
}
func (f *fakeGatehub) CreateWithdrawal(ctx context.Context, walletID, externalTransactionID string) (string, error) {
	return "", nil
}
func (f *fakeGatehub) CreateTransfer(ctx context.Context, args gatehub.CreateTransferArgs) (*gatehub_external.Transaction, error) {
	return nil, nil
}
func (f *fakeGatehub) GetTransaction(ctx context.Context, walletID, id string) (*gatehub_external.Transaction, error) {
	return nil, nil
}
func (f *fakeGatehub) ListDeliveryAddresses(ctx context.Context, walletID string) ([]gatehub_external.CustomerDeliveryAddress, error) {
	return nil, nil
}
func (f *fakeGatehub) ListCards(ctx context.Context, externalIDs gatehub.ExternalIDs) ([]gatehub_external.Card, error) {
	return nil, nil
}
func (f *fakeGatehub) GetCardApplicationProducts(ctx context.Context) ([]gatehub_external.CardApplicationProduct, error) {
	return nil, nil
}
func (f *fakeGatehub) OrderCard(ctx context.Context, args gatehub.OrderCardArgs) error { return nil }
func (f *fakeGatehub) GetExternalIDs(ctx context.Context, walletID string) (*gatehub.ExternalIDs, error) {
	return nil, nil
}
func (f *fakeGatehub) GetCardToken(ctx context.Context, args gatehub.GetCardTokenArgs) (*gatehub_external.TokenResponse, error) {
	return nil, nil
}
func (f *fakeGatehub) FreezeCard(ctx context.Context, args gatehub.FreezeCardArgs) error   { return nil }
func (f *fakeGatehub) UnfreezeCard(ctx context.Context, args gatehub.UnfreezeCardArgs) error { return nil }
func (f *fakeGatehub) BlockCard(ctx context.Context, args gatehub.BlockCardArgs) error     { return nil }
func (f *fakeGatehub) ValidateCardProductCode(ctx context.Context, cardProductCode string) error {
	return nil
}
func (f *fakeGatehub) GetPendingThreeDSConfirmations(ctx context.Context, userID string) ([]gatehub_external.PendingThreeDSConfirmation, error) {
	return nil, nil
}
func (f *fakeGatehub) ThreeDSPaymentConfirmation(ctx context.Context, userID, txID string, confirmed bool) error {
	return nil
}
func (f *fakeGatehub) RollbackReserve(ctx context.Context, txID string) error { return nil }
func (f *fakeGatehub) LinkUserToGatewayByWalletID(ctx context.Context, walletID string) error {
	return nil
}
func (f *fakeGatehub) LinkUserToGatewayByExternalID(ctx context.Context, externalID string) error {
	return nil
}

func seedRafikiWalletAddresses(t *testing.T, conn *sqlx.DB, senderPP, senderWalletID, receiverPP, receiverWalletID string) {
	t.Helper()
	_, err := conn.ExecContext(context.Background(),
		`INSERT INTO rafiki_payment_pointers (wallet_id, payment_pointer_id) VALUES ($1, $2), ($3, $4) ON CONFLICT (payment_pointer_id) DO NOTHING`,
		senderWalletID, senderPP, receiverWalletID, receiverPP)
	require.NoError(t, err)
}

func TestOutgoingPaymentCreated_Integration_HappyBatched(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	senderWalletID := uuid.NewString()
	receiverWalletID := uuid.NewString()
	senderPP := "wa_sender_" + uuid.NewString()[:8]
	receiverPP := "wa_receiver_" + uuid.NewString()[:8]
	opID := uuid.NewString()
	eventID := uuid.NewString()
	ipID := "ip_" + uuid.NewString()[:8]
	senderLinkedID := "la_sender_" + uuid.NewString()[:8]
	receiverLinkedID := "la_receiver_" + uuid.NewString()[:8]

	seedRafikiWalletAddresses(t, conn, senderPP, senderWalletID, receiverPP, receiverWalletID)

	extMock := rafiki_external_mock.NewMockClient(ctrl)
	extMock.EXPECT().GetIncomingPayment(gomock.Any(), ipID).Return(&rafiki_external.GetIncomingPaymentIncomingPayment{WalletAddressId: receiverPP}, nil)
	extMock.EXPECT().FundOutgoingPayment(gomock.Any(), opID).Return(nil)

	walletsMock := wallets_mock.NewMockClient(ctrl)
	walletsMock.EXPECT().Get(gomock.Any(), receiverWalletID).Return(&wallets.Wallet{ID: receiverWalletID}, nil)

	senderAcc := linkedaccounts.LinkedAccount{ID: senderLinkedID, WalletID: senderWalletID, Provider: "gatehub", SendCurrency: currency.EUR}
	receiverAcc := linkedaccounts.LinkedAccount{ID: receiverLinkedID, WalletID: receiverWalletID, Provider: "gatehub", ReceiveCurrency: currency.EUR}
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), senderWalletID).Return([]linkedaccounts.LinkedAccount{senderAcc}, nil)
	linkedMock.EXPECT().ListBalances(gomock.Any(), receiverWalletID).Return([]linkedaccounts.LinkedAccount{receiverAcc}, nil)

	var reserveCalled bool
	fakeGH := &fakeGatehub{
		reserveBalance: func(_ context.Context, accID, txID string, amt currency.Amount, _ time.Duration) (*gatehub.Balance, error) {
			reserveCalled = true
			assert.Equal(t, senderLinkedID, accID)
			assert.Equal(t, opID, txID)
			assert.Equal(t, uint64(50), amt.Value)
			return &gatehub.Balance{}, nil
		},
	}

	b := NewTestBackends(
		func(tb *TestBackends) { tb.SetDB(conn) },
		func(tb *TestBackends) { tb.SetExternal(extMock) },
		func(tb *TestBackends) { tb.SetWallets(walletsMock) },
		func(tb *TestBackends) { tb.SetLinkedAccounts(linkedMock) },
		func(tb *TestBackends) { tb.SetGatehub(fakeGH) },
	)

	hook := webhook{
		ID:   eventID,
		Type: "outgoing_payment.created",
		Data: mustMarshalOutgoingPaymentData(opID, senderPP, "https://example.com/incoming-payments/"+ipID, "50", "50", "0"),
	}
	err := outgoingPaymentCreated(ctx, b, hook)
	require.NoError(t, err)
	assert.True(t, reserveCalled, "ReserveBalance should be called for batched payment")

	var count int
	err = conn.GetContext(ctx, &count, "SELECT COUNT(*) FROM rafiki_outgoing_payments WHERE id = $1", opID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "rafiki_outgoing_payments should have one row (accounting)")
	var fromWallet, toWallet string
	var amount int64
	var amountAsset string
	err = conn.QueryRowContext(ctx, "SELECT from_wallet, to_wallet, amount, amount_asset FROM rafiki_outgoing_payments WHERE id = $1", opID).Scan(&fromWallet, &toWallet, &amount, &amountAsset)
	require.NoError(t, err)
	assert.Equal(t, senderWalletID, fromWallet)
	assert.Equal(t, receiverWalletID, toWallet)
	assert.Equal(t, int64(50), amount)
	assert.Equal(t, "EUR", amountAsset)
}

func TestOutgoingPaymentCreated_Integration_UnhappyInvalidJSON(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	b := NewTestBackends(func(tb *TestBackends) { tb.SetDB(conn) })

	hook := webhook{ID: "evt_1", Type: "outgoing_payment.created", Data: []byte(`{invalid}`)}
	err := outgoingPaymentCreated(ctx, b, hook)
	require.Error(t, err)
}

func TestOutgoingPaymentCreated_Integration_UnhappyInvalidDebitAmount(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	b := NewTestBackends(func(tb *TestBackends) { tb.SetDB(conn) })

	hook := webhook{
		ID:   "evt_1",
		Type: "outgoing_payment.created",
		Data: []byte(`{"id":"op_1","walletAddressId":"wa_1","state":"FUNDING","receiver":"https://example.com/ip/ip_1","debitAmount":{"value":"not_a_number","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"100","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"0","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:00:00Z"}`),
	}
	err := outgoingPaymentCreated(ctx, b, hook)
	require.Error(t, err)
}

func TestOutgoingPaymentCreated_Integration_UnhappyAccountNotFound_Cancels(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	senderWalletID := uuid.NewString()
	receiverWalletID := uuid.NewString()
	senderPP := "wa_sender_" + uuid.NewString()[:8]
	receiverPP := "wa_receiver_" + uuid.NewString()[:8]
	ipID := "ip_" + uuid.NewString()[:8]
	opID := uuid.NewString()

	seedRafikiWalletAddresses(t, conn, senderPP, senderWalletID, receiverPP, receiverWalletID)

	extMock := rafiki_external_mock.NewMockClient(ctrl)
	extMock.EXPECT().GetIncomingPayment(gomock.Any(), ipID).Return(&rafiki_external.GetIncomingPaymentIncomingPayment{WalletAddressId: receiverPP}, nil)
	extMock.EXPECT().CancelOutgoingPayment(gomock.Any(), opID, gomock.Any()).Return(nil)

	walletsMock := wallets_mock.NewMockClient(ctrl)
	walletsMock.EXPECT().Get(gomock.Any(), receiverWalletID).Return(&wallets.Wallet{ID: receiverWalletID}, nil)

	// No sender account for EUR -> getAccounts returns ErrNotFound after first ListBalances
	// Handler cancels (receiver ListBalances never called)
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), senderWalletID).Return([]linkedaccounts.LinkedAccount{}, nil)

	b := NewTestBackends(
		func(tb *TestBackends) { tb.SetDB(conn) },
		func(tb *TestBackends) { tb.SetExternal(extMock) },
		func(tb *TestBackends) { tb.SetWallets(walletsMock) },
		func(tb *TestBackends) { tb.SetLinkedAccounts(linkedMock) },
		func(tb *TestBackends) { tb.SetGatehub(&fakeGatehub{}) },
	)

	hook := webhook{
		ID:   uuid.NewString(),
		Type: "outgoing_payment.created",
		Data: mustMarshalOutgoingPaymentData(opID, senderPP, "https://example.com/incoming-payments/"+ipID, "50", "50", "0"),
	}
	err := outgoingPaymentCreated(ctx, b, hook)
	require.NoError(t, err)

	var count int
	_ = conn.GetContext(ctx, &count, "SELECT COUNT(*) FROM rafiki_outgoing_payments WHERE id = $1", opID)
	assert.Equal(t, 0, count, "should not insert when account not found (cancel path)")
}

func TestOutgoingPaymentCompleted_Integration_HappyBatched_FinalisesReserve(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	senderWalletID := uuid.NewString()
	receiverWalletID := uuid.NewString()
	senderPP := "wa_sender_" + uuid.NewString()[:8]
	opID := uuid.NewString()
	senderLinkedID := "la_sender_" + uuid.NewString()[:8]

	seedRafikiWalletAddresses(t, conn, senderPP, senderWalletID, "wa_rec", receiverWalletID)
	_, err := conn.ExecContext(ctx, `INSERT INTO rafiki_outgoing_payments (id, event_id, from_wallet, to_wallet, amount, amount_asset) VALUES ($1, $2, $3, $4, 5000, 'EUR') ON CONFLICT DO NOTHING`,
		opID, uuid.NewString(), senderWalletID, receiverWalletID)
	require.NoError(t, err)

	var finaliseCalled bool
	fakeGH := &fakeGatehub{
		finaliseReserve: func(_ context.Context, txID string) error {
			finaliseCalled = true
			assert.Equal(t, opID, txID)
			return nil
		},
	}

	extMock := rafiki_external_mock.NewMockClient(ctrl)
	extMock.EXPECT().WithdrawOutgoingPaymentLiquidity(gomock.Any(), opID, uint64(0)).Return(nil)

	senderAcc := linkedaccounts.LinkedAccount{ID: senderLinkedID, WalletID: senderWalletID, Provider: "gatehub", SendCurrency: currency.EUR}
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), senderWalletID).Return([]linkedaccounts.LinkedAccount{senderAcc}, nil)

	b := NewTestBackends(
		func(tb *TestBackends) { tb.SetDB(conn) },
		func(tb *TestBackends) { tb.SetExternal(extMock) },
		func(tb *TestBackends) { tb.SetLinkedAccounts(linkedMock) },
		func(tb *TestBackends) { tb.SetGatehub(fakeGH) },
	)

	hook := webhook{
		ID:   uuid.NewString(),
		Type: "outgoing_payment.completed",
		Data: mustMarshalOutgoingPaymentData(opID, senderPP, "https://example.com/ip/ip_1", "50", "50", "50"),
	}
	err = outgoingPaymentCompleted(ctx, b, hook)
	require.NoError(t, err)
	assert.True(t, finaliseCalled, "FinaliseReserve should be called when batched row exists")
}

// Outgoing completed with immediate payment: FinaliseReserve must NOT be called
func TestOutgoingPaymentCompleted_Integration_Unhappy_NoBatchedRow_DoesNotFinalise(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	senderWalletID := uuid.NewString()
	senderPP := "wa_sender_" + uuid.NewString()[:8]
	opID := uuid.NewString()
	senderLinkedID := "la_sender_" + uuid.NewString()[:8]

	seedRafikiWalletAddresses(t, conn, senderPP, senderWalletID, "wa_rec", uuid.NewString())

	fakeGH := &fakeGatehub{
		finaliseReserve: func(_ context.Context, _ string) error {
			t.Error("FinaliseReserve must not be called on immediate payments")
			return nil
		},
	}

	extMock := rafiki_external_mock.NewMockClient(ctrl)
	extMock.EXPECT().WithdrawOutgoingPaymentLiquidity(gomock.Any(), opID, uint64(0)).Return(nil)

	senderAcc := linkedaccounts.LinkedAccount{ID: senderLinkedID, WalletID: senderWalletID, Provider: "gatehub", SendCurrency: currency.EUR}
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), senderWalletID).Return([]linkedaccounts.LinkedAccount{senderAcc}, nil)

	b := NewTestBackends(
		func(tb *TestBackends) { tb.SetDB(conn) },
		func(tb *TestBackends) { tb.SetExternal(extMock) },
		func(tb *TestBackends) { tb.SetLinkedAccounts(linkedMock) },
		func(tb *TestBackends) { tb.SetGatehub(fakeGH) },
	)

	hook := webhook{
		ID:   uuid.NewString(),
		Type: "outgoing_payment.completed",
		Data: mustMarshalOutgoingPaymentData(opID, senderPP, "https://example.com/ip/ip_1", "50", "50", "50"),
	}
	err := outgoingPaymentCompleted(ctx, b, hook)
	require.NoError(t, err)
}

func TestOutgoingPaymentCompleted_Integration_Unhappy_InvalidSentAmount(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	b := NewTestBackends(func(tb *TestBackends) { tb.SetDB(conn) })

	hook := webhook{
		ID:   uuid.NewString(),
		Type: "outgoing_payment.completed",
		Data: []byte(`{"id":"op_1","walletAddressId":"wa_1","state":"COMPLETED","receiver":"https://example.com/ip/ip_1","debitAmount":{"value":"50","assetCode":"EUR","assetScale":2},"receiveAmount":{"value":"50","assetCode":"EUR","assetScale":2},"sentAmount":{"value":"not_a_number","assetCode":"EUR","assetScale":2},"createdAt":"2024-01-15T10:00:00Z","updatedAt":"2024-01-15T10:00:00Z"}`),
	}
	err := outgoingPaymentCompleted(ctx, b, hook)
	require.Error(t, err)
}

func TestIncomingPaymentCreated_Integration_Happy_InsertsRecord(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)

	b := NewTestBackends(func(tb *TestBackends) { tb.SetDB(conn) })

	ipID := "ip_" + uuid.NewString()[:8]
	paymentPointerID := "wa_" + uuid.NewString()[:8]
	hook := webhook{
		ID:   uuid.NewString(),
		Type: "incoming_payment.created",
		Data: mustMarshalIncomingPaymentData(ipID, paymentPointerID, "0", false),
	}
	err := incomingPaymentCreated(ctx, b, hook)
	require.NoError(t, err)

	var count int
	err = conn.GetContext(ctx, &count, "SELECT COUNT(*) FROM rafiki_incoming_payments WHERE payment_id = $1", ipID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	var completed bool
	var receivedAmt int64
	var receivedAsset string
	err = conn.QueryRowContext(ctx, "SELECT completed, received_amount, received_amount_asset FROM rafiki_incoming_payments WHERE payment_id = $1", ipID).Scan(&completed, &receivedAmt, &receivedAsset)
	require.NoError(t, err)
	assert.False(t, completed)
	assert.Equal(t, int64(0), receivedAmt)
	assert.Equal(t, "EUR", receivedAsset, "asset comes from receivedAmount when incomingAmount is nil")
}

func TestIncomingPaymentCreated_Integration_UnhappyInvalidJSON(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	b := NewTestBackends(func(tb *TestBackends) { tb.SetDB(conn) })

	hook := webhook{ID: "evt_1", Type: "incoming_payment.created", Data: []byte(`not json`)}
	err := incomingPaymentCreated(ctx, b, hook)
	require.Error(t, err)
}

func TestIncomingPaymentCompleted_Integration_Happy_AssignsAndUpdates(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiverWalletID := uuid.NewString()
	receiverPP := "wa_receiver_" + uuid.NewString()[:8]
	ipID := "ip_" + uuid.NewString()[:8]
	receiverLinkedID := "la_receiver_" + uuid.NewString()[:8]

	seedRafikiWalletAddresses(t, conn, receiverPP, receiverWalletID, "other_pp", uuid.NewString())
	_, err := conn.ExecContext(ctx, `INSERT INTO rafiki_incoming_payments (payment_id, payment_pointer_id, received_amount, received_amount_asset, completed) VALUES ($1, $2, 0, 'EUR', false) ON CONFLICT DO NOTHING`,
		ipID, receiverPP)
	require.NoError(t, err)

	var assignCalled bool
	fakeGH := &fakeGatehub{
		assignBalance: func(_ context.Context, accID, trxID string, amt currency.Amount) (*gatehub.Balance, error) {
			assignCalled = true
			assert.Equal(t, receiverLinkedID, accID)
			assert.Equal(t, ipID, trxID)
			assert.Equal(t, uint64(5000), amt.Value)
			return &gatehub.Balance{}, nil
		},
	}

	extMock := rafiki_external_mock.NewMockClient(ctrl)
	extMock.EXPECT().WithdrawIncomingPaymentLiquidity(gomock.Any(), ipID, uint64(0)).Return(nil)

	receiverAcc := linkedaccounts.LinkedAccount{ID: receiverLinkedID, WalletID: receiverWalletID, Provider: "gatehub", ReceiveCurrency: currency.EUR}
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), receiverWalletID).Return([]linkedaccounts.LinkedAccount{receiverAcc}, nil)

	b := NewTestBackends(
		func(tb *TestBackends) { tb.SetDB(conn) },
		func(tb *TestBackends) { tb.SetExternal(extMock) },
		func(tb *TestBackends) { tb.SetLinkedAccounts(linkedMock) },
		func(tb *TestBackends) { tb.SetGatehub(fakeGH) },
	)

	hook := webhook{
		ID:   uuid.NewString(),
		Type: "incoming_payment.completed",
		Data: mustMarshalIncomingPaymentData(ipID, receiverPP, "5000", true),
	}
	err = incomingPaymentCompleted(ctx, b, hook)
	require.NoError(t, err)
	assert.True(t, assignCalled)

	var completed bool
	var receivedAmt int64
	err = conn.QueryRowContext(ctx, "SELECT completed, received_amount FROM rafiki_incoming_payments WHERE payment_id = $1", ipID).Scan(&completed, &receivedAmt)
	require.NoError(t, err)
	assert.True(t, completed)
	assert.Equal(t, int64(5000), receivedAmt)
}

func TestIncomingPaymentCompleted_Integration_UnhappyInvalidReceivedAmount(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	receiverPP := "wa_" + uuid.NewString()[:8]
	receiverWalletID := uuid.NewString()
	seedRafikiWalletAddresses(t, conn, receiverPP, receiverWalletID, "other", uuid.NewString())

	b := NewTestBackends(func(tb *TestBackends) { tb.SetDB(conn) })

	hook := webhook{
		ID:   uuid.NewString(),
		Type: "incoming_payment.completed",
		Data: []byte(`{"id":"ip_1","walletAddressId":"` + receiverPP + `","createdAt":"2024-01-15T10:00:00Z","expiresAt":"2024-01-16T10:00:00Z","receivedAmount":{"value":"not_a_number","assetCode":"EUR","assetScale":2},"completed":true}`),
	}
	err := incomingPaymentCompleted(ctx, b, hook)
	require.Error(t, err)
}

func TestIncomingPaymentExpired_Integration_Happy_AssignsAndUpdates(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiverWalletID := uuid.NewString()
	receiverPP := "wa_receiver_" + uuid.NewString()[:8]
	ipID := "ip_" + uuid.NewString()[:8]
	receiverLinkedID := "la_receiver_" + uuid.NewString()[:8]

	seedRafikiWalletAddresses(t, conn, receiverPP, receiverWalletID, "other_pp", uuid.NewString())
	_, err := conn.ExecContext(ctx, `INSERT INTO rafiki_incoming_payments (payment_id, payment_pointer_id, received_amount, received_amount_asset, completed) VALUES ($1, $2, 0, 'EUR', false) ON CONFLICT DO NOTHING`,
		ipID, receiverPP)
	require.NoError(t, err)

	var assignCalled bool
	fakeGH := &fakeGatehub{
		assignBalance: func(_ context.Context, accID, trxID string, amt currency.Amount) (*gatehub.Balance, error) {
			assignCalled = true
			assert.Equal(t, receiverLinkedID, accID)
			assert.Equal(t, ipID, trxID)
			return &gatehub.Balance{}, nil
		},
	}

	extMock := rafiki_external_mock.NewMockClient(ctrl)
	extMock.EXPECT().WithdrawIncomingPaymentLiquidity(gomock.Any(), ipID, uint64(0)).Return(nil)

	receiverAcc := linkedaccounts.LinkedAccount{ID: receiverLinkedID, WalletID: receiverWalletID, Provider: "gatehub", ReceiveCurrency: currency.EUR}
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), receiverWalletID).Return([]linkedaccounts.LinkedAccount{receiverAcc}, nil)

	b := NewTestBackends(
		func(tb *TestBackends) { tb.SetDB(conn) },
		func(tb *TestBackends) { tb.SetExternal(extMock) },
		func(tb *TestBackends) { tb.SetLinkedAccounts(linkedMock) },
		func(tb *TestBackends) { tb.SetGatehub(fakeGH) },
	)

	hook := webhook{
		ID:   uuid.NewString(),
		Type: "incoming_payment.expired",
		Data: mustMarshalIncomingPaymentData(ipID, receiverPP, "2500", false),
	}
	err = incomingPaymentExpired(ctx, b, hook)
	require.NoError(t, err)
	assert.True(t, assignCalled)

	var completed bool
	var receivedAmt int64
	err = conn.QueryRowContext(ctx, "SELECT completed, received_amount FROM rafiki_incoming_payments WHERE payment_id = $1", ipID).Scan(&completed, &receivedAmt)
	require.NoError(t, err)
	assert.True(t, completed)
	assert.Equal(t, int64(2500), receivedAmt)
}

func TestIncomingPaymentPartialPaymentReceived_Integration_Happy_Assigns(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiverWalletID := uuid.NewString()
	receiverPP := "wa_receiver_" + uuid.NewString()[:8]
	ipID := "ip_" + uuid.NewString()[:8]
	receiverLinkedID := "la_receiver_" + uuid.NewString()[:8]

	seedRafikiWalletAddresses(t, conn, receiverPP, receiverWalletID, "other_pp", uuid.NewString())

	var assignCalled bool
	fakeGH := &fakeGatehub{
		assignBalance: func(_ context.Context, accID, trxID string, amt currency.Amount) (*gatehub.Balance, error) {
			assignCalled = true
			assert.Equal(t, receiverLinkedID, accID)
			assert.Equal(t, ipID, trxID)
			assert.Equal(t, uint64(1000), amt.Value)
			return &gatehub.Balance{}, nil
		},
	}

	extMock := rafiki_external_mock.NewMockClient(ctrl)
	extMock.EXPECT().WithdrawIncomingPaymentLiquidity(gomock.Any(), ipID, uint64(0)).Return(nil)

	receiverAcc := linkedaccounts.LinkedAccount{ID: receiverLinkedID, WalletID: receiverWalletID, Provider: "gatehub", ReceiveCurrency: currency.EUR}
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), receiverWalletID).Return([]linkedaccounts.LinkedAccount{receiverAcc}, nil)

	b := NewTestBackends(
		func(tb *TestBackends) { tb.SetDB(conn) },
		func(tb *TestBackends) { tb.SetExternal(extMock) },
		func(tb *TestBackends) { tb.SetLinkedAccounts(linkedMock) },
		func(tb *TestBackends) { tb.SetGatehub(fakeGH) },
	)

	hook := webhook{
		ID:   uuid.NewString(),
		Type: "incoming_payment.partial_payment_received",
		Data: mustMarshalIncomingPaymentData(ipID, receiverPP, "1000", false),
	}
	err := incomingPaymentPartialPaymentReceived(ctx, b, hook)
	require.NoError(t, err)
	assert.True(t, assignCalled)
}

// Partial payment handler increments received_amount when a row exists (created by incoming_payment.created)
func TestIncomingPaymentPartialPaymentReceived_Integration_Happy_UpdatesReceivedAmount(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	receiverWalletID := uuid.NewString()
	receiverPP := "wa_receiver_" + uuid.NewString()[:8]
	ipID := "ip_" + uuid.NewString()[:8]
	receiverLinkedID := "la_receiver_" + uuid.NewString()[:8]

	seedRafikiWalletAddresses(t, conn, receiverPP, receiverWalletID, "other_pp", uuid.NewString())
	_, err := conn.ExecContext(ctx, `INSERT INTO rafiki_incoming_payments (payment_id, payment_pointer_id, received_amount, received_amount_asset, completed) VALUES ($1, $2, 0, 'EUR', false) ON CONFLICT DO NOTHING`,
		ipID, receiverPP)
	require.NoError(t, err)

	var assignCalled bool
	fakeGH := &fakeGatehub{
		assignBalance: func(_ context.Context, accID, trxID string, amt currency.Amount) (*gatehub.Balance, error) {
			assignCalled = true
			assert.Equal(t, uint64(2500), amt.Value)
			return &gatehub.Balance{}, nil
		},
	}
	extMock := rafiki_external_mock.NewMockClient(ctrl)
	extMock.EXPECT().WithdrawIncomingPaymentLiquidity(gomock.Any(), ipID, uint64(0)).Return(nil)
	receiverAcc := linkedaccounts.LinkedAccount{ID: receiverLinkedID, WalletID: receiverWalletID, Provider: "gatehub", ReceiveCurrency: currency.EUR}
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), receiverWalletID).Return([]linkedaccounts.LinkedAccount{receiverAcc}, nil)

	b := NewTestBackends(
		func(tb *TestBackends) { tb.SetDB(conn) },
		func(tb *TestBackends) { tb.SetExternal(extMock) },
		func(tb *TestBackends) { tb.SetLinkedAccounts(linkedMock) },
		func(tb *TestBackends) { tb.SetGatehub(fakeGH) },
	)

	hook := webhook{
		ID:   uuid.NewString(),
		Type: "incoming_payment.partial_payment_received",
		Data: mustMarshalIncomingPaymentData(ipID, receiverPP, "2500", false),
	}
	err = incomingPaymentPartialPaymentReceived(ctx, b, hook)
	require.NoError(t, err)
	assert.True(t, assignCalled)

	var receivedAmt int64
	err = conn.QueryRowContext(ctx, "SELECT received_amount FROM rafiki_incoming_payments WHERE payment_id = $1", ipID).Scan(&receivedAmt)
	require.NoError(t, err)
	assert.Equal(t, int64(2500), receivedAmt, "partial handler should increment received_amount")
}

func TestIncomingPaymentPartialPaymentReceived_Integration_Unhappy_InvalidJSON(t *testing.T) {
	conn := requireTestDB(t)
	b := NewTestBackends(func(tb *TestBackends) { tb.SetDB(conn) })

	hook := webhook{ID: "evt_1", Type: "incoming_payment.partial_payment_received", Data: []byte(`not json`)}
	err := incomingPaymentPartialPaymentReceived(context.Background(), b, hook)
	require.Error(t, err)
}

func TestIncomingPaymentExpired_Integration_Unhappy_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	b := NewTestBackends(func(tb *TestBackends) { tb.SetDB(conn) })

	hook := webhook{ID: "evt_1", Type: "incoming_payment.expired", Data: []byte(`not json`)}
	err := incomingPaymentExpired(ctx, b, hook)
	require.Error(t, err)
}

func TestOutgoingPaymentCreated_Integration_UnhappySameWallet_Cancels(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	walletID := uuid.NewString()
	// Same PP for sender and receiver so both resolve to same wallet (schema allows one PP per wallet)
	senderPP := "wa_same_" + uuid.NewString()[:8]
	ipID := "ip_" + uuid.NewString()[:8]
	opID := uuid.NewString()

	_, err := conn.ExecContext(ctx, `INSERT INTO rafiki_payment_pointers (wallet_id, payment_pointer_id) VALUES ($1, $2) ON CONFLICT (payment_pointer_id) DO NOTHING`, walletID, senderPP)
	require.NoError(t, err)

	extMock := rafiki_external_mock.NewMockClient(ctrl)
	extMock.EXPECT().GetIncomingPayment(gomock.Any(), ipID).Return(&rafiki_external.GetIncomingPaymentIncomingPayment{WalletAddressId: senderPP}, nil)
	extMock.EXPECT().CancelOutgoingPayment(gomock.Any(), opID, "sending wallet cannot be the same as receiving wallet").Return(nil)

	walletsMock := wallets_mock.NewMockClient(ctrl)
	walletsMock.EXPECT().Get(gomock.Any(), walletID).Return(&wallets.Wallet{ID: walletID}, nil)

	acc := linkedaccounts.LinkedAccount{ID: "la_1", WalletID: walletID, Provider: "gatehub", SendCurrency: currency.EUR, ReceiveCurrency: currency.EUR}
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{acc}, nil).AnyTimes()

	b := NewTestBackends(
		func(tb *TestBackends) { tb.SetDB(conn) },
		func(tb *TestBackends) { tb.SetExternal(extMock) },
		func(tb *TestBackends) { tb.SetWallets(walletsMock) },
		func(tb *TestBackends) { tb.SetLinkedAccounts(linkedMock) },
		func(tb *TestBackends) { tb.SetGatehub(&fakeGatehub{}) },
	)

	hook := webhook{
		ID:   uuid.NewString(),
		Type: "outgoing_payment.created",
		Data: mustMarshalOutgoingPaymentData(opID, senderPP, "https://example.com/incoming-payments/"+ipID, "50", "50", "0"),
	}
	err = outgoingPaymentCreated(ctx, b, hook)
	require.NoError(t, err)

	var count int
	_ = conn.GetContext(ctx, &count, "SELECT COUNT(*) FROM rafiki_outgoing_payments WHERE id = $1", opID)
	assert.Equal(t, 0, count)
}

func TestOutgoingPaymentCreated_Integration_HappyImmediate(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	senderWalletID := uuid.NewString()
	receiverWalletID := uuid.NewString()
	senderPP := "wa_sender_" + uuid.NewString()[:8]
	receiverPP := "wa_receiver_" + uuid.NewString()[:8]
	opID := uuid.NewString()
	ipID := "ip_" + uuid.NewString()[:8]
	senderLinkedID := "la_sender_" + uuid.NewString()[:8]
	receiverLinkedID := "la_receiver_" + uuid.NewString()[:8]

	seedRafikiWalletAddresses(t, conn, senderPP, senderWalletID, receiverPP, receiverWalletID)

	// getAccounts is called once at start and once inside immediatePayment, so expect each external call twice
	extMock := rafiki_external_mock.NewMockClient(ctrl)
	extMock.EXPECT().GetIncomingPayment(gomock.Any(), ipID).Return(&rafiki_external.GetIncomingPaymentIncomingPayment{WalletAddressId: receiverPP}, nil).Times(2)
	extMock.EXPECT().FundOutgoingPayment(gomock.Any(), opID).Return(nil)

	walletsMock := wallets_mock.NewMockClient(ctrl)
	walletsMock.EXPECT().Get(gomock.Any(), receiverWalletID).Return(&wallets.Wallet{ID: receiverWalletID}, nil).Times(2)

	senderAcc := linkedaccounts.LinkedAccount{ID: senderLinkedID, WalletID: senderWalletID, Provider: "gatehub", SendCurrency: currency.EUR}
	receiverAcc := linkedaccounts.LinkedAccount{ID: receiverLinkedID, WalletID: receiverWalletID, Provider: "gatehub", ReceiveCurrency: currency.EUR}
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), senderWalletID).Return([]linkedaccounts.LinkedAccount{senderAcc}, nil).Times(2)
	linkedMock.EXPECT().ListBalances(gomock.Any(), receiverWalletID).Return([]linkedaccounts.LinkedAccount{receiverAcc}, nil).Times(2)

	fakeGH := &fakeGatehub{
		getBalance: func(_ context.Context, _ string) (*gatehub.Balance, error) {
			return &gatehub.Balance{Available: currency.Amount{Value: 20000, Currency: currency.EUR, Scale: 2}}, nil
		},
	}

	payMock := payments_mock.NewMockClient(ctrl)
	payMock.EXPECT().Lookup(gomock.Any(), opID).Return(nil, payments.ErrNotFound)
	payMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&payments.Payment{ID: opID, State: payments.StateCreated}, nil)
	payMock.EXPECT().Confirm(gomock.Any(), opID).Return(nil, nil, nil)

	b := NewTestBackends(
		func(tb *TestBackends) { tb.SetDB(conn) },
		func(tb *TestBackends) { tb.SetExternal(extMock) },
		func(tb *TestBackends) { tb.SetWallets(walletsMock) },
		func(tb *TestBackends) { tb.SetLinkedAccounts(linkedMock) },
		func(tb *TestBackends) { tb.SetGatehub(fakeGH) },
		func(tb *TestBackends) { tb.SetPayments(payMock) },
	)

	hook := webhook{
		ID:   uuid.NewString(),
		Type: "outgoing_payment.created",
		Data: mustMarshalOutgoingPaymentData(opID, senderPP, "https://example.com/incoming-payments/"+ipID, "150", "150", "0"),
	}
	err := outgoingPaymentCreated(ctx, b, hook)
	require.NoError(t, err)

	var count int
	_ = conn.GetContext(ctx, &count, "SELECT COUNT(*) FROM rafiki_outgoing_payments WHERE id = $1", opID)
	assert.Equal(t, 0, count, "immediate payment should not insert in rafiki_outgoing_payments")
}

func mustMarshalOutgoingPaymentData(id, walletAddrID, receiver, debit, receive, sent string) []byte {
	out := map[string]interface{}{
		"id":              id,
		"walletAddressId": walletAddrID,
		"state":           "FUNDING",
		"receiver":        receiver,
		"debitAmount":     map[string]interface{}{"value": debit, "assetCode": "EUR", "assetScale": 2},
		"receiveAmount":   map[string]interface{}{"value": receive, "assetCode": "EUR", "assetScale": 2},
		"sentAmount":      map[string]interface{}{"value": sent, "assetCode": "EUR", "assetScale": 2},
		"createdAt":      "2024-01-15T10:00:00Z",
		"updatedAt":      "2024-01-15T10:00:00Z",
	}
	b, _ := json.Marshal(out)
	return b
}

func mustMarshalIncomingPaymentData(id, walletAddrID string, receivedValue string, completed bool) []byte {
	out := map[string]interface{}{
		"id":              id,
		"walletAddressId": walletAddrID,
		"createdAt":       "2024-01-15T10:00:00Z",
		"expiresAt":       "2024-01-16T10:00:00Z",
		"completed":       completed,
		"receivedAmount":  map[string]interface{}{"value": receivedValue, "assetCode": "EUR", "assetScale": 2},
	}
	b, _ := json.Marshal(out)
	return b
}
