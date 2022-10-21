package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
)

const (
	sandboxUrl = "https://v4sandbox.machpay.com/v4"
	prodUrl    = "https://v4sandbox.machpay.com/v4" // TODO: ask what prod url is
)

var _ external.Client = Client{}

type Client struct {
	api     *http.Client
	baseUrl string
}

func New(clientID, clientSecret string) *Client {
	if clientID == "" || clientSecret == "" {
		log.Error("Machnet credentials not set.")
		if env.IsProd() || env.IsSandbox() {
			panic("Machnet credentials not set.")
		}
	}

	baseUrl := sandboxUrl
	if env.IsProd() {
		baseUrl = prodUrl
	}

	return &Client{
		api: &http.Client{
			Transport: newAuthTransport(clientID, clientSecret),
			Timeout:   5 * time.Second,
		},
		baseUrl: baseUrl,
	}
}

func (c Client) RegisterUser(ctx context.Context, user external.User) (*external.User, error) {
	payload, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Post(c.baseUrl+"/users", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var externalUser external.User
	err = json.Unmarshal(body, &externalUser)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &externalUser, nil
}

func (c Client) CreateReceiveUserBankAccount(ctx context.Context, sendUserID, receiveUserID string, acc external.ReceiveUserBankAccount) (*external.ReceiveUserBankAccount, error) {
	payload, err := json.Marshal(acc)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Post(fmt.Sprintf("%s/users/%s/receive-users/%s/accounts", c.baseUrl, sendUserID, receiveUserID), "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var externalAccount external.ReceiveUserBankAccount
	err = json.Unmarshal(body, &externalAccount)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &externalAccount, nil
}

func (c Client) UpdateUser(ctx context.Context, id string, newValues external.User) (*external.User, error) {
	if newValues.ID != "" {
		return nil, fmt.Errorf("%w Do not set ID on newValues.", external.ErrInvalidArgument)
	}
	payload, err := json.Marshal(newValues)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("%s/users/%s", c.baseUrl, id), bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var externalUser external.User
	err = json.Unmarshal(body, &externalUser)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &externalUser, nil
}

func (c Client) GetUserByID(ctx context.Context, id string) (*external.User, error) {
	resp, err := c.api.Get(fmt.Sprintf("%s/users/%s", c.baseUrl, id))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var user external.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &user, nil
}

func (c Client) InitiateKYC(ctx context.Context, userID string) (*external.InitiateKycResponse, error) {
	resp, err := c.api.Post(fmt.Sprintf("%s/users/%s/kyc", c.baseUrl, userID), "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var kyc external.InitiateKycResponse
	err = json.Unmarshal(body, &kyc)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &kyc, nil
}

func (c Client) GetVerificationStatus(ctx context.Context, userID string) (*external.VerificationStatus, error) {
	resp, err := c.api.Get(fmt.Sprintf("%s/users/%s/cip-info", c.baseUrl, userID))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var kycStatus external.VerificationStatus
	err = json.Unmarshal(body, &kycStatus)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &kycStatus, nil
}

func (c Client) GetReceiveUserList(ctx context.Context, userID string) ([]external.User, error) {
	panic("no-op")
}

func (c Client) GetFundingAccountWidgetToken(ctx context.Context, userID string) (*external.WidgetTokenResponse, error) {
	resp, err := c.api.Get(fmt.Sprintf("%s/users/%s/widget-token", c.baseUrl, userID))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var widgetToken external.WidgetTokenResponse
	err = json.Unmarshal(body, &widgetToken)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &widgetToken, nil
}

func (c Client) GetUserFundingsource(ctx context.Context, userID, fundingsourceID string) (*external.FundingSource, error) {
	resp, err := c.api.Get(fmt.Sprintf("%s/users/%s/funds", c.baseUrl, userID))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var list []external.FundingSource
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	var fs *external.FundingSource
	for _, item := range list {
		if item.ID == fundingsourceID {
			fs = &item
			break
		}
	}
	if fs == nil {
		return nil, external.ErrNotFound
	}

	return fs, nil
}

func (c Client) CreateTransaction(ctx context.Context, trx external.CreateTransactionArgs) (*external.Transaction, error) {
	payload, err := json.Marshal(trx)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/users/%s/transactions", c.baseUrl, trx.FromUserID), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var res external.Transaction
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &res, err
}

func (c Client) UpdateDeliveryRequest(ctx context.Context, request external.DeliveryRequest) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		fmt.Sprintf("%s/users/%s/transactions/delivery-requests/%s", c.baseUrl, request.UserID, request.TransactionID),
		bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	_, err = parseResponse(resp)
	return err
}

func (c Client) ListReceiveUserBankAccounts(ctx context.Context, sendUserID, receiveUserID string) ([]external.ReceiveUserBankAccount, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/users/%s/receive-users/%s/accounts", c.baseUrl, sendUserID, receiveUserID), nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var res []external.ReceiveUserBankAccount
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return res, err
}

func (c Client) GetBanks(ctx context.Context, countryCode string) ([]external.Bank, error) {
	getBanksUrl, err := url.Parse(c.baseUrl + "/banks?country=" + countryCode)
	if err != nil {
		return nil, fmt.Errorf("%w Failed to get banks url.", external.ErrInternal)
	}

	resp, err := c.api.Get(getBanksUrl.String())
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	var banks []external.Bank
	err = json.Unmarshal(body, &banks)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return banks, nil
}

func parseResponse(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, external.ErrNotFound
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, external.ErrUnauthorized
	}

	if resp.StatusCode == http.StatusBadRequest {
		return nil, fmt.Errorf("%w statusCode: %d, message: %s, body: %s", external.ErrInvalidArgument, resp.StatusCode, resp.Status, string(body))
	}

	if !isStatus2xx(resp.StatusCode) {
		return nil, fmt.Errorf("%w statusCode: %d, message: %s, body: %s", external.ErrInternal, resp.StatusCode, resp.Status, string(body))
	}

	return body, nil
}

type authTransport struct {
	baseTransport http.RoundTripper
	clientID      string
	clientSecret  string
}

func (t authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("X-Client-Id", t.clientID)
	r.Header.Set("X-Client-Secret", t.clientSecret)
	r.Header.Set("Content-Type", "application/json")
	return t.baseTransport.RoundTrip(r)
}

func newAuthTransport(clientID, clientSecret string) *authTransport {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}
	return &authTransport{
		baseTransport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func isStatus2xx(code int) bool {
	return http.StatusOK <= code && code < http.StatusMultipleChoices
}
