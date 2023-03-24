package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
)

var _ external.Client = &client{}

type client struct {
	baseUrl     string
	bearerToken string
	api         *http.Client
}

type NewClientArgs struct {
	VgsProxyURL string         `validate:"required"`
	ClientID    string         `validate:"required"`
	BearerToken string         `validate:"required"`
	CaCertPool  *x509.CertPool `validate:"required"`
}

func New(args NewClientArgs) (*client, error) {
	if err := validator.New().Struct(args); err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	proxyUrl, err := url.Parse(args.VgsProxyURL)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &client{
		baseUrl:     fmt.Sprintf("https://FQDN/v1/clients/%s", args.ClientID),
		bearerToken: args.BearerToken,
		api: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
				TLSClientConfig: &tls.Config{
					RootCAs:            args.CaCertPool,
					InsecureSkipVerify: true,
				},
			},
		},
	}, nil
}

func (c *client) setAuth(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
}

func (c *client) CreateTransaction(
	ctx context.Context, args external.CreateTransactionArgs,
) (*external.CreateTransactionResponse, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "transactions")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	payload := bytes.NewBuffer(nil)
	err = json.NewEncoder(payload).Encode(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, payload)
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
	endpoint, err := url.JoinPath(c.baseUrl, "transactions", id)
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
	endpoint, err := url.JoinPath(c.baseUrl, "accounts")
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

	payload := bytes.NewBuffer(nil)
	err = json.NewEncoder(payload).Encode(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, payload)
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
	endpoint, err := url.JoinPath(c.baseUrl, "accounts", id)
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
