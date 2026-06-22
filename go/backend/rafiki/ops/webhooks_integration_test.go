package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.temporal.io/sdk/testsuite"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/db"
	kyc_mock "github.com/interledger/interledger-app/go/backend/kyc/client/mock"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	linkedaccounts_mock "github.com/interledger/interledger-app/go/backend/linkedaccounts/client/mock"
	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/rafiki"
	rafiki_mock "github.com/interledger/interledger-app/go/backend/rafiki/client/mock"
	"github.com/interledger/interledger-app/go/backend/transactions"
	transactions_mock "github.com/interledger/interledger-app/go/backend/transactions/client/mock"
	"github.com/interledger/interledger-app/go/backend/wallets"
	wallets_mock "github.com/interledger/interledger-app/go/backend/wallets/client/mock"
	"github.com/interledger/interledger-app/go/env"
	"github.com/interledger/interledger-app/go/pacioli"
	pacioli_mock "github.com/interledger/interledger-app/go/pacioli/client/mock"

	_ "github.com/lib/pq"
)

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
	connStr := fmt.Sprintf("postgres://postgres:password@%s:%s/%%s?sslmode=disable", host, port.Port())
	return container, connStr, nil
}

func TestMain(m *testing.M) {
	env.SetRafikiNodeEnabled(true)

	if os.Getenv("DB_URL") == "" {
		ctx := context.Background()
		container, connStr, err := startPostgresContainer(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rafiki/ops: Postgres container failed to start (Docker required): %v\n", err)
			os.Exit(1)
		}
		_ = os.Setenv("DB_URL", connStr)
		defer func() { _ = container.Terminate(ctx) }()
	}
	os.Exit(m.Run())
}

func requireTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	baseString := os.Getenv("DB_URL")
	if baseString == "" {
		t.Skip("DB_URL not set; skipping integration test")
	}
	connString := baseString
	if strings.Contains(baseString, "%s") {
		connString = fmt.Sprintf(baseString, "postgres")
	}
	tryConn, err := sqlx.Connect("postgres", connString)
	if err != nil {
		t.Skipf("DB connection failed; skipping: %v", err)
	}
	_ = tryConn.Close()
	conn := db.MigrateTestDB(t, context.Background())
	return conn
}

type RafikiWorkflowSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func TestRafikiWorkflowSuite(t *testing.T) {
	suite.Run(t, new(RafikiWorkflowSuite))
}

func (s *RafikiWorkflowSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *RafikiWorkflowSuite) AfterTest(_, _ string) {
	s.env.AssertExpectations(s.T())
}

// --- RafikiIncomingPaymentFinalizedWorkflow ---------------------------------

func (s *RafikiWorkflowSuite) TestIncomingPaymentFinalized_ZeroAmount_NoOps() {
	args := RafikiIncomingPaymentFinalizedArgs{
		IncomingPayment: incomingPaymentData{
			ID:              "ip_zero",
			WalletAddressID: "wa_123",
			ReceivedAmount:  amount{Value: "0", AssetCode: "EUR", AssetScale: 2},
		},
		WebhookType: "incoming_payment.completed",
	}

	s.env.ExecuteWorkflow(RafikiIncomingPaymentFinalizedWorkflow, args)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *RafikiWorkflowSuite) TestIncomingPaymentFinalized_InvalidAmount_ReturnsError() {
	args := RafikiIncomingPaymentFinalizedArgs{
		IncomingPayment: incomingPaymentData{
			ID:              "ip_bad",
			WalletAddressID: "wa_123",
			ReceivedAmount:  amount{Value: "not_a_number", AssetCode: "EUR", AssetScale: 2},
		},
		WebhookType: "incoming_payment.completed",
	}

	s.env.ExecuteWorkflow(RafikiIncomingPaymentFinalizedWorkflow, args)
	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

func (s *RafikiWorkflowSuite) TestIncomingPaymentFinalized_HappyPath() {
	ip := incomingPaymentData{
		ID:              "ip_happy",
		WalletAddressID: "wa_recv",
		ReceivedAmount:  amount{Value: "5000", AssetCode: "EUR", AssetScale: 2},
	}
	args := RafikiIncomingPaymentFinalizedArgs{IncomingPayment: ip, WebhookType: "incoming_payment.completed"}

	var a *Activity
	accountInfo := GatehubLinkedAccountInfo{
		WalletID:   "wallet_ip_happy",
		ProviderID: "gh_la_ip_happy",
	}

	s.env.OnActivity(a.GetGatehubLinkedAccountInfo, mock.Anything, ip.WalletAddressID).Return(&accountInfo, nil)
	s.env.OnActivity(a.TransferFromIntermediaryToUser, mock.Anything, accountInfo, ip.ReceivedAmount).Return("gh_tx_123", nil)
	s.env.OnActivity(a.StoreGatehubTransferMapping, mock.Anything, "gh_tx_123", "default-test-workflow-id").Return(nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(RafikiGatehubSignalChannel, nil)
	}, time.Millisecond)

	s.env.OnActivity(a.CreateIncomingPaymentTransaction, mock.Anything, ip).Return(nil)
	s.env.OnActivity(a.CreateAndPostLedgerTransferForIncoming, mock.Anything, ip).Return(nil)
	s.env.OnActivity(a.WithdrawIncomingPaymentLiquidity, mock.Anything, ip.ID).Return(nil)

	s.env.ExecuteWorkflow(RafikiIncomingPaymentFinalizedWorkflow, args)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *RafikiWorkflowSuite) TestIncomingPaymentFinalized_TransferFails() {
	ip := incomingPaymentData{
		ID:              "ip_fail_transfer",
		WalletAddressID: "wa_recv",
		ReceivedAmount:  amount{Value: "1000", AssetCode: "EUR", AssetScale: 2},
	}
	args := RafikiIncomingPaymentFinalizedArgs{IncomingPayment: ip, WebhookType: "incoming_payment.expired"}

	var a *Activity
	accountInfo := GatehubLinkedAccountInfo{
		WalletID:   "wallet_ip_fail_transfer",
		ProviderID: "gh_la_ip_fail_transfer",
	}
	s.env.OnActivity(a.GetGatehubLinkedAccountInfo, mock.Anything, ip.WalletAddressID).Return(&accountInfo, nil)
	s.env.OnActivity(a.TransferFromIntermediaryToUser, mock.Anything, accountInfo, ip.ReceivedAmount).
		Return("", fmt.Errorf("gatehub transfer failed"))

	s.env.ExecuteWorkflow(RafikiIncomingPaymentFinalizedWorkflow, args)
	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

// --- RafikiOutgoingPaymentCreatedWorkflow -----------------------------------

func (s *RafikiWorkflowSuite) TestOutgoingPaymentCreated_HappyPath() {
	op := outgoingPaymentData{
		ID:              "op_happy",
		WalletAddressID: "wa_sender",
		State:           "FUNDING",
		Receiver:        "https://example.com/incoming-payments/ip_recv",
		DebitAmount:     amount{Value: "100", AssetCode: "EUR", AssetScale: 2},
		ReceiveAmount:   amount{Value: "100", AssetCode: "EUR", AssetScale: 2},
		SentAmount:      amount{Value: "0", AssetCode: "EUR", AssetScale: 2},
	}
	var a *Activity

	s.env.OnActivity(a.ValidateOutgoingPayment, mock.Anything, op).Return(&ValidationResult{Valid: true}, nil)
	s.env.OnActivity(a.TransferFromUserToIntermediary, mock.Anything, op.WalletAddressID, op.DebitAmount).Return("gh_tx_456", nil)
	s.env.OnActivity(a.StoreGatehubTransferMapping, mock.Anything, "gh_tx_456", "default-test-workflow-id").Return(nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(RafikiGatehubSignalChannel, nil)
	}, time.Millisecond)

	s.env.OnActivity(a.CreateOutgoingPaymentTransaction, mock.Anything, op).Return(nil)
	s.env.OnActivity(a.ReserveBalanceForOutgoing, mock.Anything, op).Return(nil)
	s.env.OnActivity(a.DepositOutgoingPaymentLiquidity, mock.Anything, op.ID).Return(nil)

	s.env.ExecuteWorkflow(RafikiOutgoingPaymentCreatedWorkflow, op)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *RafikiWorkflowSuite) TestOutgoingPaymentCreated_ValidationFails_Cancels() {
	op := outgoingPaymentData{
		ID:              "op_invalid",
		WalletAddressID: "wa_sender",
		DebitAmount:     amount{Value: "100", AssetCode: "EUR", AssetScale: 2},
	}
	var a *Activity

	s.env.OnActivity(a.ValidateOutgoingPayment, mock.Anything, op).
		Return(&ValidationResult{Valid: false, Reason: "user KYC not approved"}, nil)
	s.env.OnActivity(a.CancelOutgoingPayment, mock.Anything, op.ID, "user KYC not approved").Return(nil)

	s.env.ExecuteWorkflow(RafikiOutgoingPaymentCreatedWorkflow, op)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *RafikiWorkflowSuite) TestOutgoingPaymentCreated_ValidationError_Fails() {
	op := outgoingPaymentData{
		ID:              "op_val_err",
		WalletAddressID: "wa_sender",
		DebitAmount:     amount{Value: "100", AssetCode: "EUR", AssetScale: 2},
	}
	var a *Activity
	s.env.OnActivity(a.ValidateOutgoingPayment, mock.Anything, op).Return(nil, fmt.Errorf("db connection lost"))
	s.env.OnActivity(a.CancelOutgoingPayment, mock.Anything, op.ID, "validation error").Return(nil)

	s.env.ExecuteWorkflow(RafikiOutgoingPaymentCreatedWorkflow, op)
	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

// --- RafikiOutgoingPaymentCompletedWorkflow ---------------------------------

func (s *RafikiWorkflowSuite) TestOutgoingPaymentCompleted_HappyPath() {
	op := outgoingPaymentData{
		ID:              "op_done",
		WalletAddressID: "wa_sender",
		DebitAmount:     amount{Value: "100", AssetCode: "EUR", AssetScale: 2},
		SentAmount:      amount{Value: "100", AssetCode: "EUR", AssetScale: 2},
	}
	var a *Activity

	s.env.OnActivity(a.UpdateOutgoingPaymentTransactionState, mock.Anything, op, "Completed").Return(nil)
	s.env.OnActivity(a.PostLedgerTransferForOutgoing, mock.Anything, op).Return(nil)
	s.env.OnActivity(a.WithdrawOutgoingPaymentLiquidity, mock.Anything, op).Return(nil)

	s.env.ExecuteWorkflow(RafikiOutgoingPaymentCompletedWorkflow, op)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *RafikiWorkflowSuite) TestOutgoingPaymentCompleted_PostTransferFails() {
	op := outgoingPaymentData{
		ID:              "op_post_fail",
		WalletAddressID: "wa_sender",
		DebitAmount:     amount{Value: "100", AssetCode: "EUR", AssetScale: 2},
	}
	var a *Activity

	s.env.OnActivity(a.UpdateOutgoingPaymentTransactionState, mock.Anything, op, "Completed").Return(nil)
	s.env.OnActivity(a.PostLedgerTransferForOutgoing, mock.Anything, op).Return(fmt.Errorf("pacioli post failed"))

	s.env.ExecuteWorkflow(RafikiOutgoingPaymentCompletedWorkflow, op)
	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

// --- RafikiOutgoingPaymentFailedWorkflow ------------------------------------

func (s *RafikiWorkflowSuite) TestOutgoingPaymentFailed_HappyPath() {
	op := outgoingPaymentData{
		ID:              "op_failed",
		WalletAddressID: "wa_sender",
		DebitAmount:     amount{Value: "200", AssetCode: "EUR", AssetScale: 2},
	}
	var a *Activity
	accountInfo := GatehubLinkedAccountInfo{
		WalletID:   "wallet_op_failed",
		ProviderID: "gh_la_op_failed",
	}

	s.env.OnActivity(a.GetGatehubLinkedAccountInfo, mock.Anything, op.WalletAddressID).Return(&accountInfo, nil)
	s.env.OnActivity(a.TransferFromIntermediaryToUser, mock.Anything, accountInfo, op.DebitAmount).Return("gh_tx_refund", nil)
	s.env.OnActivity(a.StoreGatehubTransferMapping, mock.Anything, "gh_tx_refund", "default-test-workflow-id").Return(nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(RafikiGatehubSignalChannel, nil)
	}, time.Millisecond)

	s.env.OnActivity(a.UpdateOutgoingPaymentTransactionState, mock.Anything, op, "Failed").Return(nil)
	s.env.OnActivity(a.VoidLedgerTransferForOutgoing, mock.Anything, op).Return(nil)
	s.env.OnActivity(a.WithdrawOutgoingPaymentLiquidity, mock.Anything, op).Return(nil)

	s.env.ExecuteWorkflow(RafikiOutgoingPaymentFailedWorkflow, op)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *RafikiWorkflowSuite) TestOutgoingPaymentFailed_RefundTransferFails() {
	op := outgoingPaymentData{
		ID:              "op_fail_refund",
		WalletAddressID: "wa_sender",
		DebitAmount:     amount{Value: "200", AssetCode: "EUR", AssetScale: 2},
	}
	var a *Activity
	accountInfo := GatehubLinkedAccountInfo{
		WalletID:   "wallet_op_fail_refund",
		ProviderID: "gh_la_op_fail_refund",
	}
	s.env.OnActivity(a.GetGatehubLinkedAccountInfo, mock.Anything, op.WalletAddressID).Return(&accountInfo, nil)
	s.env.OnActivity(a.TransferFromIntermediaryToUser, mock.Anything, accountInfo, op.DebitAmount).
		Return("", fmt.Errorf("refund transfer failed"))

	s.env.ExecuteWorkflow(RafikiOutgoingPaymentFailedWorkflow, op)
	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

func (s *RafikiWorkflowSuite) TestOutgoingPaymentFailed_VoidFails() {
	op := outgoingPaymentData{
		ID:              "op_void_fail",
		WalletAddressID: "wa_sender",
		DebitAmount:     amount{Value: "200", AssetCode: "EUR", AssetScale: 2},
	}
	var a *Activity
	accountInfo := GatehubLinkedAccountInfo{
		WalletID:   "wallet_op_void_fail",
		ProviderID: "gh_la_op_void_fail",
	}

	s.env.OnActivity(a.GetGatehubLinkedAccountInfo, mock.Anything, op.WalletAddressID).Return(&accountInfo, nil)
	s.env.OnActivity(a.TransferFromIntermediaryToUser, mock.Anything, accountInfo, op.DebitAmount).Return("gh_tx_ok", nil)
	s.env.OnActivity(a.StoreGatehubTransferMapping, mock.Anything, "gh_tx_ok", "default-test-workflow-id").Return(nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(RafikiGatehubSignalChannel, nil)
	}, time.Millisecond)

	s.env.OnActivity(a.UpdateOutgoingPaymentTransactionState, mock.Anything, op, "Failed").Return(nil)
	s.env.OnActivity(a.VoidLedgerTransferForOutgoing, mock.Anything, op).Return(fmt.Errorf("pacioli void failed"))

	s.env.ExecuteWorkflow(RafikiOutgoingPaymentFailedWorkflow, op)
	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

// ---------------------------------------------------------------------------
// Activity unit tests
// ---------------------------------------------------------------------------

func TestValidateOutgoingPayment_KYCNotApproved(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	walletID := uuid.NewString()
	ppID := "wa_sender_" + uuid.NewString()[:8]
	op := outgoingPaymentData{
		ID:              uuid.NewString(),
		WalletAddressID: ppID,
		DebitAmount:     amount{Value: "100", AssetCode: "EUR", AssetScale: 2},
		Receiver:        "https://example.com/incoming-payments/ip_1",
	}

	ab := setupActivityBackends(t, ctrl, walletID, ppID)
	kycMock := kyc_mock.NewMockClient(ctrl)
	kycMock.EXPECT().IsKYCApproved(gomock.Any(), walletID).Return(false, nil)
	ab.SetKYC(kycMock)

	act := NewActivity(ab, gatehub.Config{})

	result, err := act.ValidateOutgoingPayment(ctx, op)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Equal(t, "user KYC not approved", result.Reason)
}

func TestValidateOutgoingPayment_SameWallet_Invalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	walletID := uuid.NewString()
	ppID := "wa_same_" + uuid.NewString()[:8]
	ipID := "ip_" + uuid.NewString()[:8]

	op := outgoingPaymentData{
		ID:              uuid.NewString(),
		WalletAddressID: ppID,
		DebitAmount:     amount{Value: "100", AssetCode: "EUR", AssetScale: 2},
		Receiver:        "https://example.com/incoming-payments/" + ipID,
	}

	ab := setupActivityBackends(t, ctrl, walletID, ppID)
	kycMock := kyc_mock.NewMockClient(ctrl)
	kycMock.EXPECT().IsKYCApproved(gomock.Any(), walletID).Return(true, nil)
	ab.SetKYC(kycMock)

	rafikiMock := rafiki_mock.NewMockClient(ctrl)
	rafikiMock.EXPECT().GetIncomingPayment(gomock.Any(), ipID).
		Return(&rafiki.IncomingPayment{WalletAddressID: ppID}, nil)
	ab.SetRafiki(rafikiMock)

	act := NewActivity(ab, gatehub.Config{})

	result, err := act.ValidateOutgoingPayment(ctx, op)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Equal(t, "sending wallet cannot be the same as receiving wallet", result.Reason)
}

// The cross-jurisdiction check in validateCurrencyMatch/findProviderCurrency
// is currently unreachable: findProviderCurrency compares currency.String() == assetCode
// and returns the currency string, so both sender and receiver values always equal
// the assetCode when non-empty, making senderCurrency != receiverCurrency impossible.
// This test only verifies the current behavior and should be updated once findProviderCurrency is refined
func TestValidateOutgoingPayment_DifferentWallets_Valid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	senderWalletID := uuid.NewString()
	receiverWalletID := uuid.NewString()
	senderPP := "wa_sender_" + uuid.NewString()[:8]
	receiverPP := "wa_receiver_" + uuid.NewString()[:8]
	ipID := "ip_" + uuid.NewString()[:8]

	op := outgoingPaymentData{
		ID:              uuid.NewString(),
		WalletAddressID: senderPP,
		DebitAmount:     amount{Value: "100", AssetCode: "EUR", AssetScale: 2},
		Receiver:        "https://example.com/incoming-payments/" + ipID,
	}

	ab := setupActivityBackendsWithReceiver(t, ctrl, senderWalletID, senderPP, receiverWalletID, receiverPP)
	kycMock := kyc_mock.NewMockClient(ctrl)
	kycMock.EXPECT().IsKYCApproved(gomock.Any(), senderWalletID).Return(true, nil)
	ab.SetKYC(kycMock)

	rafikiMock := rafiki_mock.NewMockClient(ctrl)
	rafikiMock.EXPECT().GetIncomingPayment(gomock.Any(), ipID).
		Return(&rafiki.IncomingPayment{WalletAddressID: receiverPP}, nil)
	ab.SetRafiki(rafikiMock)

	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), senderWalletID).Return([]linkedaccounts.LinkedAccount{
		{ID: "la_s", Provider: gatehub.ProviderName, SendCurrency: currency.EUR},
	}, nil)
	linkedMock.EXPECT().ListBalances(gomock.Any(), receiverWalletID).Return([]linkedaccounts.LinkedAccount{
		{ID: "la_r", Provider: gatehub.ProviderName, ReceiveCurrency: currency.EUR},
	}, nil)
	ab.SetLinkedAccounts(linkedMock)

	act := NewActivity(ab, gatehub.Config{})

	result, err := act.ValidateOutgoingPayment(ctx, op)
	require.NoError(t, err)
	assert.True(t, result.Valid, "same currency local receiver should be valid")
}

// --- Pacioli ledger activity tests -----------------------------------------

func TestCreateAndPostLedgerTransferForIncoming_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	walletID := uuid.NewString()
	ppID := "wa_recv_" + uuid.NewString()[:8]
	laID := "la_" + uuid.NewString()[:8]

	ip := incomingPaymentData{
		ID:              uuid.NewString(),
		WalletAddressID: ppID,
		ReceivedAmount:  amount{Value: "5000", AssetCode: "EUR", AssetScale: 2},
	}

	ab := setupActivityBackends(t, ctrl, walletID, ppID)
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{
		{ID: laID, Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance},
	}, nil)
	ab.SetLinkedAccounts(linkedMock)

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().CreateTransfers(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, args []pacioli.CreateTransferArgs) ([]pacioli.TransferResult, error) {
			require.Len(t, args, 1)
			assert.Equal(t, ip.ID, args[0].ID)
			assert.Equal(t, int64(5000), args[0].Amount)
			assert.Equal(t, laID, args[0].CreditAccountID)
			assert.Equal(t, gatehub.EUROpsAccount, args[0].DebitAccountID)
			assert.False(t, args[0].Pending)
			return nil, nil
		})
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.CreateAndPostLedgerTransferForIncoming(ctx, ip)
	assert.NoError(t, err)
}

func TestCreateAndPostLedgerTransferForIncoming_ExceedsCredits_Fails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	walletID := uuid.NewString()
	ppID := "wa_recv_" + uuid.NewString()[:8]

	ip := incomingPaymentData{
		ID:              uuid.NewString(),
		WalletAddressID: ppID,
		ReceivedAmount:  amount{Value: "5000", AssetCode: "EUR", AssetScale: 2},
	}

	ab := setupActivityBackends(t, ctrl, walletID, ppID)
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{
		{ID: "la", Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance},
	}, nil)
	ab.SetLinkedAccounts(linkedMock)

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().CreateTransfers(gomock.Any(), gomock.Any()).
		Return([]pacioli.TransferResult{{Index: 0, Code: pacioli.TransferExceedsCredits}}, nil)
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.CreateAndPostLedgerTransferForIncoming(ctx, ip)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ExceedsCredits")
}

func TestReserveBalanceForOutgoing_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	walletID := uuid.NewString()
	ppID := "wa_sender_" + uuid.NewString()[:8]
	laID := "la_" + uuid.NewString()[:8]

	op := outgoingPaymentData{
		ID:              uuid.NewString(),
		WalletAddressID: ppID,
		DebitAmount:     amount{Value: "1000", AssetCode: "EUR", AssetScale: 2},
	}

	ab := setupActivityBackends(t, ctrl, walletID, ppID)
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{
		{ID: laID, Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance},
	}, nil)
	ab.SetLinkedAccounts(linkedMock)

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().CreateTransfers(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, args []pacioli.CreateTransferArgs) ([]pacioli.TransferResult, error) {
			require.Len(t, args, 1)
			assert.Equal(t, op.ID, args[0].ID)
			assert.True(t, args[0].Pending, "reserve transfer must be pending")
			assert.Equal(t, laID, args[0].DebitAccountID)
			assert.Equal(t, gatehub.EUROpsAccount, args[0].CreditAccountID)
			return nil, nil
		})
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.ReserveBalanceForOutgoing(ctx, op)
	assert.NoError(t, err)
}

// --- PostTransfers activity tests ------------------------------------------

func TestPostLedgerTransferForOutgoing_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ab := NewTestActivityBackends()
	op := outgoingPaymentData{ID: uuid.NewString()}

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().PostTransfers(gomock.Any(), []string{op.ID}).Return(nil, nil)
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.PostLedgerTransferForOutgoing(context.Background(), op)
	assert.NoError(t, err)
}

func TestPostLedgerTransferForOutgoing_AlreadyPosted_Tolerated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ab := NewTestActivityBackends()
	op := outgoingPaymentData{ID: uuid.NewString()}

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().PostTransfers(gomock.Any(), []string{op.ID}).
		Return([]pacioli.TransferResult{{Index: 0, Code: pacioli.TransferPendingTransferAlreadyPosted}}, nil)
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.PostLedgerTransferForOutgoing(context.Background(), op)
	assert.NoError(t, err)
}

func TestPostLedgerTransferForOutgoing_NotFound_Fails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ab := NewTestActivityBackends()
	op := outgoingPaymentData{ID: uuid.NewString()}

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().PostTransfers(gomock.Any(), []string{op.ID}).
		Return([]pacioli.TransferResult{{Index: 0, Code: pacioli.TransferPendingTransferNotFound}}, nil)
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.PostLedgerTransferForOutgoing(context.Background(), op)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PendingTransferNotFound")
}

func TestPostLedgerTransferForOutgoing_AlreadyVoided_Fails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ab := NewTestActivityBackends()
	op := outgoingPaymentData{ID: uuid.NewString()}

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().PostTransfers(gomock.Any(), []string{op.ID}).
		Return([]pacioli.TransferResult{{Index: 0, Code: pacioli.TransferPendingTransferAlreadyVoided}}, nil)
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.PostLedgerTransferForOutgoing(context.Background(), op)
	assert.Error(t, err)
}

// --- VoidTransfers activity tests ------------------------------------------

func TestVoidLedgerTransferForOutgoing_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ab := NewTestActivityBackends()
	op := outgoingPaymentData{ID: uuid.NewString()}

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().VoidTransfers(gomock.Any(), []string{op.ID}).Return(nil, nil)
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.VoidLedgerTransferForOutgoing(context.Background(), op)
	assert.NoError(t, err)
}

func TestVoidLedgerTransferForOutgoing_AlreadyVoided_Tolerated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ab := NewTestActivityBackends()
	op := outgoingPaymentData{ID: uuid.NewString()}

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().VoidTransfers(gomock.Any(), []string{op.ID}).
		Return([]pacioli.TransferResult{{Index: 0, Code: pacioli.TransferPendingTransferAlreadyVoided}}, nil)
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.VoidLedgerTransferForOutgoing(context.Background(), op)
	assert.NoError(t, err)
}

func TestVoidLedgerTransferForOutgoing_NotFound_Tolerated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ab := NewTestActivityBackends()
	op := outgoingPaymentData{ID: uuid.NewString()}

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().VoidTransfers(gomock.Any(), []string{op.ID}).
		Return([]pacioli.TransferResult{{Index: 0, Code: pacioli.TransferPendingTransferNotFound}}, nil)
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.VoidLedgerTransferForOutgoing(context.Background(), op)
	assert.NoError(t, err)
}

func TestVoidLedgerTransferForOutgoing_Expired_Tolerated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ab := NewTestActivityBackends()
	op := outgoingPaymentData{ID: uuid.NewString()}

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().VoidTransfers(gomock.Any(), []string{op.ID}).
		Return([]pacioli.TransferResult{{Index: 0, Code: pacioli.TransferPendingTransferExpired}}, nil)
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.VoidLedgerTransferForOutgoing(context.Background(), op)
	assert.NoError(t, err)
}

func TestVoidLedgerTransferForOutgoing_AlreadyPosted_Fails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ab := NewTestActivityBackends()
	op := outgoingPaymentData{ID: uuid.NewString()}

	pacioliMock := pacioli_mock.NewMockClient(ctrl)
	pacioliMock.EXPECT().VoidTransfers(gomock.Any(), []string{op.ID}).
		Return([]pacioli.TransferResult{{Index: 0, Code: pacioli.TransferPendingTransferAlreadyPosted}}, nil)
	ab.SetPacioli(pacioliMock)

	act := &Activity{b: ab}
	err := act.VoidLedgerTransferForOutgoing(context.Background(), op)
	assert.Error(t, err, "AlreadyPosted must NOT be tolerated in the void path")
	assert.Contains(t, err.Error(), "PendingTransferAlreadyPosted")
}

// --- Transaction state activity tests --------------------------------------

func TestUpdateOutgoingPaymentTransactionState_Completed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	walletID := uuid.NewString()
	ppID := "wa_sender_" + uuid.NewString()[:8]
	trxID := "trx_" + uuid.NewString()[:8]
	transferID := "txf_" + uuid.NewString()[:8]

	op := outgoingPaymentData{ID: uuid.NewString(), WalletAddressID: ppID}

	ab := setupActivityBackends(t, ctrl, walletID, ppID)
	txMock := transactions_mock.NewMockClient(ctrl)
	txMock.EXPECT().GetTransactionByForeignID(gomock.Any(), walletID, op.ID).
		Return(&transactions.Transaction{ID: trxID}, nil)
	txMock.EXPECT().ListTransfers(gomock.Any(), trxID).
		Return([]transactions.Transfer{{ID: transferID}}, nil)
	txMock.EXPECT().SetTransferState(gomock.Any(), transferID, transactions.State("Completed")).Return(nil)
	txMock.EXPECT().SetTransactionState(gomock.Any(), trxID, transactions.State("Completed")).Return(nil)
	ab.SetTransactions(txMock)

	act := &Activity{b: ab}
	err := act.UpdateOutgoingPaymentTransactionState(ctx, op, "Completed")
	assert.NoError(t, err)
}

func TestUpdateOutgoingPaymentTransactionState_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	walletID := uuid.NewString()
	ppID := "wa_sender_" + uuid.NewString()[:8]
	trxID := "trx_" + uuid.NewString()[:8]

	op := outgoingPaymentData{ID: uuid.NewString(), WalletAddressID: ppID}

	ab := setupActivityBackends(t, ctrl, walletID, ppID)
	txMock := transactions_mock.NewMockClient(ctrl)
	txMock.EXPECT().GetTransactionByForeignID(gomock.Any(), walletID, op.ID).
		Return(&transactions.Transaction{ID: trxID}, nil)
	txMock.EXPECT().ListTransfers(gomock.Any(), trxID).
		Return([]transactions.Transfer{}, nil)
	txMock.EXPECT().SetTransactionState(gomock.Any(), trxID, transactions.State("Failed")).Return(nil)
	ab.SetTransactions(txMock)

	act := &Activity{b: ab}
	err := act.UpdateOutgoingPaymentTransactionState(ctx, op, "Failed")
	assert.NoError(t, err)
}

// --- Transaction creation activity tests -----------------------------------

func TestCreateIncomingPaymentTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	walletID := uuid.NewString()
	ppID := "wa_recv_" + uuid.NewString()[:8]
	laID := "la_" + uuid.NewString()[:8]

	ip := incomingPaymentData{
		ID:              uuid.NewString(),
		WalletAddressID: ppID,
		ReceivedAmount:  amount{Value: "1000", AssetCode: "EUR", AssetScale: 2},
	}

	ab := setupActivityBackends(t, ctrl, walletID, ppID)
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{
		{ID: laID, Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance},
	}, nil)
	ab.SetLinkedAccounts(linkedMock)

	walletsMock := wallets_mock.NewMockClient(ctrl)
	walletsMock.EXPECT().Get(gomock.Any(), walletID).Return(&wallets.Wallet{ID: walletID}, nil)
	ab.SetWallets(walletsMock)

	txMock := transactions_mock.NewMockClient(ctrl)
	txMock.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, args transactions.CreateTransactionArgs) (string, error) {
			assert.Equal(t, walletID, args.WalletID)
			assert.Equal(t, ip.ID, args.ForeignID)
			assert.Equal(t, transactions.StateCompleted, args.State)
			assert.Equal(t, transactions.TransactionTypeOpenPaymentIncoming, args.ForeignType)
			return "trx_new", nil
		})
	ab.SetTransactions(txMock)

	act := &Activity{b: ab}
	err := act.CreateIncomingPaymentTransaction(ctx, ip)
	assert.NoError(t, err)
}

func TestCreateOutgoingPaymentTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	walletID := uuid.NewString()
	ppID := "wa_sender_" + uuid.NewString()[:8]
	laID := "la_" + uuid.NewString()[:8]

	op := outgoingPaymentData{
		ID:              uuid.NewString(),
		WalletAddressID: ppID,
		DebitAmount:     amount{Value: "500", AssetCode: "EUR", AssetScale: 2},
		Receiver:        "https://example.com/incoming-payments/ip_1",
	}

	ab := setupActivityBackends(t, ctrl, walletID, ppID)
	linkedMock := linkedaccounts_mock.NewMockClient(ctrl)
	linkedMock.EXPECT().ListBalances(gomock.Any(), walletID).Return([]linkedaccounts.LinkedAccount{
		{ID: laID, Provider: gatehub.ProviderName, Type: gatehub.AccTypeBalance},
	}, nil)
	ab.SetLinkedAccounts(linkedMock)

	walletsMock := wallets_mock.NewMockClient(ctrl)
	walletsMock.EXPECT().Get(gomock.Any(), walletID).Return(&wallets.Wallet{ID: walletID}, nil)
	ab.SetWallets(walletsMock)

	txMock := transactions_mock.NewMockClient(ctrl)
	txMock.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, args transactions.CreateTransactionArgs) (string, error) {
			assert.Equal(t, walletID, args.WalletID)
			assert.Equal(t, op.ID, args.ForeignID)
			assert.Equal(t, transactions.StatePending, args.State)
			assert.Equal(t, transactions.TransactionTypeOpenOutgoingPayment, args.ForeignType)
			assert.Equal(t, payments.IdentityTypeExternalWalletURL.String(), args.DestinationIdentityType)
			return "trx_out", nil
		})
	ab.SetTransactions(txMock)

	act := &Activity{b: ab}
	err := act.CreateOutgoingPaymentTransaction(ctx, op)
	assert.NoError(t, err)
}

// --- Rafiki liquidity activity tests ---------------------------------------

func TestWithdrawIncomingPaymentLiquidity_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ipID := "ip_" + uuid.NewString()[:8]
	rafikiMock := rafiki_mock.NewMockClient(ctrl)
	rafikiMock.EXPECT().WithdrawIncomingPaymentLiquidity(gomock.Any(), ipID).Return(nil)

	ab := NewTestActivityBackends()
	ab.SetRafiki(rafikiMock)
	act := NewActivity(ab, gatehub.Config{})
	err := act.WithdrawIncomingPaymentLiquidity(context.Background(), ipID)
	assert.NoError(t, err)
}

func TestWithdrawIncomingPaymentLiquidity_Error_StillReturnsNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ipID := "ip_" + uuid.NewString()[:8]
	rafikiMock := rafiki_mock.NewMockClient(ctrl)
	rafikiMock.EXPECT().WithdrawIncomingPaymentLiquidity(gomock.Any(), ipID).
		Return(fmt.Errorf("rafiki error"))

	ab := NewTestActivityBackends()
	ab.SetRafiki(rafikiMock)
	act := NewActivity(ab, gatehub.Config{})
	err := act.WithdrawIncomingPaymentLiquidity(context.Background(), ipID)
	assert.NoError(t, err, "withdraw errors are logged but not returned")
}

func TestWithdrawOutgoingPaymentLiquidity_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	opID := uuid.NewString()
	op := outgoingPaymentData{
		ID:      opID,
		Balance: "100",
	}
	rafikiMock := rafiki_mock.NewMockClient(ctrl)
	rafikiMock.EXPECT().WithdrawOutgoingPaymentLiquidity(gomock.Any(), opID).Return(nil)

	ab := NewTestActivityBackends()
	ab.SetRafiki(rafikiMock)
	act := NewActivity(ab, gatehub.Config{})
	err := act.WithdrawOutgoingPaymentLiquidity(context.Background(), op)
	assert.NoError(t, err)
}

func TestWithdrawOutgoingPaymentLiquidity_InsufficientBalance_NoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	opID := uuid.NewString()
	op := outgoingPaymentData{
		ID:      opID,
		Balance: "0",
	}

	ab := NewTestActivityBackends()
	act := NewActivity(ab, gatehub.Config{})
	err := act.WithdrawOutgoingPaymentLiquidity(context.Background(), op)
	assert.NoError(t, err, "zero liquidity should be treated as no-op")
}

func TestDepositOutgoingPaymentLiquidity_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	opID := uuid.NewString()
	rafikiMock := rafiki_mock.NewMockClient(ctrl)
	rafikiMock.EXPECT().FundOutgoingPayment(gomock.Any(), opID).Return(nil)

	ab := NewTestActivityBackends()
	ab.SetRafiki(rafikiMock)
	act := NewActivity(ab, gatehub.Config{})
	err := act.DepositOutgoingPaymentLiquidity(context.Background(), opID)
	assert.NoError(t, err)
}

func TestDepositOutgoingPaymentLiquidity_WrongState_Ignored(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	opID := uuid.NewString()
	rafikiMock := rafiki_mock.NewMockClient(ctrl)
	rafikiMock.EXPECT().FundOutgoingPayment(gomock.Any(), opID).
		Return(fmt.Errorf("wrong state for operation"))

	ab := NewTestActivityBackends()
	ab.SetRafiki(rafikiMock)
	act := NewActivity(ab, gatehub.Config{})
	err := act.DepositOutgoingPaymentLiquidity(context.Background(), opID)
	assert.NoError(t, err, "wrong state should be silently ignored")
}

func TestCancelOutgoingPayment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	opID := uuid.NewString()
	rafikiMock := rafiki_mock.NewMockClient(ctrl)
	rafikiMock.EXPECT().CancelOutgoingPayment(gomock.Any(), opID, "test reason").Return(nil)

	ab := NewTestActivityBackends()
	ab.SetRafiki(rafikiMock)
	act := NewActivity(ab, gatehub.Config{})
	err := act.CancelOutgoingPayment(context.Background(), opID, "test reason")
	assert.NoError(t, err)
}

// --- incoming_payment.created handler test ----------

func TestIncomingPaymentCreated_InsertsRecord(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	b := NewTestBackends(func(tb *TestBackends) { tb.SetDB(conn) })

	ipID := "ip_" + uuid.NewString()[:8]
	ppID := "wa_" + uuid.NewString()[:8]
	hook := webhook{
		ID:   uuid.NewString(),
		Type: "incoming_payment.created",
		Data: mustMarshalIncomingPaymentData(ipID, ppID, "0", false),
	}
	err := incomingPaymentCreated(ctx, b, hook)
	require.NoError(t, err)

	var count int
	err = conn.GetContext(ctx, &count, "SELECT COUNT(*) FROM rafiki_incoming_payments WHERE payment_id = $1", ipID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestIncomingPaymentCreated_DuplicateIgnored(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	b := NewTestBackends(func(tb *TestBackends) { tb.SetDB(conn) })

	ipID := "ip_" + uuid.NewString()[:8]
	ppID := "wa_" + uuid.NewString()[:8]
	hook := webhook{
		ID:   uuid.NewString(),
		Type: "incoming_payment.created",
		Data: mustMarshalIncomingPaymentData(ipID, ppID, "0", false),
	}

	require.NoError(t, incomingPaymentCreated(ctx, b, hook))
	require.NoError(t, incomingPaymentCreated(ctx, b, hook), "duplicate insert should not fail")
}

func TestIncomingPaymentCreated_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	conn := requireTestDB(t)
	b := NewTestBackends(func(tb *TestBackends) { tb.SetDB(conn) })

	hook := webhook{ID: "evt_1", Type: "incoming_payment.created", Data: []byte(`not json`)}
	err := incomingPaymentCreated(ctx, b, hook)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func setupActivityBackends(t *testing.T, ctrl *gomock.Controller, walletID, ppID string) *TestActivityBackends {
	t.Helper()
	conn := requireTestDB(t)

	_, err := conn.ExecContext(context.Background(),
		`INSERT INTO rafiki_payment_pointers (wallet_id, payment_pointer_id) VALUES ($1, $2) ON CONFLICT (payment_pointer_id) DO NOTHING`,
		walletID, ppID)
	require.NoError(t, err)

	ab := NewTestActivityBackends()
	ab.SetDB(conn)
	return ab
}

func setupActivityBackendsWithReceiver(t *testing.T, ctrl *gomock.Controller, senderWalletID, senderPP, receiverWalletID, receiverPP string) *TestActivityBackends {
	t.Helper()
	conn := requireTestDB(t)

	_, err := conn.ExecContext(context.Background(),
		`INSERT INTO rafiki_payment_pointers (wallet_id, payment_pointer_id) VALUES ($1, $2), ($3, $4) ON CONFLICT (payment_pointer_id) DO NOTHING`,
		senderWalletID, senderPP, receiverWalletID, receiverPP)
	require.NoError(t, err)

	ab := NewTestActivityBackends()
	ab.SetDB(conn)
	return ab
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
