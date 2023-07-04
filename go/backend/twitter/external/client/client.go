package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/backend/twitter/external"
	"golang.org/x/oauth2"
)

var _ external.Client = &client{}

type (
	client struct {
		oauthConfig *oauth2.Config
		bearerToken string
	}

	NewClientArgs struct {
		ClientID      string
		ClientSecret  string
		BearerToken   string
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
		bearerToken: args.BearerToken,
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

	var jsonBody map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&jsonBody)
	if err != nil {
		return nil, fmt.Errorf("could not decode twitter user: %v", err)
	}

	return &twitter.TwitterUser{
		ID:       jsonBody["data"].(map[string]interface{})["id"].(string),
		Username: jsonBody["data"].(map[string]interface{})["username"].(string),
	}, nil
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

	if res.StatusCode != 201 {
		var jsonBody map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return nil, fmt.Errorf("could not decode error body: %v", err)
		}

		return nil, fmt.Errorf("could not post tweet: %v", jsonBody)
	}

	var jsonBody map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&jsonBody)
	if err != nil {
		return nil, fmt.Errorf("could not decode tweet: %v", err)
	}

	return &twitter.Tweet{
		ID:   jsonBody["data"].(map[string]interface{})["id"].(string),
		Text: jsonBody["data"].(map[string]interface{})["text"].(string),
	}, nil
}

func (c *client) GetTweet(ctx context.Context, tweetID string) (*twitter.Tweet, error) {
	println("HERE IS THE TWEET ID help: ", tweetID)
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/2/tweets/%s", tweetID), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}

	queryParams := url.Values{
		"expansions":   []string{"author_id"},
		"tweet.fields": []string{"entities"},
	}
	req.URL.RawQuery = queryParams.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
	res, err := http.DefaultClient.Do(req)
	// do error handling with res.StatusCode
	if err != nil {
		return nil, fmt.Errorf("could not get tweet: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		var jsonBody map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return nil, fmt.Errorf("could not decode error body: %v", err)
		}

		return nil, fmt.Errorf("could not get tweet: %v", jsonBody)
	}

	var tweet external.Tweet
	err = json.NewDecoder(res.Body).Decode(&tweet)
	if err != nil {
		return nil, fmt.Errorf("could not decode tweet: %v", err)
	}

	// create a list of expanded urls
	var urls []string
	for _, u := range tweet.Data.Entities.URLs {
		urls = append(urls, u.ExpandedURL)
	}

	return &twitter.Tweet{
		ID:             tweet.Data.ID,
		Text:           tweet.Data.Text,
		URLs:           urls,
		AuthorID:       tweet.Includes.Users[0].ID,
		AuthorUsername: tweet.Includes.Users[0].Username,
	}, nil
}
