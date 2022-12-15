package grpc

import (
	"fmt"
	"net"
	"testing"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/kyc"

	"github.com/jmoiron/sqlx"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/admin/auth"
	"gitlab.com/fynbos/backend/agreements"
	agreements_mock "gitlab.com/fynbos/backend/agreements/client/mock"
	"gitlab.com/fynbos/backend/country"
	country_mock "gitlab.com/fynbos/backend/country/client/mock"
	email_mock "gitlab.com/fynbos/backend/email/client/mock"
	"gitlab.com/fynbos/backend/healthcheck"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linked_accounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/fakecash"
	fakecash_mock "gitlab.com/fynbos/backend/providers/fakecash/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_mock "gitlab.com/fynbos/backend/providers/machnet/client/mock"
	"gitlab.com/fynbos/backend/signup"
	signup_mock "gitlab.com/fynbos/backend/signup/client/mock"
	"gitlab.com/fynbos/backend/supporttickets"
	support_mock "gitlab.com/fynbos/backend/supporttickets/client/mock"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"
	"gitlab.com/fynbos/backend/twilio"
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
	HealthService      healthcheck.Service
	AgreementsService  *agreements_mock.MockClient
	CountriesService   *country_mock.MockClient
	AdminAuthService   auth.Service
	UserService        user.Client
	fakecash           *fakecash_mock.MockClient
	linkedaccounts     *linked_accounts_mock.MockClient
	machnet            *machnet_mock.MockClient
	TwilioService      *twilio.MockService
	SignupService      *signup_mock.MockClient
	WaitlistClient     *waitlist_mock.MockClient
	TemporalImpl       *mocks.Client
	TicketClient       *support_mock.MockClient
	KYCClient          *kyc_mock.MockClient
	EmailClient        *email_mock.MockClient
	TransactionsClient *transactions_mock.MockClient
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

func (t TestContainer) Countries() country.Client {
	return t.CountriesService
}

func (t TestContainer) FakeCash() fakecash.Client {
	return t.fakecash
}

func (t TestContainer) LinkedAccounts() linkedaccounts.Client {
	return t.linkedaccounts
}

func (t TestContainer) Machnet() machnet.Client {
	return t.machnet
}

func (t TestContainer) HealthCheck() healthcheck.Service {
	return t.HealthService
}

func (t TestContainer) Signup() signup.Client {
	return t.SignupService
}

func (t TestContainer) SupportTickets() supporttickets.Client {
	return t.TicketClient
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
		CountriesService:   country_mock.NewMockClient(ctrl),
		AdminAuthService:   auth.NewMockService(),
		UserService:        user_mock.NewMock(),
		fakecash:           fakecash_mock.NewMockClient(ctrl),
		linkedaccounts:     linked_accounts_mock.NewMockClient(ctrl),
		machnet:            machnet_mock.NewMockClient(ctrl),
		TwilioService:      twilio.NewMockService(ctrl),
		SignupService:      signup_mock.NewMockClient(ctrl),
		WaitlistClient:     waitlist_mock.NewMockClient(ctrl),
		TicketClient:       support_mock.NewMockClient(ctrl),
		TemporalImpl:       &mocks.Client{},
		KYCClient:          kyc_mock.NewMockClient(ctrl),
		TransactionsClient: transactions_mock.NewMockClient(ctrl),
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
