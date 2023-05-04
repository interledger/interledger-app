package ops_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/twitter"
	external_mock "gitlab.com/fynbos/backend/twitter/external/client/mock"
	"gitlab.com/fynbos/backend/twitter/ops"
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

	authorization, err := ops.CreateAuthURL(ctx, b, &ops.CreateAuthURLArgs{
		ClientID:     uuid.NewString(),
		WalletID:     uuid.NewString(),
		Scopes:       []string{"tweet.read"},
		RedirectURL:  "https://fynbos.app/twitter/callback",
		AuthEndpoint: "https://twitter.com/i/oauth2/authorize",
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

	accessToken, err := ops.CreateToken(ctx, b, &twitter.CreateTokenArgs{
		AuthCode: "testAuthCode",
		State:    authorization.State,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.Equal(t, "testAccessToken", accessToken.AccessToken)
	assert.Equal(t, "testRefreshToken", accessToken.RefreshToken)
	assert.Equal(t, "testTokenType", accessToken.TokenType)
}
