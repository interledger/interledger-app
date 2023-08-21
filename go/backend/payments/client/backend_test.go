package client_test

import (
	"context"
	"testing"

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
	kyc_client "gitlab.com/fynbos/backend/kyc/client"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccount_client "gitlab.com/fynbos/backend/linkedaccounts/client"
	"gitlab.com/fynbos/backend/notify"
	notify_client "gitlab.com/fynbos/backend/notify/client"
	"gitlab.com/fynbos/backend/payments/ops"
	gmt_ops "gitlab.com/fynbos/backend/providers/gmt/ops"
	"gitlab.com/fynbos/backend/providers/tabapay"
	tabapay_client "gitlab.com/fynbos/backend/providers/tabapay/client"
	"gitlab.com/fynbos/backend/signup"
	temporal_mock "gitlab.com/fynbos/backend/temporal/mock"
	"gitlab.com/fynbos/backend/transactions"
	transaction_client "gitlab.com/fynbos/backend/transactions/client"
	"gitlab.com/fynbos/backend/twilio"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client/mock"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/vault"
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
	tabapay  tabapay.Client
	user     *user_client.MockClient
	temporal temporal_client.Client
	env      *testsuite.TestWorkflowEnvironment
}

func NewTestBackends(t *testing.T) *TestBackends {
	ctrl := gomock.NewController(t)
	em := email_mock.NewMockClient(ctrl)
	em.EXPECT().SendPaymentReceivedEmailV2(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	em.EXPECT().SendPaymentSentEmailV2(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	em.EXPECT().SendPaymentFailedEmail(gomock.Any(), gomock.Any()).AnyTimes()
	em.EXPECT().SendConnectedAccountEmail(gomock.Any(), gomock.Any()).AnyTimes()

	b := &TestBackends{
		db:    db.MigrateTestDB(t, context.Background()),
		user:  user_mock.NewMock(),
		email: em,
	}

	tp := temporal_mock.NewMockClient(ctrl)
	tp.EXPECT().ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(arg1 interface{}, arg2 interface{}, arg3 interface{}, arg4 ...interface{}) (*workflow.Execution, error) {
		require.Len(t, arg4, 1)
		b.env.ExecuteWorkflow(ops.PaymentWorkflow, arg4[0].(string))

		return nil, nil
	}).AnyTimes()
	b.temporal = tp

	tc, err := tabapay_client.New(tabapay_client.NewClientArgs{}, b)
	require.NoError(t, err)
	b.tabapay = tc

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
	env.RegisterWorkflow(gmt_ops.GMTComplianceChecksWorkflow)
	env.RegisterWorkflow(gmt_ops.GMTNotifyCompleted)
	b.env = env
}

func (b *TestBackends) Twilio() twilio.Service {
	return nil
}

func (b *TestBackends) Signup() signup.Client {
	return nil
}

func (b *TestBackends) Analytics() analytics.Client {
	return analytics_client.New(b, "")
}

func (b *TestBackends) KYC() kyc.Client {
	c, err := kyc_client.New(b, "", "")
	if err != nil {
		panic("Can't initialise kyc client")
	}
	return c
}

func (b *TestBackends) Transactions() transactions.Client {
	return transaction_client.New(b)
}

func (b *TestBackends) Tabapay() tabapay.Client {
	return b.tabapay
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
	return nil
}

func (b *TestBackends) Keys() keys.Client {
	return keys_client.New(b)
}

func (b *TestBackends) Vault() vault.Client {
	return nil
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
