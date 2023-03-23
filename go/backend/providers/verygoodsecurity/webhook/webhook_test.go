package webhook

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/analytics"
	analytics_client "gitlab.com/fynbos/backend/analytics/client"
	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/paymentpointers"
	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
	verygoodsecurity_client "gitlab.com/fynbos/backend/providers/verygoodsecurity/client"
	"gitlab.com/fynbos/backend/user"
	user_client "gitlab.com/fynbos/backend/user/client"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"go.uber.org/zap"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

type TestContainer struct {
	Ctx              context.Context
	Logger           *zap.Logger
	Db               *sqlx.DB
	Us               user.Client
	verygoodsecurity verygoodsecurity.Client
	ValidatorImpl    *validator.Validate
	Ac               analytics.Client
}

func (t TestContainer) Validator() *validator.Validate {
	return t.ValidatorImpl
}

func (t TestContainer) DB() *sqlx.DB {
	return t.Db
}

func (t TestContainer) Users() user.Client {
	return t.Us
}
func (t TestContainer) VGS() verygoodsecurity.Client {
	return t.verygoodsecurity
}

func (t TestContainer) Analytics() analytics.Client {
	return t.Ac
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{ValidatorImpl: validator.New()}
	c.Ctx = ctx
	mdb := db.MigrateTestDB(t, ctx)
	c.Db = mdb

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	c.Ac = analytics_client.New(c, "")

	c.Us = user_client.New(c, "kratosURL", "kratosAdminURL")

	c.verygoodsecurity = verygoodsecurity_client.New(c)

	return c, nil
}

func TestNewHandleInboundCard(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c, err := NewTestContainer(t, ctrl)
	if err != nil {
		s.Fatal(err)
	}
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	_, err := c.Users().CreateNewWallet(context.Background(), u.ID, "Marko Polo")
	require.NoError(t, err)

	// Create contact
	uc := &user.User{
		ID: uuid.NewString(),
	}
	contactWallet, err := c.Users().CreateNewWallet(context.Background(), uc.ID, "Alice bob")
	require.NoError(t, err)
	pp, err := paymentpointers.Parse("$fynbos.me/alice")
	require.NoError(t, err)

	c.OPClient.EXPECT().GetPaymentPointer(gomock.Any(), pp.String()).Return(&openpayments.PaymentPointer{
		ID:         uuid.NewString(),
		URL:        pp.String(),
		WalletID:   contactWallet.ID,
		Alias:      "Test",
		Asset:      "USD",
		AssetScale: 2,
	}, nil).AnyTimes()

	c.ContactsClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(
		&contacts.Contact{
			ID:             uuid.NewString(),
			Name:           contactWallet.Name,
			PaymentPointer: pp,
			WalletID:       contactWallet.ID,
		},
		nil,
	).AnyTimes()

	rpc, err := client.CreateContact(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.CreateContactRequest{
		PaymentPointer: pp.ShortString(),
	})
	require.NoError(t, err)

	assert.Equal(t, contactWallet.ID, rpc.WalletId)
}
