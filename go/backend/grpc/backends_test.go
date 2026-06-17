package grpc

import (
	"fmt"
	"net"
	"testing"

	"github.com/interledger/interledger-app/go/backend/providers/chimoney"

	"github.com/interledger/interledger-app/go/backend/providers/gatehub"

	pti "github.com/interledger/interledger-app/go/backend/providers/pti"
	"github.com/interledger/interledger-app/go/backend/providers/xago"

	"github.com/interledger/interledger-app/go/backend/rafiki"

	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/slack"

	"github.com/interledger/interledger-app/go/backend/features"
	"github.com/interledger/interledger-app/go/backend/twitter"
	"github.com/interledger/interledger-app/go/backend/wallets"
	wallets_mock "github.com/interledger/interledger-app/go/backend/wallets/client/mock"

	"github.com/interledger/interledger-app/go/backend/contacts"
	contacts_mock "github.com/interledger/interledger-app/go/backend/contacts/client/mock"
	"github.com/interledger/interledger-app/go/backend/identities"
	"github.com/interledger/interledger-app/go/backend/keys"
	keys_mock "github.com/interledger/interledger-app/go/backend/keys/client/mock"

	"github.com/interledger/interledger-app/go/backend/limits"
	limit_mock "github.com/interledger/interledger-app/go/backend/limits/client/mock"

	"github.com/interledger/interledger-app/go/backend/analytics"
	analytics_client "github.com/interledger/interledger-app/go/backend/analytics/client"

	"github.com/interledger/interledger-app/go/backend/email"
	"github.com/interledger/interledger-app/go/backend/transactions"

	"github.com/interledger/interledger-app/go/backend/kyc"

	"github.com/jmoiron/sqlx"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/interledger/interledger-app/go/backend/admin/auth"
	"github.com/interledger/interledger-app/go/backend/agreements"
	agreements_mock "github.com/interledger/interledger-app/go/backend/agreements/client/mock"
	email_mock "github.com/interledger/interledger-app/go/backend/email/client/mock"
	"github.com/interledger/interledger-app/go/backend/healthcheck"
	kyc_mock "github.com/interledger/interledger-app/go/backend/kyc/client/mock"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	linked_accounts_mock "github.com/interledger/interledger-app/go/backend/linkedaccounts/client/mock"
	rafiki_mock "github.com/interledger/interledger-app/go/backend/rafiki/client/mock"
	"github.com/interledger/interledger-app/go/backend/signup"
	signup_mock "github.com/interledger/interledger-app/go/backend/signup/client/mock"
	transactions_mock "github.com/interledger/interledger-app/go/backend/transactions/client/mock"
	"github.com/interledger/interledger-app/go/backend/twilio"
	twitter_mock "github.com/interledger/interledger-app/go/backend/twitter/client/mock"
	"github.com/interledger/interledger-app/go/backend/user"
	_user "github.com/interledger/interledger-app/go/backend/user"
	user_mock "github.com/interledger/interledger-app/go/backend/user/client/mock"
	test_utils "github.com/interledger/interledger-app/go/backend/utils"
	"github.com/interledger/interledger-app/go/backend/waitlist"
	waitlist_mock "github.com/interledger/interledger-app/go/backend/waitlist/client/mock"
	backendv1 "github.com/interledger/interledger-app/go/proto/backend/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TestContainer struct {
	HealthService      healthcheck.Service
	AgreementsService  *agreements_mock.MockClient
	AdminAuthService   auth.Service
	UserService        user.Client
	linkedaccounts     *linked_accounts_mock.MockClient
	TwilioService      *twilio.MockService
	SignupService      *signup_mock.MockClient
	WaitlistClient     *waitlist_mock.MockClient
	TemporalImpl       *mocks.Client
	KYCClient          *kyc_mock.MockClient
	EmailClient        *email_mock.MockClient
	TransactionsClient *transactions_mock.MockClient
	AnalyticsClient    analytics.Client
	ContactsClient     *contacts_mock.MockClient
	limits             *limit_mock.MockClient
	keys               *keys_mock.MockClient
	TwitterClient      *twitter_mock.MockClient
	walletImpl         *wallets_mock.MockClient
	rafiki             *rafiki_mock.MockClient
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
		HealthService:      hs,
		AgreementsService:  agreements_mock.NewMockClient(ctrl),
		AdminAuthService:   auth.NewMockService(),
		UserService:        user_mock.NewMock(),
		linkedaccounts:     linked_accounts_mock.NewMockClient(ctrl),
		TwilioService:      twilio.NewMockService(ctrl),
		SignupService:      signup_mock.NewMockClient(ctrl),
		WaitlistClient:     waitlist_mock.NewMockClient(ctrl),
		TemporalImpl:       &mocks.Client{},
		KYCClient:          kyc_mock.NewMockClient(ctrl),
		TransactionsClient: transactions_mock.NewMockClient(ctrl),
		AnalyticsClient:    analytics_client.New(nil, ""),
		ContactsClient:     contacts_mock.NewMockClient(ctrl),
		limits:             limit_mock.NewMockClient(ctrl),
		keys:               keys_mock.NewMockClient(ctrl),
		TwitterClient:      twitter_mock.NewMockClient(ctrl),
		walletImpl:         wallets_mock.NewMockClient(ctrl),
		rafiki:             rafiki_mock.NewMockClient(ctrl),
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
