package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/interledger/interledger-app/go/backend/currency"
	httplog "github.com/interledger/interledger-app/go/backend/providers/http"
	"github.com/interledger/interledger-app/go/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client interface {
	CreateWallet(ctx context.Context, req CreateWalletReq) (string, error)
	GetWallet(ctx context.Context, id string) (*WalletResp, error)
	Transfer(ctx context.Context, req TransferReq) error
	Deposit(ctx context.Context, req DepositReq) (*DepositResp, error)
	GetEstimatedFee(ctx context.Context, req EstimateFeeReq) (*EstimateFeeResp, error)
	Withdraw(ctx context.Context, req WithdrawalReq) (*WithdrawResponse, error)
	PayoutStatus(ctx context.Context, req PayoutStatusRequest) (*PayoutStatusResponse, error)
	ConvertToUSD(ctx context.Context, req ConvertToUSDRequest) (*currency.Amount, error)
	VerifyPayment(ctx context.Context, req VerifyPaymentReq) (*Payment, error)
}

type client struct {
	baseURL string
	apiKey  string
	api     *http.Client
}

func New(apiKey string, transport *http.Client) Client {
	baseURL := "https://api.chimoney.io/v0.2.4"
	if !env.IsProd() {
		baseURL = "https://api-v2-sandbox.chimoney.io/v0.2.4"
	}

	api := otelhttp.DefaultClient
	if transport != nil {
		api = transport
	}

	return &client{
		api:     api,
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// NewWithBaseURL creates a client with a custom baseURL for testing purposes
func NewWithBaseURL(baseURL string, apiKey string, transport *http.Client) Client {
	api := otelhttp.DefaultClient
	if transport != nil {
		api = transport
	}

	return &client{
		api:     api,
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func (c client) CreateWallet(ctx context.Context, req CreateWalletReq) (string, error) {
	endpoint, err := url.JoinPath(c.baseURL, "multicurrency-wallets", "create")
	if err != nil {
		return "", err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/json")
	httpReq.Header.Add("X-API-KEY", c.apiKey)

	httpResp, err := c.api.Do(httpReq)
	if err != nil {
		return "", err
	}

	body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()

	var respWrapper APIResponse
	err = json.Unmarshal(body, &respWrapper)
	if err != nil {
		return "", err
	}

	if !strings.EqualFold(respWrapper.Status, "success") || respWrapper.Error != "" {
		return "", fmt.Errorf("request failed on creating sub account with status (%s) error (%s)", respWrapper.Status, respWrapper.Error)
	}

	var resp WalletResp
	err = json.Unmarshal(respWrapper.Data, &resp)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (c client) GetWallet(ctx context.Context, id string) (*WalletResp, error) {
	endpoint, err := url.JoinPath(c.baseURL, "multicurrency-wallets", "get")
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "chimoney"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "chimoney",
		})
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?id=%s", endpoint, id), nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/json")
	httpReq.Header.Add("X-API-KEY", c.apiKey)

	httpResp, err := c.api.Do(httpReq)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var respWrapper APIResponse
	err = json.Unmarshal(body, &respWrapper)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(respWrapper.Status, "success") || respWrapper.Error != "" {
		return nil, fmt.Errorf("request failed on fetching sub account with status (%s) error (%s)", respWrapper.Status, respWrapper.Error)
	}

	var resp WalletResp
	err = json.Unmarshal(respWrapper.Data, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c client) Transfer(ctx context.Context, req TransferReq) error {
	endpoint, err := url.JoinPath(c.baseURL, "multicurrency-wallets", "transfer")
	if err != nil {
		return err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "chimoney"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "chimoney",
		})
	}

	type transferReq struct {
		SenderSubAccount    string `json:"subAccount"`
		ReceiverSubAccount  string `json:"receiver"`
		Amount              string `json:"amountToSend"`
		SourceCurrency      string `json:"originCurrency"`
		DestinationCurrency string `json:"destinationCurrency"`
		TurnOffNotification bool   `json:"turnOffNotification,omitempty"`
	}

	body, err := json.Marshal(transferReq{
		SenderSubAccount:    req.SenderSubAccount,
		ReceiverSubAccount:  req.ReceiverSubAccount,
		Amount:              req.Amount.FormatAmount(),
		SourceCurrency:      req.Amount.Currency.String(),
		DestinationCurrency: req.Amount.Currency.String(),
		TurnOffNotification: req.TurnOffNotification,
	})
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/json")
	httpReq.Header.Add("X-API-KEY", c.apiKey)

	httpResp, err := c.api.Do(httpReq)
	if err != nil {
		return err
	}

	body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	var respWrapper APIResponse
	err = json.Unmarshal(body, &respWrapper)
	if err != nil {
		return err
	}

	if !strings.EqualFold(respWrapper.Status, "success") || respWrapper.Error != "" {
		return fmt.Errorf("request failed on transfering between sub accounts with status (%s) error (%s)", respWrapper.Status, respWrapper.Error)
	}

	return nil
}

func (c client) Withdraw(ctx context.Context, req WithdrawalReq) (*WithdrawResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "payouts", "interac")
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "chimoney"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "chimoney",
		})
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/json")
	httpReq.Header.Add("X-API-KEY", c.apiKey)

	httpResp, err := c.api.Do(httpReq)
	if err != nil {
		return nil, err
	}

	body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var respWrapper APIResponse
	err = json.Unmarshal(body, &respWrapper)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(respWrapper.Status, "success") || respWrapper.Error != "" {
		return nil, fmt.Errorf("request failed on withdrawal with status (%s) error (%s)", respWrapper.Status, respWrapper.Error)
	}

	var resp WithdrawResponse
	err = json.Unmarshal(respWrapper.Data, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c client) Deposit(ctx context.Context, req DepositReq) (*DepositResp, error) {
	endpoint, err := url.JoinPath(c.baseURL, "payment", "initiate")
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "chimoney"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "chimoney",
		})
	}

	// TODO unmarshal amount
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/json")
	httpReq.Header.Add("X-API-KEY", c.apiKey)

	httpResp, err := c.api.Do(httpReq)
	if err != nil {
		return nil, err
	}

	body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var respWrapper APIResponse
	err = json.Unmarshal(body, &respWrapper)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(respWrapper.Status, "success") || respWrapper.Error != "" {
		return nil, fmt.Errorf("request failed on deposit link with status (%s) error (%s)", respWrapper.Status, respWrapper.Error)
	}

	var resp DepositResp
	err = json.Unmarshal(respWrapper.Data, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c client) ConvertToUSD(ctx context.Context, req ConvertToUSDRequest) (*currency.Amount, error) {
	endpoint, err := url.Parse(fmt.Sprintf("%s/info/convert/local-amount-to-usd", c.baseURL))
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "chimoney"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "chimoney",
		})
	}

	q := endpoint.Query()
	q.Add("amountInOriginCurrency", fmt.Sprintf("%d", req.Amount))
	q.Add("originCurrency", req.Currency)
	endpoint.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("Accept", "application/json")
	httpReq.Header.Add("X-API-KEY", c.apiKey)

	httpResp, err := c.api.Do(httpReq)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var respWrapper APIResponse
	err = json.Unmarshal(body, &respWrapper)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(respWrapper.Status, "success") || respWrapper.Error != "" {
		return nil, fmt.Errorf("request failed on converting to USD with status (%s) error (%s)", respWrapper.Status, respWrapper.Error)
	}

	var resp ConvertToUSDResponse
	err = json.Unmarshal(respWrapper.Data, &resp)
	if err != nil {
		return nil, err
	}

	amt := currency.FromFloat64(resp.AmountInUSD, currency.USD)

	return &amt, nil
}

func (c client) GetEstimatedFee(ctx context.Context, req EstimateFeeReq) (*EstimateFeeResp, error) {
	endpoint, err := url.JoinPath(c.baseURL, "info", "fee-estimate")
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "chimoney"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "chimoney",
		})
	}
	body, err := json.Marshal(req)

	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/json")
	httpReq.Header.Add("X-API-KEY", c.apiKey)

	httpResp, err := c.api.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("request failed on estimating fee: http status %d, body: %s", httpResp.StatusCode, string(body))
	}

	body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	var respWrapper APIResponse
	err = json.Unmarshal(body, &respWrapper)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(respWrapper.Status, "success") || respWrapper.Error != "" {
		return nil, fmt.Errorf("request failed on estimating fee with status (%s) error (%s)", respWrapper.Status, respWrapper.Error)
	}

	var resp EstimateFeeResp
	err = json.Unmarshal(respWrapper.Data, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c client) VerifyPayment(ctx context.Context, req VerifyPaymentReq) (*Payment, error) {
	endpoint, err := url.JoinPath(c.baseURL, "payment", "verify")
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "chimoney"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "chimoney",
		})
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/json")
	httpReq.Header.Add("X-API-KEY", c.apiKey)

	httpResp, err := c.api.Do(httpReq)
	if err != nil {
		return nil, err
	}

	body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var respWrapper APIResponse
	err = json.Unmarshal(body, &respWrapper)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(respWrapper.Status, "success") || respWrapper.Error != "" {
		return nil, fmt.Errorf("request failed on verify payment with status (%s) error (%s)", respWrapper.Status, respWrapper.Error)
	}

	var resp Payment
	err = json.Unmarshal(respWrapper.Data, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c client) PayoutStatus(ctx context.Context, req PayoutStatusRequest) (*PayoutStatusResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "payouts", "status")
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "chimoney"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "chimoney",
		})
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Accept", "application/json")
	httpReq.Header.Add("X-API-KEY", c.apiKey)

	httpResp, err := c.api.Do(httpReq)
	if err != nil {
		return nil, err
	}

	body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var respWrapper APIResponse
	err = json.Unmarshal(body, &respWrapper)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(respWrapper.Status, "success") || respWrapper.Error != "" {
		return nil, fmt.Errorf("request failed on verify payment with status (%s) error (%s)", respWrapper.Status, respWrapper.Error)
	}

	var resp PayoutStatusResponse
	err = json.Unmarshal(respWrapper.Data, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
