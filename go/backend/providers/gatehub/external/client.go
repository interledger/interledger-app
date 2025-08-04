package external

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	httplog "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

var (
	appIDHeader       = "x-gatehub-app-id"
	timestampHeader   = "x-gatehub-timestamp"
	signatureHeader   = "x-gatehub-signature"
	managedUserHeader = "x-gatehub-managed-user-uuid"
	cardAppIDHeader   = "x-gatehub-card-app-id"
)

type client struct {
	onOffRampClientID  string
	onboardingClientID string
	exchangeClientID   string
	appID              string
	cardAppID          string
	apiSecret          string
	baseURL            string
	onboardingBaseURL  string
	onOffRampBaseURL   string

	api *http.Client
}

func NewClient(appID, cardAppID, secret string, transport *http.Client) Client {
	onOffRampClientID := "f8119dfd-e563-44ee-9ae2-1e60a4fce74f"
	onboardingClientID := "4df24d1b-5796-4eec-951b-21699d61b970"
	exchangeClientID := "4e28d4df-22d7-414c-97a3-d71956df29ba"
	baseURL := "https://api.sandbox.gatehub.net"
	onboardingBaseURL := "https://onboarding.sandbox.gatehub.net"
	onOffRampBaseURL := "https://managed-ramp.sandbox.gatehub.net"
	if env.IsProd() {
		onOffRampClientID = "f4c8f30f-7fc3-4aa1-8573-520cb67565e3"
		onboardingClientID = "40a22fc5-9091-4c6f-aff6-a3fddf475b33"
		exchangeClientID = "50e7c590-f6f9-4fa9-9498-260bd978c5d6"
		baseURL = "https://api.gatehub.net"
		onboardingBaseURL = "https://onboarding.gatehub.net"
		onOffRampBaseURL = "https://managed-ramp.gatehub.net"
	}

	api := otelhttp.DefaultClient
	if transport != nil {
		api = transport
	}

	return &client{
		onOffRampClientID:  onOffRampClientID,
		onboardingClientID: onboardingClientID,
		exchangeClientID:   exchangeClientID,
		appID:              appID,
		cardAppID:          cardAppID,
		apiSecret:          secret,
		baseURL:            baseURL,
		onboardingBaseURL:  onboardingBaseURL,
		onOffRampBaseURL:   onOffRampBaseURL,
		api:                api,
	}
}

func (c *client) GetVaultID() string {
	if env.IsProd() {
		return "546ac540-4362-49cb-b639-afc5d4280d03"
	}

	return "a09a0a2c-1a3a-44c5-a1b9-603a6eea9341" // sandbox EUR vault
}

func (c *client) GetOnboardingWidget(ctx context.Context, userID string) (string, error) {
	token, err := c.IssueToken(ctx, userID, Onboarding)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s?bearer=%s", c.onboardingBaseURL, token.Token), nil
}

func (c *client) GetOnOffRampWidget(ctx context.Context, userID string, isDeposit bool) (string, error) {
	token, err := c.IssueToken(ctx, userID, OnOffRamp)
	if err != nil {
		return "", err
	}

	paymentType := "deposit"
	if !isDeposit {
		paymentType = "withdrawal"
	}

	return fmt.Sprintf("%s?paymentType=%s&bearer=%s", c.onOffRampBaseURL, paymentType, token.Token), nil
}

func (c *client) IssueToken(ctx context.Context, userID string, product Product) (*IssueTokenResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "auth", "v1", "tokens")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var clientID string
	var payload IssueTokenReqeust
	switch product {
	case Onboarding:
		clientID = c.onboardingClientID
		payload.Scope = []string{"auth", "id", "gatewayapi", "core", "storage"}
	case OnOffRamp:
		clientID = c.onOffRampClientID
		payload.Scope = []string{"auth", "id", "gatewayapi", "core"}
	case Exchange:
		clientID = c.exchangeClientID
		payload.Scope = []string{"auth", "id", "gatewayapi", "core"}
	}
	endpoint = fmt.Sprintf("%s?clientId=%s", endpoint, clientID)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(managedUserHeader, userID)
	err = c.Sign(ctx, req, time.Now(), body, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var tokenResp IssueTokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &tokenResp, nil
}

func (c *client) CreateUser(ctx context.Context, email string) (*CreateUserResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "auth", "v1", "users", "managed")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := json.Marshal(CreateUserRequest{
		Email: email,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Set("Content-Type", "application/json")
	err = c.Sign(ctx, req, time.Now(), body, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var userResp CreateUserResponse
	err = json.Unmarshal(body, &userResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &userResp, nil
}

func (c *client) GetUser(ctx context.Context, userID string) (*User, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "id", "v1", "users", userID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Set(managedUserHeader, userID)
	err = c.Sign(ctx, req, time.Now(), nil, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &user, nil
}

func (c *client) GetUserWallets(ctx context.Context, userID string) (*GetUserWalletsResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "core", "v1", "users", userID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Set(managedUserHeader, userID)
	err = c.Sign(ctx, req, time.Now(), nil, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var userWalletsResponse GetUserWalletsResponse
	err = json.Unmarshal(body, &userWalletsResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &userWalletsResponse, nil
}

func (c *client) GetWalletBalances(ctx context.Context, userID, address string) ([]WalletBalance, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "core", "v1", "wallets", address, "balances")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Set(managedUserHeader, userID)
	err = c.Sign(ctx, req, time.Now(), nil, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var ret []WalletBalance
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, nil
}

func (c *client) CreateTransaction(ctx context.Context, args CreateTransactionRequest) (*Transaction, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "core", "v1", "transactions")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set(managedUserHeader, args.SendingUserID)
	err = c.Sign(ctx, req, time.Now(), body, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var ret Transaction
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &ret, nil
}

func (c *client) GetUserTransactions(ctx context.Context, userID string) ([]Transaction, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "core", "v1", "users", userID, "transactions")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Set(managedUserHeader, userID)
	err = c.Sign(ctx, req, time.Now(), nil, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var ret []Transaction
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return ret, nil
}

func (c *client) GetTransaction(ctx context.Context, userID, id string) (*Transaction, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "core", "v1", "transactions", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Set(managedUserHeader, userID)
	err = c.Sign(ctx, req, time.Now(), nil, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var ret Transaction
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &ret, nil
}

// TODO(@radu): Pagination or fetch all by default?
func (c *client) ListCards(ctx context.Context, userID, customerID string) ([]Card, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "customers", customerID, "cards")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)

	err = c.Sign(ctx, req, time.Now(), nil, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var cards ListCardsResponse
	err = json.Unmarshal(body, &cards)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return cards.Data, nil
}

func (c *client) GetDeliveryAddresses(ctx context.Context, userID, customerID string) ([]CustomerDeliveryAddress, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", customerID, "addresses")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set(managedUserHeader, userID)
	req.Header.Set(cardAppIDHeader, c.cardAppID)

	err = c.Sign(ctx, req, time.Now(), nil, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var addrs []CustomerDeliveryAddress
	err = json.Unmarshal(body, &addrs)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return addrs, nil
}

func (c *client) GetCardApplicationProducts(ctx context.Context) ([]CardApplicationProduct, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v1", "card-applications", c.cardAppID, "card-products")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req.Header.Set(cardAppIDHeader, c.cardAppID)

	err = c.Sign(ctx, req, time.Now(), nil, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var products []CardApplicationProduct
	err = json.Unmarshal(body, &products)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return products, nil
}

func (c *client) OrderCard(ctx context.Context, userID string, gatehubWalletAddress string) error {
	type ghcard struct {
		ProductCode string `json:"productCode"`
	}

	type testStruct struct {
		WalletAddress string `json:"walletAddress"`
		Card          ghcard `json:"card"`
	}

	body, _ := json.Marshal(testStruct{
		WalletAddress: gatehubWalletAddress,
		Card: ghcard{
			ProductCode: "PMDSGWEEA",
		},
	})

	fmt.Println("test")

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "gatehub"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "gatehub",
		})
	}

	endpoint, err := url.JoinPath(c.baseURL, "cards", "v2", "cards")
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}
	fmt.Println("after-new-request")

	log.Info("userId", zap.String("id", userID))
	req.Header.Set(cardAppIDHeader, c.cardAppID)
	req.Header.Set(managedUserHeader, "0242a34b-b071-460e-9179-f4040d9895e8")
	req.Header.Set("Content-Type", "application/json")

	err = c.Sign(ctx, req, time.Now(), body, endpoint)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	fmt.Println("after-sign")

	resp, err := c.api.Do(req)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	fmt.Println(resp.StatusCode)

	err = checkResponseStatusCode(resp)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()
	fmt.Println(string(rbody[:]))

	return nil
}

func (c *client) Sign(ctx context.Context, req *http.Request, date time.Time, payload []byte, targetURL string) error {
	base := fmt.Sprintf("%d|%s|%s|%s", date.UnixMilli(), req.Method, targetURL, string(payload))
	base = strings.Trim(base, "|")
	hmac := hmac.New(sha256.New, []byte(c.apiSecret))
	_, err := hmac.Write([]byte(base))
	if err != nil {
		return err
	}

	req.Header.Set(appIDHeader, c.appID)
	req.Header.Set(timestampHeader, fmt.Sprintf("%d", date.UnixMilli()))
	req.Header.Set(signatureHeader, hex.EncodeToString(hmac.Sum(nil)))

	return nil
}

func checkResponseStatusCode(r *http.Response) error {
	if http.StatusOK <= r.StatusCode && r.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}
	origRespBody := make([]byte, len(body))
	copy(origRespBody, body)
	defer func() {
		if body != nil {
			r.Body = io.NopCloser(bytes.NewBuffer(origRespBody))
		}
	}()

	switch r.StatusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("%w %s, path=%s", ErrBadRequest, string(body), r.Request.URL.Path)
	case http.StatusUnauthorized:
		return fmt.Errorf("%w %s, path=%s", ErrUnauthorized, string(body), r.Request.URL.Path)
	case http.StatusForbidden:
		return fmt.Errorf("%w %s, path=%s", ErrForbidden, string(body), r.Request.URL.Path)
	case http.StatusNotFound:
		return fmt.Errorf("%w %s, path=%s", ErrNotFound, string(body), r.Request.URL.Path)
	case http.StatusMethodNotAllowed:
		return fmt.Errorf("%w %s, path=%s", ErrMethodNotAllowed, string(body), r.Request.URL.Path)
	case http.StatusNotAcceptable:
		return fmt.Errorf("%w %s, path=%s", ErrNotAcceptable, string(body), r.Request.URL.Path)
	case http.StatusConflict:
		return fmt.Errorf("%w %s, path=%s", ErrConflict, string(body), r.Request.URL.Path)
	case http.StatusGone:
		return fmt.Errorf("%w %s, path=%s", ErrGone, string(body), r.Request.URL.Path)
	case http.StatusUnsupportedMediaType:
		return fmt.Errorf("%w %s, path=%s", ErrUnsupportedMediatype, string(body), r.Request.URL.Path)
	case http.StatusMisdirectedRequest:
		return fmt.Errorf("%w %s, path=%s", ErrMisdirectedRequest, string(body), r.Request.URL.Path)
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("%w %s, path=%s", ErrUnprocessableEntity, string(body), r.Request.URL.Path)
	case http.StatusLocked:
		return fmt.Errorf("%w %s, path=%s", ErrLocked, string(body), r.Request.URL.Path)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w %s, path=%s", ErrTooManyRequests, string(body), r.Request.URL.Path)
	case http.StatusRequestHeaderFieldsTooLarge:
		return fmt.Errorf("%w %s, path=%s", ErrRequestHeadersTooLarge, string(body), r.Request.URL.Path)
	case http.StatusInternalServerError:
		return fmt.Errorf("%w %s, path=%s", ErrServer, string(body), r.Request.URL.Path)
	case http.StatusBadGateway:
		return fmt.Errorf("%w %s, path=%s", ErrBadGateway, string(body), r.Request.URL.Path)
	case http.StatusServiceUnavailable:
		return fmt.Errorf("%w %s, path=%s", ErrServiceUnavailable, string(body), r.Request.URL.Path)
	case http.StatusGatewayTimeout:
		return fmt.Errorf("%w %s, path=%s", ErrGatewayTimeout, string(body), r.Request.URL.Path)
	default:
		return fmt.Errorf("%w Unknown status code=%d, message=%s, path=%s", ErrInternal, r.StatusCode, string(body), r.Request.URL.Path)
	}
}
