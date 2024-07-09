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

	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client interface {
	CreateWallet(ctx context.Context, req CreateWalletReq) (string, error)
	GetWallet(ctx context.Context, id string) (*WalletResp, error)
	Transfer(ctx context.Context, req TransferReq) error
	Deposit(ctx context.Context, req DepositReq) (*DepositResp, error)
	Withdraw(ctx context.Context, req WithdrawalReq) error
}

type client struct {
	baseURL string
	apiKey  string
	api     *http.Client
}

func New(transport *http.Client) Client {
	baseURL := "https://api.chimoney.io/v0.2"
	if !env.IsProd() {
		baseURL = "https://api-v2-sandbox.chimoney.io/v0.2"
	}

	api := otelhttp.DefaultClient
	if transport != nil {
		api = transport
	}

	return &client{
		api:     api,
		baseURL: baseURL,
		apiKey:  os.Getenv("CHIMONEY_TOKEN"),
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

	type transferReq struct {
		SenderSubAccount    string `json:"subAccount"`
		ReceiverSubAccount  string `json:"receiver"`
		Amount              string `json:"amountToSend"`
		SourceCurrency      string `json:"originCurrency"`
		DestinationCurrency string `json:"destinationCurrency"`
	}

	body, err := json.Marshal(transferReq{
		SenderSubAccount:    req.SenderSubAccount,
		ReceiverSubAccount:  req.ReceiverSubAccount,
		Amount:              req.Amount.FormatAmount(),
		SourceCurrency:      req.Amount.Currency.String(),
		DestinationCurrency: req.Amount.Currency.String(),
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

func (c client) Withdraw(ctx context.Context, req WithdrawalReq) error {
	endpoint, err := url.JoinPath(c.baseURL, "payouts", "interac")
	if err != nil {
		return err
	}

	body, err := json.Marshal(req)
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
		return fmt.Errorf("request failed on withdrawal with status (%s) error (%s)", respWrapper.Status, respWrapper.Error)
	}

	return nil
}

func (c client) Deposit(ctx context.Context, req DepositReq) (*DepositResp, error) {
	endpoint, err := url.JoinPath(c.baseURL, "payouts", "interac")
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("request failed on withdrawal with status (%s) error (%s)", respWrapper.Status, respWrapper.Error)
	}

	var resp DepositResp
	err = json.Unmarshal(respWrapper.Data, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
