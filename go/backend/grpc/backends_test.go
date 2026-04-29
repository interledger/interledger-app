package grpc

import (
	"fmt"
	"net"
	"testing"

	"gitlab.com/fynbos/backend/providers/chimoney"

	"gitlab.com/fynbos/backend/providers/gatehub"

	pti "gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/rafiki"

	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/slack"

	"gitlab.com/fynbos/backend/features"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/wallets"
	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"

	"gitlab.com/fynbos/backend/contacts"
	contacts_mock "gitlab.com/fynbos/backend/contacts/client/mock"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/keys"
	keys_mock "gitlab.com/fynbos/backend/keys/client/mock"

	"gitlab.com/fynbos/backend/limits"
	limit_mock "gitlab.com/fynbos/backend/limits/client/mock"

	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/kyc"

	"github.com/jmoiron/sqlx"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/accountdeletion"
	accountdeletion_mock "gitlab.com/fynbos/backend/accountdeletion/client/mock"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	agreements_mock "gitlab.com/fynbos/backend/agreements/client/mock"
	email_mock "gitlab.com/fynbos/backend/email/client/mock"
	"gitlab.com/fynbos/backend/healthcheck"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linked_accounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	rafiki_mock "gitlab.com/fynbos/backend/rafiki/client/mock"
	"gitlab.com/fynbos/backend/signup"
	signup_mock "gitlab.com/fynbos/backend/signup/client/mock"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"
	"gitlab.com/fynbos/backend/twilio"
	twitter_mock "gitlab.com/fynbos/backend/twitter/client/mock"
	"gitlab.com/fynbos/backend/user"
	_user "gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	test_utils "gitlab.com/fynbos/backend/utils"
	"gitlab.com/fynbos/backend/waitlist"
	waitlist_mock "gitlab.com/fynbos/backend/waitlist/client/mock"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TestContainer struct {
	HealthService         healthcheck.Service
	AgreementsService     *agreements_mock.MockClient
	AdminAuthService      auth.Service
	UserService           user.Client
	linkedaccounts        *linked_accounts_mock.MockClient
	TwilioService         *twilio.MockService
	SignupService         *signup_mock.MockClient
	WaitlistClient        *waitlist_mock.MockClient
	TemporalImpl          *mocks.Client
	KYCClient             *kyc_mock.MockClient
	EmailClient           *email_mock.MockClient
	TransactionsClient    *transactions_mock.MockClient
	AnalyticsClient       analytics.Client
	ContactsClient        *contacts_mock.MockClient
	limits                *limit_mock.MockClient
	keys                  *keys_mock.MockClient
	TwitterClient         *twitter_mock.MockClient
	walletImpl            *wallets_mock.MockClient
	rafiki                *rafiki_mock.MockClient
	AccountDeletionClient *accountdeletion_mock.MockClient
}

func (t TestContainer) Xago() xago.Client {
	return nil
}

func (t TestContainer) Rafiki() rafiki.Client {
	return t.rafiki
}

func (t TestContainer) Slack() slack.Client {
	return nil
}

func (t TestContainer) Payments() payments.Client {
	return nil
}

func (t TestContainer) Wallets() wallets.Client {
	return t.walletImpl
}

func (t TestContainer) Features() features.Client {
	return nil
}

func (t TestContainer) Keys() keys.Client {
	return t.keys
}

func (t TestContainer) Identities() identities.Client {
	return nil
}

func (t TestContainer) Limits() limits.Client {
	return t.limits
}

func (t TestContainer) Email() email.Client {
	return t.EmailClient
}

func (t TestContainer) KYC() kyc.Client {
	return t.KYCClient
}

func (t TestContainer) DB() *sqlx.DB {
	return nil
}

func (t TestContainer) AdminAuth() auth.Service {
	return t.AdminAuthService
}

func (t TestContainer) Agreements() agreements.Client {
	return t.AgreementsService
}

func (t TestContainer) LinkedAccounts() linkedaccounts.Client {
	return t.linkedaccounts
}

func (t TestContainer) HealthCheck() healthcheck.Service {
	return t.HealthService
}

func (t TestContainer) Signup() signup.Client {
	return t.SignupService
}

func (t TestContainer) Temporal() client.Client {
	return t.TemporalImpl
}

func (t TestContainer) Twilio() twilio.Service {
	return t.TwilioService
}

func (t TestContainer) Users() _user.Client {
	return t.UserService
}

func (t TestContainer) Validator() *validator.Validate {
	return validator.New()
}

func (t TestContainer) Waitlist() waitlist.Client {
	return t.WaitlistClient
}

func (t TestContainer) Transactions() transactions.Client {
	return t.TransactionsClient
}

func (t TestContainer) Analytics() analytics.Client {
	return t.AnalyticsClient
}

func (t TestContainer) Contacts() contacts.Client {
	return t.ContactsClient
}

func (t TestContainer) Twitter() twitter.Client {
	return t.TwitterClient
}

func (t TestContainer) PTI() pti.Client {
	return nil
}

func (t TestContainer) Gatehub() gatehub.Client {
	return nil
}

func (t TestContainer) Chimoney() chimoney.Client {
	return nil
}

func (t TestContainer) AccountDeletion() accountdeletion.Client {
	return t.AccountDeletionClient
}

type TestContainerOption func(*TestContainer)

func NewTestContainer(t *testing.T, ctrl *gomock.Controller, opts ...TestContainerOption) *TestContainer {
	t.Cleanup(func() {
		ctrl.Finish()
	})

	hs, err := healthcheck.NewService()
	if err != nil {
		t.Fatal(err)
	}
	c := &TestContainer{
		HealthService:         hs,
		AgreementsService:     agreements_mock.NewMockClient(ctrl),
		AdminAuthService:      auth.NewMockService(),
		UserService:           user_mock.NewMock(),
		linkedaccounts:        linked_accounts_mock.NewMockClient(ctrl),
		TwilioService:         twilio.NewMockService(ctrl),
		SignupService:         signup_mock.NewMockClient(ctrl),
		WaitlistClient:        waitlist_mock.NewMockClient(ctrl),
		TemporalImpl:          &mocks.Client{},
		KYCClient:             kyc_mock.NewMockClient(ctrl),
		EmailClient:           email_mock.NewMockClient(ctrl),
		TransactionsClient:    transactions_mock.NewMockClient(ctrl),
		AnalyticsClient:       analytics_client.New(nil, ""),
		ContactsClient:        contacts_mock.NewMockClient(ctrl),
		limits:                limit_mock.NewMockClient(ctrl),
		keys:                  keys_mock.NewMockClient(ctrl),
		TwitterClient:         twitter_mock.NewMockClient(ctrl),
		walletImpl:            wallets_mock.NewMockClient(ctrl),
		rafiki:                rafiki_mock.NewMockClient(ctrl),
		AccountDeletionClient: accountdeletion_mock.NewMockClient(ctrl),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func startTestServer(
	t *testing.T,
	c *TestContainer,
) (*grpc.Server, backendv1.BackendServiceClient, backendv1.BackendServiceClient) {
	server, err := NewServer(c)
	if err != nil {
		t.Fatal(err)
	}

	port, err := test_utils.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}

	serverUrl := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", serverUrl)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := server.Serve(listener); err != nil {
			panic(fmt.Errorf("Failed to start test grpc server. %s", err))
		}
	}()

	conn, err := grpc.Dial(
		serverUrl, grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	backendClient := backendv1.NewBackendServiceClient(conn)

	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
		server.Stop()
	})

	return server, backendClient, backendClient
}
