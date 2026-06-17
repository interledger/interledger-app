package ops_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	payments_mock "github.com/interledger/interledger-app/go/backend/payments/client/mock"

	"github.com/interledger/interledger-app/go/backend/payments"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/slack"
	"github.com/interledger/interledger-app/go/backend/slack/external"
	external_mock "github.com/interledger/interledger-app/go/backend/slack/external/mock"
	"github.com/interledger/interledger-app/go/backend/slack/ops"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type TestBackends struct {
	Db             *sqlx.DB
	ExternalClient *external_mock.MockClient
	PayClient      *payments_mock.MockClient
}

func (tb *TestBackends) Payments() payments.Client {
	return tb.PayClient
}

func (tb *TestBackends) DB() *sqlx.DB {
	return tb.Db
}

func (tb *TestBackends) External() external.Client {
	return tb.ExternalClient
}

func TestCreateAuthURL(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := &TestBackends{
		Db:             db.MigrateTestDB(t, ctx),
		ExternalClient: external_mock.NewMockClient(ctrl),
	}

	b.ExternalClient.EXPECT().GetConfig().Return(&oauth2.Config{
		ClientID:     "123",
		ClientSecret: "123",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://slack.com/api/oauth2/authorize",
			TokenURL:  "https://slack.com/api/oauth2/token",
			AuthStyle: 0,
		},
		RedirectURL: "https://interledger.test/redirect/slack",
		Scopes:      []string{oidc.ScopeOpenID, "profile", "email"},
	}).AnyTimes()

	walletID := uuid.NewString()
	authURL, err := ops.CreateAuthURL(ctx, b, walletID)
	require.NoError(t, err)

	fmt.Println(authURL)

	url, err := url.Parse(authURL)
	require.NoError(t, err)
	assert.Equal(t, "slack.com", url.Host)
	assert.Equal(t, "/api/oauth2/authorize", url.Path)
	query := url.Query()
	assert.Equal(t, "123", query.Get("client_id"))
	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, "openid profile email", query.Get("scope"))
	assert.Equal(t, "https://interledger.test/redirect/slack", query.Get("redirect_uri"))
}

func TestCreateConnection(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	b := &TestBackends{
		Db:             db.MigrateTestDB(t, ctx),
		ExternalClient: external_mock.NewMockClient(ctrl),
	}

	b.ExternalClient.EXPECT().GetConfig().Return(&oauth2.Config{
		ClientID:     "123",
		ClientSecret: "123",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://slack.com/api/oauth2/authorize",
			TokenURL:  "https://slack.com/api/oauth2/token",
			AuthStyle: 0,
		},
		RedirectURL: "https://interledger.test/redirect/slack",
		Scopes:      []string{oidc.ScopeOpenID, "profile", "email"},
	}).AnyTimes()

	b.ExternalClient.EXPECT().CreateUserToken(gomock.Any(), gomock.Any()).Return(&oauth2.Token{
		AccessToken:  "access_token",
		TokenType:    "Bearer",
		RefreshToken: "refresh_token",
		Expiry:       time.Now(),
	}, &external.User{
		ID:         "user_id",
		Username:   "batman",
		TeamName:   "interledger",
		TeamDomain: "interledgertest",
		TeamID:     "team_id",
	}, nil).Times(1)

	authURL, err := ops.CreateAuthURL(ctx, b, uuid.NewString())
	require.NoError(t, err)

	fmt.Println(authURL)

	url, err := url.Parse(authURL)
	require.NoError(t, err)
	query := url.Query()
	state := query.Get("state")

	con, err := ops.CreateConnection(ctx, b, slack.CreateConnectionArgs{
		AuthCode: "auth",
		State:    state,
	})
	require.NoError(t, err)

	assert.Equal(t, "interledgertest", con.TeamDomain)
	assert.Equal(t, "interledger", con.TeamName)
	assert.Len(t, con.Scopes, 3)
	assert.Equal(t, "access_token", con.AccessToken)
	assert.Equal(t, "refresh_token", con.RefreshToken)
	assert.Equal(t, "user_id", con.UserID)
	assert.Equal(t, "batman", con.Username)
	assert.Equal(t, "team_id", con.TeamID)
}
