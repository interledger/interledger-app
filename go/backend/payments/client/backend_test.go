package client_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/providers/chimoney"

	rafiki_mock "gitlab.com/fynbos/backend/rafiki/client/mock"

	"gitlab.com/fynbos/backend/providers/gatehub"

	"gitlab.com/fynbos/backend/providers/pti"

	"gitlab.com/fynbos/backend/currency"

	pacioli_db "gitlab.com/fynbos/pacioli/db"

	xago_client "gitlab.com/fynbos/backend/providers/xago/client"
	"gitlab.com/fynbos/pacioli"
	pacioli_client "gitlab.com/fynbos/pacioli/client"

	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/rafiki"

	images_client "gitlab.com/fynbos/backend/images/client"

	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"

	limits_client "gitlab.com/fynbos/backend/limits/client"

	payments_client "gitlab.com/fynbos/backend/payments/client"

	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/payments"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/email"
	email_mock "gitlab.com/fynbos/backend/email/client/mock"
	"gitlab.com/fynbos/backend/identities"
	id_client "gitlab.com/fynbos/backend/identities/client"
	"gitlab.com/fynbos/backend/images"
	"gitlab.com/fynbos/backend/keys"
	keys_client "gitlab.com/fynbos/backend/keys/client"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccount_client "gitlab.com/fynbos/backend/linkedaccounts/client"
	"gitlab.com/fynbos/backend/notify"
	notify_client "gitlab.com/fynbos/backend/notify/client"
	"gitlab.com/fynbos/backend/payments/ops"
	"gitlab.com/fynbos/backend/signup"
	temporal_mock "gitlab.com/fynbos/backend/temporal/mock"
	"gitlab.com/fynbos/backend/transactions"
	transaction_client "gitlab.com/fynbos/backend/transactions/client"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client/mock"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	wallet_client "gitlab.com/fynbos/backend/wallets/client"
	temporal_client "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"gopkg.in/stretchr/testify.v1/require"
)

type TestBackends struct {
	db       *sqlx.DB
	email    email.Client
	user     *user_client.MockClient
	temporal temporal_client.Client
	env      *testsuite.TestWorkflowEnvironment
	kyc      kyc.Client
	xgo      xago.Client
	pac      pacioli.Client
	raf      rafiki.Client
}

func NewTestBackends(t *testing.T) *TestBackends {
	ctrl := gomock.NewController(t)
	em := email_mock.NewMockClient(ctrl)
	em.EXPECT().SendPaymentReceivedEmailV2(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	em.EXPECT().SendPaymentSentEmailV2(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	em.EXPECT().SendPaymentFailedEmail(gomock.Any(), gomock.Any()).AnyTimes()
	em.EXPECT().SendConnectedAccountEmail(gomock.Any(), gomock.Any()).AnyTimes()
	em.EXPECT().SendDepositReceivedEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	em.EXPECT().SendWithdrawalEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	_, pacDB := pacioli_db.MigrateTestDB(t, context.Background())
	b := &TestBackends{
		db:    db.MigrateTestDB(t, context.Background()),
		user:  user_mock.NewMock(),
		email: em,
		pac:   pacioli_client.NewLocal(pacDB),
	}

	tp := temporal_mock.NewMockClient(ctrl)
	tp.EXPECT().ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(arg1 interface{}, arg2 interface{}, arg3 interface{}, arg4 ...interface{}) (*workflow.Execution, error) {
		require.Len(t, arg4, 1)
		b.env.ExecuteWorkflow(ops.PaymentWorkflow, arg4[0].(string))

		return nil, nil
	}).AnyTimes()
	tp.EXPECT().SignalWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(arg0, arg1, arg2, arg3, arg4 interface{}) error {
		ch, ok := arg3.(string)
		require.True(t, ok)
		b.env.SignalWorkflow(ch, arg4)
		return nil
	}).AnyTimes()
	b.temporal = tp

	b.xgo = xago_client.New(b)

	kc := kyc_mock.NewMockClient(ctrl)
	kc.EXPECT().GetIndividualDetails(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, walletID string) (*kyc.IndividualDetails, error) {
		return &kyc.IndividualDetails{
			WalletID:     walletID,
			FirstName:    "Test",
			LastName:     "McTestFace",
			CountryCode:  "US",
			PlaceOfBirth: "US",
			Nationality:  "US",
			Gender:       kyc.GenderMale,
			DateOfBirth:  time.Date(1990, time.April, 1, 0, 0, 0, 0, time.UTC),
			Address: &kyc.Address{
				Line1:       "Lincon",
				Line2:       "Nebraska",
				Building:    "",
				Apartment:   "",
				City:        "Tallens",
				State:       "US-MO",
				ZipCode:     "9010",
				CountryCode: "US",
				PlaceID:     "",
			},
			IPAddress: "198.0.0.1",
		}, nil
	}).AnyTimes()
	kc.EXPECT().GetKYCStatus(gomock.Any(), gomock.Any()).Return(kyc.StatusLevel1, nil).AnyTimes()

	b.kyc = kc

	raf := rafiki_mock.NewMockClient(ctrl)
	raf.EXPECT().FundOutgoingPayment(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	raf.EXPECT().FinalizeWebMonetization(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	b.raf = raf

	ledgers, err := b.pac.ConfigureLedgers(context.Background(), []pacioli.ConfigureLedgerArgs{
		{
			ID:    xago.LedgerIDZAR,
			Name:  "Xago ZAR Ledger",
			Asset: currency.ZAR.String(),
			Scale: uint8(currency.ZAR.Scale()),
		},
		{
			ID:    xago.LedgerIDUSD,
			Name:  "Xago USD Ledger",
			Asset: currency.USD.String(),
			Scale: uint8(currency.USD.Scale()),
		},
		{
			ID:    pti.LedgerIDUSD,
			Name:  "PTI USD Ledger",
			Asset: currency.USD.String(),
			Scale: uint8(currency.USD.Scale()),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range ledgers {
		if l.Code != pacioli.LedgerOK {
			t.Fatal(fmt.Errorf("failed to configure pacioli ledgers (%s)", l.Code.String()))
		}
	}

	accs, err := b.pac.ConfigureAccounts(context.Background(), []pacioli.ConfigureAccountArgs{
		{
			ID:                         xago.ZAROpsAccount,
			LedgerID:                   xago.LedgerIDZAR,
			Code:                       1,
			DebitsMustNotExceedCredits: false,
			CreditsMustNotExceedDebits: false,
		},
		{
			ID:                         xago.USDOpsAccount,
			LedgerID:                   xago.LedgerIDUSD,
			Code:                       1,
			DebitsMustNotExceedCredits: false,
			CreditsMustNotExceedDebits: false,
		},
		{
			ID:                         rafiki.ZARBalanceAccount,
			LedgerID:                   xago.LedgerIDZAR,
			Code:                       1,
			DebitsMustNotExceedCredits: false,
			CreditsMustNotExceedDebits: false,
		},
		{
			ID:                         pti.USDOpsAccount,
			LedgerID:                   pti.LedgerIDUSD,
			Code:                       1,
			DebitsMustNotExceedCredits: false,
			CreditsMustNotExceedDebits: false,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, acc := range accs {
		if acc.Code != pacioli.AccountExists && acc.Code != pacioli.AccountOK {
			t.Fatal(fmt.Errorf("failed to configure pacioli accounts (%s)", acc.Code.String()))
		}
	}

	return b
}

func (b *TestBackends) RestoreTemporalEnv() {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterActivity(ops.NewActivity(b))
	env.RegisterWorkflow(ops.PaymentWorkflow)
	env.RegisterWorkflow(ops.PayinWorkflow)
	env.RegisterWorkflow(ops.PayoutWorkflow)
	env.RegisterWorkflow(ops.RollbackPayInWorkflow)
	env.RegisterWorkflow(ops.AwaitReceiverWorkflow)
	// env.RegisterActivity(pti_ops.NewActivity(b, nil))

	b.env = env
}

func (b *TestBackends) Chimoney() chimoney.Client {
	return nil
}

func (b *TestBackends) Gatehub() gatehub.Client {
	return nil
}

func (b *TestBackends) Pacioli() pacioli.Client {
	return b.pac
}

func (b *TestBackends) Xago() xago.Client {
	return b.xgo
}

func (b *TestBackends) Twilio() twilio.Service {
	return nil
}

func (b *TestBackends) Rafiki() rafiki.Client {
	return b.raf
}

func (b *TestBackends) Signup() signup.Client {
	return nil
}

func (b *TestBackends) Analytics() analytics.Client {
	return analytics_client.New(b, "")
}

func (b *TestBackends) KYC() kyc.Client {
	return b.kyc
}

func (b *TestBackends) Transactions() transactions.Client {
	return transaction_client.New(b)
}

func (b *TestBackends) LinkedAccounts() linkedaccounts.Client {
	return linkedaccount_client.New(b)
}

func (b *TestBackends) Identities() identities.Client {
	return id_client.New(b)
}

func (b *TestBackends) Twitter() twitter.Client {
	return nil
}

func (b *TestBackends) Images() images.Client {
	return images_client.New(b)
}

func (b *TestBackends) Keys() keys.Client {
	return keys_client.New(b)
}

func (b *TestBackends) Users() user.Client {
	return b.user
}

func (b *TestBackends) Validator() *validator.Validate {
	return validator.New()
}

func (b *TestBackends) DB() *sqlx.DB {
	return b.db
}

func (b *TestBackends) Temporal() temporal_client.Client {
	return b.temporal
}

func (b *TestBackends) Notify() notify.Client {
	return notify_client.New(b, "")
}

func (b *TestBackends) Email() email.Client {
	return b.email
}

func (b *TestBackends) Wallets() wallets.Client {
	return wallet_client.New(b)
}

func (b *TestBackends) Limits() limits.Client {
	return limits_client.New(b)
}

func (b *TestBackends) Payments() payments.Client {
	return payments_client.New(b)
}

func (b *TestBackends) PTI() pti.Client {
	// return pti_client.New(b)
	return nil
}
