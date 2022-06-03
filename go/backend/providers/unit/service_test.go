package unit

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	_accounts "gitlab.com/fynbos/backend/accounts"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	_country "gitlab.com/fynbos/backend/country"
	_identity "gitlab.com/fynbos/backend/identity"
	"gitlab.com/fynbos/backend/providers/noop"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestUnitProvider(s *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		err := c.Cleanup(ctx)
		if err != nil {
			return
		}
	})

	s.Run("Successfully verifies incoming webhook request", func(t *testing.T) {
		body := []byte(`{"data":[{"id":"2504140","type":"customer.created","attributes":{"createdAt":"2022-05-18T14:35:00.702Z","tags":{"userID":"02242b61-a99e-4b44-bda7-cf6a4f535a5f","test":"webhook-tag","key":"another-tag","number":"111"}},"relationships":{"customer":{"data":{"id":"344063","type":"individualCustomer"}},"application":{"data":{"id":"404728","type":"individualApplication"}}}}]}`)
		signature := "CmllgACV27KxvW0qP3fjnFfMPGg=" // key = fynbos_local_unit_webhook_token

		err = c.UnitService.VerifyWebhook(c.Ctx, body, signature)

		assert.NoError(t, err)
	})

	s.Run("Fails if signature of webhook body does not match provided one", func(t *testing.T) {
		body := []byte(`{"test":"data"}`)
		signature := "CmllgACV27KxvW0qP3fjnFfMPGg"

		err = c.UnitService.VerifyWebhook(c.Ctx, body, signature)

		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	s.Run("Successfully calls GetApplicationForm", func(t *testing.T) {
		user := _user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		form, err := c.UnitService.GetApplicationForm(c.Ctx, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, form)
		assert.Equal(t, "411479", form.ID)
		assert.Equal(t, "https://application-form.sh/DXB4GXQMBGY377CD5KQ3OWX4XJEF4Z3DQPKTMDGF77CFQM7M55WOQR5C2C3D5N2NYP52AOCSVZX6JLLGSHRLI3DXZ45R43QPDIBWUAI7KL6I7ESUPTB7C7EFURQKMZZSINKSXYQ2N63L7TFPCQVQIW6TVQQLXUYJQP6FY", form.URL)
	})
	s.Run("Successfully calls CreateApplicationForm", func(t *testing.T) {
		user := _user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		form, err := c.UnitService.CreateApplicationForm(c.Ctx, &CreateApplicationFormArgs{
			ID:      user.ID,
			Email:   user.Email,
			Country: "US",
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, form)
		assert.Equal(t, "411479", form.ID)
		assert.Equal(t, "https://application-form.sh/LJ45W6SSGO6VFFNKMLR5WPOSLH6KMSXQZPGXIPG64SLXHD5TCV4GSYXWZVUSNUEIW2KP5SZOI4RMP6IJRKLF5TTDJTU4TCLU3LQX2XFDIQAMG7TKSXHCQY3KGZ3RFEBYEQCB3GGYUGIUWBXT2ZEIOVNBG72GGNNJKMFJ6", form.URL)
	})

	s.Run("Creates and retrieves customer", func(t *testing.T) {
		accountID := uuid.NewString()
		customerID := uuid.NewString()
		customerType := "individual"
		customer, err := c.UnitService.CreateCustomer(c.Ctx, &CreateCustomerArgs{
			ID:        customerID,
			AccountID: accountID,
			Type:      customerType,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, customerID, customer.ID)
		assert.Equal(t, accountID, customer.AccountID)
		assert.Equal(t, customerType, customer.Type)

		customerByID, err := c.UnitService.GetCustomerByID(c.Ctx, customer.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, customerID, customerByID.ID)
		assert.Equal(t, accountID, customerByID.AccountID)
		assert.Equal(t, customerType, customerByID.Type)

		customerByAccountID, err := c.UnitService.GetCustomerByAccountID(ctx, accountID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, customerID, customerByAccountID.ID)
		assert.Equal(t, customerID, customerByAccountID.ID)
		assert.Equal(t, customerType, customerByAccountID.Type)
	})
}

type TestContainer struct {
	UnitMockServer     *httptest.Server
	UnitService        Service
	IdentityService    _identity.Service
	AccountService     _accounts.Service
	CountryService     _country.Service
	NoopService        noop.Service
	TransactionService account_transactions.Service
	TemporalMock       *mocks.Client
	PacioliContainer   *test_utils.PacioliContainer
	PacioliClient      pacioliv1.PacioliServiceClient
	PacioliLedgerID    uint16
	Db                 *sqlx.DB
	Logger             *zap.Logger
	Ctx                context.Context
}

func (c *TestContainer) Cleanup(ctx context.Context) error {
	c.UnitMockServer.Close()

	err := c.PacioliContainer.Terminate(ctx)
	if err != nil {
		return err
	}

	return nil
}

func NewTestContainer(ctx context.Context, s *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx
	db := test_utils.MigrateCockroachDB(s, ctx)
	c.Db = db

	c.PacioliContainer = test_utils.SetupPacioli(s, ctx)

	c.PacioliLedgerID = uint16(1)
	conn, err := grpc.Dial(c.PacioliContainer.PacioliUrl, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	pClient := pacioliv1.NewPacioliServiceClient(conn)
	c.PacioliClient = pClient

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	cs := _country.NewService(db)
	c.CountryService = cs

	is, err := _identity.NewService(_identity.ServiceArgs{
		CountryService: cs,
		Db:             db,
	})
	if err != nil {
		s.Fatal(err)
	}
	c.IdentityService = _identity.NewLoggingService(is, logger)

	as, err := _accounts.NewService(&_accounts.ServiceArgs{
		Is:              is,
		Cs:              cs,
		PacioliLedgerID: c.PacioliLedgerID,
		PacioliClient:   pClient,
		Db:              db,
	})
	if err != nil {
		return nil, err
	}
	err = as.Init(ctx)
	if err != nil {
		return nil, err
	}
	c.AccountService = _accounts.NewLoggingService(as, logger)

	np, err := noop.NewService(noop.ServiceArgs{
		LedgerID:      c.PacioliLedgerID,
		EquityAccID:   uuid.NewString(),
		PacioliTenant: "dev",
		PacioliClient: pClient,
	})
	if err != nil {
		return nil, err
	}
	err = np.Init(ctx)
	if err != nil {
		return nil, err
	}
	c.NoopService = np

	c.UnitMockServer = test_utils.SetupUnitMockServer(ctx)

	us, err := NewService(ServiceArgs{
		WebhookToken: "fynbos_local_unit_webhook_token",
		BaseURL:      c.UnitMockServer.URL,
		Token:        "test token",
		Db:           db,
	})
	if err != nil {
		return nil, err
	}
	c.UnitService = us

	ts, err := account_transactions.NewService(&account_transactions.ServiceArgs{
		AccountService: as,
		PacioliClient:  pClient,
		Db:             db,
	})
	if err != nil {
		return nil, err
	}

	c.TransactionService = ts

	temporal := &mocks.Client{}
	c.TemporalMock = temporal

	return c, nil
}
