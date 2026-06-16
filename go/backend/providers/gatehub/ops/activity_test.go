package ops_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/email"
	email_mock "gitlab.com/fynbos/backend/email/client/mock"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	la_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/ops"
	"gitlab.com/fynbos/backend/transactions"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	wallet_mock "gitlab.com/fynbos/backend/wallets/client/mock"
	"gitlab.com/fynbos/pacioli"
	pacioli_mock "gitlab.com/fynbos/pacioli/client/mock"
	temporal "go.temporal.io/sdk/client"
)

func TestSaveUser(t *testing.T) {
	b := Backends{
		db:    db.MigrateTestDB(t, context.Background()),
		users: user_mock.NewMock(),
	}
	cfg := gatehub.Config{
		AppID:                           gatehub.TestAppID,
		Secret:                          gatehub.TestSecret,
		CardAppID:                       gatehub.TestCardAppID,
		GatewayID:                       gatehub.TestGatewayID,
		CardAccountProductCode:          gatehub.TestCardAccountProductCode,
		PaywiserEuroVaultID:             gatehub.TestPaywiserEuroVaultID,
		SendingUserID:                   gatehub.TestSendingUserID,
		SendingUserAddress:              gatehub.TestSendingUserAddress,
		WebhookSecret:                   gatehub.TestWebhookSecret,
		FallbackWebhookURL:              gatehub.TestFallbackWebhookURL,
		OnOffRampClientID:               gatehub.TestOnOffRampClientID,
		OnboardingClientID:              gatehub.TestOnboardingClientID,
		ExchangeClientID:                gatehub.TestExchangeClientID,
		APIBaseURL:                      gatehub.TestAPIBaseURL,
		OnboardingBaseURL:               gatehub.TestOnboardingBaseURL,
		OnOffRampBaseURL:                gatehub.TestOnOffRampBaseURL,
		EUROpsAccount:                   gatehub.TestEUROpsAccount,
		EUROpsLedgerID:                  gatehub.TestEUROpsLedgerID,
		XagoGatehubGhOmnibusUserID:      gatehub.TestXagoGatehubGhOmnibusUserID,
		XagoGatehubGhOmnibusUserAddress: gatehub.TestXagoGatehubGhOmnibusUserAddress,
	}
	a := ops.NewActivity(b, cfg)

	walletID := uuid.NewString()
	externalID := uuid.NewString()
	err := a.SaveGatehubUser(context.Background(), walletID, externalID)
	require.NoError(t, err)

	var result string
	err = b.db.GetContext(context.Background(), &result, "SELECT external_id FROM gatehub_users WHERE wallet_id=$1;", walletID)
	require.NoError(t, err)

	err = a.SaveGatehubUser(context.Background(), walletID, externalID)
	require.NoError(t, err)

	assert.Equal(t, externalID, result)
}

func TestLinkGatehubUserToGateway(t *testing.T) {
	b := Backends{
		db:    db.MigrateTestDB(t, context.Background()),
		users: user_mock.NewMock(),
	}
	cfg := gatehub.Config{
		AppID:                           gatehub.TestAppID,
		Secret:                          gatehub.TestSecret,
		CardAppID:                       gatehub.TestCardAppID,
		GatewayID:                       gatehub.TestGatewayID,
		CardAccountProductCode:          gatehub.TestCardAccountProductCode,
		PaywiserEuroVaultID:             gatehub.TestPaywiserEuroVaultID,
		SendingUserID:                   gatehub.TestSendingUserID,
		SendingUserAddress:              gatehub.TestSendingUserAddress,
		WebhookSecret:                   gatehub.TestWebhookSecret,
		FallbackWebhookURL:              gatehub.TestFallbackWebhookURL,
		OnOffRampClientID:               gatehub.TestOnOffRampClientID,
		OnboardingClientID:              gatehub.TestOnboardingClientID,
		ExchangeClientID:                gatehub.TestExchangeClientID,
		APIBaseURL:                      gatehub.TestAPIBaseURL,
		OnboardingBaseURL:               gatehub.TestOnboardingBaseURL,
		OnOffRampBaseURL:                gatehub.TestOnOffRampBaseURL,
		EUROpsAccount:                   gatehub.TestEUROpsAccount,
		EUROpsLedgerID:                  gatehub.TestEUROpsLedgerID,
		XagoGatehubGhOmnibusUserID:      gatehub.TestXagoGatehubGhOmnibusUserID,
		XagoGatehubGhOmnibusUserAddress: gatehub.TestXagoGatehubGhOmnibusUserAddress,
	}
	a := ops.NewActivity(b, cfg)

	walletID := uuid.NewString()
	externalID := uuid.NewString()
	err := a.SaveGatehubUser(context.Background(), walletID, externalID)
	require.NoError(t, err)

	return
}

func TestGetGateHubTransactionByForeignID(t *testing.T) {
	const walletID = "wallet-123"
	const foreignID = "gh-tx-456"
	cfg := gatehub.Config{
		AppID:                  gatehub.TestAppID,
		Secret:                 gatehub.TestSecret,
		CardAppID:              gatehub.TestCardAppID,
		GatewayID:              gatehub.TestGatewayID,
		CardAccountProductCode: gatehub.TestCardAccountProductCode,
		PaywiserEuroVaultID:    gatehub.TestPaywiserEuroVaultID,
		SendingUserID:          gatehub.TestSendingUserID,
		SendingUserAddress:     gatehub.TestSendingUserAddress,
		WebhookSecret:          gatehub.TestWebhookSecret,
		FallbackWebhookURL:     gatehub.TestFallbackWebhookURL,
		OnOffRampClientID:      gatehub.TestOnOffRampClientID,
		OnboardingClientID:     gatehub.TestOnboardingClientID,
		ExchangeClientID:       gatehub.TestExchangeClientID,
		APIBaseURL:             gatehub.TestAPIBaseURL,
		OnboardingBaseURL:      gatehub.TestOnboardingBaseURL,
		OnOffRampBaseURL:       gatehub.TestOnOffRampBaseURL,
		EUROpsAccount:          gatehub.TestEUROpsAccount,
		EUROpsLedgerID:         gatehub.TestEUROpsLedgerID,
	}

	t.Run("DB error wraps ErrInternal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		tc := transactions_mock.NewMockClient(ctrl)
		tc.EXPECT().GetTransactionByForeignID(gomock.Any(), walletID, foreignID).Return(nil, errors.New("db timeout"))

		a := ops.NewActivity(Backends{tc: tc}, cfg)
		_, err := a.GetGateHubTransactionByForeignID(context.Background(), walletID, foreignID)

		require.Error(t, err)
		assert.ErrorIs(t, err, gatehub.ErrInternal)
	})

	t.Run("success returns transaction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		tc := transactions_mock.NewMockClient(ctrl)
		txID := uuid.NewString()
		tc.EXPECT().GetTransactionByForeignID(gomock.Any(), walletID, foreignID).Return(&transactions.Transaction{ID: txID}, nil)

		a := ops.NewActivity(Backends{tc: tc}, cfg)
		got, err := a.GetGateHubTransactionByForeignID(context.Background(), walletID, foreignID)

		require.NoError(t, err)
		assert.Equal(t, txID, got.ID)
	})
}

func TestRollbackGatehubWithdrawal_PacioliError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const txID = "tx-abc"
	pc := pacioli_mock.NewMockClient(ctrl)
	pc.EXPECT().VoidTransfers(gomock.Any(), []string{txID}).Return(nil, errors.New("ledger down"))

	cfg := gatehub.Config{
		AppID:                  gatehub.TestAppID,
		Secret:                 gatehub.TestSecret,
		CardAppID:              gatehub.TestCardAppID,
		GatewayID:              gatehub.TestGatewayID,
		CardAccountProductCode: gatehub.TestCardAccountProductCode,
		PaywiserEuroVaultID:    gatehub.TestPaywiserEuroVaultID,
		SendingUserID:          gatehub.TestSendingUserID,
		SendingUserAddress:     gatehub.TestSendingUserAddress,
		WebhookSecret:          gatehub.TestWebhookSecret,
		FallbackWebhookURL:     gatehub.TestFallbackWebhookURL,
		OnOffRampClientID:      gatehub.TestOnOffRampClientID,
		OnboardingClientID:     gatehub.TestOnboardingClientID,
		ExchangeClientID:       gatehub.TestExchangeClientID,
		APIBaseURL:             gatehub.TestAPIBaseURL,
		OnboardingBaseURL:      gatehub.TestOnboardingBaseURL,
		OnOffRampBaseURL:       gatehub.TestOnOffRampBaseURL,
		EUROpsAccount:          gatehub.TestEUROpsAccount,
		EUROpsLedgerID:         gatehub.TestEUROpsLedgerID,
	}
	a := ops.NewActivity(Backends{pc: pc}, cfg)
	err := a.RollbackGatehubWithdrawal(context.Background(), txID)

	require.Error(t, err)
	assert.ErrorIs(t, err, gatehub.ErrInternal)
}

func TestRollbackGatehubWithdrawal_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const txID = "tx-abc"
	pc := pacioli_mock.NewMockClient(ctrl)
	pc.EXPECT().VoidTransfers(gomock.Any(), []string{txID}).Return([]pacioli.TransferResult{}, nil)

	tc := transactions_mock.NewMockClient(ctrl)
	tc.EXPECT().SetTransactionState(gomock.Any(), txID, transactions.StateFailed).Return(nil)

	cfg := gatehub.Config{
		AppID:                  gatehub.TestAppID,
		Secret:                 gatehub.TestSecret,
		CardAppID:              gatehub.TestCardAppID,
		GatewayID:              gatehub.TestGatewayID,
		CardAccountProductCode: gatehub.TestCardAccountProductCode,
		PaywiserEuroVaultID:    gatehub.TestPaywiserEuroVaultID,
		SendingUserID:          gatehub.TestSendingUserID,
		SendingUserAddress:     gatehub.TestSendingUserAddress,
		WebhookSecret:          gatehub.TestWebhookSecret,
		FallbackWebhookURL:     gatehub.TestFallbackWebhookURL,
		OnOffRampClientID:      gatehub.TestOnOffRampClientID,
		OnboardingClientID:     gatehub.TestOnboardingClientID,
		ExchangeClientID:       gatehub.TestExchangeClientID,
		APIBaseURL:             gatehub.TestAPIBaseURL,
		OnboardingBaseURL:      gatehub.TestOnboardingBaseURL,
		OnOffRampBaseURL:       gatehub.TestOnOffRampBaseURL,
		EUROpsAccount:          gatehub.TestEUROpsAccount,
		EUROpsLedgerID:         gatehub.TestEUROpsLedgerID,
	}
	a := ops.NewActivity(Backends{pc: pc, tc: tc}, cfg)
	err := a.RollbackGatehubWithdrawal(context.Background(), txID)

	require.NoError(t, err)
}

func TestSendWithdrawalSCTITimeoutEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ec := email_mock.NewMockClient(ctrl)
	ec.EXPECT().SendSCTITimeoutEmail(gomock.Any(), "tx-1", "wallet-1", "100", "John Doe", "GB29NWBK60161331926819", "2024-01-01T10:00:00Z")

	cfg := gatehub.Config{
		AppID:                  gatehub.TestAppID,
		Secret:                 gatehub.TestSecret,
		CardAppID:              gatehub.TestCardAppID,
		GatewayID:              gatehub.TestGatewayID,
		CardAccountProductCode: gatehub.TestCardAccountProductCode,
		PaywiserEuroVaultID:    gatehub.TestPaywiserEuroVaultID,
		SendingUserID:          gatehub.TestSendingUserID,
		SendingUserAddress:     gatehub.TestSendingUserAddress,
		WebhookSecret:          gatehub.TestWebhookSecret,
		FallbackWebhookURL:     gatehub.TestFallbackWebhookURL,
		OnOffRampClientID:      gatehub.TestOnOffRampClientID,
		OnboardingClientID:     gatehub.TestOnboardingClientID,
		ExchangeClientID:       gatehub.TestExchangeClientID,
		APIBaseURL:             gatehub.TestAPIBaseURL,
		OnboardingBaseURL:      gatehub.TestOnboardingBaseURL,
		OnOffRampBaseURL:       gatehub.TestOnOffRampBaseURL,
		EUROpsAccount:          gatehub.TestEUROpsAccount,
		EUROpsLedgerID:         gatehub.TestEUROpsLedgerID,
	}
	a := ops.NewActivity(Backends{ec: ec}, cfg)
	err := a.SendWithdrawalSCTITimeoutEmail(context.Background(), "tx-1", "wallet-1", "100", "John Doe", "GB29NWBK60161331926819", "2024-01-01T10:00:00Z")

	require.NoError(t, err)
}

func TestSendWithdrawalRejectedEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ec := email_mock.NewMockClient(ctrl)
	ec.EXPECT().SendGatehubWithdrawalRejectedEmail(gomock.Any(), "tx-1", "wallet-1", "100", "EUR", "GB29NWBK60161331926819", "John Doe")

	cfg := gatehub.Config{
		AppID:                  gatehub.TestAppID,
		Secret:                 gatehub.TestSecret,
		CardAppID:              gatehub.TestCardAppID,
		GatewayID:              gatehub.TestGatewayID,
		CardAccountProductCode: gatehub.TestCardAccountProductCode,
		PaywiserEuroVaultID:    gatehub.TestPaywiserEuroVaultID,
		SendingUserID:          gatehub.TestSendingUserID,
		SendingUserAddress:     gatehub.TestSendingUserAddress,
		WebhookSecret:          gatehub.TestWebhookSecret,
		FallbackWebhookURL:     gatehub.TestFallbackWebhookURL,
		OnOffRampClientID:      gatehub.TestOnOffRampClientID,
		OnboardingClientID:     gatehub.TestOnboardingClientID,
		ExchangeClientID:       gatehub.TestExchangeClientID,
		APIBaseURL:             gatehub.TestAPIBaseURL,
		OnboardingBaseURL:      gatehub.TestOnboardingBaseURL,
		OnOffRampBaseURL:       gatehub.TestOnOffRampBaseURL,
		EUROpsAccount:          gatehub.TestEUROpsAccount,
		EUROpsLedgerID:         gatehub.TestEUROpsLedgerID,
	}
	a := ops.NewActivity(Backends{ec: ec}, cfg)
	err := a.SendWithdrawalRejectedEmail(context.Background(), "tx-1", "wallet-1", "100", "EUR", "GB29NWBK60161331926819", "John Doe")

	require.NoError(t, err)
}

type Backends struct {
	db    *sqlx.DB
	users *user_mock.MockClient
	la    *la_mock.MockClient
	wc    *wallet_mock.MockClient
	tc    *transactions_mock.MockClient
	pc    pacioli.Client
	ec    email.Client
}

func (b Backends) Payments() payments.Client {
	return nil
}

func (b Backends) Users() user.Client {
	return b.users
}

func (b Backends) DB() *sqlx.DB {
	return b.db
}

func (b Backends) Temporal() temporal.Client {
	return nil
}

func (b Backends) LinkedAccounts() linkedaccounts.Client {
	return b.la
}

func (b Backends) Wallets() wallets.Client {
	return b.wc
}

func (b Backends) Pacioli() pacioli.Client {
	return b.pc
}

func (b Backends) KYC() kyc.Client {
	return nil
}

func (b Backends) Transactions() transactions.Client {
	return b.tc
}

func (b Backends) Email() email.Client {
	return b.ec
}

func (b Backends) LinkGatehubUserToGateway() transactions.Client {
	return nil
}

func (b Backends) Notify() notify.Client {
	return nil
}
