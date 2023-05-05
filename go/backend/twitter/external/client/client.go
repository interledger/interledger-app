package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/twitter/external"
	"golang.org/x/oauth2"
)

var _ external.Client = &client{}

type (
	client struct {
		oauthConfig *oauth2.Config
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
			AuthURL:   args.AuthEndpoint,
			TokenURL:  args.TokenEndpoint,
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}

	return &client{
		oauthConfig: oauthConfig,
	}
}

func (c *client) CreateToken(ctx context.Context, args *external.CreateTokenArgs) (*oauth2.Token, error) {
	token, err := c.oauthConfig.Exchange(ctx, args.AuthCode, oauth2.SetAuthURLParam("code_verifier", args.CodeVerifier))
	if err != nil {
		return nil, fmt.Errorf("could not exchange auth code for token: %v", err)
	}

	return token, nil
}

func (c *client) GetAuthorizedUser(ctx context.Context, token *oauth2.Token) (*twitter.TwitterUser, error) {
	res, err := c.oauthConfig.Client(ctx, token).Get("https://api.twitter.com/2/users/me")
	if err != nil {
		return nil, fmt.Errorf("could not get authorized user: %v", err)
	}
	defer res.Body.Close()

	var user *twitter.TwitterUser
	err = json.NewDecoder(res.Body).Decode(user)
	if err != nil {
		return nil, fmt.Errorf("could not decode twitter user: %v", err)
	}

	return user, nil
}

func (c *client) PostTweet(ctx context.Context, token *oauth2.Token, text string) (*twitter.Tweet, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(map[string]interface{}{
		"text": text,
	})
	if err != nil {
		return nil, fmt.Errorf("could not encode tweet: %v", err)
	}

	res, err := c.oauthConfig.Client(ctx, token).Post("https://api.twitter.com/2/tweets", "application/json", &buf)
	if err != nil {
		return nil, fmt.Errorf("could not post tweet: %v", err)
	}
	defer res.Body.Close()

	var tweet *twitter.Tweet
	err = json.NewDecoder(res.Body).Decode(tweet)
	if err != nil {
		return nil, fmt.Errorf("could not decode tweet: %v", err)
	}

	return tweet, nil
}
