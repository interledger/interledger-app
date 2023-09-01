package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.com/fynbos/backend/discord"
	"gitlab.com/fynbos/backend/discord/external"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
)

var _ external.Client = &client{}

type (
	client struct {
		oauthConfig *oauth2.Config
		web         *http.Client
	}

	NewClientArgs struct {
		ClientID      string
		ClientSecret  string
		AuthEndpoint  string
		TokenEndpoint string
		RedirectURL   string
	}
)

func New(args *NewClientArgs) *client {
	oauthConfig := &oauth2.Config{
		ClientID:     args.ClientID,
		ClientSecret: args.ClientSecret,
		RedirectURL:  args.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  args.AuthEndpoint,
			TokenURL: args.TokenEndpoint,
		},
	}

	return &client{
		oauthConfig: oauthConfig,
		web:         otelhttp.DefaultClient,
	}
}

func (c *client) CreateToken(ctx context.Context, args *external.CreateTokenArgs) (*oauth2.Token, error) {
	reqCtx := context.WithValue(ctx, oauth2.HTTPClient, c.web)
	token, err := c.oauthConfig.Exchange(reqCtx, args.AuthCode)
	if err != nil {
		return nil, fmt.Errorf("could not exchange auth code for token: %v", err)
	}

	return token, nil
}

func (c *client) GetAuthorizedUser(ctx context.Context, token *oauth2.Token) (*discord.User, error) {
	reqCtx := context.WithValue(ctx, oauth2.HTTPClient, c.web)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "https://discord.com/api/oauth2/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}
	res, err := c.oauthConfig.Client(reqCtx, token).Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not get authorized user: %v", err)
	}
	defer res.Body.Close()

	var jsonBody map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&jsonBody)
	if err != nil {
		return nil, fmt.Errorf("could not decode discord user: %v", err)
	}

	return &discord.User{
		ID:         jsonBody["user"].(map[string]interface{})["id"].(string),
		Username:   jsonBody["user"].(map[string]interface{})["username"].(string),
		GlobalName: jsonBody["user"].(map[string]interface{})["global_name"].(string),
	}, nil
}
