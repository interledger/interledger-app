package ops_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/discord"
	external_mock "gitlab.com/fynbos/backend/discord/external/client/mock"
	"gitlab.com/fynbos/backend/discord/ops"
	"golang.org/x/oauth2"
)

func TestCreateAuthURL(t *testing.T) {
	t.Skip("Discord is deprecated")
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := &TestBackends{
		Db:             db.MigrateTestDB(t, ctx),
		ExternalClient: external_mock.NewMockClient(ctrl),
	}

	authURL, err := ops.CreateAuthURL(ctx, b, ops.CreateAuthURLArgs{
		ClientID:     "test",
		WalletID:     uuid.NewString(),
		Scopes:       []string{"identify"},
		RedirectURL:  "https://interledger.test/connect/discord",
		AuthEndpoint: "https://discord.com/api/oauth2/authorize",
		State:        "state",
	})
	require.NoError(t, err)

	url, err := url.Parse(authURL)
	require.NoError(t, err)
	assert.Equal(t, "discord.com", url.Host)
	assert.Equal(t, "/api/oauth2/authorize", url.Path)
	query := url.Query()
	assert.Equal(t, "test", query.Get("client_id"))
	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, "identify", query.Get("scope"))
	assert.Equal(t, "state", query.Get("state"))
	assert.Equal(t, "https://interledger.test/connect/discord", query.Get("redirect_uri"))
}

func TestCreateToken(t *testing.T) {
	t.Skip("Discord is deprecated")
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := &TestBackends{
		Db:             db.MigrateTestDB(t, ctx),
		ExternalClient: external_mock.NewMockClient(ctrl),
	}

	walletID := uuid.NewString()
	_, err := ops.CreateAuthURL(ctx, b, ops.CreateAuthURLArgs{
		ClientID:     "test",
		WalletID:     walletID,
		Scopes:       []string{"identify"},
		RedirectURL:  "https://interledger.test/connect/discord",
		AuthEndpoint: "https://discord.com/api/oauth2/authorize",
		State:        "state",
	})
	require.NoError(t, err)

	b.ExternalClient.EXPECT().CreateToken(gomock.Any(), gomock.Any()).Return(&oauth2.Token{
		AccessToken:  "testAccessToken",
		TokenType:    "testTokenType",
		RefreshToken: "testRefreshToken",
		Expiry:       time.Now(),
	}, nil)

	b.ExternalClient.EXPECT().GetAuthorizedUser(gomock.Any(), gomock.Any()).Return(&discord.User{
		ID:       "testID",
		Username: "testUsername",
	}, nil)

	connection, err := ops.CreateConnection(ctx, b, discord.CreateConnectionArgs{
		AuthCode: "testAuthCode",
		State:    "state",
	})
	require.NoError(t, err)
	assert.Equal(t, "testAccessToken", connection.AccessToken)
	assert.Equal(t, "testID", connection.UserID)
	assert.Equal(t, "testRefreshToken", connection.RefreshToken)
	assert.Equal(t, "testTokenType", connection.TokenType)
	assert.Equal(t, "testUsername", connection.Username)

	connections, err := ops.GetWalletConnections(ctx, b, walletID)
	require.NoError(t, err)
	require.Len(t, connections, 1)
	assert.Equal(t, "testAccessToken", connections[0].AccessToken)
	assert.Equal(t, "testID", connections[0].UserID)
	assert.Equal(t, "testUsername", connections[0].Username)
	assert.Equal(t, walletID, connections[0].WalletID)
}
