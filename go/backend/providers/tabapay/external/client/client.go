package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	httplog "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var _ external.Client = &client{}

const basisTheoryProxyUrl = "https://api.basistheory.com/proxy"

type client struct {
	basisTheoryApiKey string
	baseUrl           string
	bearerToken       string
	clientID          string
	api               *http.Client
}

type NewClientArgs struct {
	BasisTheoryProxyApiKey string
	ClientID               string
	BearerToken            string
	Transport              http.RoundTripper
}

func New(args NewClientArgs) (*client, error) {
	baseUrl := "https://api.sandbox.tabapay.net:10443"
	if env.IsProd() {
		baseUrl = "https://api.tabapay.net:10443"
	}
	t := args.Transport
	if t == nil {
		t = otelhttp.NewTransport(http.DefaultTransport)
	}

	return &client{
		baseUrl:           baseUrl,
		bearerToken:       args.BearerToken,
		basisTheoryApiKey: args.BasisTheoryProxyApiKey,
		clientID:          args.ClientID,
		api: &http.Client{
			Transport: t,
			Timeout:   95 * time.Second, // set high as Tabapay may be waiting for transactions process
		},
	}, nil
}

func (c *client) setAuth(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
}

func (c *client) setProxy(r *http.Request) {
	r.Header.Set("BT-PROXY-URL", c.baseUrl)
	r.Header.Set("BT-API-KEY", c.basisTheoryApiKey)
}

func (c *client) CreateTransaction(
	ctx context.Context, args external.CreateTransactionArgs,
) (*external.CreateTransactionResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "tabapay"
	}

	endpoint, err := url.JoinPath(c.baseUrl, "v1", "clients", c.clientID, "transactions")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	c.setAuth(req)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var transactionResp external.CreateTransactionResponse
	err = json.NewDecoder(resp.Body).Decode(&transactionResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &transactionResp, nil
}

func (c *client) RetrieveTransaction(
	ctx context.Context, id string,
) (*external.RetrieveTransactionResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "tabapay"
	}
	endpoint, err := url.JoinPath(c.baseUrl, "v1", "clients", c.clientID, "transactions", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	c.setAuth(req)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var retrieveResp external.RetrieveTransactionResponse
	err = json.NewDecoder(resp.Body).Decode(&retrieveResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &retrieveResp, nil
}

func (c *client) CreateAccount(
	ctx context.Context, args external.CreateAccountArgs,
) (*external.CreateAccountResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "tabapay"
	}

	endpoint, err := url.JoinPath(basisTheoryProxyUrl, "v1", "clients", c.clientID, "accounts")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	if args.RejectDuplicateCard && args.OKToAddDuplicateCard {
		return nil, fmt.Errorf("%w use either RejectDuplicateCard or OKToAddDuplicateCard.", external.ErrInternal)
	}

	if args.RejectDuplicateCard {
		endpoint = fmt.Sprintf("%s?RejectDuplicateCard=", endpoint)
	}

	if args.OKToAddDuplicateCard {
		endpoint = fmt.Sprintf("%s?OKToAddDuplicateCard=", endpoint)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	c.setAuth(req)
	c.setProxy(req)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var accountResp external.CreateAccountResponse
	err = json.NewDecoder(resp.Body).Decode(&accountResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &accountResp, nil
}

func (c *client) RetrieveAccount(
	ctx context.Context, id string,
) (*external.RetrieveAccountResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "tabapay"
	}

	endpoint, err := url.JoinPath(c.baseUrl, "v1", "clients", c.clientID, "accounts", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	c.setAuth(req)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var retrieveResp external.RetrieveAccountResponse
	err = json.NewDecoder(resp.Body).Decode(&retrieveResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &retrieveResp, nil
}

func (c *client) QueryCard(
	ctx context.Context, args external.QueryCardArgs,
) (*external.QueryCardResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "tabapay"
	}

	endpoint, err := url.JoinPath(basisTheoryProxyUrl, "v1", "clients", c.clientID, "cards")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	c.setAuth(req)
	c.setProxy(req)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var cardResp external.QueryCardResponse
	err = json.NewDecoder(resp.Body).Decode(&cardResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &cardResp, nil
}

func (c *client) Init3DS(ctx context.Context, args external.Init3DSArgs) (*external.Init3DSResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "tabapay"
	}

	endpoint, err := url.JoinPath(c.baseUrl, "v2", "clients", c.clientID, "3ds", "init")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	c.setAuth(req)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var initResp external.Init3DSResponse
	err = json.NewDecoder(resp.Body).Decode(&initResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &initResp, nil
}

func (c *client) Lookup3DS(ctx context.Context, args external.Lookup3DSArgs) (*external.Lookup3DSResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "tabapay"
	}

	endpoint, err := url.JoinPath(c.baseUrl, "v2", "clients", c.clientID, "3ds", "lookup")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	c.setAuth(req)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var lookupResp external.Lookup3DSResponse
	err = json.NewDecoder(resp.Body).Decode(&lookupResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &lookupResp, nil
}

func (c *client) Authenticate3DS(ctx context.Context, args external.Authenticate3DSArgs) (*external.Authenticate3DSResponse, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "tabapay"
	}

	endpoint, err := url.JoinPath(c.baseUrl, "v2", "clients", c.clientID, "3ds", "authenticate")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	c.setAuth(req)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var authResp external.Authenticate3DSResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &authResp, nil
}

func checkResponseStatusCode(r *http.Response) error {
	if http.StatusOK <= r.StatusCode && r.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	switch r.StatusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("%w %s, path=%s", external.ErrBadRequest, string(body), r.Request.URL.Path)
	case http.StatusUnauthorized:
		return fmt.Errorf("%w %s, path=%s", external.ErrUnauthorized, string(body), r.Request.URL.Path)
	case http.StatusForbidden:
		return fmt.Errorf("%w %s, path=%s", external.ErrForbidden, string(body), r.Request.URL.Path)
	case http.StatusNotFound:
		return fmt.Errorf("%w %s, path=%s", external.ErrNotFound, string(body), r.Request.URL.Path)
	case http.StatusMethodNotAllowed:
		return fmt.Errorf("%w %s, path=%s", external.ErrMethodNotAllowed, string(body), r.Request.URL.Path)
	case http.StatusNotAcceptable:
		return fmt.Errorf("%w %s, path=%s", external.ErrNotAcceptable, string(body), r.Request.URL.Path)
	case http.StatusConflict:
		return fmt.Errorf("%w %s, path=%s", external.ErrConflict, string(body), r.Request.URL.Path)
	case http.StatusGone:
		return fmt.Errorf("%w %s, path=%s", external.ErrGone, string(body), r.Request.URL.Path)
	case http.StatusUnsupportedMediaType:
		return fmt.Errorf("%w %s, path=%s", external.ErrUnsupportedMediatype, string(body), r.Request.URL.Path)
	case http.StatusMisdirectedRequest:
		return fmt.Errorf("%w %s, path=%s", external.ErrMisdirectedRequest, string(body), r.Request.URL.Path)
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("%w %s, path=%s", external.ErrUnprocessableEntity, string(body), r.Request.URL.Path)
	case http.StatusLocked:
		return fmt.Errorf("%w %s, path=%s", external.ErrLocked, string(body), r.Request.URL.Path)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w %s, path=%s", external.ErrTooManyRequests, string(body), r.Request.URL.Path)
	case http.StatusRequestHeaderFieldsTooLarge:
		return fmt.Errorf("%w %s, path=%s", external.ErrRequestHeadersTooLarge, string(body), r.Request.URL.Path)
	case http.StatusInternalServerError:
		return fmt.Errorf("%w %s, path=%s", external.ErrServer, string(body), r.Request.URL.Path)
	case http.StatusBadGateway:
		return fmt.Errorf("%w %s, path=%s", external.ErrBadGateway, string(body), r.Request.URL.Path)
	case http.StatusServiceUnavailable:
		return fmt.Errorf("%w %s, path=%s", external.ErrServiceUnavailable, string(body), r.Request.URL.Path)
	case http.StatusGatewayTimeout:
		return fmt.Errorf("%w %s, path=%s", external.ErrGatewayTimeout, string(body), r.Request.URL.Path)
	default:
		return fmt.Errorf("%w Unknown status code=%d, message=%s, path=%s", external.ErrInternal, r.StatusCode, string(body), r.Request.URL.Path)
	}
}
