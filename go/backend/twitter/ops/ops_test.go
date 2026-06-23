package ops_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/twitter"
	external_mock "github.com/interledger/interledger-app/go/backend/twitter/external/client/mock"
	"github.com/interledger/interledger-app/go/backend/twitter/ops"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestCreateAuthURL(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(func(tb *ops.TestBackends) {
		tb.Db = db.MigrateTestDB(t, ctx)
	})

	url, err := ops.CreateAuthURL(ctx, b, &ops.CreateAuthURLArgs{
		ClientID:     uuid.NewString(),
		WalletID:     uuid.NewString(),
		Scopes:       []string{"tweet.read"},
		RedirectURL:  "https://fynbos.app/twitter/callback",
		AuthEndpoint: "https://twitter.com/i/oauth2/authorize",
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	fmt.Println(url)
}

func TestCreateToken(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := ops.NewTestBackends(func(tb *ops.TestBackends) {
		tb.Db = db.MigrateTestDB(t, ctx)
		tb.ExternalClient = external_mock.NewMockClient(ctrl)
	})

	state := uuid.NewString()

	_, err := ops.CreateAuthURL(ctx, b, &ops.CreateAuthURLArgs{
		ClientID:     uuid.NewString(),
		WalletID:     uuid.NewString(),
		Scopes:       []string{"tweet.read"},
		RedirectURL:  "https://fynbos.app/twitter/callback",
		AuthEndpoint: "https://twitter.com/i/oauth2/authorize",
		State:        state,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	b.ExternalClient.EXPECT().CreateToken(gomock.Any(), gomock.Any()).Return(&oauth2.Token{
		AccessToken:  "testAccessToken",
		TokenType:    "testTokenType",
		RefreshToken: "testRefreshToken",
		Expiry:       time.Now(),
	}, nil)

	b.ExternalClient.EXPECT().GetAuthorizedUser(gomock.Any(), gomock.Any()).Return(&twitter.TwitterUser{
		ID:       "testTwitterID",
		Username: "testUsername",
	}, nil)

	connection, err := ops.CreateConnection(ctx, b, &twitter.CreateConnectionArgs{
		AuthCode: "testAuthCode",
		State:    state,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.Equal(t, "testAccessToken", connection.AccessToken)
	assert.Equal(t, "testTwitterID", connection.UserID)
	assert.Equal(t, "testRefreshToken", connection.RefreshToken)
	assert.Equal(t, "testTokenType", connection.TokenType)
	assert.Equal(t, "testUsername", connection.Username)
}

func TestGetWalletConnections(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := ops.NewTestBackends(func(tb *ops.TestBackends) {
		tb.Db = db.MigrateTestDB(t, ctx)
		tb.ExternalClient = external_mock.NewMockClient(ctrl)
	})

	walletId := uuid.NewString()
	state := uuid.NewString()

	_, err := ops.CreateAuthURL(ctx, b, &ops.CreateAuthURLArgs{
		ClientID:     uuid.NewString(),
		WalletID:     walletId,
		Scopes:       []string{"tweet.read", "tweet.write"},
		RedirectURL:  "https://fynbos.app/twitter/callback",
		AuthEndpoint: "https://twitter.com/i/oauth2/authorize",
		State:        state,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	b.ExternalClient.EXPECT().CreateToken(gomock.Any(), gomock.Any()).Return(&oauth2.Token{
		AccessToken:  "testAccessToken",
		TokenType:    "testTokenType",
		RefreshToken: "testRefreshToken",
		Expiry:       time.Now(),
	}, nil)

	b.ExternalClient.EXPECT().GetAuthorizedUser(gomock.Any(), gomock.Any()).Return(&twitter.TwitterUser{
		ID:       "testTwitterID",
		Username: "testUsername",
	}, nil)

	_, err = ops.CreateConnection(ctx, b, &twitter.CreateConnectionArgs{
		AuthCode: "testAuthCode",
		State:    state,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	connections, err := ops.GetWalletConnections(ctx, b, walletId)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.Equal(t, "testAccessToken", connections[0].AccessToken)
	assert.Equal(t, "testTwitterID", connections[0].UserID)
	assert.Equal(t, "testUsername", connections[0].Username)
	assert.Equal(t, walletId, connections[0].WalletID)
}
