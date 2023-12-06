package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	ptiClientIDHeader = "x-pti-client-id"
)

type client struct {
	baseURL      string
	clientID     string
	clientSecret string
	api          *http.Client
}

var _ Client = client{}

type ClientArgs struct {
	ClientID     string
	ClientSecret string
	Transport    *http.Client
	BaseURL      string
}

func New(args ClientArgs) Client {
	// TODO: auth

	base := "https://api.pearsurge.io/v0"
	if args.BaseURL != "" {
		base = args.BaseURL
	} else if env.IsLocal() {
		base = "http://pti.mock"
	}

	transport := otelhttp.DefaultClient
	if args.Transport != nil {
		transport = args.Transport
	}

	return &client{
		baseURL:      base,
		clientID:     args.ClientID,
		clientSecret: args.ClientSecret,
		api:          transport,
	}
}

func (c client) CreateUser(ctx context.Context, args CreateUserArgs) (string, error) {
	url, err := url.JoinPath(c.baseURL, "users")
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)

	resp, err := c.api.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var userResp CreateUserResponse
	err = json.Unmarshal(body, &userResp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return userResp.ID, nil
}

func (c client) CreateWallet(ctx context.Context, args CreateWalletArgs) (*Wallet, error) {
	url, err := url.JoinPath(c.baseURL, "users", args.UserID, "wallets")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var wallet Wallet
	err = json.Unmarshal(body, &wallet)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &wallet, nil
}

func (c client) GetWallet(ctx context.Context, id string) (*Wallet, error) {
	url, err := url.JoinPath(c.baseURL, "wallets", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var wallet Wallet
	err = json.Unmarshal(body, &wallet)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &wallet, nil
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
