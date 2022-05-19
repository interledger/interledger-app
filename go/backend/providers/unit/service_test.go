package unit

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
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
		data := []byte(`{"test":"data"}`)

		mac := hmac.New(sha1.New, []byte("test-webhook-token"))
		mac.Write(data)

		sha := hex.EncodeToString(mac.Sum(nil))

		req, err := http.NewRequest("POST", "http://fynbos.test/webhooks/unit", bytes.NewBuffer(data))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("x-unit-signature", sha)
		req.Header.Set("Content-Type", "application/json")

		err = c.UnitService.VerifyWebhook(c.Ctx, req)

		assert.NoError(t, err)
	})

	s.Run("Fails if webhook tokens does not match", func(t *testing.T) {
		data := []byte(`{"test":"data"}`)

		mac := hmac.New(sha1.New, []byte("malicious-webhook-token"))
		mac.Write(data)

		sha := hex.EncodeToString(mac.Sum(nil))

		req, err := http.NewRequest("POST", "http://fynbos.test/webhooks/unit", bytes.NewBuffer(data))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("x-unit-signature", sha)
		req.Header.Set("Content-Type", "application/json")

		err = c.UnitService.VerifyWebhook(c.Ctx, req)

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
	Crdb               *test_utils.CockroachDBContainer
	Ctx                context.Context
}

func (c *TestContainer) Cleanup(ctx context.Context) error {
	err := c.Db.Close()
	if err != nil {
		return err
	}

	err = c.Crdb.Container.Terminate(ctx)
	if err != nil {
		return err
	}

	c.UnitMockServer.Close()

	err = c.PacioliContainer.Terminate(ctx)
	if err != nil {
		return err
	}

	return nil
}

func NewTestContainer(ctx context.Context, s *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		return nil, err
	}
	c.Crdb = crdb

	// the tests are run in serial. We use a global connection for
	// each of the tests.
	db, err := sqlx.Connect("postgres", crdb.URI)
	if err != nil {
		return nil, err
	}
	c.Db = db

	pacioliContainer, err := test_utils.SetupPacioli(ctx)
	if err != nil {
		return nil, err
	}
	c.PacioliContainer = pacioliContainer

	c.PacioliLedgerID = uint16(1)
	conn, err := grpc.Dial(pacioliContainer.PacioliUrl, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		WebhookToken: "test-webhook-token",
		BaseURL:      c.UnitMockServer.URL,
		Token:        "test token",
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
