package ops_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/linkedaccounts"
	linkedaccounts_mock "gitlab.com/fynbos/backend/linkedaccounts/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	external_client "gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
	"gitlab.com/fynbos/backend/providers/machnet/ops"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestCreateAndGetUser(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	walletID := NewWallet(t, b)

	args := machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: uuid.NewString(),
	}
	user, err := ops.CreateUser(context.Background(), b, args)
	require.NoError(t, err)
	require.Equal(t, args.ExternalID, user.ID)
	require.Equal(t, walletID, user.WalletID)

	freshUser, err := ops.GetUserByWalletID(context.Background(), b, walletID)
	require.NoError(t, err)
	require.Equal(t, args.ExternalID, freshUser.ID)
	require.Equal(t, walletID, freshUser.WalletID)

	freshUserByID, err := ops.GetUserByID(context.Background(), b, freshUser.ID)
	require.NoError(t, err)
	require.Equal(t, args.ExternalID, freshUserByID.ID)
	require.Equal(t, walletID, freshUserByID.WalletID)

	noUser, err := ops.GetUserByWalletID(context.Background(), b, uuid.NewString())
	assert.Nil(t, noUser)
	assert.ErrorIs(t, err, machnet.ErrNotFound)

	noUserByID, err := ops.GetUserByID(context.Background(), b, uuid.NewString())
	assert.Nil(t, noUserByID)
	assert.ErrorIs(t, err, machnet.ErrNotFound)
}

func TestGetWidgetToken(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	walletID := NewWallet(t, b)
	user := NewMachnetUser(t, b, walletID)

	token, err := ops.GetWidgetToken(context.Background(), b, walletID)
	require.NoError(t, err)

	assert.Equal(t, "machnet-widget-token", token.Value)
	assert.Equal(t, int(15), token.ExpiresInMinutes)
	assert.Equal(t, user.ID, token.UserID)

	// non-existent user
	token, err = ops.GetWidgetToken(context.Background(), b, uuid.NewString())
	require.Nil(t, token)
	assert.ErrorIs(t, err, machnet.ErrNotFound)
}

func TestHandleUserCardAddedEvent(t *testing.T) {
	t.Parallel()
	b := NewTestBackends(t)
	walletID := NewWallet(t, b)
	user := NewMachnetUser(t, b, walletID)
	externalFundingsource, err := b.external.CreateUserFundingsource(
		context.Background(),
		external.FundingSource{
			ID:                 uuid.NewString(),
			UserID:             user.ID,
			AccountNumber:      "9991",
			FundingsourceName:  "VISA-9991",
			FundingsourceType:  "CARD",
			InstitutionName:    "VISA",
			VerificationStatus: external.StatusVerified,
		},
	)
	require.NoError(t, err)
	userCardAddedEvent := external.Event{
		ID:             uuid.NewString(),
		EventName:      external.UserCardAdded,
		ResourceID:     externalFundingsource.ID,
		UserID:         externalFundingsource.UserID,
		SubscriptionID: uuid.NewString(),
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Payload:        []byte("{}"),
	}

	b.linkedaccounts.EXPECT().Create(gomock.Any(), &linkedaccounts.CreateArgs{
		WalletID:   walletID,
		Name:       externalFundingsource.FundingsourceName,
		Provider:   machnet.ProviderName,
		ProviderID: externalFundingsource.ID,
		Type:       external.TypeCard,
		Mask:       externalFundingsource.AccountNumber,
	}).Return(
		&linkedaccounts.LinkedAccount{ID: uuid.NewString(), WalletId: walletID},
		nil,
	).Times(1)

	err = ops.HandleEvent(context.Background(), b, userCardAddedEvent)
	require.NoError(t, err)
}

func NewWallet(t *testing.T, b backends) string {
	walletID := uuid.NewString()
	_, err := b.DB().Exec(
		"INSERT INTO wallets (id, name) VALUES ($1, $2);",
		walletID,
		"test",
	)
	require.NoError(t, err)

	return walletID
}
func NewMachnetUser(t *testing.T, b backends, walletID string) *machnet.User {
	externalUser, err := b.External().RegisterUser(context.Background(), external.User{
		Type: external.SendUser,
	})
	require.NoError(t, err)
	user, err := ops.CreateUser(context.Background(), b, machnet.CreateArgs{
		WalletID:   walletID,
		ExternalID: externalUser.ID,
	})
	require.NoError(t, err)

	return user
}

func NewTestBackends(t *testing.T) backends {
	ctrl := gomock.NewController(t)
	return backends{
		db:             test_utils.MigrateCockroachDB(t, context.Background()),
		external:       external_client.New(),
		linkedaccounts: linkedaccounts_mock.NewMockClient(ctrl),
	}
}

type backends struct {
	db             *sqlx.DB
	external       *external_client.Client
	linkedaccounts *linkedaccounts_mock.MockClient
	users    user.Client
	kycImpl  kyc.Client
}

func (b backends) Users() user.Client {
	return b.users
}

func (b backends) KYC() kyc.Client {
	return b.kycImpl
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) External() external.Client {
	return b.external
}

func (b backends) LinkedAccounts() linkedaccounts.Client {
	return b.linkedaccounts
}
