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

	"gitlab.com/fynbos/backend/providers/mx/external"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var _ external.Client = &client{}

type client struct {
	baseUrl string
	api     *http.Client
}

func New(clientID, apiKey string, opts ...func(*client)) *client {
	c := &client{
		api: &http.Client{
			Transport: otelhttp.NewTransport(newAuthTransport(clientID, apiKey)),
			Timeout:   5 * time.Second,
		},
	}
	if env.IsProd() {
		c.baseUrl = "https://api.mx.com"
	} else {
		c.baseUrl = "https://int-api.mx.com"
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type authTransport struct {
	clientID string
	apiKey   string
}

func (t authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/vnd.mx.api.v1+json")
	r.SetBasicAuth(t.clientID, t.apiKey)
	return http.DefaultTransport.RoundTrip(r)
}

func newAuthTransport(clientID, apiKey string) *authTransport {
	return &authTransport{
		clientID: clientID,
		apiKey:   apiKey,
	}
}

func (c *client) CreateUser(ctx context.Context, id string) (*external.User, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "users")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	payload := map[string]map[string]string{
		"user": {
			"id": id,
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var user external.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &user, nil
}

func (c *client) GetWidgetURL(ctx context.Context, args external.GetWidgetURLArgs) (*external.WidgetURL, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "users", args.UserGuid, "widget_urls")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	payload := map[string]external.GetWidgetURLArgs{
		"widget_url": args,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var widget external.GetWidgerURLResponse
	err = json.NewDecoder(resp.Body).Decode(&widget)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &widget.WidgetURL, nil
}

func (c *client) ListUsersByID(ctx context.Context, id string) (*external.ListUsersResponse, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "users")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s?id=%s", endpoint, url.QueryEscape(id)), nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var listResponse external.ListUsersResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &listResponse, nil
}

func (c *client) ListAccountOwnersByMember(ctx context.Context, userGuid, memberGuid string) (*external.ListAccountOwnersResponse, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "users", userGuid, "members", memberGuid, "account_owners")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var listResponse external.ListAccountOwnersResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &listResponse, nil
}

func (c *client) ListAccountNumbersByMember(ctx context.Context, userGuid, memberGuid string) (*external.ListAccountNumbersResponse, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "users", userGuid, "members", memberGuid, "account_numbers")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var listResponse external.ListAccountNumbersResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &listResponse, nil
}

func (c *client) ListAccountsByMember(ctx context.Context, userGuid, memberGuid string) (*external.ListAccountsResponse, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "users", userGuid, "members", memberGuid, "accounts")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var listResponse external.ListAccountsResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &listResponse, nil
}

func (c *client) ReadUsersAccount(ctx context.Context, userGuid, accountGuid string) (*external.Account, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "users", userGuid, "accounts", accountGuid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var accountResponse struct{ Account external.Account }
	err = json.NewDecoder(resp.Body).Decode(&accountResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return &accountResponse.Account, nil
}

func (c *client) ListUsers(ctx context.Context) ([]external.User, error) {
	endpoint, err := url.JoinPath(c.baseUrl, "users")
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?records_per_page=100", nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	var listResponse external.ListUsersResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	if err != nil {
		return nil, fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	return listResponse.Users, nil
}

func (c *client) DeleteUser(ctx context.Context, guid string) error {
	endpoint, err := url.JoinPath(c.baseUrl, "users", guid)
	if err != nil {
		return fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("%w %s", external.ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return fmt.Errorf("%w %s", external.ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return err
	}

	return nil
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
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("%w %s, path=%s", external.ErrUnprocessableEntity, string(body), r.Request.URL.Path)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusGatewayTimeout:
		return fmt.Errorf("%w %s, path=%s", external.ErrServer, string(body), r.Request.URL.Path)
	case http.StatusServiceUnavailable:
		return fmt.Errorf("%w %s, path=%s", external.ErrServiceUnavailable, string(body), r.Request.URL.Path)
	default:
		return fmt.Errorf("%w Unknown status code=%d, message=%s, path=%s", external.ErrInternal, r.StatusCode, string(body), r.Request.URL.Path)
	}
}
