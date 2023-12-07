package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	httplog "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client interface {
	CreateIntent(ctx context.Context, args CreateIntentReq) (string, error)
	GetIntent(ctx context.Context, intentID string) (*Intent, error)
	CreateAccessToken(ctx context.Context, intentID, walletID string) (*AccessToken, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*AccessToken, error)
}

type client struct {
	api          *http.Client
	baseURL      string
	clientID     string
	clientSecret string
}

func New(transport *http.Client) Client {
	baseURL := "https://api-sandbox.astra.finance/v1"
	if transport == nil {
		transport = otelhttp.DefaultClient
	}
	return &client{
		api:          transport,
		baseURL:      baseURL,
		clientID:     os.Getenv("ASTRA_CLIENT_ID"),
		clientSecret: os.Getenv("ASTRA_CLIENT_SECRET"),
	}
}

func (c client) RefreshAccessToken(ctx context.Context, refreshToken string) (*AccessToken, error) {
	reqURL, err := url.JoinPath(c.baseURL, "oauth", "token")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(fmt.Sprintf("grant_type=refresh_token&refresh_token=%s&redirect_uri=%s", refreshToken, env.GetUrl())))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to refresh astra user token (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp AccessToken
	err = json.Unmarshal(respBody, &tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (c client) CreateAccessToken(ctx context.Context, intentID, walletID string) (*AccessToken, error) {
	reqURL, err := url.JoinPath(c.baseURL, "partner", "identity", "verification")
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(GetVerificationTokenReq{
		Provider: "fynbos", // TODO: check
		ProviderData: struct {
			CustomerID string `json:"customer_id"`
		}{
			CustomerID: walletID,
		},
		ClientID:     c.clientID,
		UserIntentID: intentID,
	})
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "astra"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "astra",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create astra user token (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var codeResp VerificationTokenResp
	err = json.Unmarshal(respBody, &codeResp)
	if err != nil {
		return nil, err
	}

	// Exchange the code for a token

	reqURL, err = url.JoinPath(c.baseURL, "oauth", "token")
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s", codeResp.Token, env.GetUrl())))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = c.api.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create astra user token (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp AccessToken
	err = json.Unmarshal(respBody, &tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (c client) CreateIntent(ctx context.Context, args CreateIntentReq) (string, error) {
	reqURL, err := url.JoinPath(c.baseURL, "user_intent")
	if err != nil {
		return "", err
	}

	reqBody, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "astra"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "astra",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.api.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create astra user intent (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var respData CreateIntentResp
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return "", err
	}

	return respData.ID, nil
}

func (c client) GetIntent(ctx context.Context, intentID string) (*Intent, error) {
	reqURL, err := url.JoinPath(c.baseURL, "user_intent", intentID)
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "astra"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "astra",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Accept", "application/json")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get astra user intent (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respData Intent
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}
