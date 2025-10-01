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
	CreateBusinessProfile(ctx context.Context, args CreateBusinessUserReq) (string, error)
	CreateIntent(ctx context.Context, args CreateIntentReq) (string, error)
	GetIntent(ctx context.Context, intentID string) (*Intent, error)
	CreateAccessToken(ctx context.Context, intentID, walletID string) (*AccessToken, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*AccessToken, error)
	AddCard(ctx context.Context, token string, args CreateCardArgs) (*UserCard, error)
	LookupCard(ctx context.Context, token, cardID string) (*UserCard, error)
	AddAccount(ctx context.Context, token string, args CreateAccountArgs) (*UserAccount, error)
	LookupAccount(ctx context.Context, token, accountID string) (*UserAccount, error)
	CardToAccount(ctx context.Context, token string, args CardToAccountArgs) (*CardToAccountResp, error)
	AccountToCard(ctx context.Context, token string, args AccountToCardArgs) (*AccountToCardResp, error)
	GetTransfer(ctx context.Context, token, transferID string) (*Transaction, error)
	GetRoutine(ctx context.Context, token, routineID string) (*Routine, error)
	CodeExchange(ctx context.Context, code string) (*AccessToken, error)
}

func (c client) CodeExchange(ctx context.Context, code string) (*AccessToken, error) {
	reqURL, err := url.JoinPath(c.baseURL, "oauth", "token")
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", os.Getenv("ASTRA_CODE_EXCHANGE_REDIRECT"))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(data.Encode()))
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println("get token", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create astra user token (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
	}

	var tokenResp AccessToken
	err = json.Unmarshal(respBody, &tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

var basisTheoryProxyUrl = "https://api.basistheory.com/proxy"

type client struct {
	api               *http.Client
	baseURL           string
	secureBaseURL     string
	clientID          string
	clientSecret      string
	institutionID     string
	basisTheoryAPIKey string
}

func New(transport *http.Client) Client {
	baseURL := "https://api-sandbox.astra.finance/v1"
	secureBaseURL := "https://secure.api-sandbox.astra.finance/v1"
	institutionID := "astra_ins_131"
	if env.IsProd() {
		baseURL = "https://api.astra.finance/v1"
		secureBaseURL = "https://secure.api.astra.finance/v1"
		institutionID = "astra_ins_131" // TODO
	} else if env.IsLocal() {
		baseURL = "http://mockbos:8080/astra/v1"
		secureBaseURL = "http://mockbos:8080/astra/v1"
		institutionID = "astra_ins_131" // TODO
		basisTheoryProxyUrl = "http://mockbos:8080/basistheory"
	}

	if transport == nil {
		transport = otelhttp.DefaultClient
	}
	return &client{
		api:               transport,
		baseURL:           baseURL,
		secureBaseURL:     secureBaseURL,
		institutionID:     institutionID,
		clientID:          os.Getenv("ASTRA_CLIENT_ID"),
		clientSecret:      os.Getenv("ASTRA_CLIENT_SECRET"),
		basisTheoryAPIKey: os.Getenv("BASISTHEORY_API_KEY"),
	}
}

func (c client) CreateBusinessProfile(ctx context.Context, args CreateBusinessUserReq) (string, error) {
	reqURL, err := url.JoinPath(c.baseURL, "business_profile")
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create astra business user (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
	}

	var respData CreateIntentResp
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return "", err
	}

	return respData.ID, nil
}

func (c client) RefreshAccessToken(ctx context.Context, refreshToken string) (*AccessToken, error) {
	reqURL, err := url.JoinPath(c.baseURL, "oauth", "token")
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")
	data.Set("redirect_uri", env.GetUrl())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(data.Encode()))
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to refresh astra user token (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create astra user token (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
	}

	var codeResp VerificationTokenResp
	err = json.Unmarshal(respBody, &codeResp)
	if err != nil {
		return nil, err
	}

	// Exchange the code for a token

	reqURL, err = url.JoinPath(c.baseURL, "partner", "identity", "token")
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("token", codeResp.Token)
	data.Set("user_consent_captured", "true")

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(data.Encode()))
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

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create astra user token (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create astra user intent (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get astra user intent (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
	}

	var respData Intent
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c client) AddCard(ctx context.Context, token string, args CreateCardArgs) (*UserCard, error) {
	reqURL, err := url.JoinPath(c.secureBaseURL, "cards")
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

	reqJs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	proxyURL := basisTheoryProxyUrl
	if env.IsLocal() {
		proxyURL = fmt.Sprintf("%s/cards", c.baseURL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyURL, bytes.NewReader(reqJs))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	c.setBasisTheoryProxy(req, reqURL)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create astra user card (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
	}

	var respData UserCard
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c client) LookupCard(ctx context.Context, token, cardID string) (*UserCard, error) {
	reqURL, err := url.JoinPath(c.baseURL, "cards", cardID)
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
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get astra user card (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
	}

	var respData UserCard
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c client) AddAccount(ctx context.Context, token string, args CreateAccountArgs) (*UserAccount, error) {
	reqURL, err := url.JoinPath(c.secureBaseURL, "accounts", "create")
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

	args.InstitutionID = c.institutionID

	reqJs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(reqJs))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create astra user account (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
	}

	var respData UserAccount
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c client) LookupAccount(ctx context.Context, token, accountID string) (*UserAccount, error) {
	reqURL, err := url.JoinPath(c.baseURL, "accounts", accountID)
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
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get astra user account (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
	}

	var respData UserAccount
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c client) AccountToCard(ctx context.Context, token string, args AccountToCardArgs) (*AccountToCardResp, error) {
	reqURL, err := url.JoinPath(c.baseURL, "routines", "account-to-card")
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

	reqJs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(reqJs))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if args.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", args.IdempotencyKey)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create asrta card to account routine (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
	}

	var respData AccountToCardResp
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c client) CardToAccount(ctx context.Context, token string, args CardToAccountArgs) (*CardToAccountResp, error) {
	reqURL, err := url.JoinPath(c.baseURL, "routines", "card-to-account")
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

	reqJs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(reqJs))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if args.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", args.IdempotencyKey)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create astra card to account (%d - %s - %s - %s)", resp.StatusCode, resp.Header.Get("request-id"), resp.Status, string(respBody))
	}

	var respData CardToAccountResp
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c client) GetTransfer(ctx context.Context, token, transferID string) (*Transaction, error) {
	reqURL, err := url.JoinPath(c.baseURL, "transfers", transferID)
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
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get astra user transfer (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
	}

	var respData Transaction
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c client) GetRoutine(ctx context.Context, token, routineID string) (*Routine, error) {
	reqURL, err := url.JoinPath(c.baseURL, "routines", routineID)
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
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get astra user routine (%d - %s - %s)", resp.StatusCode, resp.Status, string(respBody))
	}

	var respData Routine
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c client) setBasisTheoryProxy(r *http.Request, proxyURL string) {
	r.Header.Set("BT-PROXY-URL", proxyURL)
	r.Header.Set("BT-API-KEY", c.basisTheoryAPIKey)
}
