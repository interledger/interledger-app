package unit

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/identity"
	_user "gitlab.com/fynbos/backend/user"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestUnitProvider(s *testing.T) {
	s.Parallel()
	ctx := context.Background()
	c := NewTestContainer(s)

	s.Run("Successfully verifies incoming webhook request", func(t *testing.T) {
		body := []byte(`{"data":[{"id":"2504140","type":"customer.created","attributes":{"createdAt":"2022-05-18T14:35:00.702Z","tags":{"userID":"02242b61-a99e-4b44-bda7-cf6a4f535a5f","test":"webhook-tag","key":"another-tag","number":"111"}},"relationships":{"customer":{"data":{"id":"344063","type":"individualCustomer"}},"application":{"data":{"id":"404728","type":"individualApplication"}}}}]}`)
		signature := "CmllgACV27KxvW0qP3fjnFfMPGg=" // key = fynbos_local_unit_webhook_token

		err := c.UnitService.VerifyWebhook(ctx, body, signature)

		assert.NoError(t, err)
	})

	s.Run("Fails if signature of webhook body does not match provided one", func(t *testing.T) {
		body := []byte(`{"test":"data"}`)
		signature := "CmllgACV27KxvW0qP3fjnFfMPGg"

		err := c.UnitService.VerifyWebhook(ctx, body, signature)

		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	s.Run("Successfully calls GetApplicationForm", func(t *testing.T) {
		user := _user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}

		form, err := c.UnitService.GetApplicationForm(ctx, user.ID)
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

		form, err := c.UnitService.CreateApplicationForm(ctx, &CreateApplicationFormArgs{
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

	s.Run("Successfully calls CreateApplication", func(t *testing.T) {
		user := _user.User{
			ID:    uuid.NewString(),
			Email: faker.Email(),
		}
		deviceFingerprints := make([]string, 1)
		deviceFingerprints = append(deviceFingerprints, "Some randon fingerprint")

		c.IdentityService.EXPECT().Get(gomock.Any(), user.ID).Return(&identity.Identity{
			ID:           user.ID,
			MobileNumber: faker.E164PhoneNumber(),
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			Country:      "US",
			Email:        user.Email,
		}, nil).Times(1)

		form, err := c.UnitService.CreateApplication(ctx, &CreateApplicationArgs{
			Ssn:                faker.Phonenumber(),
			DateOfBirth:        faker.Date(),
			Street:             faker.FirstName(),
			Street2:            faker.FirstName(),
			City:               faker.FirstName(),
			State:              faker.FirstName(),
			PostalCode:         faker.FirstName(),
			IpAddress:          faker.FirstName(),
			UserID:             user.ID,
			DeviceFingerprints: deviceFingerprints,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, form)
		assert.Equal(t, form.FynbosUserId, "106a75e9-de77-4e25-9561-faffe59d7814")
		assert.Equal(t, form.Status, "AwaitingDocuments")
	})
	s.Run("Creates and retrieves customer", func(t *testing.T) {
		accountID := uuid.NewString()
		customerID := uuid.NewString()
		customerType := "individual"
		customer, err := c.UnitService.CreateCustomer(ctx, &CreateCustomerArgs{
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

		customerByID, err := c.UnitService.GetCustomerByID(ctx, customer.ID)
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

func TestCreateAndGetCounterParty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	c := NewTestContainer(t)

	args := &CreateCounterPartyArgs{
		FundingsourceID: uuid.NewString(),
		Name:            "test name",
		UnitCustomerID:  uuid.NewString(),
		RoutingNumber:   faker.CCNumber(),
		AccountNumber:   faker.CCNumber(),
		AccountType:     faker.CCType(),
		Type:            "person",
		IdempotencyKey:  "test",
	}

	unitCounterparty, err := c.UnitService.CreateCounterParty(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.FundingsourceID, args.FundingsourceID)
	assert.NotEqual(t, "", unitCounterparty.ID)

	freshCounterParty, err := c.UnitService.GetCounterPartyByFundingsourceID(ctx, args.FundingsourceID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, args.FundingsourceID, freshCounterParty.FundingsourceID)
	assert.NotEqual(t, "", freshCounterParty.ID)
}

type TestContainer struct {
	Ctrl            *gomock.Controller
	UnitMockServer  *httptest.Server
	UnitService     Service
	IdentityService *identity.MockService
	Db              *sqlx.DB
}

func NewTestContainer(s *testing.T) *TestContainer {
	c := &TestContainer{}
	db := test_utils.MigrateCockroachDB(s, context.Background())
	c.Db = db
	c.Ctrl = gomock.NewController(s)
	identityService := identity.NewMockService(c.Ctrl)
	c.IdentityService = identityService
	c.UnitMockServer = test_utils.SetupUnitMockServer(context.Background())
	us, err := NewService(ServiceArgs{
		WebhookToken:    "fynbos_local_unit_webhook_token",
		BaseURL:         c.UnitMockServer.URL,
		Token:           "test token",
		Db:              db,
		IdentityService: identityService,
	})
	if err != nil {
		s.Fatal(err)
	}
	c.UnitService = us

	s.Cleanup(func() {
		c.UnitMockServer.Close()
	})

	return c
}
