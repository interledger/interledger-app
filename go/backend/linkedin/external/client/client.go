package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"gitlab.com/fynbos/backend/linkedin"
	"gitlab.com/fynbos/backend/linkedin/external"
	"golang.org/x/oauth2"
	"net/http"
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
			AuthURL:  args.AuthEndpoint,
			TokenURL: args.TokenEndpoint,
		},
	}

	return &client{
		oauthConfig: oauthConfig,
	}
}

func (c *client) CreateToken(ctx context.Context, authCode string) (*oauth2.Token, error) {
	token, err := c.oauthConfig.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("could not exchange auth code for token: %v", err)
	}

	return token, nil
}

// https://learn.microsoft.com/en-us/linkedin/shared/integrations/people/profile-api?context=linkedin%2Fconsumer%2Fcontext#retrieve-current-members-profile
func (c *client) GetAuthorizedUser(ctx context.Context, token *oauth2.Token) (*linkedin.User, error) {
	res, err := c.oauthConfig.Client(ctx, token).Get("https://api.linkedin.com/v2/me")
	if err != nil {
		return nil, fmt.Errorf("could not get authorized linkedin user: %v", err)
	}
	defer res.Body.Close()

	var jsonBody map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&jsonBody)
	if err != nil {
		return nil, fmt.Errorf("could not decode linkedin user: %v", err)
	}

	return &linkedin.User{
		ID:       jsonBody["id"].(string),
		Username: jsonBody["vanityName"].(string),
	}, nil
}

// https://learn.microsoft.com/en-us/linkedin/consumer/integrations/self-serve/share-on-linkedin?context=linkedin%2Fconsumer%2Fcontext
func (c *client) Post(ctx context.Context, connection *linkedin.Connection, text string) (string, error) {
	token := &oauth2.Token{
		AccessToken:  connection.AccessToken,
		RefreshToken: connection.RefreshToken,
		TokenType:    connection.TokenType,
		Expiry:       connection.Expiry,
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(map[string]interface{}{
		"author":         "urn:li:person:" + connection.UserID,
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]interface{}{
					"text": text,
				},
				"shareMediaCategory": "NONE",
			},
		},
		"visibility": map[string]interface{}{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	})
	if err != nil {
		return "", fmt.Errorf("could not encode post: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.linkedin.com/v2/ugcPosts", nil)
	if err != nil {
		return "", fmt.Errorf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	res, err := c.oauthConfig.Client(ctx, token).Post("https://api.linkedin.com/v2/ugcPosts", "application/json", &buf)
	if err != nil {
		return "", fmt.Errorf("could not post: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		var jsonBody map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&jsonBody)
		if err != nil {
			return "", fmt.Errorf("could not decode error body: %v", err)
		}

		return "", fmt.Errorf("could not post: %v", jsonBody)
	}

	var jsonBody map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&jsonBody)
	if err != nil {
		return "", fmt.Errorf("could not decode post: %v", err)
	}

	id := res.Header.Get("X-RestLi-Id")
	if id == "" {
		return "", fmt.Errorf("could not get post id")
	}

	return id, nil
}

func (c *client) GetPost(ctx context.Context, url string) (*linkedin.Post, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get post: %v", err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not parse post: %v", err)
	}

	post := doc.Find("script[type='application/ld+json']").First().Text()
	var jsonBody map[string]interface{}
	err = json.Unmarshal([]byte(post), &jsonBody)
	if err != nil {
		return nil, fmt.Errorf("could not decode post: %v", err)
	}

	// TODO: handle multiple URLs and check if the post exist
	return &linkedin.Post{
		URLs:   []string{jsonBody["sharedContent"].(map[string]interface{})["url"].(string)},
		Text:   jsonBody["articleBody"].(string),
		Author: jsonBody["author"].(map[string]interface{})["url"].(string),
	}, nil
}
